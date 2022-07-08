package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"sync"
	"time"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/grpctools"
	"github.com/caster8013/logv2rayfullstack/model"
	routers "github.com/caster8013/logv2rayfullstack/routers"
	"github.com/caster8013/logv2rayfullstack/routine"
	"github.com/caster8013/logv2rayfullstack/v2ray"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/shomali11/parallelizer"
	"github.com/urfave/cli/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type User = model.User
type Traffic = model.Traffic
type TrafficAtPeriod = model.TrafficAtPeriod

var (
	BOOT_MODE      = os.Getenv("BOOT_MODE")
	NODE_TYPE      = os.Getenv("NODE_TYPE")
	CURRENT_DOMAIN = os.Getenv("CURRENT_DOMAIN")
)

var cronInstance *cron.Cron

func init() {

	cronInstance = cron.New()
	cronInstance.Start()
}

func main() {

	logFile, err := os.OpenFile("./log_file.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logFile)

	app := &cli.App{
		Name:    "logv2rayfullstack",
		Version: "0.1.0",
		Usage:   "A simple CLI program to manage logv2ray backend",
		Authors: []*cli.Author{
			{Name: "Caster Won", Email: "warley8013@gmail.com"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "flag",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "appoint flag",
				// Destination: &flag,
			},
		},
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
					group.Add(RunGRPCServer)

					err := group.Wait()

					return err
				},
			},
			{
				Name:    "cron",
				Aliases: []string{"c"},
				Usage:   "run cron job",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "mongo",
				Aliases: []string{"db"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}},
				},
				Usage: "manage mongoDB",
				Action: func(c *cli.Context) error {

					tag := c.Args().First()
					name := c.String("name")

					switch tag {

					case "test":
						var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
						defer cancel()

						var adminUser User
						userCollection := database.OpenCollection(database.Client, "USERS")
						err := userCollection.FindOne(ctx, bson.M{"email": "casterasadmin"}).Decode(&adminUser)
						if err != nil {
							fmt.Printf("Error: %v\n", err)
							return err
						}

						adminUser.NodeInUseStatus["sl.undervineyard.com"] = true
						adminUser.NodeInUseStatus["w8.undervineyard.com"] = true
						adminUser.ProduceSuburl()
						filter := bson.D{primitive.E{Key: "email", Value: "casterasadmin"}}
						update := bson.M{"$set": bson.M{"status": v2ray.PLAIN, "node_in_use_status": adminUser.NodeInUseStatus, "suburl": adminUser.Suburl}}

						_, err = userCollection.UpdateOne(ctx, filter, update)
						if err != nil {
							msg := "database user info update failed."
							fmt.Printf("%s", msg)
							return err
						}

						return nil

					case "updateNodeInUseStatus":
						var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
						defer cancel()

						var current = time.Now().Local()

						var adminUser User
						userCollection := database.OpenCollection(database.Client, "USERS")
						err := userCollection.FindOne(ctx, bson.M{"email": "casterasadmin"}).Decode(&adminUser)
						if err != nil {
							fmt.Printf("Error: %v\n", err)
						}

						if adminUser.NodeGlobalList == nil {
							adminUser.NodeGlobalList = make(map[string]string)
						}

						allUsers, err := database.GetFullInfosForAllUsers_ForInternalUse()
						if err != nil {
							fmt.Printf("Error: %v\n", err)
						}

						for _, user := range allUsers {
							if user.Role == "admin" {
								user.NodeGlobalList = adminUser.NodeGlobalList
							}

							user.ProduceNodeInUse(adminUser.NodeGlobalList)
							user.UpdatedAt = current

							_, err = userCollection.ReplaceOne(ctx, bson.M{"user_id": user.User_id}, user)
							if err != nil {
								fmt.Printf("Error: %v\n", err)
							}

						}

						return nil

					case "adduserproperty":
						database.AddDBUserProperty()
						return nil

					case "emptydb":
						database.DelUsersTable()
						database.DelUsersTable()
						fmt.Println("user info, user traffic tables deleted in success!")
						return nil

					case "update":
						var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
						defer cancel()

						var current = time.Now().Local()
						var current_year = current.Format("2006")
						var current_month = current.Format("200601")
						var current_day = current.Format("20060102")

						allUsers, err := database.GetFullInfosForAllUsers_ForInternalUse()
						if err != nil {
							fmt.Printf("Error: %v\n", err)
						}

						var checkEleInArray = func(str string, ts []TrafficAtPeriod) (bool, int) {
							for index, ele := range ts {
								if ele.Period == str {
									return true, index
								}
							}
							return false, -1
						}

						for _, user := range allUsers {

							user.Usedtraffic = 0
							user.UsedByCurrentDay = TrafficAtPeriod{}
							user.UsedByCurrentMonth = TrafficAtPeriod{}
							user.UsedByCurrentYear = TrafficAtPeriod{}
							user.TrafficByDay = []TrafficAtPeriod{}
							user.TrafficByMonth = []TrafficAtPeriod{}
							user.TrafficByYear = []TrafficAtPeriod{}

							userCollection := database.OpenCollection(database.Client, "USERS")

							userTrafficCollection := database.OpenCollection(database.Client, user.Email)
							userTrafficFilter := bson.D{{}}
							cur, err := userTrafficCollection.Find(ctx, userTrafficFilter)
							if err != nil {
								fmt.Printf("Error: %v\n", err)
							}

							for cur.Next(ctx) {

								var t model.TrafficInDB
								err := cur.Decode(&t)
								if err != nil {
									fmt.Printf("Error: %v\n", err)
								}

								var at_time = t.CreatedAt
								var at_year = at_time.Format("2006")
								var at_month = at_time.Format("200601")
								var at_day = at_time.Format("20060102")

								user.Usedtraffic += t.Total
								user.CreatedAt = at_time

								if at_day == current_day {
									user.UsedByCurrentDay.Amount += t.Total
									if val, ok := user.UsedByCurrentDay.UsedByDomain[t.Domain]; ok {
										user.UsedByCurrentDay.UsedByDomain[t.Domain] = val + t.Total
									} else {
										user.UsedByCurrentDay.Period = at_day
										if user.UsedByCurrentDay.UsedByDomain == nil {
											user.UsedByCurrentDay.UsedByDomain = make(map[string]int64)
										}
										user.UsedByCurrentDay.UsedByDomain[t.Domain] = t.Total
									}
								} else {
									if ok, index := checkEleInArray(at_day, user.TrafficByDay); ok {
										user.TrafficByDay[index].Amount += t.Total
										if val, ok := user.TrafficByDay[index].UsedByDomain[t.Domain]; ok {
											user.TrafficByDay[index].UsedByDomain[t.Domain] = val + t.Total
										} else {
											if user.TrafficByDay[index].UsedByDomain == nil {
												user.TrafficByDay[index].UsedByDomain = make(map[string]int64)
											}
											user.TrafficByDay[index].UsedByDomain[t.Domain] = t.Total
										}
									} else {
										user.TrafficByDay = append(user.TrafficByDay, TrafficAtPeriod{Period: at_day, Amount: t.Total, UsedByDomain: map[string]int64{t.Domain: t.Total}})
									}
								}

								if at_month == current_month {
									user.UsedByCurrentMonth.Amount += t.Total
									if val, ok := user.UsedByCurrentMonth.UsedByDomain[t.Domain]; ok {
										user.UsedByCurrentMonth.UsedByDomain[t.Domain] = val + t.Total
									} else {
										user.UsedByCurrentMonth.Period = at_month
										if user.UsedByCurrentMonth.UsedByDomain == nil {
											user.UsedByCurrentMonth.UsedByDomain = make(map[string]int64)
										}
										user.UsedByCurrentMonth.UsedByDomain[t.Domain] = t.Total
									}
								} else {
									if ok, index := checkEleInArray(at_month, user.TrafficByMonth); ok {
										user.TrafficByMonth[index].Amount += t.Total
										if val, ok := user.TrafficByMonth[index].UsedByDomain[t.Domain]; ok {
											user.TrafficByMonth[index].UsedByDomain[t.Domain] = val + t.Total
										} else {
											if user.TrafficByMonth[index].UsedByDomain == nil {
												user.TrafficByMonth[index].UsedByDomain = make(map[string]int64)
											}
											user.TrafficByMonth[index].UsedByDomain[t.Domain] = t.Total
										}
									} else {
										user.TrafficByMonth = append(user.TrafficByMonth, TrafficAtPeriod{Period: at_month, Amount: t.Total, UsedByDomain: map[string]int64{t.Domain: t.Total}})
									}
								}

								if at_year == current_year {
									user.UsedByCurrentYear.Amount += t.Total
									if val, ok := user.UsedByCurrentYear.UsedByDomain[t.Domain]; ok {
										user.UsedByCurrentYear.UsedByDomain[t.Domain] = val + t.Total
									} else {
										user.UsedByCurrentYear.Period = at_year
										if user.UsedByCurrentYear.UsedByDomain == nil {
											user.UsedByCurrentYear.UsedByDomain = make(map[string]int64)
										}
										user.UsedByCurrentYear.UsedByDomain[t.Domain] = t.Total
									}
								} else {
									if ok, index := checkEleInArray(at_year, user.TrafficByYear); ok {
										user.TrafficByYear[index].Amount += t.Total
										if val, ok := user.TrafficByYear[index].UsedByDomain[t.Domain]; ok {
											user.TrafficByYear[index].UsedByDomain[t.Domain] = val + t.Total
										} else {
											if user.TrafficByYear[index].UsedByDomain == nil {
												user.TrafficByYear[index].UsedByDomain = make(map[string]int64)
											}
											user.TrafficByYear[index].UsedByDomain[t.Domain] = t.Total
										}
									} else {
										user.TrafficByYear = append(user.TrafficByYear, TrafficAtPeriod{Period: at_year, Amount: t.Total, UsedByDomain: map[string]int64{t.Domain: t.Total}})
									}
								}

							}
							defer cur.Close(ctx)

							upsert := true
							userCollection.FindOneAndReplace(ctx, bson.M{"_id": user.ID}, user, &options.FindOneAndReplaceOptions{Upsert: &upsert})
						}

						return nil

					default:
						fmt.Println(name)
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

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func RunGRPCServer() {
	if CURRENT_DOMAIN == "sl.undervineyard.com" {
		grpctools.GrpcServer("0.0.0.0:80")
	} else {
		grpctools.GrpcServer("0.0.0.0:50051")
	}
}

func V2rayProcess() {

	V2RAY := os.Getenv("V2RAY")
	V2RAY_CONFIG := os.Getenv("V2RAY_CONFIG")

	var cmd = exec.Command(V2RAY, "-config", V2RAY_CONFIG)
	if err := cmd.Run(); err != nil {
		log.Panic("Panic: ", err)
	}
}

func runServer() {
	// wait v2ray process to be ready.
	time.Sleep(time.Second)

	var V2_API_ADDRESS = os.Getenv("V2_API_ADDRESS")
	var V2_API_PORT = os.Getenv("V2_API_PORT")

	var SERVER_ADDRESS = os.Getenv("SERVER_ADDRESS")
	var SERVER_PORT = os.Getenv("SERVER_PORT")

	var projections = bson.D{
		{Key: "_id", Value: 0},
		{Key: "token", Value: 0},
		{Key: "password", Value: 0},
		{Key: "refresh_token", Value: 0},
		{Key: "used_by_current_day", Value: 0},
		{Key: "used_by_current_month", Value: 0},
		{Key: "used_by_current_year", Value: 0},
		{Key: "traffic_by_day", Value: 0},
		{Key: "traffic_by_month", Value: 0},
		{Key: "traffic_by_year", Value: 0},
		{Key: "suburl", Value: 0},
	}
	allUsersInDB, _ := database.GetPartialInfosForAllUsers(projections)
	if len(allUsersInDB) != 0 {

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic(err)
		}
		defer cmdConn.Close()

		var wg sync.WaitGroup
		wg.Add(len(allUsersInDB))

		for _, user := range allUsersInDB {
			go func(user User) {
				defer wg.Done()
				if user.Status == "plain" && user.NodeInUseStatus[CURRENT_DOMAIN] {
					NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
					NHSClient.AddUser(user)
				}
			}(*user)
		}
		wg.Wait()
	}
	// add cron
	routine.Cron_loggingJobs(cronInstance)

	router := gin.New()

	// Enables automatic redirection if the current route can’t be matched but a
	// handler for the path with (without) the trailing slash exists.
	router.RedirectTrailingSlash = true

	if BOOT_MODE == "" {
		router.Use(cors.Default())
		// router.Use(middleware.CORS())
	}
	router.Use(gin.Logger())
	router.Use(recoverFromError)

	routers.AuthRoutes(router)
	routers.UserRoutes(router)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "status: 404! no route found."})
	})

	router.Run(fmt.Sprintf("%s:%s", SERVER_ADDRESS, SERVER_PORT))

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
