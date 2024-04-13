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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Traffic         = model.Traffic
	User            = model.User
	TrafficAtPeriod = model.TrafficAtPeriod
	NodeAtPeriod    = model.NodeAtPeriod
	CurrentNode     = model.CurrentNode
)

var (
	CURRENT_DOMAIN         = os.Getenv("CURRENT_DOMAIN")
	V2_API_ADDRESS         = os.Getenv("V2_API_ADDRESS")
	V2_API_PORT            = os.Getenv("V2_API_PORT")
	CRON_INTERVAL_BY_HOUR  = os.Getenv("CRON_INTERVAL_BY_HOUR")
	CRON_INTERVAL_BY_DAY   = os.Getenv("CRON_INTERVAL_BY_DAY")
	CRON_INTERVAL_BY_MONTH = os.Getenv("CRON_INTERVAL_BY_MONTH")
	CRON_INTERVAL_BY_YEAR  = os.Getenv("CRON_INTERVAL_BY_YEAR")
	userCollection         = database.OpenCollection(database.Client, "USERS")
	trafficCollection      = database.OpenCollection(database.Client, "TRAFFIC")
	nodesCollection        = database.OpenCollection(database.Client, "NODES")
)

// traffic: {Name: "tom", Total: 100}
func CronLoggingByUser(traffic Traffic) {

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var current = time.Now().Local()
	var current_year = current.Format("2006")
	var current_month = current.Format("200601")
	var current_day = current.Format("20060102")

	var projections = bson.D{
		{Key: "email", Value: 1},
		{Key: "credit", Value: 1},
		{Key: "used", Value: 1},
		{Key: "path", Value: 1},
		{Key: "status", Value: 1},
		{Key: "used_by_current_day", Value: 1},
		{Key: "used_by_current_month", Value: 1},
		{Key: "used_by_current_year", Value: 1},
		{Key: "traffic_by_day", Value: 1},
		{Key: "traffic_by_month", Value: 1},
		{Key: "traffic_by_year", Value: 1},
	}

	user, err := database.GetUserByName(traffic.Name, projections)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	trafficCollection.InsertOne(ctx, model.TrafficInDB{
		Total:     traffic.Total,
		CreatedAt: current,
		Domain:    CURRENT_DOMAIN,
		Email:     traffic.Name,
	})

	if user.UsedByCurrentDay.Period == current_day {
		user.UsedByCurrentDay.Amount += traffic.Total
		user.UsedByCurrentDay.UsedByDomain[CURRENT_DOMAIN] += traffic.Total
	} else if user.UsedByCurrentDay.Period < current_day {
		user.TrafficByDay = append(user.TrafficByDay, user.UsedByCurrentDay)
		user.UsedByCurrentDay.Period = current_day
		user.UsedByCurrentDay.Amount = traffic.Total
		user.UsedByCurrentDay.UsedByDomain = map[string]int64{}
		user.UsedByCurrentDay.UsedByDomain[CURRENT_DOMAIN] = traffic.Total
	}

	if user.UsedByCurrentMonth.Period == current_month {
		user.UsedByCurrentMonth.Amount += traffic.Total
		user.UsedByCurrentMonth.UsedByDomain[CURRENT_DOMAIN] += traffic.Total
	} else if user.UsedByCurrentMonth.Period < current_month {
		user.TrafficByMonth = append(user.TrafficByMonth, user.UsedByCurrentMonth)
		user.UsedByCurrentMonth.Period = current_month
		user.UsedByCurrentMonth.Amount = traffic.Total
		user.UsedByCurrentMonth.UsedByDomain = map[string]int64{}
		user.UsedByCurrentMonth.UsedByDomain[CURRENT_DOMAIN] = traffic.Total
	}

	if user.UsedByCurrentYear.Period == current_year {
		user.UsedByCurrentYear.Amount += traffic.Total
		user.UsedByCurrentYear.UsedByDomain[CURRENT_DOMAIN] += traffic.Total
	} else if user.UsedByCurrentYear.Period < current_year {
		user.TrafficByYear = append(user.TrafficByYear, user.UsedByCurrentYear)
		user.UsedByCurrentYear.Period = current_year
		user.UsedByCurrentYear.Amount = traffic.Total
		user.UsedByCurrentYear.UsedByDomain = map[string]int64{}
		user.UsedByCurrentYear.UsedByDomain[CURRENT_DOMAIN] = traffic.Total
	}

	user.Usedtraffic += traffic.Total

	filter := bson.D{primitive.E{Key: "email", Value: traffic.Name}}
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "used_by_current_day", Value: user.UsedByCurrentDay},
		primitive.E{Key: "used_by_current_month", Value: user.UsedByCurrentMonth},
		primitive.E{Key: "used_by_current_year", Value: user.UsedByCurrentYear},
		primitive.E{Key: "traffic_by_day", Value: user.TrafficByDay},
		primitive.E{Key: "traffic_by_month", Value: user.TrafficByMonth},
		primitive.E{Key: "traffic_by_year", Value: user.TrafficByYear},
		primitive.E{Key: "used", Value: user.Usedtraffic},
		primitive.E{Key: "updated_at", Value: current},
	}}}

	result := userCollection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		log.Printf("Error: %v", result.Err())
	}

}

func CronLoggingByNode(traffics []Traffic) {

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var current = time.Now().Local()
	var current_year = current.Format("2006")
	var current_month = current.Format("200601")
	var current_day = current.Format("20060102")

	// query the node by domain
	filter := bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
	projection := bson.D{
		{Key: "node_at_current_year", Value: 1},
		{Key: "node_at_current_month", Value: 1},
		{Key: "node_at_current_day", Value: 1},
		{Key: "node_by_year", Value: 1},
		{Key: "node_by_month", Value: 1},
		{Key: "node_by_day", Value: 1},
		{Key: "domain", Value: 1},
		{Key: "status", Value: 1},
	}
	var queriedNode CurrentNode
	err := nodesCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&queriedNode)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	for _, traffic := range traffics {

		if queriedNode.NodeAtCurrentDay.Period == current_day {
			queriedNode.NodeAtCurrentDay.Amount += traffic.Total
			queriedNode.NodeAtCurrentDay.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		} else if queriedNode.NodeAtCurrentDay.Period < current_day {
			queriedNode.NodeByDay = append(queriedNode.NodeByDay, queriedNode.NodeAtCurrentDay)
			queriedNode.NodeAtCurrentDay.Period = current_day
			queriedNode.NodeAtCurrentDay.Amount = traffic.Total
			queriedNode.NodeAtCurrentDay.UserTrafficAtPeriod = map[string]int64{}
			queriedNode.NodeAtCurrentDay.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		}

		if queriedNode.NodeAtCurrentMonth.Period == current_month {
			queriedNode.NodeAtCurrentMonth.Amount += traffic.Total
			queriedNode.NodeAtCurrentMonth.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		} else if queriedNode.NodeAtCurrentMonth.Period < current_month {
			queriedNode.NodeByMonth = append(queriedNode.NodeByMonth, queriedNode.NodeAtCurrentMonth)
			queriedNode.NodeAtCurrentMonth.Period = current_month
			queriedNode.NodeAtCurrentMonth.Amount = traffic.Total
			queriedNode.NodeAtCurrentMonth.UserTrafficAtPeriod = map[string]int64{}
			queriedNode.NodeAtCurrentMonth.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		}

		if queriedNode.NodeAtCurrentYear.Period == current_year {
			queriedNode.NodeAtCurrentYear.Amount += traffic.Total
			queriedNode.NodeAtCurrentYear.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		} else if queriedNode.NodeAtCurrentYear.Period < current_year {
			queriedNode.NodeByYear = append(queriedNode.NodeByYear, queriedNode.NodeAtCurrentYear)
			queriedNode.NodeAtCurrentYear.Period = current_year
			queriedNode.NodeAtCurrentYear.Amount = traffic.Total
			queriedNode.NodeAtCurrentYear.UserTrafficAtPeriod = map[string]int64{}
			queriedNode.NodeAtCurrentYear.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		}

	}

	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "node_at_current_day", Value: queriedNode.NodeAtCurrentDay},
		primitive.E{Key: "node_at_current_month", Value: queriedNode.NodeAtCurrentMonth},
		primitive.E{Key: "node_at_current_year", Value: queriedNode.NodeAtCurrentYear},
		primitive.E{Key: "node_by_day", Value: queriedNode.NodeByDay},
		primitive.E{Key: "node_by_month", Value: queriedNode.NodeByMonth},
		primitive.E{Key: "node_by_year", Value: queriedNode.NodeByYear},
		primitive.E{Key: "updated_at", Value: current},
	}}}

	result := nodesCollection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		log.Printf("Error: %v", result.Err())
	}

}

func Cron_loggingJobs(c *cron.Cron, instance *box.Box) {

	c.AddFunc("0 */15 * * * *", func() {
		// c.AddFunc("0 * * * * *", func() {

		usageData, err := thirdparty.GetUsageDataOfAllUsers(instance)
		if err != nil {
			log.Printf("error getting usage data: %v\n", err)
		}

		if len(usageData) != 0 {
			// by user
			for _, perUser := range usageData {
				CronLoggingByUser(perUser)
			}

			// by node
			CronLoggingByNode(usageData)
		}

		log.Printf("logging user&node every 15 Mins: %v", time.Now().Local().Format("2006010215"))
	})

	c.AddFunc("5 0 0 * * *", func() {
		// c.AddFunc("5 */2 * * *", func() {

		var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var current = time.Now().Local()
		var current_year = current.Format("2006")
		var current_month = current.Format("200601")
		var current_day = current.Format("20060102")

		var filter = bson.D{{}}
		var projections = bson.D{
			{Key: "email", Value: 1},
			{Key: "credit", Value: 1},
			{Key: "used", Value: 1},
			{Key: "path", Value: 1},
			{Key: "status", Value: 1},
			{Key: "used_by_current_day", Value: 1},
			{Key: "used_by_current_month", Value: 1},
			{Key: "used_by_current_year", Value: 1},
			{Key: "traffic_by_day", Value: 1},
			{Key: "traffic_by_month", Value: 1},
			{Key: "traffic_by_year", Value: 1},
		}
		cursor, err := userCollection.Find(ctx, filter, options.Find().SetProjection(projections))
		if err != nil {
			log.Printf("error occured while finding users")
		}

		for cursor.Next(ctx) {
			var user User
			err := cursor.Decode(&user)
			if err != nil {
				log.Printf("error occured while decoding users")
			}

			if user.UsedByCurrentDay.Period == current_day && user.UsedByCurrentMonth.Period == current_month && user.UsedByCurrentYear.Period == current_year {
				continue
			}

			if user.UsedByCurrentDay.Period < current_day {
				user.TrafficByDay = append(user.TrafficByDay, user.UsedByCurrentDay)
				user.UsedByCurrentDay.Period = current_day
				user.UsedByCurrentDay.Amount = 0
				user.UsedByCurrentDay.UsedByDomain = map[string]int64{}
			}

			if user.UsedByCurrentMonth.Period < current_month {
				user.TrafficByMonth = append(user.TrafficByMonth, user.UsedByCurrentMonth)
				user.UsedByCurrentMonth.Period = current_month
				user.UsedByCurrentMonth.Amount = 0
				user.UsedByCurrentMonth.UsedByDomain = map[string]int64{}
			}

			if user.UsedByCurrentYear.Period < current_year {
				user.TrafficByYear = append(user.TrafficByYear, user.UsedByCurrentYear)
				user.UsedByCurrentYear.Period = current_year
				user.UsedByCurrentYear.Amount = 0
				user.UsedByCurrentYear.UsedByDomain = map[string]int64{}
			}

			filter := bson.D{primitive.E{Key: "email", Value: user.Email}}
			upsert := true
			after := options.After
			opt := options.FindOneAndUpdateOptions{
				ReturnDocument: &after,
				Upsert:         &upsert,
			}
			update := bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "used_by_current_day", Value: user.UsedByCurrentDay},
				primitive.E{Key: "used_by_current_month", Value: user.UsedByCurrentMonth},
				primitive.E{Key: "used_by_current_year", Value: user.UsedByCurrentYear},
				primitive.E{Key: "traffic_by_day", Value: user.TrafficByDay},
				primitive.E{Key: "traffic_by_month", Value: user.TrafficByMonth},
				primitive.E{Key: "traffic_by_year", Value: user.TrafficByYear},
				primitive.E{Key: "updated_at", Value: current},
			}}}

			result := userCollection.FindOneAndUpdate(ctx, filter, update, &opt)
			if result.Err() != nil {
				log.Printf("Error: %v", result.Err())
			}

		}

		if err := cursor.Err(); err != nil {
			log.Printf("error occured while cursoring users")
		}
		cursor.Close(ctx)

		log.Printf("update user credit every day: %v", time.Now().Local().Format("2006010215"))
	})

}
