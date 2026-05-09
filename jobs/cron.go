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

// upsertTrafficPeriod 在指定的周期集合（user_traffic_periods 或 node_traffic_periods）
// 中按 (ownerKey=ownerVal, kind, period) 唯一键累加 traffic。
//
// 集合的唯一索引保证了高并发下原子安全；首次写入会触发 upsert。
func upsertTrafficPeriod(
	ctx context.Context,
	periodCol *mongo.Collection,
	ownerKey string, // "email_as_id" 或 "domain_as_id"
	ownerVal string,
	kind string, // "daily" | "monthly" | "yearly"
	period string, // 对应粒度的字符串
	traffic int64,
	now time.Time,
) error {
	filter := bson.M{
		ownerKey: ownerVal,
		"kind":   kind,
		"period": period,
	}
	update := bson.M{
		"$inc": bson.M{"traffic": traffic},
		"$set": bson.M{"updated_at": now},
	}
	upsert := true
	_, err := periodCol.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
	return err
}

// LogUserTraffic 写入单个用户的本次采样流量。
//
// 拆分集合后流程被显著简化：
//  1. 累加用户主文档的 used 字段（仅 $inc，不再 Find 整个文档）；
//  2. 在 user_traffic_periods 集合中按日/月/年 upsert 累加 traffic。
//
// 入参 collection 仍是 USER_TRAFFIC_LOGS 集合，保持调用点签名不变。
func LogUserTraffic(collection *mongo.Collection, email string, timestamp time.Time, traffic int64) error {

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	now := time.Now()
	date := timestamp.Format("20060102")
	month := timestamp.Format("200601")
	year := timestamp.Format("2006")

	// 1) 用户主文档：累加 used、刷新 updated_at
	if _, err := collection.UpdateOne(
		ctx,
		bson.M{"email_as_id": email},
		bson.M{
			"$inc": bson.M{"used": traffic},
			"$set": bson.M{"updated_at": now},
		},
	); err != nil {
		log.Printf("更新用户主文档 used 失败: %v", err)
		return err
	}

	// 2) 周期集合：按日/月/年分别 upsert
	periodCol := database.GetCollection(model.UserTrafficPeriod{})
	for _, item := range []struct {
		kind   string
		period string
	}{
		{"daily", date},
		{"monthly", month},
		{"yearly", year},
	} {
		if err := upsertTrafficPeriod(ctx, periodCol, "email_as_id", email, item.kind, item.period, traffic, now); err != nil {
			log.Printf("用户流量周期 upsert 失败 email=%s kind=%s period=%s err=%v",
				email, item.kind, item.period, err)
			return err
		}
	}
	return nil
}

// LogNodeTraffic 写入单个节点的本次采样流量。
//
// 与 LogUserTraffic 同理：节点主文档仅更新 updated_at，
// 周期数据全部下沉到 node_traffic_periods 集合。
func LogNodeTraffic(collection *mongo.Collection, domain string, timestamp time.Time, traffic int64) error {

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	now := time.Now()
	date := timestamp.Format("20060102")
	month := timestamp.Format("200601")
	year := timestamp.Format("2006")

	// 1) 节点主文档：刷新 updated_at（只在已存在时更新；不在则跳过，避免凭空创建节点）
	if _, err := collection.UpdateOne(
		ctx,
		bson.M{"domain_as_id": domain},
		bson.M{"$set": bson.M{"updated_at": now}},
	); err != nil {
		log.Printf("更新节点主文档 updated_at 失败: %v", err)
		return err
	}

	// 2) 周期集合：按日/月/年分别 upsert
	periodCol := database.GetCollection(model.NodeTrafficPeriod{})
	for _, item := range []struct {
		kind   string
		period string
	}{
		{"daily", date},
		{"monthly", month},
		{"yearly", year},
	} {
		if err := upsertTrafficPeriod(ctx, periodCol, "domain_as_id", domain, item.kind, item.period, traffic, now); err != nil {
			log.Printf("节点流量周期 upsert 失败 domain=%s kind=%s period=%s err=%v",
				domain, item.kind, item.period, err)
			return err
		}
	}
	return nil
}

// Cron_loggingJobs 注册定时任务：每 15 分钟将 sing-box 中累积的流量数据写入 MongoDB
func Cron_loggingJobs(c *cron.Cron, instance *box.Box) {

	c.AddFunc("0 * * * * *", func() {
		// c.AddFunc("0 */15 * * * *", func() {

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
			log.Printf("用户流量记录: %v %v", perUser.Name, perUser.Total)
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
