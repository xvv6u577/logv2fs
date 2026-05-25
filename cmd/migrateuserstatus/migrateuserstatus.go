package migrateuserstatus

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
)

// NewMigrateUserStatusCmd 将历史 status=deleted 批量迁移为 disabled。
func NewMigrateUserStatusCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "migrate-user-status",
		Short: "将 USER_TRAFFIC_LOGS 中 status=deleted 迁移为 disabled",
		Long: `一次性数据迁移：把手动禁用遗留的 deleted 状态统一改为 disabled。

默认 --dry-run=true 仅统计数量；真正执行请加 --dry-run=false。`,
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
	filter := bson.M{"status": "deleted"}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatalf("统计 deleted 用户失败: %v", err)
	}

	mode := "DRY-RUN"
	if !dryRun {
		mode = "REAL-RUN"
	}
	log.Printf("=== migrate-user-status [%s] 待迁移 deleted 用户: %d ===", mode, count)

	if dryRun || count == 0 {
		if dryRun {
			log.Printf("提示：要真正执行请加 --dry-run=false")
		}
		return
	}

	res, err := col.UpdateMany(ctx, filter, bson.M{"$set": bson.M{"status": "disabled"}})
	if err != nil {
		log.Fatalf("迁移失败: %v", err)
	}

	log.Printf("迁移完成: matched=%d modified=%d", res.MatchedCount, res.ModifiedCount)
}
