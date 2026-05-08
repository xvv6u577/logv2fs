package jobs

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/robfig/cron"
	box "github.com/sagernet/sing-box"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"github.com/xvv6u577/logv2fs/singbox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Traffic = model.Traffic
)

func getCurrentDomain() string {
	return os.Getenv("CURRENT_DOMAIN")
}

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
			"used":       beforeUpdate.Used + traffic,
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

// Cron_loggingJobs 注册定时任务：每 15 分钟将 sing-box 中累积的流量数据写入 MongoDB
func Cron_loggingJobs(c *cron.Cron, instance *box.Box) {

	c.AddFunc("0 */15 * * * *", func() {

		timesteamp := time.Now().Local()
		usageData, err := singbox.UsageDataOfAll(instance)
		if err != nil {
			log.Printf("获取使用数据时出错: %v\n", err)
			return
		}

		if len(usageData) == 0 {
			log.Printf("没有流量数据需要记录: %v", timesteamp.Format("20060102 15:04:05"))
			return
		}

		for _, perUser := range usageData {
			// perUser = traffic: {Name: "tom", Total: 100}
			if err := LogUserTraffic(database.GetCollection(model.UserTrafficLogs{}), perUser.Name, timesteamp, perUser.Total); err != nil {
				log.Printf("用户流量记录失败: %v\n", err)
			}

			if err := LogNodeTraffic(database.GetCollection(model.NodeTrafficLogs{}), getCurrentDomain(), timesteamp, perUser.Total); err != nil {
				log.Printf("节点流量记录失败: %v\n", err)
			}
		}
		log.Printf("流量记录完成: %v 用户=%d", timesteamp.Format("20060102 15:04:05"), len(usageData))

	})

}
