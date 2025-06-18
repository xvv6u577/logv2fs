package cron

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/robfig/cron"
	box "github.com/sagernet/sing-box"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	thirdparty "github.com/xvv6u577/logv2fs/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type (
	Traffic = model.Traffic
	// PostgreSQL相关类型别名
	UserTrafficLogsPG = model.UserTrafficLogsPG
	NodeTrafficLogsPG = model.NodeTrafficLogsPG
	TrafficLogEntry   = model.TrafficLogEntry
	DailyLogEntry     = model.DailyLogEntry
	MonthlyLogEntry   = model.MonthlyLogEntry
	YearlyLogEntry    = model.YearlyLogEntry
)

var (
	currentDomain = os.Getenv("CURRENT_DOMAIN")
	// MongoDB 集合
	// userCollection = database.OpenCollection(database.Client, "USERS")
	// nodesCollection = database.OpenCollection(database.Client, "NODES")
	nodeTrafficLogs = database.OpenCollection(database.Client, "NODE_TRAFFIC_LOGS")
	userTrafficLogs = database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")
)

// 检查是否使用PostgreSQL
func isUsingPostgreSQL() bool {
	return database.IsUsingPostgres()
}

// PostgreSQL版本的用户流量记录函数
func LogUserTrafficPG(db *gorm.DB, email string, timestamp time.Time, traffic int64) error {
	var date = timestamp.Format("20060102")
	var month = timestamp.Format("200601")
	var year = timestamp.Format("2006")

	var userLog UserTrafficLogsPG

	// 查找用户记录，如果不存在会在后续的Upsert中创建
	result := db.Where("email_as_id = ?", email).First(&userLog)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Printf("查找用户流量记录失败: %v\n", result.Error)
		return result.Error
	}

	isNewRecord := result.Error == gorm.ErrRecordNotFound

	if isNewRecord {
		// 创建新的用户记录
		userLog = UserTrafficLogsPG{
			EmailAsId: email,
			Name:      email, // 默认使用email作为name
			Status:    "plain",
			Role:      "normal",
			Used:      traffic,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		// 初始化日志数组
		hourlyLogs, _ := json.Marshal([]TrafficLogEntry{{Timestamp: timestamp, Traffic: traffic}})
		dailyLogs, _ := json.Marshal([]DailyLogEntry{{Date: date, Traffic: traffic}})
		monthlyLogs, _ := json.Marshal([]MonthlyLogEntry{{Month: month, Traffic: traffic}})
		yearlyLogs, _ := json.Marshal([]YearlyLogEntry{{Year: year, Traffic: traffic}})

		userLog.HourlyLogs = datatypes.JSON(hourlyLogs)
		userLog.DailyLogs = datatypes.JSON(dailyLogs)
		userLog.MonthlyLogs = datatypes.JSON(monthlyLogs)
		userLog.YearlyLogs = datatypes.JSON(yearlyLogs)

		return db.Create(&userLog).Error
	}

	// 更新现有记录
	// 解析现有的日志数据
	var hourlyLogs []TrafficLogEntry
	var dailyLogs []DailyLogEntry
	var monthlyLogs []MonthlyLogEntry
	var yearlyLogs []YearlyLogEntry

	json.Unmarshal([]byte(userLog.HourlyLogs), &hourlyLogs)
	json.Unmarshal([]byte(userLog.DailyLogs), &dailyLogs)
	json.Unmarshal([]byte(userLog.MonthlyLogs), &monthlyLogs)
	json.Unmarshal([]byte(userLog.YearlyLogs), &yearlyLogs)

	// 添加小时级记录
	hourlyLogs = append(hourlyLogs, TrafficLogEntry{Timestamp: timestamp, Traffic: traffic})

	// 更新或创建日级记录
	found := false
	for i := range dailyLogs {
		if dailyLogs[i].Date == date {
			dailyLogs[i].Traffic += traffic
			found = true
			break
		}
	}
	if !found {
		dailyLogs = append(dailyLogs, DailyLogEntry{Date: date, Traffic: traffic})
	}

	// 更新或创建月级记录
	found = false
	for i := range monthlyLogs {
		if monthlyLogs[i].Month == month {
			monthlyLogs[i].Traffic += traffic
			found = true
			break
		}
	}
	if !found {
		monthlyLogs = append(monthlyLogs, MonthlyLogEntry{Month: month, Traffic: traffic})
	}

	// 更新或创建年级记录
	found = false
	for i := range yearlyLogs {
		if yearlyLogs[i].Year == year {
			yearlyLogs[i].Traffic += traffic
			found = true
			break
		}
	}
	if !found {
		yearlyLogs = append(yearlyLogs, YearlyLogEntry{Year: year, Traffic: traffic})
	}

	// 将更新后的数据转换为JSON
	hourlyJSON, _ := json.Marshal(hourlyLogs)
	dailyJSON, _ := json.Marshal(dailyLogs)
	monthlyJSON, _ := json.Marshal(monthlyLogs)
	yearlyJSON, _ := json.Marshal(yearlyLogs)

	// 更新数据库记录
	return db.Model(&userLog).Updates(map[string]interface{}{
		"used":         gorm.Expr("used + ?", traffic),
		"updated_at":   timestamp,
		"hourly_logs":  datatypes.JSON(hourlyJSON),
		"daily_logs":   datatypes.JSON(dailyJSON),
		"monthly_logs": datatypes.JSON(monthlyJSON),
		"yearly_logs":  datatypes.JSON(yearlyJSON),
	}).Error
}

// PostgreSQL版本的节点流量记录函数
func LogNodeTrafficPG(db *gorm.DB, domain string, timestamp time.Time, traffic int64) error {
	var date = timestamp.Format("20060102")
	var month = timestamp.Format("200601")
	var year = timestamp.Format("2006")

	var nodeLog NodeTrafficLogsPG

	// 查找节点记录
	result := db.Where("domain_as_id = ?", domain).First(&nodeLog)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Printf("查找节点流量记录失败: %v\n", result.Error)
		return result.Error
	}

	isNewRecord := result.Error == gorm.ErrRecordNotFound

	if isNewRecord {
		// 创建新的节点记录
		nodeLog = NodeTrafficLogsPG{
			DomainAsId: domain,
			Remark:     domain, // 默认使用domain作为remark
			Status:     "active",
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
		}

		// 初始化日志数组
		hourlyLogs, _ := json.Marshal([]TrafficLogEntry{{Timestamp: timestamp, Traffic: traffic}})
		dailyLogs, _ := json.Marshal([]DailyLogEntry{{Date: date, Traffic: traffic}})
		monthlyLogs, _ := json.Marshal([]MonthlyLogEntry{{Month: month, Traffic: traffic}})
		yearlyLogs, _ := json.Marshal([]YearlyLogEntry{{Year: year, Traffic: traffic}})

		nodeLog.HourlyLogs = datatypes.JSON(hourlyLogs)
		nodeLog.DailyLogs = datatypes.JSON(dailyLogs)
		nodeLog.MonthlyLogs = datatypes.JSON(monthlyLogs)
		nodeLog.YearlyLogs = datatypes.JSON(yearlyLogs)

		return db.Create(&nodeLog).Error
	}

	// 更新现有记录 - 类似用户记录的逻辑
	var hourlyLogs []TrafficLogEntry
	var dailyLogs []DailyLogEntry
	var monthlyLogs []MonthlyLogEntry
	var yearlyLogs []YearlyLogEntry

	json.Unmarshal([]byte(nodeLog.HourlyLogs), &hourlyLogs)
	json.Unmarshal([]byte(nodeLog.DailyLogs), &dailyLogs)
	json.Unmarshal([]byte(nodeLog.MonthlyLogs), &monthlyLogs)
	json.Unmarshal([]byte(nodeLog.YearlyLogs), &yearlyLogs)

	// 添加小时级记录
	hourlyLogs = append(hourlyLogs, TrafficLogEntry{Timestamp: timestamp, Traffic: traffic})

	// 更新或创建日级记录
	found := false
	for i := range dailyLogs {
		if dailyLogs[i].Date == date {
			dailyLogs[i].Traffic += traffic
			found = true
			break
		}
	}
	if !found {
		dailyLogs = append(dailyLogs, DailyLogEntry{Date: date, Traffic: traffic})
	}

	// 更新或创建月级记录
	found = false
	for i := range monthlyLogs {
		if monthlyLogs[i].Month == month {
			monthlyLogs[i].Traffic += traffic
			found = true
			break
		}
	}
	if !found {
		monthlyLogs = append(monthlyLogs, MonthlyLogEntry{Month: month, Traffic: traffic})
	}

	// 更新或创建年级记录
	found = false
	for i := range yearlyLogs {
		if yearlyLogs[i].Year == year {
			yearlyLogs[i].Traffic += traffic
			found = true
			break
		}
	}
	if !found {
		yearlyLogs = append(yearlyLogs, YearlyLogEntry{Year: year, Traffic: traffic})
	}

	// 将更新后的数据转换为JSON
	hourlyJSON, _ := json.Marshal(hourlyLogs)
	dailyJSON, _ := json.Marshal(dailyLogs)
	monthlyJSON, _ := json.Marshal(monthlyLogs)
	yearlyJSON, _ := json.Marshal(yearlyLogs)

	// 更新数据库记录
	return db.Model(&nodeLog).Updates(map[string]interface{}{
		"updated_at":   timestamp,
		"hourly_logs":  datatypes.JSON(hourlyJSON),
		"daily_logs":   datatypes.JSON(dailyJSON),
		"monthly_logs": datatypes.JSON(monthlyJSON),
		"yearly_logs":  datatypes.JSON(yearlyJSON),
	}).Error
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
			"used":       gorm.Expr("used + ?", traffic),
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

func Cron_loggingJobs(c *cron.Cron, instance *box.Box) {

	// cron job by 15 mins - 支持MongoDB和PostgreSQL两种数据库
	c.AddFunc("0 */15 * * * *", func() {

		timesteamp := time.Now().Local()
		usageData, err := thirdparty.UsageDataOfAll(instance)
		if err != nil {
			log.Printf("获取使用数据时出错: %v\n", err)
			return
		}

		if len(usageData) == 0 {
			log.Printf("没有流量数据需要记录: %v", timesteamp.Format("20060102 15:04:05"))
			return
		}

		// 根据环境变量决定使用哪种数据库
		usePostgreSQL := isUsingPostgreSQL()

		if usePostgreSQL {
			log.Printf("使用PostgreSQL记录流量数据...")
			db := database.GetPostgresDB()
			if db == nil {
				log.Printf("PostgreSQL数据库连接不可用，回退到MongoDB")
				usePostgreSQL = false
			} else {
				// 使用PostgreSQL记录流量
				for _, perUser := range usageData {

					// 记录用户流量
					if err := LogUserTrafficPG(db, perUser.Name, timesteamp, perUser.Total); err != nil {
						log.Printf("PostgreSQL用户流量记录失败: %v\n", err)
					}

					// 记录节点流量
					if err := LogNodeTrafficPG(db, currentDomain, timesteamp, perUser.Total); err != nil {
						log.Printf("PostgreSQL节点流量记录失败: %v\n", err)
					}
				}
				log.Printf("PostgreSQL流量记录完成: %v 用户=%d", timesteamp.Format("20060102 15:04:05"), len(usageData))
			}
		}

		if !usePostgreSQL {
			log.Printf("使用MongoDB记录流量数据...")
			// 使用原有的MongoDB逻辑
			for _, perUser := range usageData {

				// perUser = traffic: {Name: "tom", Total: 100}
				if err := LogUserTraffic(userTrafficLogs, perUser.Name, timesteamp, perUser.Total); err != nil {
					log.Printf("MongoDB用户流量记录失败: %v\n", err)
				}

				if err := LogNodeTraffic(nodeTrafficLogs, currentDomain, timesteamp, perUser.Total); err != nil {
					log.Printf("MongoDB节点流量记录失败: %v\n", err)
				}
			}
			log.Printf("MongoDB流量记录完成: %v 用户=%d", timesteamp.Format("20060102 15:04:05"), len(usageData))
		}

	})

}
