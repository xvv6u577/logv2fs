package routine

import (
	"context"
	"fmt"
	"log"
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

func Cron_loggingV2TrafficByUser(traffic Traffic) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// write traffic record to DB

	col := database.OpenCollection(database.Client, traffic.Name)
	userCollection := database.OpenCollection(database.Client, "USERS")

	filter := bson.D{primitive.E{Key: "email", Value: traffic.Name}}
	user := &User{}
	userCollection.FindOne(ctx, filter).Decode(user)

	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
	if err != nil {
		log.Panic("Panic: ", err)
	}
	NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)

	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	col.InsertOne(ctx, model.TrafficInDB{
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

func Cron_loggingV2TrafficAll_everyHour(c *cron.Cron) {

	c.AddFunc("0 */5 * * * *", func() {

		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%d", v2ray.V2_API_ADDRESS, v2ray.V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic("Panic: ", err)
		}

		NSSClient := v2ray.NewStatsServiceClient(cmdConn)
		all, _ := NSSClient.GetAllUserTraffic(true)
		if len(all) != 0 {
			for _, item := range all {
				Cron_loggingV2TrafficByUser(item)
			}
		}
	})
}
