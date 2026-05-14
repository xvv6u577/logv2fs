package migratenodes

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
)

// NewMigrateNodesCmd 返回一次性迁移命令：
// 将 NODE_TRAFFIC_LOGS 中的节点主文档元数据合并到 subscription_nodes。
func NewMigrateNodesCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "migrate-nodes",
		Short: "把 NODE_TRAFFIC_LOGS 节点元数据合并到 subscription_nodes",
		Long: `把 NODE_TRAFFIC_LOGS 中的 status / remark / created_at / updated_at
按 domain_as_id 合并到 subscription_nodes 中同 domain 的 reality / hysteria2 节点。

默认开启 --dry-run（仅打印计划）。要真正执行请显式传入 --dry-run=false。
本命令不会迁移 daily/monthly/yearly 流量数组，也不会 drop NODE_TRAFFIC_LOGS。`,
		Run: func(cmd *cobra.Command, args []string) {
			run(dryRun)
		},
	}

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
	log.Printf("=== migrate-nodes 启动 [%s] ===", mode)

	srcCol := database.GetCollection(model.NodeTrafficLogs{})
	dstCol := database.GetCollection(model.SubscriptionNode{})

	cur, err := srcCol.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("[node] Find NODE_TRAFFIC_LOGS 失败: %v", err)
		return
	}
	defer cur.Close(ctx)

	var total, updated, skipped, notFound int

	for cur.Next(ctx) {
		var oldNode model.NodeTrafficLogs
		if err := cur.Decode(&oldNode); err != nil {
			log.Printf("[node] Decode 失败: %v", err)
			skipped++
			continue
		}
		total++

		if oldNode.Domain_As_Id == "" {
			log.Printf("[node] 跳过缺失 domain_as_id 的旧文档: _id=%s", oldNode.ID.Hex())
			skipped++
			continue
		}

		filter := bson.M{
			"domain": oldNode.Domain_As_Id,
			"type":   bson.M{"$in": []string{"reality", "hysteria2"}},
		}
		update := bson.M{"$set": buildUpdate(oldNode)}

		if dryRun {
			count, err := dstCol.CountDocuments(ctx, filter)
			if err != nil {
				log.Printf("[dry-run] 统计 subscription_nodes 失败 domain=%s err=%v", oldNode.Domain_As_Id, err)
				skipped++
				continue
			}
			if count == 0 {
				log.Printf("[dry-run] 未找到同域名可统计订阅节点: domain=%s", oldNode.Domain_As_Id)
				notFound++
				continue
			}
			log.Printf("[dry-run] 计划更新 subscription_nodes: domain=%s matched=%d status=%s",
				oldNode.Domain_As_Id, count, normalizedStatus(oldNode.Status))
			updated += int(count)
			continue
		}

		res, err := dstCol.UpdateMany(ctx, filter, update)
		if err != nil {
			log.Printf("[node] 更新 subscription_nodes 失败 domain=%s err=%v", oldNode.Domain_As_Id, err)
			skipped++
			continue
		}
		if res.MatchedCount == 0 {
			log.Printf("[node] 未找到同域名可统计订阅节点: domain=%s", oldNode.Domain_As_Id)
			notFound++
			continue
		}
		updated += int(res.ModifiedCount)
	}

	if err := cur.Err(); err != nil {
		log.Printf("[node] Cursor 错误: %v", err)
	}

	log.Printf("[node] 汇总: 旧节点=%d 更新=%d 跳过=%d 未找到订阅节点=%d",
		total, updated, skipped, notFound)
	log.Printf("=== migrate-nodes 结束 [%s] ===", mode)
	if dryRun {
		log.Printf("提示：以上为预演结果。要真正执行请加 --dry-run=false")
	}
}

func buildUpdate(oldNode model.NodeTrafficLogs) bson.M {
	update := bson.M{
		"status": normalizedStatus(oldNode.Status),
	}
	if oldNode.Remark != "" {
		update["remark"] = oldNode.Remark
	}
	if !oldNode.CreatedAt.IsZero() {
		update["created_at"] = oldNode.CreatedAt
	}
	if !oldNode.UpdatedAt.IsZero() {
		update["updated_at"] = oldNode.UpdatedAt
	}
	return update
}

func normalizedStatus(status string) string {
	if status == "inactive" {
		return "inactive"
	}
	return "active"
}
