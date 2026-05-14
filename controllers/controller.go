package controllers

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"unicode"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	uuid "github.com/nu7hatch/gouuid"
	"gopkg.in/yaml.v2"

	"github.com/xvv6u577/logv2fs/database"

	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/middleware"

	"github.com/xvv6u577/logv2fs/model"
	singboxctl "github.com/xvv6u577/logv2fs/singbox"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// validatePasswordStrength 校验密码复杂度。
// 规则（KISS）：
//   - 长度 >= 8
//   - 至少包含 1 个字母（Unicode 字母，含中文/CJK）
//   - 至少包含 1 个数字
//
// 调用方：EditUser 用户主动改密时使用。
// 注意：SignUp() 当前直接用邮箱作为初始密码，绕过此校验；如未来允许自选密码请同步加上。
// 错误信息使用中文，前端会直接把 err.response.data.error 展示给用户。
func validatePasswordStrength(pwd string) error {
	if len(pwd) < 8 {
		return errors.New("密码长度至少需要 8 个字符")
	}
	var hasLetter, hasDigit bool
	for _, r := range pwd {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("密码必须同时包含字母和数字")
	}
	return nil
}

var (
	validate = validator.New()
)

type (
	TrafficAtPeriod  = model.TrafficAtPeriod
	Node             = model.Node
	SubscriptionNode = model.SubscriptionNode
	SingboxYAML      = model.SingboxYAML
	SingboxJSON      = model.SingboxJSON
	RealityJSON      = model.RealityJSON
	Hysteria2JSON    = model.Hysteria2JSON
	RealityYAML      = model.RealityYAML
	Hysteria2YAML    = model.Hysteria2YAML
	CFVlessJSON      = model.CFVlessJSON
	CFVlessYAML      = model.CFVlessYAML
	UserTrafficLogs  = model.UserTrafficLogs
	NodeTrafficLogs  = model.NodeTrafficLogs
)

func getPublicKey() string {
	return os.Getenv("PUBLIC_KEY")
}

func getShortID() string {
	return os.Getenv("SHORT_ID")
}

func getCredit() int64 {
	credit, _ := strconv.ParseInt(os.Getenv("CREDIT"), 10, 64)
	return credit
}

// HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

// VerifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or passowrd is incorrect"
		check = false
	}

	return check, msg
}

// Renews the user tokens when they login
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	Updated_at := time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := database.GetCollection(model.UserTrafficLogs{}).UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$set", Value: updateObj}},
		&opt,
	)
	defer cancel()

	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

}

// lookupUserPeriodStage 构造一个 $lookup 聚合 stage，把 user_traffic_periods 集合中
// 当前用户对应粒度（kind）的周期记录，重新组装为前端兼容形态注入到结果文档。
//
// 参数:
//   - kind:        "daily" | "monthly" | "yearly"
//   - periodAlias: 输出数组中表示周期的字段名（"date" / "month" / "year"），保持与旧前端一致
//   - asField:     注入到结果文档的字段名（"daily_logs" / "monthly_logs" / "yearly_logs"）
//   - limit:       仅保留最近的多少条；<=0 表示不限
//
// 关联键：USER_TRAFFIC_LOGS.email_as_id == user_traffic_periods.email_as_id
func lookupUserPeriodStage(kind, periodAlias, asField string, limit int) bson.D {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{{Key: "$and", Value: bson.A{
				bson.D{{Key: "$eq", Value: bson.A{"$email_as_id", "$$email"}}},
				bson.D{{Key: "$eq", Value: bson.A{"$kind", kind}}},
			}}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "period", Value: -1}}}},
	}
	if limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})
	}
	pipeline = append(pipeline, bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: periodAlias, Value: "$period"},
		{Key: "traffic", Value: 1},
	}}})

	return bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: model.UserTrafficPeriod{}.CollectionName()},
		{Key: "let", Value: bson.D{{Key: "email", Value: "$email_as_id"}}},
		{Key: "pipeline", Value: pipeline},
		{Key: "as", Value: asField},
	}}}
}

// check if a string in a slice
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// CreateUser is the api used to tget a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user UserTrafficLogs
		var current = time.Now()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			log.Printf("validate error: %v", validationErr)
			return
		}

		user_email := helper.SanitizeStr(user.Email_As_Id)
		count, err := database.GetCollection(model.UserTrafficLogs{}).CountDocuments(context.TODO(), bson.M{"email_as_id": user_email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			log.Printf("Checking email error: %s", err.Error())
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
			log.Printf("this email already exists")
			return
		}

		if user.Name == "" {
			user.Name = user_email
		}

		password := HashPassword(user_email)
		user.Password = password

		user.CreatedAt = current
		user.UpdatedAt = current

		uuidV4, _ := uuid.NewV4()
		user.UUID = uuidV4.String()

		user_role := "plain"
		user.Used = 0

		if user.Credit == 0 {
			user.Credit = getCredit()
		}

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(user_email, user.UUID, user.Name, user_role, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		// 周期级流量数据已拆分到 user_traffic_periods 集合，新建用户无需在主文档预置空数组。
		// 后续读接口会通过 $lookup 自动返回空数组保持前端兼容。

		_, err = database.GetCollection(model.UserTrafficLogs{}).InsertOne(context.Background(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while inserting user traffic logs: %v", err)
			return
		}

		// 异步把新用户推到 sing-box 控制端口，不阻塞 HTTP 响应。
		// 即便推送失败（控制端口未配/网络异常），DB 仍是事实源，singbox 重启时
		// 会通过 UpdateOptionsFromMongoDB 自动重建状态。
		go singboxctl.SafeAddUser(singboxctl.AddUserRequest{
			EmailAsId: user.Email_As_Id,
			UUID:      user.UUID,
			UserId:    user.User_id,
		})

		c.JSON(http.StatusOK, gin.H{"message": "user " + user.Name + " created successfully"})
	}
}

// Login 处理用户登录。
// 安全要点：
//  1. 失败信息统一为"用户名或密码错误"，避免账号枚举。
//  2. IP 维度限流由路由层中间件完成；这里再做账号维度限流。
//  3. 不再把 mongo 错误原文返回给客户端。
//  4. 仅返回 access token，不再回写整个用户对象。
const genericLoginFailMsg = "invalid username or password"

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var boundUser, foundUser UserTrafficLogs

		if err := c.BindJSON(&boundUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			log.Printf("login bind error: %v", err)
			return
		}

		sanitizedEmail := helper.SanitizeStr(boundUser.Email_As_Id)

		// 账号维度限流：撞库防护。即便攻击者更换 IP，也会卡在这里。
		if !middleware.LoginAccountAllow(sanitizedEmail) {
			c.Header("Retry-After", "300")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, please try later"})
			return
		}

		err := database.GetCollection(model.UserTrafficLogs{}).
			FindOne(ctx, bson.M{"email_as_id": sanitizedEmail}).
			Decode(&foundUser)
		if err != nil {
			// 不区分"用户不存在"与"密码错误"，避免账号枚举攻击。
			c.JSON(http.StatusUnauthorized, gin.H{"error": genericLoginFailMsg})
			log.Printf("login lookup error for %s: %v", sanitizedEmail, err)
			return
		}

		passwordIsValid, _ := VerifyPassword(boundUser.Password, foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": genericLoginFailMsg})
			log.Printf("password mismatch for %s", sanitizedEmail)
			return
		}

		token, refreshToken, err := helper.GenerateAllTokens(sanitizedEmail, foundUser.UUID, foundUser.Name, foundUser.Role, foundUser.User_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			log.Printf("token generation failed: %v", err)
			return
		}

		UpdateAllTokens(token, refreshToken, foundUser.User_id)

		// 仅回写 token 字段，不暴露 refresh_token / 用户其他字段。
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func EditUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// 从路径参数获取用户名
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var user, foundUser UserTrafficLogs

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		// 不需要验证整个结构体，因为我们只是部分更新

		err := database.GetCollection(model.UserTrafficLogs{}).FindOne(ctx, bson.M{"email_as_id": helper.SanitizeStr(name)}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("user not found: %s", name)
			return
		}

		newFoundUser := bson.M{}

		// 允许编辑 name, role, password 和 remark
		if foundUser.Role != user.Role && user.Role != "" {
			newFoundUser["role"] = user.Role
			log.Printf("Updating role from %s to %s", foundUser.Role, user.Role)
		}

		if foundUser.Name != user.Name && user.Name != "" {
			newFoundUser["name"] = user.Name
			log.Printf("Updating name from %s to %s", foundUser.Name, user.Name)
		}

		// 添加备注更新支持（允许设置为空字符串）
		if foundUser.Remark != user.Remark {
			newFoundUser["remark"] = user.Remark
			log.Printf("Updating remark from '%s' to '%s'", foundUser.Remark, user.Remark)
		}

		// 密码更新支持：
		//   - 仅当 user.Password 非空时才尝试更新。
		//   - 走统一的密码复杂度校验（>= 8 位、字母+数字），不达标直接 400。
		//   - 之前的实现是 len >= 6 才更新、否则静默忽略，会导致用户以为改了密码实际没改，这里一并修复。
		if user.Password != "" {
			if err := validatePasswordStrength(user.Password); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				log.Printf("EditUser password strength check failed for %s: %v", foundUser.Email_As_Id, err)
				return
			}
			hashedPassword := HashPassword(user.Password)
			newFoundUser["password"] = hashedPassword
			log.Printf("Updating password for user %s", foundUser.Email_As_Id)
		}

		if len(newFoundUser) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no new value in post data."})
			log.Printf("no new value in post data.")
			return
		}

		// 添加更新时间
		newFoundUser["updated_at"] = time.Now()

		var updatedUser UserTrafficLogs
		err = database.GetCollection(model.UserTrafficLogs{}).FindOneAndUpdate(
			ctx,
			bson.M{"email_as_id": helper.SanitizeStr(name)},
			bson.M{"$set": newFoundUser},
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&updatedUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error updating user: %v", err)
			return
		}

		log.Printf("User %s updated successfully", updatedUser.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": updatedUser})
	}
}

func DeleteUserByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := c.Param("name")
		log.Printf("Attempting to delete user: %s", name)

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var user UserTrafficLogs
		filter := bson.M{"email_as_id": name}
		var projections = bson.D{
			{Key: "email_as_id", Value: 1},
			{Key: "name", Value: 1},
		}
		err := database.GetCollection(model.UserTrafficLogs{}).FindOne(context.TODO(), filter, options.FindOne().SetProjection(projections)).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("DeleteUserByUserName - user not found: %s, error: %s", name, err.Error())
			return
		}

		// delete user from userTrafficLogsCol
		result, err := database.GetCollection(model.UserTrafficLogs{}).DeleteOne(context.TODO(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("DeleteUserByUserName - delete failed: %s", err.Error())
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not deleted"})
			log.Printf("DeleteUserByUserName - no documents deleted for user: %s", name)
			return
		}

		// 同步清理拆分出去的周期流量数据，避免新建同名用户时残留历史
		if _, err := database.GetCollection(model.UserTrafficPeriod{}).DeleteMany(
			context.TODO(),
			bson.M{"email_as_id": name},
		); err != nil {
			log.Printf("DeleteUserByUserName - cleanup user_traffic_periods failed: %s", err.Error())
		}

		// 异步通知 sing-box 把该用户从认证列表彻底移除
		go singboxctl.SafeRemoveUser(user.Email_As_Id)

		log.Printf("Delete user %s successfully!", user.Name)
		c.JSON(http.StatusOK, gin.H{"message": "Delete user " + user.Name + " successfully!"})
	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// 列表场景仅取每种粒度最近的 10 条，足够前端展示且控制响应体大小。
		// 用户主文档已不再保存 daily/monthly/yearly_logs，全部通过 $lookup 从 user_traffic_periods 注入。
		pipeline := mongo.Pipeline{
			{{Key: "$project", Value: bson.D{
				{Key: "email_as_id", Value: 1},
				{Key: "uuid", Value: 1},
				{Key: "name", Value: 1},
				{Key: "role", Value: 1},
				{Key: "status", Value: 1},
				{Key: "used", Value: 1},
				{Key: "remark", Value: 1},
				{Key: "updated_at", Value: 1},
			}}},
			lookupUserPeriodStage("daily", "date", "daily_logs", 10),
			lookupUserPeriodStage("monthly", "month", "monthly_logs", 10),
			lookupUserPeriodStage("yearly", "year", "yearly_logs", 10),
		}

		cursor, err := database.GetCollection(model.UserTrafficLogs{}).Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetAllUsers: %s", err.Error())
			return
		}
		defer cursor.Close(ctx)

		var results []UserTrafficLogs
		if err = cursor.All(ctx, &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetAllUsers: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, results)

	}
}

func GetUserByName() gin.HandlerFunc {
	return func(c *gin.Context) {

		name := c.Param("name")

		if err := helper.MatchUserTypeAndName(c, name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("GetUserByName: %s", err.Error())
			return
		}

		// 单用户详情：通过聚合管道注入 daily/monthly/yearly_logs，全部周期不截断。
		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.D{{Key: "email_as_id", Value: name}}}},
			{{Key: "$project", Value: bson.D{
				{Key: "email_as_id", Value: 1},
				{Key: "used", Value: 1},
				{Key: "uuid", Value: 1},
				{Key: "name", Value: 1},
				{Key: "status", Value: 1},
				{Key: "role", Value: 1},
				{Key: "remark", Value: 1},
				{Key: "credit", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "updated_at", Value: 1},
			}}},
			lookupUserPeriodStage("daily", "date", "daily_logs", 0),
			lookupUserPeriodStage("monthly", "month", "monthly_logs", 0),
			lookupUserPeriodStage("yearly", "year", "yearly_logs", 0),
		}

		cursor, err := database.GetCollection(model.UserTrafficLogs{}).Aggregate(context.Background(), pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetUserByName aggregate: %s", err.Error())
			return
		}
		defer cursor.Close(context.Background())

		var users []UserTrafficLogs
		if err := cursor.All(context.Background(), &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetUserByName cursor.All: %s", err.Error())
			return
		}
		if len(users) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			log.Printf("GetUserByName: user %s not found", name)
			return
		}

		c.JSON(http.StatusOK, users[0])
	}
}

func GetSubscripionURL() gin.HandlerFunc {
	return func(c *gin.Context) {

		var subscription []byte
		var err error
		name := helper.SanitizeStr(c.Param("name"))

		var activeGlobalNodes []SubscriptionNode

		// 查询所有节点并按权重升序排序
		cur, err := database.GetCollection(model.SubscriptionNode{}).Find(
			context.TODO(),
			bson.D{{Key: "status", Value: bson.D{{Key: "$ne", Value: "inactive"}}}},
			options.Find().SetSort(bson.D{{Key: "weight", Value: 1}}), // 按权重升序排序
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting active global nodes"})
			log.Printf("Getting active global nodes error: %s", err.Error())
			return
		}
		defer cur.Close(context.Background())

		cur.All(context.Background(), &activeGlobalNodes)

		// projections include status, user_id, uuid,
		var projections = bson.D{
			{Key: "status", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "uuid", Value: 1},
		}
		var user UserTrafficLogs
		err = database.GetCollection(model.UserTrafficLogs{}).FindOne(context.TODO(), bson.M{"email_as_id": name}, options.FindOne().SetProjection(projections)).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("GetSubscripionURL error: %v", err)
			return
		}

		if user.Status == "plain" {
			var sub string
			for _, node := range activeGlobalNodes {

				// 格式化IP地址以支持IPv6
				formattedIP := helper.FormatIPForURL(node.IP)

				if node.Type == "reality" {
					if len(sub) == 0 {
						sub = "vless://" + user.UUID + "@" + formattedIP + ":" + node.SERVER_PORT + "?encryption=none&flow=xtls-rprx-vision&security=reality&sni=www.microsoft.com&fp=chrome&pbk=" + getPublicKey() + "&sid=" + getShortID() + "&type=tcp&headerType=none#" + node.Remark
					} else {
						sub = sub + "\n" + "vless://" + user.UUID + "@" + formattedIP + ":" + node.SERVER_PORT + "?encryption=none&flow=xtls-rprx-vision&security=reality&sni=www.microsoft.com&fp=chrome&pbk=" + getPublicKey() + "&sid=" + getShortID() + "&type=tcp&headerType=none#" + node.Remark
					}
				}

				if node.Type == "hysteria2" {
					if len(sub) == 0 {
						sub = "hysteria2://" + user.User_id + "@" + formattedIP + ":" + node.SERVER_PORT + "?insecure=1&sni=bing.com#" + node.Remark
					} else {
						sub = sub + "\n" + "hysteria2://" + user.User_id + "@" + formattedIP + ":" + node.SERVER_PORT + "?insecure=1&sni=bing.com#" + node.Remark
					}
				}

				if node.Type == "vlessCDN" {
					if len(sub) == 0 {
						sub = "vless://" + node.UUID + "@" + formattedIP + ":" + node.SERVER_PORT + "?encryption=none&security=tls&sni=" + node.Domain + "&fp=randomized&type=ws&host=" + node.Domain + "&path=%2F%3Fed%3D2048#" + node.Remark
					} else {
						sub = sub + "\n" + "vless://" + node.UUID + "@" + formattedIP + ":" + node.SERVER_PORT + "?encryption=none&security=tls&sni=" + node.Domain + "&fp=randomized&type=ws&host=" + node.Domain + "&path=%2F%3Fed%3D2048#" + node.Remark
					}
				}
			}

			subscription = []byte(b64.StdEncoding.EncodeToString([]byte(sub)))
		} else {
			subscription, err = os.ReadFile(helper.CurrentPath() + "/config/error.txt")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("GetSubscripionURL error: %v", err)
				// return
			}
		}

		c.Data(http.StatusOK, "text/plain", subscription)
	}
}

// ReturnSingboxJson
func ReturnSingboxJson() gin.HandlerFunc {
	return func(c *gin.Context) {

		name := helper.SanitizeStr(c.Param("name"))

		var err error
		var jsonFile []byte
		var singboxJSON = SingboxJSON{}
		var user UserTrafficLogs

		var projections = bson.D{
			{Key: "status", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "uuid", Value: 1},
		}
		err = database.GetCollection(model.UserTrafficLogs{}).FindOne(context.TODO(), bson.M{"email_as_id": name}, options.FindOne().SetProjection(projections)).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("ReturnSingboxJson failed: %s", err.Error())
			return
		}

		var subscriptionNodes []SubscriptionNode

		// 查询所有节点并按权重升序排序
		cur, err := database.GetCollection(model.SubscriptionNode{}).Find(
			context.TODO(),
			bson.D{{Key: "status", Value: bson.D{{Key: "$ne", Value: "inactive"}}}},
			options.Find().SetSort(bson.D{{Key: "weight", Value: 1}}), // 按权重升序排序
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting active global nodes"})
			log.Printf("Getting active global nodes error: %s", err.Error())
			return
		}
		defer cur.Close(context.Background())

		cur.All(context.Background(), &subscriptionNodes)

		if user.Status == "plain" {

			jsonFile, err = os.ReadFile(helper.CurrentPath() + "/config/template_singbox.json")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = json.Unmarshal(jsonFile, &singboxJSON)
			if err != nil {
				log.Printf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// append reality and hysteria2 nodes to outbounds in jsonfile.
			for _, node := range subscriptionNodes {

				server_port, _ := strconv.Atoi(node.SERVER_PORT)
				var outboundTags = []string{
					"manual-select",
					"auto",
					"WeChat",
					"Apple",
					"Microsoft",
				}

				if node.Type == "reality" {

					for i, outbound := range singboxJSON.Outbounds {
						if outboundMap, ok := outbound.(map[string]interface{}); ok {
							if Contains(outboundTags, outboundMap["tag"].(string)) || (node.EnableOpenai) && outboundMap["tag"] == "Openai" {
								if outbounds, ok := singboxJSON.Outbounds[i].(map[string]interface{}); ok {
									if outboundsList, ok := outbounds["outbounds"].([]interface{}); ok {
										singboxJSON.Outbounds[i].(map[string]interface{})["outbounds"] = append(outboundsList, node.Remark)
									}
								}
							}
						}
					}

					singboxJSON.Outbounds = append(singboxJSON.Outbounds, RealityJSON{
						Tag:            node.Remark,
						Type:           "vless",
						UUID:           user.UUID,
						ServerPort:     server_port,
						Flow:           "xtls-rprx-vision",
						PacketEncoding: "xudp",
						Server:         helper.FormatIPForURL(node.IP),
						TLS: struct {
							Enabled    bool   `json:"enabled"`
							ServerName string `json:"server_name"`
							Utls       struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							} `json:"utls"`
							Reality struct {
								Enabled   bool   `json:"enabled"`
								PublicKey string `json:"public_key"`
								ShortID   string `json:"short_id"`
							} `json:"reality"`
						}{
							Enabled:    true,
							ServerName: "www.microsoft.com",
							Utls: struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							}{
								Enabled:     true,
								Fingerprint: "chrome",
							},
							Reality: struct {
								Enabled   bool   `json:"enabled"`
								PublicKey string `json:"public_key"`
								ShortID   string `json:"short_id"`
							}{
								Enabled:   true,
								PublicKey: getPublicKey(),
								ShortID:   getShortID(),
							},
						},
					})
				}

				if node.Type == "hysteria2" {

					for i, outbound := range singboxJSON.Outbounds {
						if outboundMap, ok := outbound.(map[string]interface{}); ok {
							if Contains(outboundTags, outboundMap["tag"].(string)) || (node.EnableOpenai) && outboundMap["tag"] == "Openai" {
								if outbounds, ok := singboxJSON.Outbounds[i].(map[string]interface{}); ok {
									if outboundsList, ok := outbounds["outbounds"].([]interface{}); ok {
										singboxJSON.Outbounds[i].(map[string]interface{})["outbounds"] = append(outboundsList, node.Remark)
									}
								}
							}
						}
					}

					singboxJSON.Outbounds = append(singboxJSON.Outbounds, Hysteria2JSON{
						Tag:        node.Remark,
						Type:       "hysteria2",
						Server:     helper.FormatIPForURL(node.IP),
						ServerPort: server_port,
						UpMbps:     100,
						DownMbps:   100,
						Password:   user.User_id,
						TLS: struct {
							Enabled    bool     `json:"enabled"`
							ServerName string   `json:"server_name"`
							Insecure   bool     `json:"insecure"`
							Alpn       []string `json:"alpn"`
						}{
							Enabled:    true,
							ServerName: "bing.com",
							Insecure:   true,
							Alpn:       []string{"h3"},
						},
					})
				}

				if node.Type == "vlessCDN" {

					for i, outbound := range singboxJSON.Outbounds {
						if outboundMap, ok := outbound.(map[string]interface{}); ok {
							if Contains(outboundTags, outboundMap["tag"].(string)) || (node.EnableOpenai) && outboundMap["tag"] == "Openai" {
								if outbounds, ok := singboxJSON.Outbounds[i].(map[string]interface{}); ok {
									if outboundsList, ok := outbounds["outbounds"].([]interface{}); ok {
										singboxJSON.Outbounds[i].(map[string]interface{})["outbounds"] = append(outboundsList, node.Remark)
									}
								}
							}
						}
					}

					singboxJSON.Outbounds = append(singboxJSON.Outbounds, CFVlessJSON{
						Tag:        node.Remark,
						Type:       "vless",
						Server:     helper.FormatIPForURL(node.IP),
						ServerPort: server_port,
						UUID:       node.UUID,
						Flow:       "",
						TLS: struct {
							Enabled    bool   `json:"enabled"`
							ServerName string `json:"server_name"`
							Insecure   bool   `json:"insecure"`
							Utls       struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							} `json:"utls"`
						}{
							Enabled:    true,
							ServerName: node.Domain,
							Insecure:   false,
							Utls: struct {
								Enabled     bool   `json:"enabled"`
								Fingerprint string `json:"fingerprint"`
							}{
								Enabled:     true,
								Fingerprint: "chrome",
							},
						},
						Multiplex: struct {
							Enabled    bool   `json:"enabled"`
							Protocol   string `json:"protocol"`
							MaxStreams int    `json:"max_streams"`
						}{
							Enabled:    false,
							Protocol:   "smux",
							MaxStreams: 32,
						},
						PacketEncoding: "xudp",
						Transport: struct {
							Type    string `json:"type"`
							Path    string `json:"path"`
							Headers struct {
								Host string `json:"Host"`
							} `json:"headers"`
						}{
							Type: "ws",
							Path: "/?ed=2048",
							Headers: struct {
								Host string `json:"Host"`
							}{
								Host: node.Domain,
							},
						},
					})

				}
			}

		} else {
			jsonFile, err = os.ReadFile(helper.CurrentPath() + "/config/error.json")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = json.Unmarshal(jsonFile, &singboxJSON)
			if err != nil {
				log.Printf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, singboxJSON)
	}
}

// ReturnVergeYAML: return yaml file
func ReturnVergeYAML() gin.HandlerFunc {
	return func(c *gin.Context) {

		name := helper.SanitizeStr(c.Param("name"))

		var err error
		var yamlFile []byte
		var singboxYAML = SingboxYAML{}

		var projections = bson.D{
			{Key: "status", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "uuid", Value: 1},
		}
		var user UserTrafficLogs
		err = database.GetCollection(model.UserTrafficLogs{}).FindOne(context.TODO(), bson.M{"email_as_id": name}, options.FindOne().SetProjection(projections)).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("ReturnVergeYAML failed: %s", err.Error())
			return
		}

		var subscriptionNodes []SubscriptionNode
		// 查询所有节点并按权重升序排序
		cur, err := database.GetCollection(model.SubscriptionNode{}).Find(
			context.TODO(),
			bson.D{{Key: "status", Value: bson.D{{Key: "$ne", Value: "inactive"}}}},
			options.Find().SetSort(bson.D{{Key: "weight", Value: 1}}), // 按权重升序排序
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting active global nodes"})
			log.Printf("Getting active global nodes error: %s", err.Error())
			return
		}
		defer cur.Close(context.Background())

		cur.All(context.Background(), &subscriptionNodes)

		if user.Status == "plain" {
			yamlFile, err = os.ReadFile(helper.CurrentPath() + "/config/template_verge.yaml")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = yaml.Unmarshal(yamlFile, &singboxYAML)
			if err != nil {
				log.Printf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// append reality and hysteria2 nodes to outbounds in yamlfile.
			for _, node := range subscriptionNodes {

				server_port, _ := strconv.Atoi(node.SERVER_PORT)
				if node.Type == "reality" {

					for i, proxy := range singboxYAML.ProxyGroups {
						if proxy.Type == "select" || proxy.Type == "url-test" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, node.Remark)
						}
					}

					singboxYAML.Proxies = append(singboxYAML.Proxies, RealityYAML{
						Name:              node.Remark,
						Type:              "vless",
						Server:            helper.FormatIPForURL(node.IP),
						Port:              server_port,
						UUID:              user.UUID,
						Network:           "tcp",
						UDP:               true,
						TLS:               true,
						Flow:              "xtls-rprx-vision",
						Servername:        "www.microsoft.com",
						ClientFingerprint: "chrome",
						RealityOpts: struct {
							PublicKey string `yaml:"public-key"`
							ShortID   string `yaml:"short-id"`
						}{
							PublicKey: getPublicKey(),
							ShortID:   getShortID(),
						},
					})
				}

				if node.Type == "hysteria2" {

					for i, proxy := range singboxYAML.ProxyGroups {
						if proxy.Type == "select" || proxy.Type == "url-test" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, node.Remark)
						}
					}

					singboxYAML.Proxies = append(singboxYAML.Proxies, Hysteria2YAML{
						Name:           node.Remark,
						Type:           "hysteria2",
						Server:         helper.FormatIPForURL(node.IP),
						Port:           server_port,
						Password:       user.User_id,
						Sni:            "bing.com",
						SkipCertVerify: true,
						Alpn:           []string{"h3"},
					})
				}

				if node.Type == "vlessCDN" {

					for i, proxy := range singboxYAML.ProxyGroups {
						if proxy.Type == "select" || proxy.Type == "url-test" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, node.Remark)
						}
					}

					singboxYAML.Proxies = append(singboxYAML.Proxies, CFVlessYAML{
						Name:              node.Remark,
						Type:              "vless",
						Server:            helper.FormatIPForURL(node.IP),
						Port:              server_port,
						UUID:              node.UUID,
						Network:           "ws",
						TLS:               true,
						UDP:               false,
						Servername:        node.Domain,
						ClientFingerprint: "chrome",
						WsOpts: struct {
							Path    string `yaml:"path"`
							Headers struct {
								Host string `yaml:"Host"`
							} `yaml:"headers"`
						}{
							Path: node.PATH,
							Headers: struct {
								Host string `yaml:"Host"`
							}{
								Host: node.Domain,
							},
						},
					})
				}
			}

			// if DIRECT type is not at the end of singboxYAML.ProxyGroups at select type, set it to the end.
			for i, proxy := range singboxYAML.ProxyGroups {
				if proxy.Type == "select" {
					for j, p := range proxy.Proxies {
						if p == "DIRECT" {
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies[:j], singboxYAML.ProxyGroups[i].Proxies[j+1:]...)
							singboxYAML.ProxyGroups[i].Proxies = append(singboxYAML.ProxyGroups[i].Proxies, "DIRECT")
						}
					}
				}
			}

		} else {
			yamlFile, err = os.ReadFile(helper.CurrentPath() + "/config/error.yaml")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error: %v", err)
				return
			}

			err = yaml.Unmarshal(yamlFile, &singboxYAML)
			if err != nil {
				log.Fatalf("error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.YAML(http.StatusOK, singboxYAML)
	}
}

// DisableUser 禁用用户 - 将用户状态设为deleted
func DisableUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var foundUser UserTrafficLogs
		err := database.GetCollection(model.UserTrafficLogs{}).FindOne(ctx, bson.M{"email_as_id": helper.SanitizeStr(name)}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("user not found: %s", name)
			return
		}

		// 不允许禁用管理员账户
		if foundUser.Role == "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot disable admin user"})
			log.Printf("attempted to disable admin user: %s", name)
			return
		}

		// 更新用户状态为deleted
		updateData := bson.M{
			"status":     "deleted",
			"updated_at": time.Now(),
		}

		var updatedUser UserTrafficLogs
		err = database.GetCollection(model.UserTrafficLogs{}).FindOneAndUpdate(
			ctx,
			bson.M{"email_as_id": helper.SanitizeStr(name)},
			bson.M{"$set": updateData},
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&updatedUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error disabling user: %v", err)
			return
		}

		// 异步通知 sing-box 把该用户从认证列表中移除
		go singboxctl.SafeDisableUser(updatedUser.Email_As_Id)

		log.Printf("User %s disabled successfully", updatedUser.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User " + updatedUser.Name + " disabled successfully"})
	}
}

// EnableUser 启用用户 - 将用户状态设为plain
func EnableUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var foundUser UserTrafficLogs
		err := database.GetCollection(model.UserTrafficLogs{}).FindOne(ctx, bson.M{"email_as_id": helper.SanitizeStr(name)}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("user not found: %s", name)
			return
		}

		// 更新用户状态为plain
		updateData := bson.M{
			"status":     "plain",
			"updated_at": time.Now(),
		}

		var updatedUser UserTrafficLogs
		err = database.GetCollection(model.UserTrafficLogs{}).FindOneAndUpdate(
			ctx,
			bson.M{"email_as_id": helper.SanitizeStr(name)},
			bson.M{"$set": updateData},
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&updatedUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error enabling user: %v", err)
			return
		}

		// 异步通知 sing-box 把该用户重新注册到认证列表
		go singboxctl.SafeEnableUser(singboxctl.AddUserRequest{
			EmailAsId: updatedUser.Email_As_Id,
			UUID:      updatedUser.UUID,
			UserId:    updatedUser.User_id,
		})

		log.Printf("User %s enabled successfully", updatedUser.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User " + updatedUser.Name + " enabled successfully"})
	}
}
