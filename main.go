package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"time"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	routers "github.com/caster8013/logv2rayfullstack/routers"
	"github.com/caster8013/logv2rayfullstack/routine"
	"github.com/caster8013/logv2rayfullstack/v2ray"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/shomali11/parallelizer"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

const (
	SERVER_ADDRESS = "127.0.0.1"
	SERVER_PORT    = 8079
)

type User = model.User

var cronInstance *cron.Cron

// var collection *mongo.Collection

func init() {

	cronInstance = cron.New()
	cronInstance.Start()
	// collection = database.OpenCollection(database.Client, "USERS")
}

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
						err := database.AddDBUserProperty()
						return err

					case "emptyuserall":
						err := database.EmptyUsersInfoInDB()
						return err

					case "deldbs":
						err := database.DeleteUsersDBs()
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
	cmd := exec.Command("/opt/homebrew/bin/v2ray", "-config", "/Users/guestuser2/Desktop/v2ray-2-instances/transit-server/config.json")
	// cmd := exec.Command("/usr/local/bin/v2ray", "-config", "/Users/caster/Desktop/config.json")

	if err := cmd.Run(); err != nil {
		log.Panic("Panic: ", err)
	}
}

func runServer() {
	// wait v2ray process to be ready.
	time.Sleep(time.Second)

	// add users in databases to v2ray service
	allUsersInDB, _ := database.GetAllUsersInfo()
	if len(allUsersInDB) != 0 {

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic(err)
		}

		for _, user := range allUsersInDB {
			if user.Status == "plain" {
				NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
				NHSClient.AddUser(*user)
			}
		}
	}

	// add cron
	routine.Cron_loggingV2TrafficAll_everyHour(cronInstance)

	// gin.SetMode(gin.ReleaseMode)

	// default:
	// router := gin.Default()

	// customized:
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(recoverFromError)

	// user auth tutorial:
	// https://dev.to/joojodontoh/build-user-authentication-in-golang-with-jwt-and-mongodb-2igd
	routers.AuthRoutes(router)
	routers.UserRoutes(router)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found."})
	})

	router.Run(fmt.Sprintf("%s:%d", SERVER_ADDRESS, SERVER_PORT))

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
