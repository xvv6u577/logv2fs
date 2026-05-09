// Package migratetrafficlogs 提供一次性数据迁移命令：把 USER_TRAFFIC_LOGS / NODE_TRAFFIC_LOGS
// 主文档中遗留的 daily_logs / monthly_logs / yearly_logs / hourly_logs 嵌套数组，
// 拆出到独立的 user_traffic_periods / node_traffic_periods 集合，并清理老字段。
//
// 设计原则：
//   - 默认开启 --dry-run，避免误操作；用户必须显式 --dry-run=false 才会写库。
//   - 写入使用 upsert + $setOnInsert，配合周期表的 (owner, kind, period) 唯一索引，
//     保证多次重跑不会重复累加。已存在的记录会被跳过、不被覆盖。
//   - 清理老字段使用 $unset，幂等。
//
// 使用示例：
//
//	./main migrate-traffic-logs                    # 默认仅打印计划
//	./main migrate-traffic-logs --dry-run=false    # 真正执行
package migratetrafficlogs

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// periodSpec 描述一种粒度的拆分规则
type periodSpec struct {
	kind        string // "daily" | "monthly" | "yearly"
	bsonField   string // 老文档里的字段名: "daily_logs" / "monthly_logs" / "yearly_logs"
	periodAlias string // 老子文档里的周期字段名: "date" / "month" / "year"
}

var allSpecs = []periodSpec{
	{kind: "daily", bsonField: "daily_logs", periodAlias: "date"},
	{kind: "monthly", bsonField: "monthly_logs", periodAlias: "month"},
	{kind: "yearly", bsonField: "yearly_logs", periodAlias: "year"},
}

// 老字段统一清理列表（包含已废弃的 hourly_logs）
var legacyFieldsToUnset = []string{"daily_logs", "monthly_logs", "yearly_logs", "hourly_logs"}

// NewMigrateTrafficLogsCmd 返回 cobra 命令
func NewMigrateTrafficLogsCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "migrate-traffic-logs",
		Short: "把用户/节点主文档中的 daily/monthly/yearly_logs 拆到独立集合，并清理 hourly_logs",
		Long: `把 USER_TRAFFIC_LOGS / NODE_TRAFFIC_LOGS 主文档中遗留的嵌套日志数组
拆分到 user_traffic_periods / node_traffic_periods 两张新集合，并 $unset 老字段。

默认开启 --dry-run（仅打印计划）。要真正执行请显式传入 --dry-run=false。

可重复执行，依赖 (owner, kind, period) 唯一索引避免重复插入。`,
		Run: func(cmd *cobra.Command, args []string) {
			run(dryRun)
		},
	}

	// 默认 true，避免误操作
	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "仅打印将要进行的变更，不实际写库")
	return cmd
}

func run(dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	mode := "REAL-RUN"
	if dryRun {
		mode = "DRY-RUN"
	}
	log.Printf("=== migrate-traffic-logs 启动 [%s] ===", mode)

	migrateUsers(ctx, dryRun)
	migrateNodes(ctx, dryRun)

	log.Printf("=== migrate-traffic-logs 结束 [%s] ===", mode)
	if dryRun {
		log.Printf("提示：以上为预演结果。要真正执行请加 --dry-run=false")
	}
}

// migrateUsers 处理 USER_TRAFFIC_LOGS
func migrateUsers(ctx context.Context, dryRun bool) {
	srcCol := database.GetCollection(model.UserTrafficLogs{})
	dstCol := database.GetCollection(model.UserTrafficPeriod{})

	cur, err := srcCol.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{
		"email_as_id":  1,
		"daily_logs":   1,
		"monthly_logs": 1,
		"yearly_logs":  1,
		"hourly_logs":  1,
	}))
	if err != nil {
		log.Printf("[user] Find 失败: %v", err)
		return
	}
	defer cur.Close(ctx)

	var totalUsers, totalInsert, totalSkip, totalUnset int

	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			log.Printf("[user] Decode 失败: %v", err)
			continue
		}
		email, _ := doc["email_as_id"].(string)
		if email == "" {
			log.Printf("[user] 跳过缺失 email_as_id 的文档: _id=%v", doc["_id"])
			continue
		}
		totalUsers++

		ins, skip := splitAndUpsertPeriods(ctx, dstCol, doc, "email_as_id", email, dryRun)
		totalInsert += ins
		totalSkip += skip

		if !dryRun {
			if _, err := srcCol.UpdateOne(ctx, bson.M{"email_as_id": email}, bson.M{
				"$unset": unsetExpr(),
			}); err != nil {
				log.Printf("[user] $unset 老字段失败 email=%s err=%v", email, err)
			} else {
				totalUnset++
			}
		} else if hasAnyLegacyField(doc) {
			totalUnset++
		}
	}

	log.Printf("[user] 汇总: 用户=%d 新写入=%d 已存在跳过=%d 清理老字段=%d",
		totalUsers, totalInsert, totalSkip, totalUnset)
}

// migrateNodes 处理 NODE_TRAFFIC_LOGS
func migrateNodes(ctx context.Context, dryRun bool) {
	srcCol := database.GetCollection(model.NodeTrafficLogs{})
	dstCol := database.GetCollection(model.NodeTrafficPeriod{})

	cur, err := srcCol.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{
		"domain_as_id": 1,
		"daily_logs":   1,
		"monthly_logs": 1,
		"yearly_logs":  1,
		"hourly_logs":  1,
	}))
	if err != nil {
		log.Printf("[node] Find 失败: %v", err)
		return
	}
	defer cur.Close(ctx)

	var totalNodes, totalInsert, totalSkip, totalUnset int

	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			log.Printf("[node] Decode 失败: %v", err)
			continue
		}
		domain, _ := doc["domain_as_id"].(string)
		if domain == "" {
			log.Printf("[node] 跳过缺失 domain_as_id 的文档: _id=%v", doc["_id"])
			continue
		}
		totalNodes++

		ins, skip := splitAndUpsertPeriods(ctx, dstCol, doc, "domain_as_id", domain, dryRun)
		totalInsert += ins
		totalSkip += skip

		if !dryRun {
			if _, err := srcCol.UpdateOne(ctx, bson.M{"domain_as_id": domain}, bson.M{
				"$unset": unsetExpr(),
			}); err != nil {
				log.Printf("[node] $unset 老字段失败 domain=%s err=%v", domain, err)
			} else {
				totalUnset++
			}
		} else if hasAnyLegacyField(doc) {
			totalUnset++
		}
	}

	log.Printf("[node] 汇总: 节点=%d 新写入=%d 已存在跳过=%d 清理老字段=%d",
		totalNodes, totalInsert, totalSkip, totalUnset)
}

// splitAndUpsertPeriods 把单个主文档的 daily/monthly/yearly_logs 全部拆出，
// 用 upsert+$setOnInsert 写入周期集合。返回 (新写入条数, 已存在跳过条数)。
//
// 唯一索引 (ownerKey, kind, period) 是关键的去重保险。已存在的 (ownerKey,kind,period)
// 走 $setOnInsert 的语义：matchedCount=1 但 upsertedCount=0，traffic 不会被覆盖。
func splitAndUpsertPeriods(
	ctx context.Context,
	dstCol *mongo.Collection,
	doc bson.M,
	ownerKey string,
	ownerVal string,
	dryRun bool,
) (insertCount, skipCount int) {
	now := time.Now()

	for _, spec := range allSpecs {
		entries, ok := doc[spec.bsonField].(bson.A)
		if !ok || len(entries) == 0 {
			continue
		}

		for _, raw := range entries {
			entry, ok := raw.(bson.M)
			if !ok {
				continue
			}
			period, _ := entry[spec.periodAlias].(string)
			if period == "" {
				continue
			}
			traffic := toInt64(entry["traffic"])

			if dryRun {
				log.Printf("[dry-run] 计划写入: %s=%s kind=%s period=%s traffic=%d",
					ownerKey, ownerVal, spec.kind, period, traffic)
				insertCount++ // dry-run 时统一按"将插入"计
				continue
			}

			filter := bson.M{
				ownerKey: ownerVal,
				"kind":   spec.kind,
				"period": period,
			}
			update := bson.M{
				"$setOnInsert": bson.M{
					"_id":      primitive.NewObjectID(),
					ownerKey:   ownerVal,
					"kind":     spec.kind,
					"period":   period,
					"traffic":  traffic,
					"updated_at": now,
				},
			}
			upsert := true
			res, err := dstCol.UpdateOne(ctx, filter, update, &options.UpdateOptions{Upsert: &upsert})
			if err != nil {
				log.Printf("[migrate] upsert 失败 %s=%s kind=%s period=%s err=%v",
					ownerKey, ownerVal, spec.kind, period, err)
				continue
			}
			if res.UpsertedCount > 0 {
				insertCount++
			} else {
				skipCount++
			}
		}
	}
	return
}

// unsetExpr 构造统一的 $unset 表达式
func unsetExpr() bson.M {
	m := bson.M{}
	for _, f := range legacyFieldsToUnset {
		m[f] = ""
	}
	return m
}

// hasAnyLegacyField 仅在 dry-run 计数时使用
func hasAnyLegacyField(doc bson.M) bool {
	for _, f := range legacyFieldsToUnset {
		if _, ok := doc[f]; ok {
			return true
		}
	}
	return false
}

// toInt64 兼容 BSON 解码出的 int32 / int64 / float64
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}
