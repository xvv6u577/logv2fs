package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/shomali11/parallelizer"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

func main() {

	app := &cli.App{
		Name:  "logv2rayfullstack",
		Usage: "A simple CLI program to manage logv2ray backend",
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "run backend server",
				Action: func(c *cli.Context) error {

					group := parallelizer.NewGroup()
					defer group.Close()

					group.Add(V2rayProcess)
					group.Add(runServer)

					err := group.Wait()

					return err
				},
			},
			{
				Name:    "cron",
				Aliases: []string{"c"},
				Usage:   "run cron job",
				Action: func(c *cli.Context) error {
					// CronTest()
					return nil
				},
			},
			{
				Name:    "mongo",
				Aliases: []string{"db"},
				Usage:   "manage mongoDB",
				Action: func(c *cli.Context) error {

					tag := c.Args().First()

					switch tag {

					case "tweet":
						err := AddDBUserProperty()
						return err

					case "emptyuserall":
						err := EmptyUsersInfoInDB()
						return err

					case "deldbs":
						err := DeleteUsersDBs()
						return err

					default:
						fmt.Println(tag)
					}

					return nil
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "command test",
				Action: func(c *cli.Context) error {
					fmt.Println("added task: ", c.Args().First(), c.Args().Get(2))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func V2rayProcess() {
	cmd := exec.Command("/usr/local/bin/v2ray", "-config", "/Users/caster/Desktop/v2ray-2-instances/transit-server/config.json")
	// cmd := exec.Command("/usr/local/bin/v2ray", "-config", "/Users/caster/Desktop/config.json")

	if err := cmd.Run(); err != nil {
		log.Panic("Panic: ", err)
	}
}

func runServer() {
	// wait v2ray process to be ready.
	time.Sleep(time.Second)

	// add users in databases to v2ray service
	allUsersInDB, _ := GetAllUsersInfo()
	if len(allUsersInDB) != 0 {

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic(err)
		}

		for _, user := range allUsersInDB {
			if user.Status == "plain" {
				NHSClient := NewHandlerServiceClient(cmdConn, user.Path)
				NHSClient.AddUser(*user)
			}
		}
	}

	// add cron
	Cron_loggingV2TrafficAll_everyHour()

	// gin.SetMode(gin.ReleaseMode)

	// default:
	// router := gin.Default()

	// customized:
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(recoverFromError)

	router.GET("/v1/user/all", getAllUsers)           // mongo
	router.GET("/v1/user/:name", getUserByName)       // mongo
	router.GET("/v1/adduser/:name", addUserByName)    // v2api
	router.GET("/v1/deluser/:name", deleteUserByName) // v2api

	// curl http://localhost:8079/v1/postuser \
	// --include \
	// --header "Content-Type: application/json" \
	// --request "POST" \
	// --data '{"email": "email","password": "email","uuid": "98a131b0-69a5-41ef-9339-d6dbcabaa773", "path": "ray"}'
	router.POST("/v1/postuser", postUserByJSON) // v2api

	// ———————————————————————————————————————————————————
	router.GET("/v1/traffic/all", getAllUserTraffic)  // v2api
	router.GET("/v1/traffic/:name", getTrafficByUser) // v2api

	router.Use(static.Serve("/", static.LocalFile("./frontend/build/", false)))

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found."})
	})

	router.Run(fmt.Sprintf("%s:%d", SERVER_ADDRESS, SERVER_PORT))

}

func getTrafficByUser(c *gin.Context) {
	name := c.Param("name")

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
	if err != nil {
		log.Panic(err)
	}

	NSSClient := NewStatsServiceClient(cmdConn)
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

func getAllUserTraffic(c *gin.Context) {

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
	if err != nil {
		log.Panic("Panic: ", err)
	}

	NSSClient := NewStatsServiceClient(cmdConn)

	allTraffic, err := NSSClient.GetAllUserTraffic(false)
	if err != nil {
		log.Panic(err)
	}

	c.JSON(http.StatusOK, allTraffic)
}

func getAllUsers(c *gin.Context) {

	allUsers, _ := GetAllUsersInfo()
	if len(allUsers) == 0 {
		c.JSON(http.StatusOK, []User{})
		return
	}
	c.JSON(http.StatusOK, allUsers)
}

func addUserByName(c *gin.Context) {

	var errors error
	name := c.Param("name")

	uuidV4, err := uuid.NewV4()
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(name), 8)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user := User{
		Path:          "ray",
		Email:         name,
		ID:            uuidV4.String(),
		CreatedAt:     now,
		UpdatedAt:     now,
		Credittraffic: 1073741824,
		Password:      string(bytes),
		Status:        "plain",
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	NHSClient := NewHandlerServiceClient(cmdConn, user.Path)
	err = NHSClient.AddUser(user)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	err = CreateUserByName(&user)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	if errors != nil {
		fmt.Println("Error: ", errors.Error())

		c.JSON(http.StatusInternalServerError, gin.H{"message": errors.Error()})

		return
	}

	fmt.Println(name, "created at v2ray and database.")

	c.JSON(http.StatusCreated, gin.H{"message": "user " + name + " created at v2ray and database."})
}

func postUserByJSON(c *gin.Context) {

	var errors error

	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user := User{
		CreatedAt:     now,
		UpdatedAt:     now,
		Credittraffic: 1073741824,
		Status:        "plain",
	}

	if err := c.BindJSON(&user); err != nil {
		errors = multierror.Append(errors, err)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		errors = multierror.Append(errors, err)
	}
	user.Password = string(bytes)

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	NHSClient := NewHandlerServiceClient(cmdConn, user.Path)
	err = NHSClient.AddUser(user)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	err = CreateUserByName(&user)
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	if errors != nil {
		fmt.Println("Error: ", errors.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": errors.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created at v2ray and database."})
}

func deleteUserByName(c *gin.Context) {

	var errors error
	name := c.Param("name")

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
	if err != nil {
		errors = multierror.Append(errors, err)
	}

	NHSClient := NewHandlerServiceClient(cmdConn, INBOUND_TAG)
	err_DelUser := NHSClient.DelUser(name)
	if err_DelUser != nil {
		errors = multierror.Append(errors, err)
	}

	err_DeleteUserByName := DeleteUserByName(name)
	if err_DeleteUserByName != nil {
		errors = multierror.Append(errors, err)
	}

	if errors != nil {
		fmt.Println("Error: ", errors.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": errors.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted from v2ray, info updated in database."})
}

func getUserByName(c *gin.Context) {
	name := c.Param("name")

	user, err := GetUserByName(name)
	if err != nil {
		log.Panic("Panic: ", err)
	}

	c.JSON(http.StatusOK, user)
}

func recoverFromError(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {

			// 打印错误堆栈信息
			log.Panicf("Panic: %v\n", r)
			debug.PrintStack()

			// 用json封装信息返回
			c.JSON(200, gin.H{"code": 4444, "message": "Server internal error!"})
		}
	}()

	// 加载完defer recover, 继续后续接口调用
	c.Next()
}
