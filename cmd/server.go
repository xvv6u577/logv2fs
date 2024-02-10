/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/shomali11/parallelizer"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/grpctools"
	"github.com/xvv6u577/logv2fs/middleware"
	"github.com/xvv6u577/logv2fs/model"
	routers "github.com/xvv6u577/logv2fs/routers"
	"github.com/xvv6u577/logv2fs/v2ray"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc"
)

type (
	Traffic         = model.Traffic
	TrafficAtPeriod = model.TrafficAtPeriod
	YamlTemplate    = model.YamlTemplate
	Proxies         = model.Proxies
	Headers         = model.Headers
	WsOpts          = model.WsOpts
	ProxyGroups     = model.ProxyGroups
)

var (
	NODE_TYPE      = os.Getenv("NODE_TYPE")
	CURRENT_DOMAIN = os.Getenv("CURRENT_DOMAIN")
	SERVER_ADDRESS = os.Getenv("SERVER_ADDRESS")
	SERVER_PORT    = os.Getenv("SERVER_PORT")
	V2_API_ADDRESS = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT    = os.Getenv("V2_API_PORT")
	V2RAY          = os.Getenv("V2RAY")
	V2RAY_CONFIG   = os.Getenv("V2RAY_CONFIG")
	GRPC_PORT      = os.Getenv("GRPC_PORT")
	GIN_MODE       = os.Getenv("GIN_MODE")
	cronInstance   *cron.Cron
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run server",
	Long:  `run server, which is the entry of this program.`,
	Run: func(cmd *cobra.Command, args []string) {

		logFile, err := os.OpenFile("./logs/log_file.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			log.Fatalln(err)
		}
		log.SetOutput(logFile)

		group := parallelizer.NewGroup()
		defer group.Close()

		group.Add(V2rayProcess)

		group.Add(func() {
			log.Printf("Server runs at %s:%s", SERVER_ADDRESS, SERVER_PORT)
			err := RunServer().Run(fmt.Sprintf("%s:%s", SERVER_ADDRESS, SERVER_PORT))
			if err != nil {
				log.Panic("Panic: ", err)
			}
		})

		group.Add(RunGRPCServer)

		group.Wait()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	cronInstance = cron.New()
	cronInstance.Start()
}

func RunGRPCServer() {
	if err := grpctools.GrpcServer(fmt.Sprintf("0.0.0.0:%s", GRPC_PORT), false); err != nil {
		log.Panic("GrpcServer panic: ", err)
	}
}

func V2rayProcess() {
	log.Printf("V2ray process runs at 8070, 10000, 10001, 10002")
	var cmd = exec.Command(V2RAY, "-config", V2RAY_CONFIG)
	if err := cmd.Run(); err != nil {
		log.Panic("Panic: ", err)
	}
}

func RunServer() *gin.Engine {

	time.Sleep(time.Second)

	var projections = bson.D{
		{Key: "email", Value: 1},
		{Key: "name", Value: 1},
		{Key: "path", Value: 1},
		{Key: "status", Value: 1},
		{Key: "uuid", Value: 1},
		{Key: "node_in_use_status", Value: 1},
		{Key: "role", Value: 1},
		{Key: "used", Value: 1},
		{Key: "credit", Value: 1},
	}

	allUsersInDB, _ := database.GetAllUsersPortionInfo(projections)
	if len(allUsersInDB) != 0 {

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic(err)
		}
		defer cmdConn.Close()

		var wg sync.WaitGroup
		for _, user := range allUsersInDB {
			if user.Status == "plain" && user.NodeInUseStatus[CURRENT_DOMAIN] {
				wg.Add(1)
				go func(user User) {
					defer wg.Done()
					NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
					NHSClient.AddUser(user)
				}(*user)
			}
		}
		wg.Wait()
	}
	// add cron
	// localCron.Cron_loggingJobs(cronInstance)

	if GIN_MODE == "release" {
		gin.SetMode(gin.ReleaseMode)

	}
	router := gin.New()

	// Enables automatic redirection if the current route can’t be matched but a
	// handler for the path with (without) the trailing slash exists.
	router.RedirectTrailingSlash = true

	router.Use(middleware.CORS())
	router.Use(gin.Logger())
	router.Use(recoverFromError)

	routers.PublicRoutes(router)
	routers.AuthorizedRoutes(router)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "status: 404! no route found."})
	})

	return router
}

func recoverFromError(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Panicf("Panic: %v\n", r)
			debug.PrintStack()

			c.JSON(200, gin.H{"code": 4444, "message": "Server internal error!"})
		}
	}()

	c.Next()
}
