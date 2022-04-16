package routine

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	"github.com/caster8013/logv2rayfullstack/v2ray"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

type Traffic = model.Traffic
type User = model.User
type TrafficAtPeriod = model.TrafficAtPeriod

var V2_API_ADDRESS = os.Getenv("V2_API_ADDRESS")
var V2_API_PORT = os.Getenv("V2_API_PORT")
var CRON_INTERVAL_BY_HOUR = os.Getenv("CRON_INTERVAL_BY_HOUR")
var CRON_INTERVAL_BY_DAY = os.Getenv("CRON_INTERVAL_BY_DAY")
var CRON_INTERVAL_BY_MONTH = os.Getenv("CRON_INTERVAL_BY_MONTH")
var CRON_INTERVAL_BY_YEAR = os.Getenv("CRON_INTERVAL_BY_YEAR")
var NODE_TYPE = os.Getenv("NODE_TYPE")

func Cron_loggingV2TrafficByUser(traffic Traffic) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	CURRENT_DOMAIN := os.Getenv("CURRENT_DOMAIN")
	var current = time.Now()
	// var current_year = current.Local().Format("2006")
	// var current_month = current.Local().Format("200601")
	var current_day = current.Local().Format("20060102")

	// write traffic record to DB
	userTrafficCollection := database.OpenCollection(database.Client, traffic.Name)
	userCollection := database.OpenCollection(database.Client, "USERS")

	filter := bson.D{primitive.E{Key: "email", Value: traffic.Name}}
	user := &User{}
	err := userCollection.FindOne(ctx, filter).Decode(user)
	if err != nil {
		log.Panic("Panic: ", err)
	}

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
	if err != nil {
		log.Panic("Panic: ", err)
	}
	NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)

	userTrafficCollection.InsertOne(ctx, model.TrafficInDB{
		Total:     traffic.Total,
		CreatedAt: current,
		Domain:    CURRENT_DOMAIN,
	})

	// 超额的话，删除用户。之后，Usedtraffic += Total
	if user.Status == "plain" {

		var update bson.D

		if traffic.Total+int64(user.Usedtraffic) > int64(user.Credittraffic) {
			NHSClient.DelUser(user.Email)

			update = bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "used_by_current_day", Value: primitive.D{
					primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentDay.Amount)},
					primitive.E{Key: "period", Value: current_day},
				}},
				primitive.E{Key: "used", Value: traffic.Total + int64(user.Usedtraffic)},
				primitive.E{Key: "updated_at", Value: current},
				primitive.E{Key: "status", Value: "overdue"},
			}}}
		} else {
			update = bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "used_by_current_day", Value: primitive.D{
					primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentDay.Amount)},
					primitive.E{Key: "period", Value: current_day},
				}},
				primitive.E{Key: "used", Value: traffic.Total + int64(user.Usedtraffic)},
				primitive.E{Key: "updated_at", Value: current},
			}}}
		}

		userCollection.FindOneAndUpdate(ctx, filter, update)
	}

}

func Log_basicAction() error {
	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
	if err != nil {
		return err
	}

	NSSClient := v2ray.NewStatsServiceClient(cmdConn)
	allUserTraffic, err := NSSClient.GetAllUserTraffic(true)
	if err != nil {
		return err
	}

	if len(allUserTraffic) != 0 {
		for _, trafficPerUser := range allUserTraffic {
			if trafficPerUser.Total != 0 {
				Cron_loggingV2TrafficByUser(trafficPerUser)
			}
		}
	}

	return nil
}

func Cron_loggingJobs(c *cron.Cron) {

	c.AddFunc(CRON_INTERVAL_BY_HOUR, func() {
		Log_basicAction()
		log.Printf("Written by hour")
	})

	if NODE_TYPE == "main" {

		c.AddFunc(CRON_INTERVAL_BY_DAY, func() {
			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			var current = time.Now()
			var next = current.Add(2 * time.Minute)
			var current_month = current.Local().Format("200601")
			var next_day = next.Local().Format("20060102")

			filter := bson.D{{}}
			userCollection := database.OpenCollection(database.Client, "USERS")
			cursor, err := userCollection.Find(ctx, filter)
			if err != nil {
				log.Panic("Panic: ", err)
			}

			for cursor.Next(ctx) {
				var currentUser User
				err := cursor.Decode(&currentUser)
				if err != nil {
					log.Panic("Panic: ", err)
				}

				singleUserFilter := bson.D{primitive.E{Key: "email", Value: currentUser.Email}}
				var update bson.D
				if currentUser.UsedByCurrentDay.Amount == 0 {

					update = bson.D{primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "used_by_current_day", Value: primitive.D{
							primitive.E{Key: "amount", Value: 0},
							primitive.E{Key: "period", Value: next_day},
						}},
						primitive.E{Key: "updated_at", Value: current},
					}}}

				} else {

					trafficByDay := currentUser.TrafficByDay
					trafficByDay = append(trafficByDay, currentUser.UsedByCurrentDay)
					update = bson.D{primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "used_by_current_month", Value: primitive.D{
							primitive.E{Key: "amount", Value: currentUser.UsedByCurrentMonth.Amount + currentUser.UsedByCurrentDay.Amount},
							primitive.E{Key: "period", Value: current_month},
						}},
						primitive.E{Key: "used_by_current_day", Value: primitive.D{
							primitive.E{Key: "amount", Value: 0},
							primitive.E{Key: "period", Value: next_day},
						}},
						primitive.E{Key: "traffic_by_day", Value: trafficByDay},
						primitive.E{Key: "updated_at", Value: current},
					}}}
				}

				userCollection.FindOneAndUpdate(ctx, singleUserFilter, update)
			}

			cursor.Close(ctx)

			log.Printf("Written by day")
		})

		c.AddFunc(CRON_INTERVAL_BY_MONTH, func() {

			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			// 2022-01-01 00:01:00 +0800 CST
			var current = time.Now()
			// 2021-12-31 23:59:00 +0800 CST
			var last = current.Add(-2 * time.Minute)
			var last_year = last.Local().Format("2006")
			var current_month = current.Local().Format("200601")

			filter := bson.D{{}}
			userCollection := database.OpenCollection(database.Client, "USERS")
			cursor, err := userCollection.Find(ctx, filter)
			if err != nil {
				log.Panic("Panic: ", err)
			}

			for cursor.Next(ctx) {
				var currentUser User
				err := cursor.Decode(&currentUser)
				if err != nil {
					log.Panic("Panic: ", err)
				}

				singleUserFilter := bson.D{primitive.E{Key: "email", Value: currentUser.Email}}
				trafficByMonth := currentUser.TrafficByMonth
				trafficByMonth = append(trafficByMonth, currentUser.UsedByCurrentMonth)
				update := bson.D{primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "used_by_current_year", Value: primitive.D{
						primitive.E{Key: "amount", Value: currentUser.UsedByCurrentYear.Amount + currentUser.UsedByCurrentMonth.Amount},
						primitive.E{Key: "period", Value: last_year},
					}},
					primitive.E{Key: "used_by_current_month", Value: primitive.D{
						primitive.E{Key: "amount", Value: 0},
						primitive.E{Key: "period", Value: current_month},
					}},
					primitive.E{Key: "traffic_by_month", Value: trafficByMonth},
					primitive.E{Key: "updated_at", Value: last},
				}}}

				userCollection.FindOneAndUpdate(ctx, singleUserFilter, update)
			}
			cursor.Close(ctx)

			log.Printf("Written by month")
		})

		c.AddFunc(CRON_INTERVAL_BY_YEAR, func() {
			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			// 2022-01-01 00:01:30 +0800 CST
			var current = time.Now()
			// 2021-12-31 23:59:30 +0800 CST
			var last = current.Add(-2 * time.Minute)
			var current_year = current.Local().Format("2006")

			filter := bson.D{{}}
			userCollection := database.OpenCollection(database.Client, "USERS")
			cursor, err := userCollection.Find(ctx, filter)
			if err != nil {
				log.Panic("Panic: ", err)
			}

			for cursor.Next(ctx) {
				var currentUser User
				err := cursor.Decode(&currentUser)
				if err != nil {
					log.Panic("Panic: ", err)
				}

				singleUserFilter := bson.D{primitive.E{Key: "email", Value: currentUser.Email}}
				trafficByYear := currentUser.TrafficByYear
				trafficByYear = append(trafficByYear, currentUser.UsedByCurrentYear)
				update := bson.D{primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "used_by_current_year", Value: primitive.D{
						primitive.E{Key: "amount", Value: 0},
						primitive.E{Key: "period", Value: current_year},
					}},
					primitive.E{Key: "traffic_by_year", Value: trafficByYear},
					primitive.E{Key: "updated_at", Value: last},
				}}}

				userCollection.FindOneAndUpdate(ctx, singleUserFilter, update)
			}

			cursor.Close(ctx)

			log.Printf("Written by year")
		})
	}

}
