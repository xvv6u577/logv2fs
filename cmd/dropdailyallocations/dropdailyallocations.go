// Package dropdailyallocations 提供一次性数据清理命令：
// 把已经被新计费实现废弃的 daily_payment_allocations 集合整张 drop 掉。
//
// 设计原则：
//   - 默认开启 --dry-run，避免误操作；用户必须显式 --dry-run=false 才会真正 drop。
//   - 真正的事实表 payment_records 不动，分摊视图改为接口实时计算（见 controllers/payment.go）。
//   - 命令是幂等的：集合不存在时（已经 drop 过）也能正常退出。
//
// 使用示例：
//
//	./main drop-daily-allocations                    # 默认仅打印计划
//	./main drop-daily-allocations --dry-run=false    # 真正执行 drop
package dropdailyallocations

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
)

// 老集合名（与已删除的 model.DailyPaymentAllocation.CollectionName() 保持一致）
const legacyCollectionName = "daily_payment_allocations"

// NewDropDailyAllocationsCmd 返回 cobra 命令
func NewDropDailyAllocationsCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "drop-daily-allocations",
		Short: "删除废弃的 daily_payment_allocations 集合",
		Long: `计费模块改造后，daily_payment_allocations（每日费用分摊表）已被
完全废弃 —— 所有月/年统计都直接基于 payment_records 实时算出。

本命令把这张老集合整张 drop 掉，立即释放磁盘 / 备份空间。
默认开启 --dry-run（仅打印计划）。要真正执行请显式传入 --dry-run=false。

可重复执行，集合不存在时也会正常退出。`,
		Run: func(cmd *cobra.Command, args []string) {
			run(dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", true, "仅打印将要进行的变更，不实际删除集合")
	return cmd
}

func run(dryRun bool) {
	mode := "REAL-RUN"
	if dryRun {
		mode = "DRY-RUN"
	}
	log.Printf("=== drop-daily-allocations 启动 [%s] ===", mode)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := database.GetMongoDBClient()
	db := client.Database("logV2rayTrafficDB")

	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{"name": legacyCollectionName})
	if err != nil {
		log.Printf("列举集合失败: %v", err)
		return
	}
	if len(collections) == 0 {
		log.Printf("集合 %s 不存在（可能已经 drop 过），直接退出", legacyCollectionName)
		return
	}

	coll := db.Collection(legacyCollectionName)

	// 顺手统计当前文档数，方便日志留痕
	count, err := coll.EstimatedDocumentCount(ctx)
	if err != nil {
		log.Printf("估算集合大小失败（不影响后续）：%v", err)
	} else {
		log.Printf("集合 %s 当前估算文档数: %d", legacyCollectionName, count)
	}

	if dryRun {
		log.Printf("[dry-run] 计划 drop 集合: %s", legacyCollectionName)
		log.Printf("=== drop-daily-allocations 结束 [DRY-RUN] ===")
		log.Printf("提示：以上为预演结果。要真正执行请加 --dry-run=false")
		return
	}

	if err := coll.Drop(ctx); err != nil {
		log.Printf("drop 集合 %s 失败: %v", legacyCollectionName, err)
		return
	}
	log.Printf("已 drop 集合: %s", legacyCollectionName)
	log.Printf("=== drop-daily-allocations 结束 [REAL-RUN] ===")
}
