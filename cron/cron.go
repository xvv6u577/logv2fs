package cron

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/robfig/cron"
	box "github.com/sagernet/sing-box"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	thirdparty "github.com/xvv6u577/logv2fs/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Traffic     = model.Traffic
	TrafficInDB = model.TrafficInDB
)

var (
	currentDomain = os.Getenv("CURRENT_DOMAIN")
	// userCollection = database.OpenCollection(database.Client, "USERS")
	// nodesCollection = database.OpenCollection(database.Client, "NODES")
	trafficCollection = database.OpenCollection(database.Client, "TRAFFIC")
	nodeTrafficLogs   = database.OpenCollection(database.Client, "NODE_TRAFFIC_LOGS")
	userTrafficLogs   = database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")
)

// traffic: {Name: "tom", Total: 100}
func LogUserTraffic(collection *mongo.Collection, email string, timestamp time.Time, traffic int64) error {

	var ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var date = timestamp.Format("20060102")
	var month = timestamp.Format("200601")
	var year = timestamp.Format("2006")

	var beforeUpdate model.UserTrafficLogs
	filter := bson.M{"email_as_id": email}

	err := collection.FindOne(ctx, filter).Decode(&beforeUpdate)
	if err != nil {
		log.Printf("error getting user traffic logs: %v\n", err)
	}

	filters := []interface{}{}
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$inc":  bson.M{},
		"$push": bson.M{},
	}

	// check if date exists in daily_logs
	var found bool
	for _, daily := range beforeUpdate.DailyLogs {
		if daily.Date == date {
			found = true
			break
		}
	}
	if !found {
		update["$push"].(bson.M)["daily_logs"] = bson.M{
			"date":    date,
			"traffic": traffic,
		}

	} else {
		update["$inc"].(bson.M)["daily_logs.$[daily].traffic"] = traffic
		filters = append(filters, bson.M{"daily.date": date})
	}

	// check if month exists in monthly_logs
	for _, monthly := range beforeUpdate.MonthlyLogs {
		if monthly.Month == month {
			found = true
			break
		}
	}
	if !found {
		update["$push"].(bson.M)["monthly_logs"] = bson.M{
			"month":   month,
			"traffic": traffic,
		}
	} else {
		update["$inc"].(bson.M)["monthly_logs.$[monthly].traffic"] = traffic
		filters = append(filters, bson.M{"monthly.month": month})
	}

	// check if year exists in yearly_logs
	for _, yearly := range beforeUpdate.YearlyLogs {
		if yearly.Year == year {
			found = true
			break
		}
	}
	if !found {
		update["$push"].(bson.M)["yearly_logs"] = bson.M{
			"year":    year,
			"traffic": traffic,
		}
	} else {
		update["$inc"].(bson.M)["yearly_logs.$[yearly].traffic"] = traffic
		filters = append(filters, bson.M{"yearly.year": year})
	}

	arrayFilters := options.ArrayFilters{
		Filters: filters,
	}

	upsert := true
	updateOptions := options.UpdateOptions{
		ArrayFilters: &arrayFilters,
		Upsert:       &upsert,
	}

	_, err = collection.UpdateOne(ctx, filter, update, &updateOptions)
	return err

}

func LogNodeTraffic(collection *mongo.Collection, domain string, timestamp time.Time, traffic int64) error {

	var ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var date = timestamp.Format("20060102")
	var month = timestamp.Format("200601")
	var year = timestamp.Format("2006")

	var beforeUpdate model.NodeTrafficLogs
	filter := bson.M{"domain_as_id": domain}

	err := collection.FindOne(ctx, filter).Decode(&beforeUpdate)
	if err != nil {
		log.Printf("error getting node traffic logs: %v\n", err)
	}

	filters := []interface{}{}
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$inc":  bson.M{},
		"$push": bson.M{},
	}

	// check if date exists in daily_logs
	var found bool
	for _, daily := range beforeUpdate.DailyLogs {
		if daily.Date == date {
			found = true
			break
		}
	}
	if !found {
		update["$push"].(bson.M)["daily_logs"] = bson.M{
			"date":    date,
			"traffic": traffic,
		}

	} else {
		update["$inc"].(bson.M)["daily_logs.$[daily].traffic"] = traffic
		filters = append(filters, bson.M{"daily.date": date})
	}

	// check if month exists in monthly_logs
	for _, monthly := range beforeUpdate.MonthlyLogs {
		if monthly.Month == month {
			found = true
			break
		}
	}
	if !found {
		update["$push"].(bson.M)["monthly_logs"] = bson.M{
			"month":   month,
			"traffic": traffic,
		}
	} else {
		update["$inc"].(bson.M)["monthly_logs.$[monthly].traffic"] = traffic
		filters = append(filters, bson.M{"monthly.month": month})
	}

	// check if year exists in yearly_logs
	for _, yearly := range beforeUpdate.YearlyLogs {
		if yearly.Year == year {
			found = true
			break
		}
	}
	if !found {
		update["$push"].(bson.M)["yearly_logs"] = bson.M{
			"year":    year,
			"traffic": traffic,
		}
	} else {
		update["$inc"].(bson.M)["yearly_logs.$[yearly].traffic"] = traffic
		filters = append(filters, bson.M{"yearly.year": year})
	}

	arrayFilters := options.ArrayFilters{
		Filters: filters,
	}

	upsert := true
	updateOptions := options.UpdateOptions{
		ArrayFilters: &arrayFilters,
		Upsert:       &upsert,
	}

	_, err = collection.UpdateOne(ctx, filter, update, &updateOptions)
	return err

}

// insert traffic data of all users into traffic collection
func insertTrafficData(traffic Traffic, timestamp time.Time) error {

	_, err := trafficCollection.InsertOne(context.TODO(), TrafficInDB{
		ID:        primitive.NewObjectID(),
		CreatedAt: timestamp,
		Email:     traffic.Name,
		Total:     traffic.Total,
		Domain:    currentDomain,
	})
	return err

}

func Cron_loggingJobs(c *cron.Cron, instance *box.Box) {

	// cron job by 15 mins
	c.AddFunc("0 */15 * * * *", func() {

		timesteamp := time.Now().Local()
		usageData, err := thirdparty.UsageDataOfAll(instance)
		if err != nil {
			log.Printf("error getting usage data: %v\n", err)
		}

		if len(usageData) != 0 {

			for _, perUser := range usageData {

				// perUser = traffic: {Name: "tom", Total: 100}
				if err := insertTrafficData(perUser, timesteamp); err != nil {
					log.Printf("error inserting traffic data: %v\n", err)
				}

				if err := LogUserTraffic(userTrafficLogs, perUser.Name, timesteamp, perUser.Total); err != nil {
					log.Printf("error logging user traffic: %v\n", err)
				}

				if err := LogNodeTraffic(nodeTrafficLogs, currentDomain, timesteamp, perUser.Total); err != nil {
					log.Printf("error logging node traffic: %v\n", err)
				}
			}

		}

		log.Printf("logging user&node traffic: %v", time.Now().Local().Format("20060102"))
	})

}
