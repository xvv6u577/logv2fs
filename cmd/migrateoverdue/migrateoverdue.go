package migrateoverdue

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
)

// NewMigrateOverdueCmd 将历史 status=overdue 批量迁移为 disabled。
// 缴费功能已从 logv2fs 剥离；欠费状态不再由本系统维护。
// 本命令只改 USER_TRAFFIC_LOGS.status，不调用节点广播（与 migrate-user-status 一致）。
// 运维流程见 docs/REMOVE_PAYMENT_RECORDS.md。
func NewMigrateOverdueCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "migrate-overdue-status",
		Short: "将 USER_TRAFFIC_LOGS 中 status=overdue 迁移为 disabled",
		Long: `一次性数据迁移：把缴费剥离前遗留的 overdue（欠费停用）统一改为 disabled。

默认 --dry-run=true 仅统计数量；真正执行请加 --dry-run=false。

注意：本命令不 drop payment_records，也不 mongodump。
集合备份与删除请按 docs/REMOVE_PAYMENT_RECORDS.md 由运维手工执行。`,
		Run: func(cmd *cobra.Command, args []string) {
			run(dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "仅统计将要迁移的文档数，不实际写库")
	return cmd
}

func run(dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	col := database.GetCollection(model.UserTrafficLogs{})
	filter := bson.M{"status": "overdue"}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatalf("统计 overdue 用户失败: %v", err)
	}

	mode := "DRY-RUN"
	if !dryRun {
		mode = "REAL-RUN"
	}
	log.Printf("=== migrate-overdue-status [%s] 待迁移 overdue 用户: %d ===", mode, count)

	if dryRun || count == 0 {
		if dryRun {
			log.Printf("提示：要真正执行请加 --dry-run=false")
		}
		return
	}

	// 只写 status + updated_at，与手动禁用字段对齐；不广播节点（运维可在迁移后按需重启/同步）。
	now := time.Now()
	res, err := col.UpdateMany(ctx, filter, bson.M{"$set": bson.M{
		"status":     "disabled",
		"updated_at": now,
	}})
	if err != nil {
		log.Fatalf("迁移失败: %v", err)
	}

	log.Printf("迁移完成: matched=%d modified=%d", res.MatchedCount, res.ModifiedCount)
}
