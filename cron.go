package main

import (
	"fmt"
	"log"
	"time"

	"github.com/caster8013/logv2rayfullstack/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

func Cron_loggingV2TrafficByUser(traffic Traffic) {

	// write traffic record to DB
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Panic(err)
	}
	col := client.Database("logV2rayTrafficDB").Collection(traffic.Name)
	userCollection := client.Database("logV2rayTrafficDB").Collection("USERS")

	filter := bson.D{primitive.E{Key: "email", Value: traffic.Name}}
	user := &User{}
	userCollection.FindOne(ctx, filter).Decode(user)

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
	if err != nil {
		log.Panic("Panic: ", err)
	}
	NHSClient := NewHandlerServiceClient(cmdConn, user.Path)

	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	col.InsertOne(ctx, types.TrafficInDB{
		Total:     traffic.Total,
		CreatedAt: now,
	})

	// 超额的话，删除用户。之后，Usedtraffic += Total
	if user.Status == "plain" {
		now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		if int64(user.Usedtraffic) > int64(user.Credittraffic) {
			NHSClient.DelUser(user.Email)

			update := bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "used", Value: traffic.Total + int64(user.Usedtraffic)},
				primitive.E{Key: "updated_at", Value: now},
				primitive.E{Key: "status", Value: "overdue"},
			}}}
			userCollection.FindOneAndUpdate(ctx, filter, update)
		} else {
			update := bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "used", Value: traffic.Total + int64(user.Usedtraffic)},
				primitive.E{Key: "updated_at", Value: now},
			}}}
			userCollection.FindOneAndUpdate(ctx, filter, update)
		}
	}

}

func Cron_loggingV2TrafficAll_everyHour() {

	cronInstance.AddFunc("0 */5 * * * *", func() {

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", API_ADDRESS, API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic("Panic: ", err)
		}

		NSSClient := NewStatsServiceClient(cmdConn)
		all, _ := NSSClient.GetAllUserTraffic(true)
		if len(all) != 0 {
			for _, item := range all {
				Cron_loggingV2TrafficByUser(item)
			}
		}
	})
}
