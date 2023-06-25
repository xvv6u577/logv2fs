package controllers

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/grpc"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/routine"
	"github.com/caster8013/logv2rayfullstack/v2ray"

	helper "github.com/caster8013/logv2rayfullstack/helpers"
	sanitize "github.com/caster8013/logv2rayfullstack/sanitize"

	"github.com/caster8013/logv2rayfullstack/grpctools"
	"github.com/caster8013/logv2rayfullstack/model"
	yamlTools "github.com/caster8013/logv2rayfullstack/yaml"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var (
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "USERS")
	// nodeCollection *mongo.Collection = database.OpenCollection(database.Client, "NODES")
	globalCollection *mongo.Collection = database.OpenCollection(database.Client, "GLOBAL")
	validate                           = validator.New()
	V2_API_ADDRESS                     = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT                        = os.Getenv("V2_API_PORT")
	NODE_TYPE                          = os.Getenv("NODE_TYPE")
	CURRENT_DOMAIN                     = os.Getenv("CURRENT_DOMAIN")
	MIXED_PORT                         = os.Getenv("MIXED_PORT")
	ADMINUSERID                        = os.Getenv("ADMINUSERID")
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
)

type DomainInfo struct {
	Domain       string `json:"domain"`
	ExpiredDate  string `json:"expired_date"`
	DaysToExpire int    `json:"days_to_expire"`
	IsInUVP      bool   `json:"is_in_uvp"`
}

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

//CreateUser is the api used to tget a single user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"sign up error": err.Error()})
			log.Printf("%s", err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var user model.User
		var current = time.Now()

		CREDIT := os.Getenv("CREDIT")

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

		user_email := sanitize.SanitizeStr(user.Email)
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user_email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			log.Printf("error occured while checking for the email: %s", err.Error())
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
			log.Printf("this email already exists")
			return
		}

		var adminUser model.User
		var userId = c.GetString("uid")
		if userId == "" {
			userId = ADMINUSERID
		}

		err = userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&adminUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOne error: %v", err)
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

		user.NodeGlobalList = adminUser.NodeGlobalList
		user.ProduceNodeInUse(adminUser.NodeGlobalList)
		user_role := sanitize.SanitizeStr(user.Role)
		// if user_role == "admin" {
		// }

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
			Period: current_day,
			Amount: 0,
		}
		user.UsedByCurrentMonth = TrafficAtPeriod{
			Period: current_month,
			Amount: 0,
		}
		user.UsedByCurrentYear = TrafficAtPeriod{
			Period: current_year,
			Amount: 0,
		}

		user.TrafficByDay = []TrafficAtPeriod{}
		user.TrafficByMonth = []TrafficAtPeriod{}
		user.TrafficByYear = []TrafficAtPeriod{}

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(user_email, user.UUID, user.Path, user_role, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		var wg sync.WaitGroup
		var waitQueueLength = 3

		if NODE_TYPE == "local" {

			wg.Add(waitQueueLength + 1)
			go func() {
				defer wg.Done()
				err = grpctools.GrpcClientToAddUser("0.0.0.0", MIXED_PORT, user, true)
				if err != nil {
					log.Printf("v2ray add user failed: \n%v", err.Error())
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

		go func() {
			defer wg.Done()
			_, err = userCollection.InsertOne(ctx, user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error occured while inserting user: %v", err)
				return
			}
		}()

		go func() {
			defer wg.Done()
			err = database.Client.Database("logV2rayTrafficDB").CreateCollection(ctx, user.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error occured while creating collection for user %s", user_email)
				return
			}
		}()

		go func() {
			defer wg.Done()
			err = yamlTools.GenerateOneYAML(user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error occured while generating yaml: %v", err)
				return
			}
		}()

		wg.Wait()
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

		sanitized_email := sanitize.SanitizeStr(boundUser.Email)
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

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
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

func AddNode() gin.HandlerFunc {
	return func(c *gin.Context) {

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"sign up error": err.Error()})
			log.Printf("%s", err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		// var current = time.Now()
		var domains map[string]string

		if err := c.BindJSON(&domains); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		// search domain in NODES collection, if not exist, add it.
		// for domain := range domains {
		// 	var foundNode model.Node
		// 	err := nodeCollection.FindOne(ctx, bson.M{"domain": domain}).Decode(&foundNode)
		// 	if err != nil {
		// 		if err.Error() == "mongo: no documents in result" {
		// 			var node CurrentNode
		// 			node.Domain = domain
		// 			node.Status = "active"
		// 			node.NodeAtCurrentYear = NodeAtPeriod{
		// 				Period:              current.Local().Format("2006"),
		// 				Amount:              0,
		// 				UserTrafficAtPeriod: map[string]int64{},
		// 			}
		// 			node.NodeAtCurrentMonth = NodeAtPeriod{
		// 				Period:              current.Local().Format("200601"),
		// 				Amount:              0,
		// 				UserTrafficAtPeriod: map[string]int64{},
		// 			}
		// 			node.NodeAtCurrentDay = NodeAtPeriod{
		// 				Period:              current.Local().Format("20060102"),
		// 				Amount:              0,
		// 				UserTrafficAtPeriod: map[string]int64{},
		// 			}
		// 			node.NodeByYear = []NodeAtPeriod{}
		// 			node.NodeByMonth = []NodeAtPeriod{}
		// 			node.NodeByDay = []NodeAtPeriod{}
		// 			node.CreatedAt = current
		// 			node.UpdatedAt = current

		// 			_, err = nodeCollection.InsertOne(ctx, node)
		// 			if err != nil {
		// 				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 				log.Printf("error occured while inserting node: %v", err)
		// 				return
		// 			}

		// 		}
		// 	}
		// }

		var projections = bson.D{
			{Key: "node_global_list", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "role", Value: 1},
			{Key: "email", Value: 1},
			{Key: "user_id", Value: 1},
		}
		allUsers, err := database.GetPartialInfosForAllUsers(projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		for _, user := range allUsers {
			// if user.Role == "admin" {
			// }
			user.NodeGlobalList = domains

			user.ProduceNodeInUse(domains)
			filter := bson.D{primitive.E{Key: "user_id", Value: user.User_id}}
			update := bson.M{"$set": bson.M{"updated_at": time.Now(), "node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl, "node_global_list": user.NodeGlobalList}}
			_, err = userCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("%v", err.Error())
				return
			}

			err = yamlTools.GenerateOneByQuery(user.Email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				log.Printf("error occured while generating yaml: %v", err)
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Congrats! Nodes updated in success!"})
	}
}

func EditUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("%s", err.Error())
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

		err := userCollection.FindOne(ctx, bson.M{"email": sanitize.SanitizeStr(user.Email)}).Decode(&foundUser)
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
			bson.M{"email": sanitize.SanitizeStr(user.Email)},
			bson.M{"$set": newFoundUser},
			options.FindOneAndUpdate().SetUpsert(true),
		).Decode(&replacedDocument)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		err = userCollection.FindOne(ctx, bson.M{"email": sanitize.SanitizeStr(user.Email)}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

//GetUser is the api used to get a single user
func GetUserByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var user model.User
		userId := sanitize.SanitizeStr(c.Param("user_id"))

		if err := helper.MatchUserTypeAndUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("%s", err.Error())
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while listing user items")
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func TakeItOfflineByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("%s", err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		name := sanitize.SanitizeStr(c.Param("name"))
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
			{Key: "node_global_list", Value: 1},
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
			err = grpctools.GrpcClientToDeleteUser("0.0.0.0", MIXED_PORT, user, true)
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

		user.ProduceSuburl()
		filter := bson.D{primitive.E{Key: "email", Value: name}}
		update := bson.M{"$set": bson.M{"status": v2ray.DELETE, "node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}
		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		err = yamlTools.GenerateOneYAML(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while generating yaml: %v", err)
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
			log.Printf("%s", err.Error())
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		name := sanitize.SanitizeStr(c.Param("name"))
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
			{Key: "node_global_list", Value: 1},
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
			err = grpctools.GrpcClientToAddUser("0.0.0.0", MIXED_PORT, user, true)
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

		user.ProduceSuburl()
		filter := bson.D{primitive.E{Key: "email", Value: name}}
		update := bson.M{"$set": bson.M{"status": v2ray.PLAIN, "node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}

		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		err = yamlTools.GenerateOneYAML(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while generating yaml: %v", err)
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
			log.Printf("%s", err.Error())
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
			{Key: "node_global_list", Value: 1},
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
				err = grpctools.GrpcClientToDeleteUser("0.0.0.0", MIXED_PORT, user, true)
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

func GetTrafficByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if err := helper.MatchUserTypeAndName(c, name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("%s", err.Error())
			return
		}

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("grpc dial error: %v", err)
			return
		}

		NSSClient := v2ray.NewStatsServiceClient(cmdConn)
		uplink, err := NSSClient.GetUserUplink(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetUserUplink failed: %v", err)
			return
		}

		downlink, err := NSSClient.GetUserDownlink(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Get user %s downlink failed.", sanitize.SanitizeStr(name))
			return
		}

		c.JSON(http.StatusOK, gin.H{"uplink": uplink, "downlink": downlink})
	}
}

func GetAllUserTraffic() gin.HandlerFunc {
	return func(c *gin.Context) {

		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("%s", err.Error())
			return
		}

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("%s", err.Error())
			return
		}

		NSSClient := v2ray.NewStatsServiceClient(cmdConn)

		allTraffic, err := NSSClient.GetAllUserTraffic(false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("GetAllUserTraffic failed: %s", err.Error())
			return
		}

		c.JSON(http.StatusOK, allTraffic)
	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("%s", err.Error())
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
			{Key: "node_global_list", Value: 1},
			{Key: "used", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
			{Key: "used_by_current_day", Value: 0},
			{Key: "used_by_current_month", Value: 0},
			{Key: "used_by_current_year", Value: 0},
		}

		allUsers, err := database.GetPartialInfosForAllUsers(projections)
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
			{Key: "node_global_list", Value: 1},
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

func GetSubscripionURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		var projections = bson.D{
			{Key: "suburl", Value: 1},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("GetUserByName failed: %s", err.Error())
			return
		}

		c.String(http.StatusOK, user.Suburl)
	}
}

func WriteToDB() gin.HandlerFunc {
	return func(c *gin.Context) {

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = routine.Log_basicAction()
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

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		email := sanitize.SanitizeStr(c.Request.URL.Query().Get("email"))
		node := sanitize.SanitizeStr(c.Request.URL.Query().Get("node"))

		var projections = bson.D{
			{Key: "token", Value: 0},
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "node_global_list", Value: 1},
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

				err = grpctools.GrpcClientToDeleteUser("0.0.0.0", MIXED_PORT, user, true)
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
		filter := bson.D{primitive.E{Key: "email", Value: email}}
		update := bson.M{"$set": bson.M{"node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}

		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		err = yamlTools.GenerateOneYAML(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while generating yaml: %v", err)
			return
		}

		log.Printf("Disable user: %v, node: %v by hand!", sanitize.SanitizeStr(email), sanitize.SanitizeStr(node))
		c.JSON(http.StatusOK, gin.H{"message": "Disable user: " + email + " at node: " + node + " successfully!"})
	}
}

func EnableNodePerUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		email := sanitize.SanitizeStr(c.Request.URL.Query().Get("email"))
		node := sanitize.SanitizeStr(c.Request.URL.Query().Get("node"))

		var projections = bson.D{
			{Key: "token", Value: 0},
			{Key: "email", Value: 1},
			{Key: "path", Value: 1},
			{Key: "name", Value: 1},
			{Key: "node_in_use_status", Value: 1},
			{Key: "uuid", Value: 1},
			{Key: "role", Value: 1},
			{Key: "status", Value: 1},
			{Key: "suburl", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "node_global_list", Value: 1},
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
				err = grpctools.GrpcClientToAddUser("0.0.0.0", MIXED_PORT, user, true)
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
		filter := bson.D{primitive.E{Key: "email", Value: email}}
		update := bson.M{"$set": bson.M{"node_in_use_status": user.NodeInUseStatus, "suburl": user.Suburl}}

		_, err = userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		err = yamlTools.GenerateOneYAML(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while generating yaml: %v", err)
			return
		}

		log.Printf("Enable user: %v, node: %v by hand!", sanitize.SanitizeStr(email), sanitize.SanitizeStr(node))
		c.JSON(http.StatusOK, gin.H{"message": "Enable user: " + email + " at node: " + node + " successfully!"})
	}
}

func GetDomainInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// query GlobalVariable
		var domainInfos = []DomainInfo{}
		var foundGlobal GlobalVariable
		var filter = bson.D{primitive.E{Key: "name", Value: "GLOBAL"}}
		err = globalCollection.FindOne(ctx, filter).Decode(&foundGlobal)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOne error: %v", err)
			return
		}
		tempArray, err := buildDomainInfo(foundGlobal.DomainList, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while parsing domain info: %v", err)
			return
		}
		domainInfos = append(domainInfos, tempArray...)

		// query user
		var adminUser model.User
		var userId = ADMINUSERID
		err = userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&adminUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOne error: %v", err)
			return
		}
		tempArray, err = buildDomainInfo(adminUser.NodeGlobalList, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while parsing domain info: %v", err)
			return
		}
		domainInfos = append(domainInfos, tempArray...)

		c.JSON(http.StatusOK, domainInfos)
	}
}

func buildDomainInfo(domains map[string]string, isInUvp bool) ([]DomainInfo, error) {

	var minorDomainInfos []DomainInfo
	var port = "443"
	var conf = &tls.Config{
		// InsecureSkipVerify: true,
	}

	for _, domain := range domains {
		if domain == "localhost" {
			continue
		}

		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", domain+":"+port, conf)
		if err != nil {
			log.Printf("error occured while parsing domain info: %v", err)
			return nil, err
		}
		err = conn.VerifyHostname(domain)
		if err != nil {
			log.Printf("error occured while parsing domain info: %v", err)
			return nil, err
		}
		expiry := conn.ConnectionState().PeerCertificates[0].NotAfter

		minorDomainInfos = append(minorDomainInfos, DomainInfo{
			IsInUVP:      isInUvp,
			Domain:       domain,
			ExpiredDate:  expiry.Local().Format("2006-01-02 15:04:05"),
			DaysToExpire: int(time.Until(expiry).Hours() / 24),
		})
		defer conn.Close()
	}

	return minorDomainInfos, nil
}

func UpdateDomainInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

		err := helper.CheckUserType(c, "admin")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var tempDomainList map[string]string
		err = c.BindJSON(&tempDomainList)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// replace domain_list in GlobalVariable in globalCollection with tempDomainList
		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var replacedDocument GlobalVariable
		err = globalCollection.FindOneAndUpdate(ctx,
			bson.M{"name": "GLOBAL"},
			bson.M{"$set": bson.M{"domain_list": tempDomainList}},
			options.FindOneAndUpdate().SetReturnDocument(1),
		).Decode(&replacedDocument)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("FindOneAndUpdate error: %v", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Update GLOBAL domain list successfully!"})
	}
}
