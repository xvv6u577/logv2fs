package controllers

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	uuid "github.com/nu7hatch/gouuid"
	"gopkg.in/yaml.v2"

	localCron "github.com/caster8013/logv2rayfullstack/cron"
	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/v2ray"

	helper "github.com/caster8013/logv2rayfullstack/helpers"

	yamlTools "github.com/caster8013/logv2rayfullstack/config"
	"github.com/caster8013/logv2rayfullstack/grpctools"
	"github.com/caster8013/logv2rayfullstack/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	userCollection   *mongo.Collection = database.OpenCollection(database.Client, "USERS")
	nodeCollection   *mongo.Collection = database.OpenCollection(database.Client, "NODES")
	globalCollection *mongo.Collection = database.OpenCollection(database.Client, "GLOBAL")
	validate                           = validator.New()
	V2_API_ADDRESS                     = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT                        = os.Getenv("V2_API_PORT")
	NODE_TYPE                          = os.Getenv("NODE_TYPE")
	CURRENT_DOMAIN                     = os.Getenv("CURRENT_DOMAIN")
	MIXED_PORT                         = os.Getenv("MIXED_PORT")
	ADMINUSERID                        = os.Getenv("ADMINUSERID")
	CREDIT                             = os.Getenv("CREDIT")
)

type (
	User            = model.User
	TrafficAtPeriod = model.TrafficAtPeriod
	Node            = model.Node
	YamlTemplate    = model.YamlTemplate
	Proxies         = model.Proxies
	Headers         = model.Headers
	WsOpts          = model.WsOpts
	ProxyGroups     = model.ProxyGroups
	CurrentNode     = model.CurrentNode
	NodeAtPeriod    = model.NodeAtPeriod
	GlobalVariable  = model.GlobalVariable
	Domain          = model.Domain
	DomainInfo      = model.DomainInfo
	SingboxYAML     = model.SingboxYAML
	SingboxJSON     = model.SingboxJSON
)

//HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

//VerifyPassword checks the input password while verifying it with the passward in the DB.
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
	_, err := userCollection.UpdateOne(
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

//CreateUser is the api used to tget a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var user model.User
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

		user_email := helper.SanitizeStr(user.Email)
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user_email})
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

		// get ActiveGlobalNodes from globalCollection, seperate out vmess nodes and put them into vmessNodes
		var globalVariable GlobalVariable
		err = globalCollection.FindOne(ctx, bson.M{"name": "GLOBAL"}).Decode(&globalVariable)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting globalVariable"})
			log.Printf("Getting globalVariable error: %s", err.Error())
			return
		}

		if user.Name == "" {
			user.Name = user_email
		}

		if user.Path == "" {
			user.Path = "ray"
		}

		password := HashPassword(user.Password)
		user.Password = password

		user.CreatedAt = current
		user.UpdatedAt = current

		if user.UUID == "" {
			uuidV4, _ := uuid.NewV4()
			user.UUID = uuidV4.String()
		}
		user_role := helper.SanitizeStr(user.Role)

		if user.Credittraffic == 0 {
			credit, _ := strconv.ParseInt(CREDIT, 10, 64)
			user.Credittraffic = credit
		}
		if user.Usedtraffic == 0 {
			user.Usedtraffic = 0
		}

		var current_year = current.Local().Format("2006")
		var current_month = current.Local().Format("200601")
		var current_day = current.Local().Format("20060102")

		user.UsedByCurrentDay = TrafficAtPeriod{
			Period:       current_day,
			Amount:       0,
			UsedByDomain: map[string]int64{},
		}
		user.UsedByCurrentMonth = TrafficAtPeriod{
			Period:       current_month,
			Amount:       0,
			UsedByDomain: map[string]int64{},
		}
		user.UsedByCurrentYear = TrafficAtPeriod{
			Period:       current_year,
			Amount:       0,
			UsedByDomain: map[string]int64{},
		}

		user.TrafficByDay = []TrafficAtPeriod{}
		user.TrafficByMonth = []TrafficAtPeriod{}
		user.TrafficByYear = []TrafficAtPeriod{}

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(user_email, user.UUID, user.Path, user_role, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken
		user.UpdateNodeStatusInUse(globalVariable.ActiveGlobalNodes)

		var wg sync.WaitGroup
		var waitQueueLength = 0

		if NODE_TYPE == "local" {
			wg.Add(waitQueueLength + 1)
			go func() {
				defer wg.Done()
				err = grpctools.GrpcClientToAddUser("0.0.0.0", "50051", user, false)
				if err != nil {
					log.Printf("error occured while adding user: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}()
		} else {
			wg.Add(waitQueueLength + helper.CountNodesInUse(user.NodeInUseStatus))
			for node, available := range user.NodeInUseStatus {
				if available {
					go func(domain string) {
						defer wg.Done()
						grpctools.GrpcClientToAddUser(domain, MIXED_PORT, user, true)
					}(node)
				}
			}
		}

		wg.Wait()

		user.ProduceSuburl(globalVariable.ActiveGlobalNodes)
		err = user.GenerateYAML(globalVariable.ActiveGlobalNodes)
		if err != nil {
			log.Printf("error occured while generating yaml: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while inserting user: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user " + user.Name + " created at v2ray and database."})
	}
}

//Login is the api used to get a single user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var boundUser model.User
		var foundUser model.User

		if err := c.BindJSON(&boundUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		sanitized_email := helper.SanitizeStr(boundUser.Email)
		err := userCollection.FindOne(ctx, bson.M{"email": sanitized_email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		passwordIsValid, msg := VerifyPassword(boundUser.Password, foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("password is not valid: %s", msg)
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(sanitized_email, foundUser.UUID, foundUser.Path, foundUser.Role, foundUser.User_id)

		UpdateAllTokens(token, refreshToken, foundUser.User_id)
		var projections = bson.D{
			{Key: "token", Value: 1},
		}

		user, err := database.GetUserByName(boundUser.Email, projections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func GetUserSimpleInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		var name string
		if c.Query("name") != "" {
			name = c.Query("name")
		} else {
			name = c.Param("name")
		}

		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "path", Value: 1},
			{Key: "node_in_use_status", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"email": user.Email, "uuid": user.UUID, "path": user.Path, "nodeinuse": user.NodeInUseStatus})
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

		var user, foundUser model.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": helper.SanitizeStr(user.Email)}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email is incorrect"})
			log.Printf("email is incorrect")
			return
		}

		var replacedDocument bson.M
		newFoundUser := bson.M{}

		if foundUser.Role != user.Role {
			newFoundUser["role"] = user.Role
		}

		if user.Password != "" {
			password := HashPassword(user.Password)
			if foundUser.Password != user.Password && foundUser.Password != password {
				newFoundUser["password"] = password
			}
		}
		if foundUser.Name != user.Name {
			newFoundUser["name"] = user.Name
		}
		if foundUser.Usedtraffic != user.Usedtraffic {
			newFoundUser["used"] = user.Usedtraffic
		}
		if foundUser.Credittraffic != user.Credittraffic {
			newFoundUser["credit"] = user.Credittraffic
		}

		if len(newFoundUser) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no new value in post data."})
			log.Printf("no new value in post data.")
			return
		}

		err = userCollection.FindOneAndUpdate(
			ctx,
			bson.M{"email": helper.SanitizeStr(user.Email)},
			bson.M{"$set": newFoundUser},
			options.FindOneAndUpdate().SetUpsert(true),
		).Decode(&replacedDocument)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func TakeItOfflineByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var globalVariable GlobalVariable
		err := globalCollection.FindOne(ctx, bson.M{"name": "GLOBAL"}).Decode(&globalVariable)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting globalVariable"})
			log.Printf("Getting globalVariable error: %s", err.Error())
			return
		}

		name := helper.SanitizeStr(c.Param("name"))
		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "used", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		if NODE_TYPE == "local" {
			err = grpctools.GrpcClientToDeleteUser("0.0.0.0", "50051", user, false)
			if err != nil {
				msg := "v2ray take user back online failed."
				log.Panicf("%v", msg)
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			for node, enable := range user.NodeInUseStatus {
				if enable {
					user.NodeInUseStatus[node] = false
				}
			}
		} else {
			var wg sync.WaitGroup
			for node, available := range user.NodeInUseStatus {
				if available {
					wg.Add(1)
					go func(domain string) {
						defer wg.Done()
						grpctools.GrpcClientToDeleteUser(domain, MIXED_PORT, user, true)
					}(node)
					user.NodeInUseStatus[node] = false
				}
			}
			wg.Wait()
		}

		user.Status = v2ray.DELETE
		user.UpdateNodeStatusInUse(globalVariable.ActiveGlobalNodes)
		user.ProduceSuburl(globalVariable.ActiveGlobalNodes)
		err = user.GenerateYAML(globalVariable.ActiveGlobalNodes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("TakeItOfflineByUserName generating yaml error: %v", err)
			return
		}

		filter := bson.D{primitive.E{Key: "email", Value: name}}
		update := bson.M{"$set": bson.M{"status": v2ray.DELETE, "node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		log.Printf("user %s is offline", user.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User " + user.Name + " is offline!"})
	}
}

func TakeItOnlineByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		name := helper.SanitizeStr(c.Param("name"))

		var globalVariable GlobalVariable
		err := globalCollection.FindOne(ctx, bson.M{"name": "GLOBAL"}).Decode(&globalVariable)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting globalVariable"})
			log.Printf("Getting globalVariable error: %s", err.Error())
			return
		}

		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "used", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}

		user, err := database.GetUserByName(name, projections)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		if NODE_TYPE == "local" {
			err = grpctools.GrpcClientToAddUser("0.0.0.0", "50051", user, false)
			if err != nil {
				msg := "v2ray take user back online failed."
				log.Panicf("%v", msg)
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			for node, enable := range user.NodeInUseStatus {
				if !enable {
					user.NodeInUseStatus[node] = true
				}
			}
		} else {
			var wg sync.WaitGroup
			for node, available := range user.NodeInUseStatus {
				if !available {
					wg.Add(1)
					go func(domain string) {
						defer wg.Done()
						grpctools.GrpcClientToAddUser(node, MIXED_PORT, user, true)
					}(node)
					user.NodeInUseStatus[node] = true
				}
			}
			wg.Wait()
		}

		user.Status = v2ray.PLAIN
		user.UpdateNodeStatusInUse(globalVariable.ActiveGlobalNodes)
		user.ProduceSuburl(globalVariable.ActiveGlobalNodes)
		err = user.GenerateYAML(globalVariable.ActiveGlobalNodes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("TakeItOnlineByUserName generating yaml error: %v", err)
			return
		}

		filter := bson.D{primitive.E{Key: "email", Value: name}}
		update := bson.M{"$set": bson.M{"status": v2ray.PLAIN, "node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		log.Printf("user %s is online", user.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User " + user.Name + " is online!"})
	}
}

func DeleteUserByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := c.Param("name")
		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "used", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		if user.Status == "plain" {
			if NODE_TYPE == "local" {
				err = grpctools.GrpcClientToDeleteUser("0.0.0.0", "50051", user, false)
				if err != nil {
					msg := "v2ray take user offline failed."
					log.Panicf("%v", msg)
					c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
					return
				}
				for node, enable := range user.NodeInUseStatus {
					if enable {
						user.NodeInUseStatus[node] = false
					}
				}
			} else {
				for node, available := range user.NodeInUseStatus {
					if available {
						grpctools.GrpcClientToDeleteUser(node, MIXED_PORT, user, true)
					}
				}
			}
		}

		err = database.DeleteUserByName(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("DeleteUserByUserName: %s", err.Error())
			return
		}

		err = yamlTools.RemoveOne(user.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Remove yaml file by user name: %s", err.Error())
			return
		}

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

		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "used", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
			{Key: "used_by_current_day", Value: 1},
			{Key: "used_by_current_month", Value: 1},
			{Key: "used_by_current_year", Value: 1},
		}

		allUsers, err := database.GetAllUsersPartialInfo(projections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetAllUsers failed: %s", err.Error())
			return
		}

		if len(allUsers) == 0 {
			c.JSON(http.StatusOK, []User{})
			return
		}

		if NODE_TYPE == "local" {
			for _, user := range allUsers {
				user.NodeInUseStatus = map[string]bool{CURRENT_DOMAIN: user.NodeInUseStatus[CURRENT_DOMAIN]}
			}
			c.JSON(http.StatusOK, allUsers)
			return
		}

		c.JSON(http.StatusOK, allUsers)
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

		var projections = bson.D{
			{Key: "used_by_current_day", Value: 1},
			{Key: "used_by_current_month", Value: 1},
			{Key: "used_by_current_year", Value: 1},
			{Key: "traffic_by_day", Value: 1},
			{Key: "traffic_by_month", Value: 1},
			{Key: "traffic_by_year", Value: 1},
			{Key: "used", Value: 1},
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Get user by name failed: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func WriteToDB() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := localCron.Log_basicAction()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Write to DB failed: %s", err.Error())
			return
		}

		log.Println("Write to DB by hand!")
		c.JSON(http.StatusOK, gin.H{"message": "Write to DB successfully!"})
	}
}

func DisableNodePerUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		email := helper.SanitizeStr(c.Request.URL.Query().Get("email"))
		node := helper.SanitizeStr(c.Request.URL.Query().Get("node"))

		var globalVariable GlobalVariable
		err := globalCollection.FindOne(ctx, bson.M{"name": "GLOBAL"}).Decode(&globalVariable)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting globalVariable"})
			log.Printf("Getting globalVariable error: %s", err.Error())
			return
		}

		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "used", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}
		user, err := database.GetUserByName(email, projections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Get user by name failed: %s", err.Error())
			return
		}

		if NODE_TYPE == "local" {
			if CURRENT_DOMAIN == node {

				err = grpctools.GrpcClientToDeleteUser("0.0.0.0", "50051", user, false)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Delete user on node failed: %s", err.Error())
					return
				}

			}
		} else {
			grpctools.GrpcClientToDeleteUser(node, MIXED_PORT, user, true)
		}

		user.DeleteNodeInUse(node)
		user.UpdateNodeStatusInUse(globalVariable.ActiveGlobalNodes)
		user.ProduceSuburl(globalVariable.ActiveGlobalNodes)
		user.GenerateYAML(globalVariable.ActiveGlobalNodes)

		filter := bson.D{primitive.E{Key: "email", Value: email}}
		update := bson.M{"$set": bson.M{"node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		log.Printf("Disable user: %v, node: %v by hand!", helper.SanitizeStr(email), helper.SanitizeStr(node))
		c.JSON(http.StatusOK, gin.H{"message": "Disable user: " + email + " at node: " + node + " successfully!"})
	}
}

func EnableNodePerUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		email := helper.SanitizeStr(c.Request.URL.Query().Get("email"))
		node := helper.SanitizeStr(c.Request.URL.Query().Get("node"))

		var globalVariable GlobalVariable
		err := globalCollection.FindOne(ctx, bson.M{"name": "GLOBAL"}).Decode(&globalVariable)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while getting globalVariable"})
			log.Printf("Getting globalVariable error: %s", err.Error())
			return
		}

		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "used", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}
		user, err := database.GetUserByName(email, projections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Get user by name failed: %s", err.Error())
			return
		}

		if NODE_TYPE == "local" {
			if CURRENT_DOMAIN == node {
				err = grpctools.GrpcClientToAddUser("0.0.0.0", "50051", user, false)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					log.Printf("Add user on node failed: %s", err.Error())
					return
				}
			}
		} else {
			grpctools.GrpcClientToAddUser(node, MIXED_PORT, user, true)
		}

		user.AddNodeInUse(node)
		user.UpdateNodeStatusInUse(globalVariable.ActiveGlobalNodes)
		user.ProduceSuburl(globalVariable.ActiveGlobalNodes)
		user.GenerateYAML(globalVariable.ActiveGlobalNodes)

		filter := bson.D{primitive.E{Key: "email", Value: email}}
		update := bson.M{"$set": bson.M{"node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		log.Printf("Enable user: %v, node: %v by hand!", helper.SanitizeStr(email), helper.SanitizeStr(node))
		c.JSON(http.StatusOK, gin.H{"message": "Enable user: " + email + " at node: " + node + " successfully!"})
	}
}

func GetSubscripionURL() gin.HandlerFunc {
	return func(c *gin.Context) {

		name := helper.SanitizeStr(c.Param("name"))
		var projections = bson.D{
			{Key: "status", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("GetUserByName failed: %s", err.Error())
			return
		}

		if user.Status != "plain" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "there is problem with this user, please contact admin"})
			log.Printf(user.Name + ": GetSubscripionURL Error!")
			return
		}

		file, err := os.ReadFile(helper.CurrentPath() + "/sing-box-full-platform/sing-box.txt")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetSubscripionURL error: %v", err)
			return
		}

		c.Data(http.StatusOK, "text/plain", file)
	}
}

// ReturnSingboxJson
func ReturnSingboxJson() gin.HandlerFunc {
	return func(c *gin.Context) {

		name := helper.SanitizeStr(c.Param("name"))
		var projections = bson.D{
			{Key: "status", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("ReturnSingboxJson failed: %s", err.Error())
			return
		}

		if user.Status != "plain" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "there is problem with this user, please contact admin"})
			log.Printf(user.Name + ": ReturnSingboxJson Error!")
			return
		}

		// read json file from sing-box-full-platform/sing-box.json, and return it.
		var singboxJSON = SingboxJSON{}
		jsonFile, err := os.ReadFile(helper.CurrentPath() + "/sing-box-full-platform/sing-box.json")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		err = json.Unmarshal(jsonFile, &singboxJSON)
		if err != nil {
			log.Fatalf("error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// return json file
		c.JSON(http.StatusOK, singboxJSON)
	}
}

// ReturnVergeYAML: return yaml file
func ReturnVergeYAML() gin.HandlerFunc {
	return func(c *gin.Context) {

		name := helper.SanitizeStr(c.Param("name"))

		var projections = bson.D{
			{Key: "status", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("ReturnSingboxJson failed: %s", err.Error())
			return
		}

		if user.Status != "plain" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "there is problem with this user, please contact admin"})
			log.Printf(user.Name + ": ReturnVergeYAML Error!")
			return
		}

		var singboxYAML = SingboxYAML{}
		yamlFile, err := os.ReadFile(helper.CurrentPath() + "/sing-box-full-platform/sing-box.yaml")
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

		// return yaml file as attachment.
		c.Header("Content-Disposition", "attachment; filename="+name+".yaml")
		c.Data(http.StatusOK, "application/octet-stream", yamlFile)
	}
}
