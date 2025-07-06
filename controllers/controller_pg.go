package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	validatePG = validator.New()
)

// HashPasswordPG is used to encrypt the password before it is stored in the DB (PostgreSQL version)
func HashPasswordPG(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

// VerifyPasswordPG checks the input password while verifying it with the password in the DB (PostgreSQL version).
func VerifyPasswordPG(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or password is incorrect"
		check = false
	}

	return check, msg
}

// PostgreSQL版本的用户操作函数

// SignUpPG 创建新用户 - PostgreSQL版本
func SignUpPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var user UserTrafficLogs
		var current = time.Now()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		validationErr := validatePG.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			log.Printf("validate error: %v", validationErr)
			return
		}

		user_email := helper.SanitizeStr(user.Email_As_Id)

		// 检查用户是否已存在
		var count int64
		db.Model(&model.UserTrafficLogsPG{}).Where("email_as_id = ?", user_email).Count(&count)
		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
			log.Printf("this email already exists")
			return
		}

		if user.Name == "" {
			user.Name = user_email
		}

		password := HashPasswordPG(user_email)

		// 创建PostgreSQL用户记录
		pgUser := model.UserTrafficLogsPG{
			ID:        uuid.New(),
			EmailAsId: user_email,
			Password:  password,
			Name:      user.Name,
			Role:      user.Role,
			Status:    "plain",
			CreatedAt: current,
			UpdatedAt: current,
			UUID:      uuid.New().String(),
			Used:      0,
		}

		user_role := "plain"
		if user.Credit == 0 {
			credit, _ := strconv.ParseInt(os.Getenv("CREDIT"), 10, 64)
			pgUser.Credit = credit
		} else {
			pgUser.Credit = user.Credit
		}

		token, refreshToken, _ := helper.GenerateAllTokens(user_email, pgUser.UUID, pgUser.Name, user_role, pgUser.ID.String())
		pgUser.Token = &token
		pgUser.RefreshToken = &refreshToken

		// 初始化空的日志数组
		hourlyLogs, _ := json.Marshal([]model.TrafficLogEntry{})
		dailyLogs, _ := json.Marshal([]model.DailyLogEntry{})
		monthlyLogs, _ := json.Marshal([]model.MonthlyLogEntry{})
		yearlyLogs, _ := json.Marshal([]model.YearlyLogEntry{})

		pgUser.HourlyLogs = datatypes.JSON(hourlyLogs)
		pgUser.DailyLogs = datatypes.JSON(dailyLogs)
		pgUser.MonthlyLogs = datatypes.JSON(monthlyLogs)
		pgUser.YearlyLogs = datatypes.JSON(yearlyLogs)

		// 保存到数据库
		if err := db.Create(&pgUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occurred while inserting user traffic logs: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user " + pgUser.Name + " created successfully"})
	}
}

// LoginPG 用户登录 - PostgreSQL版本
func LoginPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetPostgresDB()
		var boundUser UserTrafficLogs
		var pgUser model.UserTrafficLogsPG

		if err := c.BindJSON(&boundUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		sanitized_email := helper.SanitizeStr(boundUser.Email_As_Id)

		// 查找用户
		if err := db.Where("email_as_id = ?", sanitized_email).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			log.Printf("error: %v", err)
			return
		}

		passwordIsValid, msg := VerifyPasswordPG(boundUser.Password, pgUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("password is not valid: %s", msg)
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(sanitized_email, pgUser.UUID, pgUser.Name, pgUser.Role, pgUser.ID.String())

		// 更新令牌
		db.Model(&model.UserTrafficLogsPG{}).
			Where("email_as_id = ?", sanitized_email).
			Updates(map[string]interface{}{
				"token":         token,
				"refresh_token": refreshToken,
				"updated_at":    time.Now(),
			})

		// 返回令牌
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

// EditUserPG 编辑用户 - PostgreSQL版本
func EditUserPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		name := c.Param("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var user UserTrafficLogs
		var pgUser model.UserTrafficLogsPG

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		// 查找用户
		if err := db.Where("email_as_id = ?", helper.SanitizeStr(name)).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("user not found: %s", name)
			return
		}

		updates := map[string]interface{}{"updated_at": time.Now()}
		updateCount := 0

		// 允许编辑 name, role 和 password
		if pgUser.Role != user.Role && user.Role != "" {
			updates["role"] = user.Role
			updateCount++
			log.Printf("Updating role from %s to %s", pgUser.Role, user.Role)
		}

		if pgUser.Name != user.Name && user.Name != "" {
			updates["name"] = user.Name
			updateCount++
			log.Printf("Updating name from %s to %s", pgUser.Name, user.Name)
		}

		// 添加密码更新支持
		if user.Password != "" && len(user.Password) >= 6 {
			hashedPassword := HashPasswordPG(user.Password)
			updates["password"] = hashedPassword
			updateCount++
			log.Printf("Updating password for user %s", pgUser.EmailAsId)
		}

		if updateCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no new value in post data."})
			log.Printf("no new value in post data.")
			return
		}

		// 更新用户
		if err := db.Model(&model.UserTrafficLogsPG{}).Where("email_as_id = ?", helper.SanitizeStr(name)).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error updating user: %v", err)
			return
		}

		// 获取更新后的用户
		db.Where("email_as_id = ?", helper.SanitizeStr(name)).First(&pgUser)

		log.Printf("User %s updated successfully", pgUser.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": pgUser})
	}
}

// DeleteUserByUserNamePG 删除用户 - PostgreSQL版本
func DeleteUserByUserNamePG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		name := c.Param("name")
		log.Printf("Attempting to delete user: %s", name)

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var pgUser model.UserTrafficLogsPG

		// 查找用户
		if err := db.Select("email_as_id, name").Where("email_as_id = ?", name).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("DeleteUserByUserName - user not found: %s, error: %s", name, err.Error())
			return
		}

		// 删除用户
		result := db.Delete(&model.UserTrafficLogsPG{}, "email_as_id = ?", name)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			log.Printf("DeleteUserByUserName - delete failed: %s", result.Error.Error())
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not deleted"})
			log.Printf("DeleteUserByUserName - no rows deleted for user: %s", name)
			return
		}

		log.Printf("Delete user %s successfully!", pgUser.Name)
		c.JSON(http.StatusOK, gin.H{"message": "Delete user " + pgUser.Name + " successfully!"})
	}
}

// GetAllUsersPG 获取所有用户 - PostgreSQL版本
func GetAllUsersPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		var users []model.UserTrafficLogsPG

		// 查询所有用户，只选择需要的字段
		query := `SELECT email_as_id, uuid, name, role, status, used, updated_at, daily_logs, monthly_logs, yearly_logs 
				  FROM user_traffic_logs`

		if err := db.Raw(query).Scan(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetAllUsers: %s", err.Error())
			return
		}

		// 处理结果，限制日志数量
		type UserWithLimitedLogs struct {
			EmailAsId   string         `json:"email_as_id"`
			UUID        string         `json:"uuid"`
			Name        string         `json:"name"`
			Role        string         `json:"role"`
			Status      string         `json:"status"`
			Used        int64          `json:"used"`
			UpdatedAt   time.Time      `json:"updated_at"`
			DailyLogs   datatypes.JSON `json:"daily_logs"`
			MonthlyLogs datatypes.JSON `json:"monthly_logs"`
			YearlyLogs  datatypes.JSON `json:"yearly_logs"`
		}

		var results []UserWithLimitedLogs
		for _, user := range users {
			// 处理日志数据，限制为最近10条
			dailyLogs := limitJsonLogs(user.DailyLogs, 10)
			monthlyLogs := limitJsonLogs(user.MonthlyLogs, 10)
			yearlyLogs := limitJsonLogs(user.YearlyLogs, 10)

			results = append(results, UserWithLimitedLogs{
				EmailAsId:   user.EmailAsId,
				UUID:        user.UUID,
				Name:        user.Name,
				Role:        user.Role,
				Status:      user.Status,
				Used:        user.Used,
				UpdatedAt:   user.UpdatedAt,
				DailyLogs:   dailyLogs,
				MonthlyLogs: monthlyLogs,
				YearlyLogs:  yearlyLogs,
			})
		}

		c.JSON(http.StatusOK, results)
	}
}

// GetUserByNamePG 通过名称获取用户 - PostgreSQL版本
func GetUserByNamePG() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if err := helper.MatchUserTypeAndName(c, name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("GetUserByName: %s", err.Error())
			return
		}

		db := database.GetPostgresDB()
		var user model.UserTrafficLogsPG

		query := `SELECT email_as_id, used, uuid, name, status, role, credit, daily_logs, monthly_logs, yearly_logs, created_at, updated_at
				  FROM user_traffic_logs
				  WHERE email_as_id = ?`

		// 查询用户
		if err := db.Raw(query, name).Scan(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetUserByName: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// 辅助函数：限制JSON日志数量
func limitJsonLogs(jsonData datatypes.JSON, limit int) datatypes.JSON {
	if len(jsonData) == 0 {
		return jsonData
	}

	var data []interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return jsonData
	}

	if len(data) <= limit {
		return jsonData
	}

	// 只保留最近的limit条记录
	limitedData := data[len(data)-limit:]

	result, err := json.Marshal(limitedData)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return jsonData
	}

	return datatypes.JSON(result)
}

// DisableUserPG 禁用用户 - PostgreSQL版本
func DisableUserPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		name := c.Param("name")
		log.Printf("Attempting to disable user: %s", name)

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var pgUser model.UserTrafficLogsPG

		// 查找用户
		if err := db.Where("email_as_id = ?", helper.SanitizeStr(name)).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("DisableUser - user not found: %s, error: %s", name, err.Error())
			return
		}

		// 不允许禁用管理员账户
		if pgUser.Role == "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot disable admin user"})
			log.Printf("attempted to disable admin user: %s", name)
			return
		}

		// 更新用户状态为deleted
		updates := map[string]interface{}{
			"status":     "deleted",
			"updated_at": time.Now(),
		}

		if err := db.Model(&model.UserTrafficLogsPG{}).Where("email_as_id = ?", helper.SanitizeStr(name)).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error disabling user: %v", err)
			return
		}

		// 获取更新后的用户
		db.Where("email_as_id = ?", helper.SanitizeStr(name)).First(&pgUser)

		log.Printf("User %s disabled successfully", pgUser.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User " + pgUser.Name + " disabled successfully"})
	}
}

// EnableUserPG 启用用户 - PostgreSQL版本
func EnableUserPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db := database.GetPostgresDB()
		name := c.Param("name")
		log.Printf("Attempting to enable user: %s", name)

		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
			return
		}

		var pgUser model.UserTrafficLogsPG

		// 查找用户
		if err := db.Where("email_as_id = ?", helper.SanitizeStr(name)).First(&pgUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			log.Printf("EnableUser - user not found: %s, error: %s", name, err.Error())
			return
		}

		// 更新用户状态为plain
		updates := map[string]interface{}{
			"status":     "plain",
			"updated_at": time.Now(),
		}

		if err := db.Model(&model.UserTrafficLogsPG{}).Where("email_as_id = ?", helper.SanitizeStr(name)).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error enabling user: %v", err)
			return
		}

		// 获取更新后的用户
		db.Where("email_as_id = ?", helper.SanitizeStr(name)).First(&pgUser)

		log.Printf("User %s enabled successfully", pgUser.Name)
		c.JSON(http.StatusOK, gin.H{"message": "User " + pgUser.Name + " enabled successfully"})
	}
}

// UpdateAllTokensPG 更新用户令牌 - PostgreSQL版本
func UpdateAllTokensPG(db *gorm.DB, signedToken string, signedRefreshToken string, userId string) {
	updates := map[string]interface{}{
		"token":         signedToken,
		"refresh_token": signedRefreshToken,
		"updated_at":    time.Now(),
	}

	db.Model(&model.UserTrafficLogsPG{}).
		Where("user_id = ?", userId).
		Updates(updates)
}
