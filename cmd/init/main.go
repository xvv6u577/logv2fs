package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/caster8013/logv2rayfullstack/database"
	helper "github.com/caster8013/logv2rayfullstack/helpers"
	"github.com/caster8013/logv2rayfullstack/model"
	"github.com/caster8013/logv2rayfullstack/sanitize"
	yamlTools "github.com/caster8013/logv2rayfullstack/yaml"
	"github.com/go-playground/validator/v10"
	uuid "github.com/nu7hatch/gouuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "USERS")
	validate                         = validator.New()
	V2_API_ADDRESS                   = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT                      = os.Getenv("V2_API_PORT")
	NODE_TYPE                        = os.Getenv("NODE_TYPE")
	CURRENT_DOMAIN                   = os.Getenv("CURRENT_DOMAIN")
	CREDIT                           = os.Getenv("CREDIT")
	INITUSER                         = os.Getenv("INITUSER")
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

func main() {

	var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var user = model.User{
		Email:    INITUSER,
		Password: INITUSER,
		Role:     "admin",
		Path:     "ray",
		Status:   "plain",
		NodeGlobalList: map[string]string{
			"init": CURRENT_DOMAIN,
		},
	}
	var current = time.Now()

	validationErr := validate.Struct(user)
	if validationErr != nil {
		fmt.Printf("validate error: %v", validationErr)
		return
	}

	user_email := sanitize.SanitizeStr(user.Email)
	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user_email})
	if err != nil {
		fmt.Printf("error occured while checking for the email: %s", err.Error())
		return
	}

	if count > 0 {
		fmt.Printf("this email already exists")
		return
	}

	if user.Name == "" {
		user.Name = user_email
	}

	if user.Path == "" {
		user.Path = "ray"
	}

	password := controllers.HashPassword(user.Password)
	user.Password = password

	user.CreatedAt = current
	user.UpdatedAt = current

	if user.UUID == "" {
		uuidV4, _ := uuid.NewV4()
		user.UUID = uuidV4.String()
	}

	user.ProduceNodeInUse(user.NodeGlobalList)
	user_role := sanitize.SanitizeStr(user.Role)

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

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		fmt.Printf("error occured while inserting user: %v", err)
		return
	}

	err = database.Client.Database("logV2rayTrafficDB").CreateCollection(ctx, user.Email)
	if err != nil {
		fmt.Printf("error occured while creating collection for user %s", user_email)
		return
	}

	err = yamlTools.GenerateOneYAML(user)
	if err != nil {
		fmt.Printf("error occured while generating yaml: %v", err)
		return
	}

	println("user created successfully")
}
