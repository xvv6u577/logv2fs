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
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
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
	NODE_TYPE              = os.Getenv("NODE_TYPE")
	userCollection         = database.OpenCollection(database.Client, "USERS")
	trafficCollection      = database.OpenCollection(database.Client, "TRAFFIC")
	nodesCollection        = database.OpenCollection(database.Client, "NODES")
)

func CronLoggingByUser(traffic Traffic) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var current = time.Now().Local()
	var current_year = current.Format("2006")
	var current_month = current.Format("200601")
	var current_day = current.Format("20060102")

	var projections = bson.D{
		{Key: "email", Value: 1},
		{Key: "credittraffic", Value: 1},
		{Key: "usedtraffic", Value: 1},
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
		log.Panic("Panic: ", err)
	}

	trafficCollection.InsertOne(ctx, model.TrafficInDB{
		Total:     traffic.Total,
		CreatedAt: current,
		Domain:    CURRENT_DOMAIN,
		Email:     traffic.Name,
	})

	// if NodeAtCurrentDay.Period not equal to current_day, append NodeAtCurrentDay to NodeByDay, then empty NodeAtCurrentDay.

	if user.UsedByCurrentDay.Period != current_day {
		trafficByDay := user.TrafficByDay
		trafficByDay = append(trafficByDay, user.UsedByCurrentDay)
		var usedByDomain = make(map[string]int64)
		user.UsedByCurrentDay = TrafficAtPeriod{
			Period:       current_day,
			Amount:       0,
			UsedByDomain: usedByDomain,
		}
		user.TrafficByDay = trafficByDay
	}

	// if NodeAtCurrentMonth.Period not equal to current_month, append NodeAtCurrentMonth to NodeByMonth, then empty NodeAtCurrentMonth.
	if user.UsedByCurrentMonth.Period != current_month {
		trafficByMonth := user.TrafficByMonth
		trafficByMonth = append(trafficByMonth, user.UsedByCurrentMonth)
		var usedByDomain = make(map[string]int64)
		user.UsedByCurrentMonth = TrafficAtPeriod{
			Period:       current_month,
			Amount:       0,
			UsedByDomain: usedByDomain,
		}
		user.TrafficByMonth = trafficByMonth
	}

	// if NodeAtCurrentYear.Period not equal to current_year, append NodeAtCurrentYear to NodeByYear, then empty NodeAtCurrentYear.
	if user.UsedByCurrentYear.Period != current_year {
		trafficByYear := user.TrafficByYear
		trafficByYear = append(trafficByYear, user.UsedByCurrentYear)
		var usedByDomain = make(map[string]int64)
		user.UsedByCurrentYear = TrafficAtPeriod{
			Period:       current_year,
			Amount:       0,
			UsedByDomain: usedByDomain,
		}
		user.TrafficByYear = trafficByYear
	}

	filter := bson.D{primitive.E{Key: "email", Value: traffic.Name}}
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	// if user.UsedByCurrentDay.UsedByDomain is nil, make it.
	if user.UsedByCurrentDay.UsedByDomain == nil {
		user.UsedByCurrentDay.UsedByDomain = make(map[string]int64)
	}
	user.UsedByCurrentDay.UsedByDomain[CURRENT_DOMAIN] += traffic.Total
	// if user.UsedByCurrentMonth.UsedByDomain is nil, make it.
	if user.UsedByCurrentMonth.UsedByDomain == nil {
		user.UsedByCurrentMonth.UsedByDomain = make(map[string]int64)
	}
	user.UsedByCurrentMonth.UsedByDomain[CURRENT_DOMAIN] += traffic.Total
	// if user.UsedByCurrentYear.UsedByDomain is nil, make it.
	if user.UsedByCurrentYear.UsedByDomain == nil {
		user.UsedByCurrentYear.UsedByDomain = make(map[string]int64)
	}
	user.UsedByCurrentYear.UsedByDomain[CURRENT_DOMAIN] += traffic.Total

	var update = bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "used_by_current_day", Value: primitive.D{
			primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentDay.Amount)},
			primitive.E{Key: "period", Value: current_day},
			primitive.E{Key: "used_by_domain", Value: user.UsedByCurrentDay.UsedByDomain},
		}},
		primitive.E{Key: "used_by_current_month", Value: primitive.D{
			primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentMonth.Amount)},
			primitive.E{Key: "period", Value: current_month},
			primitive.E{Key: "used_by_domain", Value: user.UsedByCurrentMonth.UsedByDomain},
		}},
		primitive.E{Key: "used_by_current_year", Value: primitive.D{
			primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentYear.Amount)},
			primitive.E{Key: "period", Value: current_year},
			primitive.E{Key: "used_by_domain", Value: user.UsedByCurrentYear.UsedByDomain},
		}},
		primitive.E{Key: "used", Value: traffic.Total + int64(user.Usedtraffic)},
		primitive.E{Key: "updated_at", Value: current},
	}}}

	if traffic.Total+int64(user.Usedtraffic) > int64(user.Credittraffic) {
		cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
		if err != nil {
			log.Panic("Panic: ", err)
		}
		NHSClient := v2ray.NewHandlerServiceClient(cmdConn, user.Path)
		NHSClient.DelUser(user.Email)

		// set status to overdue to variable update.
		update = bson.D{primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "used_by_current_day", Value: primitive.D{
				primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentDay.Amount)},
				primitive.E{Key: "period", Value: current_day},
				primitive.E{Key: "used_by_domain", Value: user.UsedByCurrentDay.UsedByDomain},
			}},
			primitive.E{Key: "used_by_current_month", Value: primitive.D{
				primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentMonth.Amount)},
				primitive.E{Key: "period", Value: current_month},
				primitive.E{Key: "used_by_domain", Value: user.UsedByCurrentMonth.UsedByDomain},
			}},
			primitive.E{Key: "used_by_current_year", Value: primitive.D{
				primitive.E{Key: "amount", Value: traffic.Total + int64(user.UsedByCurrentYear.Amount)},
				primitive.E{Key: "period", Value: current_year},
				primitive.E{Key: "used_by_domain", Value: user.UsedByCurrentYear.UsedByDomain},
			}},
			primitive.E{Key: "used", Value: traffic.Total + int64(user.Usedtraffic)},
			primitive.E{Key: "status", Value: "overdue"},
			primitive.E{Key: "updated_at", Value: current},
		}}}
	}

	result := userCollection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		log.Printf("Error: %v", result.Err())
	}
}

func CronLoggingByNode(traffics []Traffic) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
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
	}
	var queriedNode CurrentNode
	err := nodesCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&queriedNode)
	if err != nil {
		log.Panic("Panic: ", err)
	}

	// merge currentNodeAtPeriod into NodeAtCurrentDay, NodeAtCurrentMonth, NodeAtCurrentYear of queriedNode
	for _, traffic := range traffics {
		queriedNode.UpdatedAt = current
		queriedNode.NodeAtCurrentDay.Period = current_day
		queriedNode.NodeAtCurrentMonth.Period = current_month
		queriedNode.NodeAtCurrentYear.Period = current_year
		queriedNode.NodeAtCurrentDay.Amount += traffic.Total
		queriedNode.NodeAtCurrentMonth.Amount += traffic.Total
		queriedNode.NodeAtCurrentYear.Amount += traffic.Total
		queriedNode.NodeAtCurrentDay.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		queriedNode.NodeAtCurrentMonth.UserTrafficAtPeriod[traffic.Name] += traffic.Total
		queriedNode.NodeAtCurrentYear.UserTrafficAtPeriod[traffic.Name] += traffic.Total
	}

	// upsert the queriedNode into nodesCollection
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}
	filter = bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "node_at_current_day", Value: queriedNode.NodeAtCurrentDay},
		primitive.E{Key: "node_at_current_month", Value: queriedNode.NodeAtCurrentMonth},
		primitive.E{Key: "node_at_current_year", Value: queriedNode.NodeAtCurrentYear},
		primitive.E{Key: "updated_at", Value: current},
	}}}
	result := nodesCollection.FindOneAndUpdate(ctx, filter, update, &opt)
	if result.Err() != nil {
		log.Printf("Error: %v", result.Err())
	}

}

func Log_basicAction() error {
	cmdConn, err := grpc.Dial(fmt.Sprintf("%s:%s", V2_API_ADDRESS, V2_API_PORT), grpc.WithInsecure())
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	NSSClient := v2ray.NewStatsServiceClient(cmdConn)
	allUserTraffic, err := NSSClient.GetAllUserTraffic(true)
	if err != nil {
		log.Printf("Error: %v", err)
		return err
	}

	// log allUserTraffic by user
	if len(allUserTraffic) != 0 {
		for _, perUser := range allUserTraffic {
			CronLoggingByUser(perUser)
		}
	}

	// log allUserTraffic by node
	CronLoggingByNode(allUserTraffic)

	return nil
}

func Cron_loggingJobs(c *cron.Cron) {

	c.AddFunc(CRON_INTERVAL_BY_HOUR, func() {
		Log_basicAction()
		log.Printf("Written by hour: %v", time.Now().Local().Format("2006010215"))
	})

	if NODE_TYPE == "main" || NODE_TYPE == "local" {

		c.AddFunc(CRON_INTERVAL_BY_DAY, func() {
			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			var current = time.Now()
			var next = current.Add(2 * time.Minute)
			// var current_month = current.Local().Format("200601")
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
				var usedByDomain = make(map[string]int64)
				if currentUser.UsedByCurrentDay.Amount == 0 {

					update = bson.D{primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "used_by_current_day", Value: primitive.D{
							primitive.E{Key: "amount", Value: 0},
							primitive.E{Key: "period", Value: next_day},
							primitive.E{Key: "used_by_domain", Value: usedByDomain},
						}},
						primitive.E{Key: "updated_at", Value: current},
					}}}

				} else {

					trafficByDay := currentUser.TrafficByDay
					trafficByDay = append(trafficByDay, currentUser.UsedByCurrentDay)
					update = bson.D{primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "used_by_current_day", Value: primitive.D{
							primitive.E{Key: "amount", Value: 0},
							primitive.E{Key: "period", Value: next_day},
							primitive.E{Key: "used_by_domain", Value: usedByDomain},
						}},
						primitive.E{Key: "traffic_by_day", Value: trafficByDay},
						primitive.E{Key: "updated_at", Value: current},
					}}}
				}

				userCollection.FindOneAndUpdate(ctx, singleUserFilter, update)
			}

			cursor.Close(ctx)

			// query the node by domain
			filter = bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
			projection := bson.D{
				// {Key: "node_at_current_year", Value: 1},
				// {Key: "node_at_current_month", Value: 1},
				{Key: "node_at_current_day", Value: 1},
				// {Key: "node_by_year", Value: 1},
				// {Key: "node_by_month", Value: 1},
				{Key: "node_by_day", Value: 1},
			}
			var queriedNode CurrentNode
			err = nodesCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&queriedNode)
			if err != nil {
				log.Panic("Panic: ", err)
			}

			// append NodeAtCurrentDay to NodeByDay, then empty NodeAtCurrentDay.
			queriedNode.NodeByDay = append(queriedNode.NodeByDay, queriedNode.NodeAtCurrentDay)
			var usedByDomain = make(map[string]int64)
			queriedNode.NodeAtCurrentDay = NodeAtPeriod{
				Period:              next_day,
				Amount:              0,
				UserTrafficAtPeriod: usedByDomain,
			}

			// upsert the queriedNode into nodesCollection
			upsert := true
			after := options.After
			opt := options.FindOneAndUpdateOptions{
				ReturnDocument: &after,
				Upsert:         &upsert,
			}
			filter = bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
			update := bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "node_at_current_day", Value: queriedNode.NodeAtCurrentDay},
				primitive.E{Key: "node_by_day", Value: queriedNode.NodeByDay},
				primitive.E{Key: "updated_at", Value: current},
			}}}
			result := nodesCollection.FindOneAndUpdate(ctx, filter, update, &opt)
			if result.Err() != nil {
				log.Printf("Error: %v", result.Err())
			}

			log.Printf("Written by day: %v", next_day)
		})

		c.AddFunc(CRON_INTERVAL_BY_MONTH, func() {

			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			// 2022-01-01 00:01:00 +0800 CST
			var current = time.Now()
			// 2021-12-31 23:59:00 +0800 CST
			var last = current.Add(-2 * time.Minute)
			// var last_year = last.Local().Format("2006")
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
				var update bson.D
				var usedByDomain = make(map[string]int64)
				if currentUser.UsedByCurrentMonth.Amount == 0 {

					update = bson.D{primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "used_by_current_month", Value: primitive.D{
							primitive.E{Key: "amount", Value: 0},
							primitive.E{Key: "period", Value: current_month},
							primitive.E{Key: "used_by_domain", Value: usedByDomain},
						}},
						primitive.E{Key: "updated_at", Value: last},
					}}}

				} else {

					trafficByMonth := currentUser.TrafficByMonth
					trafficByMonth = append(trafficByMonth, currentUser.UsedByCurrentMonth)
					update = bson.D{primitive.E{Key: "$set", Value: bson.D{
						primitive.E{Key: "used_by_current_month", Value: primitive.D{
							primitive.E{Key: "amount", Value: 0},
							primitive.E{Key: "period", Value: current_month},
							primitive.E{Key: "used_by_domain", Value: usedByDomain},
						}},
						primitive.E{Key: "traffic_by_month", Value: trafficByMonth},
						primitive.E{Key: "updated_at", Value: last},
					}}}
				}

				userCollection.FindOneAndUpdate(ctx, singleUserFilter, update)
			}
			cursor.Close(ctx)

			// query the node by domain
			filter = bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
			projection := bson.D{
				// {Key: "node_at_current_year", Value: 1},
				{Key: "node_at_current_month", Value: 1},
				// {Key: "node_at_current_day", Value: 1},
				// {Key: "node_by_year", Value: 1},
				{Key: "node_by_month", Value: 1},
				// {Key: "node_by_day", Value: 1},
			}
			var queriedNode CurrentNode
			err = nodesCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&queriedNode)
			if err != nil {
				log.Panic("Panic: ", err)
			}

			// append NodeAtCurrentMonth to NodeByMonth, then empty NodeAtCurrentMonth.
			queriedNode.NodeByMonth = append(queriedNode.NodeByMonth, queriedNode.NodeAtCurrentMonth)
			var usedByDomain = make(map[string]int64)
			queriedNode.NodeAtCurrentMonth = NodeAtPeriod{
				Period:              current_month,
				Amount:              0,
				UserTrafficAtPeriod: usedByDomain,
			}

			// upsert the queriedNode into nodesCollection
			upsert := true
			after := options.After
			opt := options.FindOneAndUpdateOptions{
				ReturnDocument: &after,
				Upsert:         &upsert,
			}
			filter = bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
			update := bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "node_at_current_month", Value: queriedNode.NodeAtCurrentMonth},
				primitive.E{Key: "node_by_month", Value: queriedNode.NodeByMonth},
				primitive.E{Key: "updated_at", Value: current},
			}}}
			result := nodesCollection.FindOneAndUpdate(ctx, filter, update, &opt)
			if result.Err() != nil {
				log.Printf("Error: %v", result.Err())
			}

			log.Printf("Written by month: %v", current_month)
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
				var usedByDomain = make(map[string]int64)
				update := bson.D{primitive.E{Key: "$set", Value: bson.D{
					primitive.E{Key: "used_by_current_year", Value: primitive.D{
						primitive.E{Key: "amount", Value: 0},
						primitive.E{Key: "period", Value: current_year},
						primitive.E{Key: "used_by_domain", Value: usedByDomain},
					}},
					primitive.E{Key: "traffic_by_year", Value: trafficByYear},
					primitive.E{Key: "updated_at", Value: last},
				}}}

				userCollection.FindOneAndUpdate(ctx, singleUserFilter, update)
			}

			cursor.Close(ctx)

			// query the node by domain
			filter = bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
			projection := bson.D{
				{Key: "node_at_current_year", Value: 1},
				// {Key: "node_at_current_month", Value: 1},
				// {Key: "node_at_current_day", Value: 1},
				{Key: "node_by_year", Value: 1},
				// {Key: "node_by_month", Value: 1},
				// {Key: "node_by_day", Value: 1},
			}
			var queriedNode CurrentNode
			err = nodesCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&queriedNode)
			if err != nil {
				log.Panic("Panic: ", err)
			}

			// append NodeAtCurrentYear to NodeByYear, then empty NodeAtCurrentYear.
			queriedNode.NodeByYear = append(queriedNode.NodeByYear, queriedNode.NodeAtCurrentYear)
			var usedByDomain = make(map[string]int64)
			queriedNode.NodeAtCurrentYear = NodeAtPeriod{
				Period:              current_year,
				Amount:              0,
				UserTrafficAtPeriod: usedByDomain,
			}

			// upsert the queriedNode into nodesCollection
			upsert := true
			after := options.After
			opt := options.FindOneAndUpdateOptions{
				ReturnDocument: &after,
				Upsert:         &upsert,
			}
			filter = bson.D{primitive.E{Key: "domain", Value: CURRENT_DOMAIN}}
			update := bson.D{primitive.E{Key: "$set", Value: bson.D{
				primitive.E{Key: "node_at_current_year", Value: queriedNode.NodeAtCurrentYear},
				primitive.E{Key: "node_by_year", Value: queriedNode.NodeByYear},
				primitive.E{Key: "updated_at", Value: current},
			}}}
			result := nodesCollection.FindOneAndUpdate(ctx, filter, update, &opt)
			if result.Err() != nil {
				log.Printf("Error: %v", result.Err())
			}

			log.Printf("Written by year: %v", current_year)
		})
	}

}
