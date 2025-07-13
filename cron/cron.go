package cron

import (
	"context"
	"fmt"
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
)

type (
	Traffic = model.Traffic
	// PostgreSQL相关类型别名
	UserTrafficLogsPG = model.UserTrafficLogsPG
	NodeTrafficLogsPG = model.NodeTrafficLogsPG
	DailyLogEntry     = model.DailyLogEntry
	MonthlyLogEntry   = model.MonthlyLogEntry
	YearlyLogEntry    = model.YearlyLogEntry
)

// UserTrafficRequest 定义调用 upsert_user_traffic_log 函数的请求参数
type UserTrafficRequest struct {
	Email     string    `json:"p_email"`
	Timestamp time.Time `json:"p_timestamp"`
	Traffic   int64     `json:"p_traffic"`
}

// NodeTrafficRequest 定义调用 upsert_node_traffic_log 函数的请求参数
type NodeTrafficRequest struct {
	Domain    string    `json:"p_domain"`
	Timestamp time.Time `json:"p_timestamp"`
	Traffic   int64     `json:"p_traffic"`
}

var (
	currentDomain = os.Getenv("CURRENT_DOMAIN")
	// MongoDB 集合
	nodeTrafficLogs = database.GetCollection(model.NodeTrafficLogs{})
	userTrafficLogs = database.GetCollection(model.UserTrafficLogs{})
)

// 检查是否使用PostgreSQL
func isUsingPostgreSQL() bool {
	return database.IsUsingPostgres()
}

// PostgreSQL版本的用户流量记录函数
// 优化版本：使用 Supabase RPC 调用方式执行流量记录
func LogUserTrafficPG(email string, timestamp time.Time, traffic int64) error {
	// 获取 Supabase 客户端
	supaClient := database.GetSupabaseClient()
	if supaClient == nil {
		log.Printf("Supabase 客户端初始化失败")
		return fmt.Errorf("Supabase 客户端初始化失败")
	}

	// 创建上下文
	ctx := context.Background()

	// 准备请求参数
	userRequest := UserTrafficRequest{
		Email:     email,
		Timestamp: timestamp,
		Traffic:   traffic,
	}

	// 使用 Supabase RPC 方法调用 upsert_user_traffic_log 函数
	rpcBuilder := supaClient.DB.RPC("upsert_user_traffic_log", userRequest)

	// 执行 RPC 调用
	err := rpcBuilder.Execute(ctx, nil)
	if err != nil {
		log.Printf("用户流量记录 RPC 调用失败: %v", err)
		return err
	}

	log.Printf("用户流量记录成功 - 用户: %s, 流量: %d, 时间: %s",
		email, traffic, timestamp.Format("2006-01-02 15:04:05"))
	return nil
}

// PostgreSQL版本的节点流量记录函数
// 优化版本：使用 Supabase RPC 调用方式执行流量记录
func LogNodeTrafficPG(domain string, timestamp time.Time, traffic int64) error {
	// 获取 Supabase 客户端
	supaClient := database.GetSupabaseClient()
	if supaClient == nil {
		log.Printf("Supabase 客户端初始化失败")
		return fmt.Errorf("Supabase 客户端初始化失败")
	}

	// 创建上下文
	ctx := context.Background()

	// 准备请求参数
	nodeRequest := NodeTrafficRequest{
		Domain:    domain,
		Timestamp: timestamp,
		Traffic:   traffic,
	}

	// 使用 Supabase RPC 方法调用 upsert_node_traffic_log 函数
	rpcBuilder := supaClient.DB.RPC("upsert_node_traffic_log", nodeRequest)

	// 执行 RPC 调用
	err := rpcBuilder.Execute(ctx, nil)
	if err != nil {
		log.Printf("节点流量记录 RPC 调用失败: %v", err)
		return err
	}

	log.Printf("节点流量记录成功 - 节点: %s, 流量: %d, 时间: %s",
		domain, traffic, timestamp.Format("2006-01-02 15:04:05"))
	return nil
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

func Cron_loggingJobs(c *cron.Cron, instance *box.Box) {

	// cron job by 12 hours - 支持MongoDB和PostgreSQL两种数据库
	// c.AddFunc("0 0 */12 * * *", func() {

	c.AddFunc("0 */15 * * * *", func() {
		// 15 mins - 支持MongoDB和PostgreSQL两种数据库

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
			// 使用 Supabase RPC 调用方式记录流量
			for _, perUser := range usageData {

				// 记录用户流量
				if err := LogUserTrafficPG(perUser.Name, timesteamp, perUser.Total); err != nil {
					log.Printf("PostgreSQL用户流量记录失败: %v\n", err)
				}

				// 记录节点流量
				if err := LogNodeTrafficPG(currentDomain, timesteamp, perUser.Total); err != nil {
					log.Printf("PostgreSQL节点流量记录失败: %v\n", err)
				}
			}
			log.Printf("PostgreSQL流量记录完成: %v 用户=%d", timesteamp.Format("20060102 15:04:05"), len(usageData))
		} else {
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
