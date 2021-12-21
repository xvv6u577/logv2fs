package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/grpc"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/v2ray"

	helper "github.com/caster8013/logv2rayfullstack/helpers"

	"github.com/caster8013/logv2rayfullstack/model"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "USERS")
var validate = validator.New()

type User = model.User

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
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var user model.User

		CREDIT := os.Getenv("CREDIT")

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email already exists"})
			return
		}

		if user.Name == "" {
			user.Name = user.Email
		}

		password := HashPassword(user.Password)
		user.Password = password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		if user.UUID == "" {
			uuidV4, _ := uuid.NewV4()
			user.UUID = uuidV4.String()
		}

		if user.Credittraffic == 0 {
			credit, _ := strconv.ParseInt(CREDIT, 10, 64)
			user.Credittraffic = credit
		}
		if user.Usedtraffic == 0 {
			user.Usedtraffic = 0
		}

		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(user.Email, user.UUID, user.Path, user.Role, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		_, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		_ = database.Client.Database("logV2rayTrafficDB").CreateCollection(ctx, user.Email)

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			msg := "v2ray connection failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
		err = NHSClient.AddUser(user)
		if err != nil {
			msg := "v2ray add user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		fmt.Println(user.Email, "created at v2ray and database.")

		c.JSON(http.StatusOK, gin.H{"message": "user " + user.Email + " created at v2ray and database."})

	}
}

//Login is the api used to tget a single user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user model.User
		var foundUser model.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or passowrd is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(user.Password, foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}
		token, refreshToken, _ := helper.GenerateAllTokens(foundUser.Email, foundUser.UUID, foundUser.Path, foundUser.Role, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// recordPerPage := 10
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		// startIndex := (page - 1) * recordPerPage
		startIndex, _ := strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}
		var allusers []bson.M
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allusers[0])

	}
}

//GetUser is the api used to get a single user
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		if err := helper.MatchUserTypeAndUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user model.User

		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)

	}
}

func TakeItOfflineByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := c.Param("name")

		user, err := database.GetUserByName(name)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			msg := "v2ray connection failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
		err = NHSClient.DelUser(name)
		if err != nil {
			msg := "v2ray delete user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		err = database.UpdateUserStatusByName(name, v2ray.DELETE)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user deleted from v2ray, info updated in database."})
	}
}

func TakeItOnlineByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := c.Param("name")

		user, err := database.GetUserByName(name)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			msg := "v2ray connection failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
		err = NHSClient.AddUser(user)
		if err != nil {
			msg := "v2ray take user back online failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		err = database.UpdateUserStatusByName(name, v2ray.PLAIN)
		if err != nil {
			msg := "database user info update failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user added from v2ray, info updated in database."})
	}
}

func DeleteUserByUserName() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		name := c.Param("name")

		user, err := database.GetUserByName(name)
		if err != nil {
			msg := "database get user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			msg := "v2ray connection failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
		err = NHSClient.DelUser(name)
		if err != nil {
			msg := "v2ray delete user failed."
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		err = database.DeleteUserByName(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user deleted from v2ray, info deleted in database."})
	}
}

func GetTrafficByUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if err := helper.MatchUserTypeAndName(c, name); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic(err)
		}

		NSSClient := v2ray.NewStatsServiceClient(cmdConn)
		uplink, err := NSSClient.GetUserUplink(name)
		if err != nil {
			log.Panic(err)
		}

		downlink, err := NSSClient.GetUserDownlink(name)
		if err != nil {
			log.Panic(err)
		}

		c.JSON(http.StatusOK, gin.H{"uplink": uplink, "downlink": downlink})
	}
}

func GetAllUserTraffic() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic("Panic: ", err)
		}

		NSSClient := v2ray.NewStatsServiceClient(cmdConn)

		allTraffic, err := NSSClient.GetAllUserTraffic(false)
		if err != nil {
			log.Panic(err)
		}

		c.JSON(http.StatusOK, allTraffic)
	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		allUsers, _ := database.GetAllUsersInfo()
		if len(allUsers) == 0 {
			c.JSON(http.StatusOK, []User{})
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
			return
		}

		user, err := database.GetUserByName(name)
		if err != nil {
			log.Panic("Panic: ", err)
		}

		c.JSON(http.StatusOK, user)
	}
}

// func addUserByName(c *gin.Context) {

// 	var errors error
// 	name := c.Param("name")

// 	uuidV4, err := uuid.NewV4()
// 	if err != nil {
// 		errors = multierror.Append(errors, err)
// 	}

// 	bytes, err := bcrypt.GenerateFromPassword([]byte(name), 8)
// 	if err != nil {
// 		errors = multierror.Append(errors, err)
// 	}

// 	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
// 	user := User{
// 		Path:          "ray",
// 		Email:         name,
// 		UUID:          uuidV4.String(),
// 		CreatedAt:     now,
// 		UpdatedAt:     now,
// 		Credittraffic: 1073741824,
// 		Password:      string(bytes),
// 		Status:        "plain",
// 	}

// 	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
// 	if err != nil {
// 		errors = multierror.Append(errors, err)
// 	}

// 	NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
// 	err = NHSClient.AddUser(user)
// 	if err != nil {
// 		errors = multierror.Append(errors, err)
// 	}

// 	err = database.CreateUserByName(&user)
// 	if err != nil {
// 		errors = multierror.Append(errors, err)
// 	}

// 	if errors != nil {
// 		fmt.Println("Error: ", errors.Error())

// 		c.JSON(http.StatusInternalServerError, gin.H{"message": errors.Error()})

// 		return
// 	}

// 	fmt.Println(name, "created at v2ray and database.")

// 	c.JSON(http.StatusCreated, gin.H{"message": "user " + name + " created at v2ray and database."})
// }
