package controllers

import (
	"context"
	"fmt"
	"log"
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

// var Domains = map[string]string{"w8": "w8.undervineyard.com", "rm": "rm.undervineyard.com"}
var (
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "USERS")
	validate                         = validator.New()
	V2_API_ADDRESS                   = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT                      = os.Getenv("V2_API_PORT")
	NODE_TYPE                        = os.Getenv("NODE_TYPE")
	CURRENT_DOMAIN                   = os.Getenv("CURRENT_DOMAIN")
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
			userId = "6253a0f5b3829a0c7281aca6"
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

		user.ProduceNodeInUse(adminUser.NodeGlobalList)
		user_role := sanitize.SanitizeStr(user.Role)
		if user_role == "admin" {
			user.NodeGlobalList = adminUser.NodeGlobalList
		}

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

		if NODE_TYPE == "local" {

			cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
			if err != nil {
				log.Printf("v2ray connection failed: %v", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
			err = NHSClient.AddUser(user)
			if err != nil {
				log.Panicf("v2ray add user failed: %v", err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		} else {

			domainsLen := len(user.NodeInUseStatus)
			wg.Add(domainsLen)
			for node, available := range user.NodeInUseStatus {

				if available {
					go func(domain string) {
						defer wg.Done()
						if domain == "sel.undervineyard.com" {
							grpctools.GrpcClientToAddUser(domain, "80", user)
						} else {
							grpctools.GrpcClientToAddUser(domain, "50051", user)
						}
					}(node)
				}

			}
		}

		_, err = userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while inserting user: %v", err)
			return
		}

		err = database.Client.Database("logV2rayTrafficDB").CreateCollection(ctx, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while creating collection for user %s", user_email)
			return
		}

		wg.Wait()

		err = yamlTools.GenerateOne(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("error occured while generating yaml: %v", err)
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

		boundUser_email := sanitize.SanitizeStr(boundUser.Email)
		err := userCollection.FindOne(ctx, bson.M{"email": boundUser_email}).Decode(&foundUser)
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

		token, refreshToken, _ := helper.GenerateAllTokens(boundUser_email, foundUser.UUID, foundUser.Path, foundUser.Role, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		var projections = bson.D{
			{Key: "_id", Value: 0},
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

		var projections = bson.D{}
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
		var domains map[string]string

		if err := c.BindJSON(&domains); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		var projections = bson.D{
			{Key: "used_by_current_year", Value: 0},
			{Key: "used_by_current_month", Value: 0},
			{Key: "used_by_current_day", Value: 0},
			{Key: "traffic_by_year", Value: 0},
			{Key: "traffic_by_month", Value: 0},
			{Key: "traffic_by_day", Value: 0},
			{Key: "password", Value: 0},
			{Key: "refresh_token", Value: 0},
			{Key: "token", Value: 0},
		}
		allUsers, err := database.GetPartialInfosForAllUsers(projections)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("error: %v", err)
			return
		}

		for _, user := range allUsers {
			if user.Role == "admin" {
				user.NodeGlobalList = domains
			}

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

		c.JSON(http.StatusOK, gin.H{"message": "node added successfully"})
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
			{Key: "used_by_current_year", Value: 0},
			{Key: "used_by_current_month", Value: 0},
			{Key: "used_by_current_day", Value: 0},
			{Key: "traffic_by_year", Value: 0},
			{Key: "traffic_by_month", Value: 0},
			{Key: "password", Value: 0},
			{Key: "refresh_token", Value: 0},
			{Key: "token", Value: 0},
		}
		user, err := database.GetUserByName(name, projections)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		var wg sync.WaitGroup

		if NODE_TYPE == "local" {

			cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
			if err != nil {
				msg := "v2ray connection failed."
				log.Panicf("%v", msg)
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}

			NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
			err = NHSClient.DelUser(user.Email)
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

			domainsLen := len(user.NodeInUseStatus)
			wg.Add(domainsLen)
			for node, available := range user.NodeInUseStatus {

				if available {
					go func(domain string) {
						defer wg.Done()
						if domain == "sel.undervineyard.com" {
							grpctools.GrpcClientToDeleteUser(domain, "80", user)
						} else {
							grpctools.GrpcClientToDeleteUser(domain, "50051", user)
						}
					}(node)
					user.NodeInUseStatus[node] = false
				}

			}
		}

		wg.Wait()
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

		err = yamlTools.GenerateOneByQuery(user.Email)
		if err != nil {
			msg := "yaml generate failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

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
			{Key: "used_by_current_year", Value: 0},
			{Key: "used_by_current_month", Value: 0},
			{Key: "used_by_current_day", Value: 0},
			{Key: "traffic_by_year", Value: 0},
			{Key: "traffic_by_month", Value: 0},
			{Key: "password", Value: 0},
			{Key: "refresh_token", Value: 0},
			{Key: "token", Value: 0},
		}

		user, err := database.GetUserByName(name, projections)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		var wg sync.WaitGroup

		if NODE_TYPE == "local" {

			cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
			if err != nil {
				msg := "v2ray connection failed."
				log.Panicf("%v", msg)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
			err = NHSClient.AddUser(user)
			if err != nil {
				msg := "v2ray take user offline failed."
				log.Panicf("%v", msg)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for node, enable := range user.NodeInUseStatus {
				if !enable {
					user.NodeInUseStatus[node] = true
				}
			}

		} else {

			domainsLen := len(user.NodeInUseStatus)
			wg.Add(domainsLen)
			for node, available := range user.NodeInUseStatus {

				if !available {
					go func(domain string) {
						defer wg.Done()
						if domain == "sel.undervineyard.com" {
							grpctools.GrpcClientToAddUser(domain, "80", user)
						} else {
							grpctools.GrpcClientToAddUser(domain, "50051", user)
						}
					}(node)
					user.NodeInUseStatus[node] = true
				}

			}
		}

		wg.Wait()
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

		err = yamlTools.GenerateOneByQuery(user.Email)
		if err != nil {
			msg := "yaml generate failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

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
		var projections = bson.D{}

		user, err := database.GetUserByName(name, projections)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		if user.Status == "plain" {

			var wg sync.WaitGroup

			if NODE_TYPE == "local" {
				cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
				if err != nil {
					msg := "v2ray connection failed."
					log.Panicf("%v", msg)
					c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
					return
				}

				NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
				err = NHSClient.DelUser(user.Email)
				if err != nil {
					msg := "v2ray take user back online failed."
					log.Panicf("%v", msg)
					c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
					return
				}
			} else {

				domainsLen := len(user.NodeInUseStatus)
				wg.Add(domainsLen)
				for node, available := range user.NodeInUseStatus {

					if available {
						go func(domain string) {
							defer wg.Done()
							if domain == "sel.undervineyard.com" {
								grpctools.GrpcClientToDeleteUser(domain, "80", user)
							} else {
								grpctools.GrpcClientToDeleteUser(domain, "50051", user)
							}
						}(node)
					}

				}
			}
			wg.Wait()

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
			{Key: "_id", Value: 0},
			{Key: "token", Value: 0},
			{Key: "password", Value: 0},
			{Key: "refresh_token", Value: 0},
			{Key: "traffic_by_year", Value: 0},
			{Key: "traffic_by_month", Value: 0},
			{Key: "traffic_by_day", Value: 0},
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
			{Key: "_id", Value: 0},
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

		var projections = bson.D{}
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
			{Key: "used_by_current_year", Value: 0},
			{Key: "used_by_current_month", Value: 0},
			{Key: "used_by_current_day", Value: 0},
			{Key: "traffic_by_year", Value: 0},
			{Key: "traffic_by_month", Value: 0},
			{Key: "password", Value: 0},
			{Key: "refresh_token", Value: 0},
			{Key: "token", Value: 0},
		}
		user, err := database.GetUserByName(email, projections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Get user by name failed: %s", err.Error())
			return
		}

		if NODE_TYPE == "local" {
			if CURRENT_DOMAIN == node {
				cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
				if err != nil {
					msg := "v2ray connection failed."
					log.Panicf("%v", msg)
					c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
					return
				}

				NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
				err = NHSClient.DelUser(email)
				if err != nil {
					msg := "v2ray take user back online failed."
					log.Panicf("%v", msg)
					c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
					return
				}
			}
		} else {
			if node == "sel.undervineyard.com" {
				grpctools.GrpcClientToDeleteUser(node, "80", user)
			} else {
				grpctools.GrpcClientToDeleteUser(node, "50051", user)
			}
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

		err = yamlTools.GenerateOneByQuery(user.Email)
		if err != nil {
			msg := "yaml generate failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
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
			{Key: "used_by_current_year", Value: 0},
			{Key: "used_by_current_month", Value: 0},
			{Key: "used_by_current_day", Value: 0},
			{Key: "traffic_by_year", Value: 0},
			{Key: "traffic_by_month", Value: 0},
			{Key: "password", Value: 0},
			{Key: "refresh_token", Value: 0},
			{Key: "token", Value: 0},
		}
		user, err := database.GetUserByName(email, projections)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Get user by name failed: %s", err.Error())
			return
		}

		if NODE_TYPE == "local" {
			if CURRENT_DOMAIN == node {
				cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
				if err != nil {
					msg := "v2ray connection failed."
					log.Printf("%v", msg)
					c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
					return
				}

				NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
				err = NHSClient.AddUser(user)
				if err != nil {
					msg := "v2ray add user failed."
					log.Printf("%v", msg)
					c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
					return
				}
			}
		} else {
			if node == "sel.undervineyard.com" {
				grpctools.GrpcClientToAddUser(node, "80", user)
			} else {
				grpctools.GrpcClientToAddUser(node, "50051", user)
			}
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

		err = yamlTools.GenerateOneByQuery(user.Email)
		if err != nil {
			msg := "yaml generate failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			log.Printf("%s", msg)
			return
		}

		log.Printf("Enable user: %v, node: %v by hand!", sanitize.SanitizeStr(email), sanitize.SanitizeStr(node))
		c.JSON(http.StatusOK, gin.H{"message": "Enable user: " + email + " at node: " + node + " successfully!"})
	}
}
