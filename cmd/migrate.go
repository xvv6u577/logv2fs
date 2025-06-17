/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"gorm.io/gorm"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "将MongoDB数据迁移到PostgreSQL数据库 (混合设计模式)",
	Long: `这个命令将执行从MongoDB到PostgreSQL的完整数据迁移。
	
混合设计策略:
- 核心字段使用关系型设计，便于查询和维护
- 时间序列数据使用JSONB存储，保持灵活性
- 支持增量迁移和断点续传
	
支持的迁移类型:
- schema: 仅创建PostgreSQL表结构
- data: 仅迁移数据 (需要先创建表结构)
- full: 完整迁移 (表结构+数据，默认选项)

使用示例:
  # 完整迁移 (推荐)
  ./logv2fs migrate --type=full

  # 仅创建表结构
  ./logv2fs migrate --type=schema

  # 仅迁移数据，批量大小500，跳过重复记录
  ./logv2fs migrate --type=data --batch-size=500 --skip-existing
`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		migrationType, _ := cmd.Flags().GetString("type")
		batchSize, _ := cmd.Flags().GetInt("batch-size")
		skipExisting, _ := cmd.Flags().GetBool("skip-existing")

		log.Printf("🚀 开始执行数据库迁移，类型: %s", migrationType)

		// 初始化统计信息
		stats := &model.MigrationStats{
			StartTime: time.Now(),
			Errors:    []string{},
		}

		// 执行迁移
		switch migrationType {
		case "schema":
			err := migrateSchema()
			if err != nil {
				log.Fatalf("❌ 模式迁移失败: %v", err)
			}
			log.Println("✅ 模式迁移完成")
		case "data":
			err := migrateData(batchSize, skipExisting, stats)
			if err != nil {
				log.Fatalf("❌ 数据迁移失败: %v", err)
			}
		case "full":
			// 先创建模式
			err := migrateSchema()
			if err != nil {
				log.Fatalf("❌ 模式迁移失败: %v", err)
			}
			log.Println("✅ 模式迁移完成")

			// 再迁移数据
			err = migrateData(batchSize, skipExisting, stats)
			if err != nil {
				log.Fatalf("❌ 数据迁移失败: %v", err)
			}
		default:
			log.Fatalf("❌ 不支持的迁移类型: %s", migrationType)
		}

		// 输出迁移统计信息
		stats.EndTime = time.Now()
		printMigrationStats(stats)
	},
}

// migrateSchema 创建PostgreSQL表结构
func migrateSchema() error {
	log.Println("📋 开始创建PostgreSQL数据库和表结构...")

	// 创建数据库（如果不存在）
	err := database.CreateDatabaseIfNotExists()
	if err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	// 初始化PostgreSQL连接
	db := database.InitPostgreSQL()

	// 启用PostgreSQL扩展
	err = enablePostgresExtensions(db)
	if err != nil {
		log.Printf("⚠️  启用PostgreSQL扩展失败: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.DomainPG{},
		&model.NodeTrafficLogsPG{},
		&model.UserTrafficLogsPG{},
	)
	if err != nil {
		return fmt.Errorf("自动迁移失败: %v", err)
	}

	// 创建必要的索引
	err = createCustomIndexes(db)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	log.Println("✅ PostgreSQL表结构创建完成")
	return nil
}

// migrateData 迁移数据从MongoDB到PostgreSQL
func migrateData(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	log.Println("📦 开始数据迁移...")

	// 获取数据库连接
	mongoClient := database.Client
	postgresDB := database.GetPostgresDB()

	// 验证连接
	if mongoClient == nil {
		return fmt.Errorf("MongoDB连接未初始化")
	}
	if postgresDB == nil {
		return fmt.Errorf("PostgreSQL连接未初始化")
	}

	// 迁移Domains (先迁移Domain，因为其他表可能依赖它)
	err := migrateDomainData(batchSize, skipExisting, stats)
	if err != nil {
		return fmt.Errorf("Domain迁移失败: %v", err)
	}

	// 迁移NodeTrafficLogs
	err = migrateNodeTrafficLogsData(batchSize, skipExisting, stats)
	if err != nil {
		return fmt.Errorf("NodeTrafficLogs迁移失败: %v", err)
	}

	// 迁移UserTrafficLogs
	err = migrateUserTrafficLogsData(batchSize, skipExisting, stats)
	if err != nil {
		return fmt.Errorf("UserTrafficLogs迁移失败: %v", err)
	}

	log.Println("✅ 数据迁移完成")
	return nil
}

// migrateNodeTrafficLogsData 迁移NodeTrafficLogs数据
func migrateNodeTrafficLogsData(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	return migrateNodeTrafficLogsDataImpl(batchSize, skipExisting, stats)
}

// migrateUserTrafficLogsData 迁移UserTrafficLogs数据
func migrateUserTrafficLogsData(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	return migrateUserTrafficLogsDataImpl(batchSize, skipExisting, stats)
}

// enablePostgresExtensions 启用必要的PostgreSQL扩展
func enablePostgresExtensions(db *gorm.DB) error {
	log.Println("🔧 启用PostgreSQL扩展...")

	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"", // UUID生成函数
		"CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"",  // 加密函数
	}

	for _, ext := range extensions {
		if err := db.Exec(ext).Error; err != nil {
			log.Printf("⚠️  启用扩展失败: %s, 错误: %v", ext, err)
			// 继续执行，某些扩展可能已经存在或者不是必需的
		}
	}

	log.Println("✅ PostgreSQL扩展启用完成")
	return nil
}

// createCustomIndexes 创建自定义索引
func createCustomIndexes(db *gorm.DB) error {
	log.Println("🔍 创建自定义索引...")

	// 创建Domain表的索引
	err := createDomainIndexes(db)
	if err != nil {
		return fmt.Errorf("创建Domain索引失败: %v", err)
	}

	// 创建NodeTrafficLogs表的索引
	err = createNodeTrafficLogsIndexes(db)
	if err != nil {
		return fmt.Errorf("创建NodeTrafficLogs索引失败: %v", err)
	}

	// 创建UserTrafficLogs表的索引
	err = createUserTrafficLogsIndexes(db)
	if err != nil {
		return fmt.Errorf("创建UserTrafficLogs索引失败: %v", err)
	}

	// 创建JSONB字段的GIN索引
	err = createJSONBIndexes(db)
	if err != nil {
		return fmt.Errorf("创建JSONB索引失败: %v", err)
	}

	// 创建时间范围查询索引
	err = createTimeIndexes(db)
	if err != nil {
		return fmt.Errorf("创建时间索引失败: %v", err)
	}

	log.Println("✅ 自定义索引创建完成")
	return nil
}

// createJSONBIndexes 为JSONB字段创建GIN索引
func createJSONBIndexes(db *gorm.DB) error {
	log.Println("📄 为JSONB字段创建GIN索引...")

	jsonbIndexes := []string{
		// UserTrafficLogs的JSONB索引
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_hourly_logs ON user_traffic_logs USING GIN (hourly_logs)",
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_daily_logs ON user_traffic_logs USING GIN (daily_logs)",
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_monthly_logs ON user_traffic_logs USING GIN (monthly_logs)",
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_yearly_logs ON user_traffic_logs USING GIN (yearly_logs)",

		// NodeTrafficLogs的JSONB索引
		"CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_hourly_logs ON node_traffic_logs USING GIN (hourly_logs)",
		"CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_daily_logs ON node_traffic_logs USING GIN (daily_logs)",
		"CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_monthly_logs ON node_traffic_logs USING GIN (monthly_logs)",
		"CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_yearly_logs ON node_traffic_logs USING GIN (yearly_logs)",
	}

	for _, indexSQL := range jsonbIndexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️  创建JSONB索引失败: %s, 错误: %v", indexSQL, err)
		}
	}

	log.Println("✅ JSONB索引创建完成")
	return nil
}

// createTimeIndexes 创建时间相关索引
func createTimeIndexes(db *gorm.DB) error {
	log.Println("⏰ 创建时间相关索引...")

	timeIndexes := []string{
		// 创建时间索引
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_created_at ON user_traffic_logs(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_created_at ON node_traffic_logs(created_at)",

		// 更新时间索引
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_updated_at ON user_traffic_logs(updated_at)",
		"CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_updated_at ON node_traffic_logs(updated_at)",

		// 复合索引 - 状态+时间
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_status_created_at ON user_traffic_logs(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_status_created_at ON node_traffic_logs(status, created_at)",
	}

	for _, indexSQL := range timeIndexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️  创建时间索引失败: %s, 错误: %v", indexSQL, err)
		}
	}

	log.Println("✅ 时间索引创建完成")
	return nil
}

// printMigrationStats 打印迁移统计信息
func printMigrationStats(stats *model.MigrationStats) {
	duration := stats.EndTime.Sub(stats.StartTime)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 数据迁移统计报告")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏱️  迁移耗时: %v\n", duration)
	fmt.Printf("🔗 域名记录: %d\n", stats.DomainRecordsMigrated)
	fmt.Printf("🌐 节点记录: %d\n", stats.NodeRecordsMigrated)
	fmt.Printf("👥 用户记录: %d\n", stats.UserRecordsMigrated)

	totalRecords := stats.DomainRecordsMigrated + stats.NodeRecordsMigrated + stats.UserRecordsMigrated
	fmt.Printf("📈 总记录数: %d\n", totalRecords)

	if duration.Seconds() > 0 {
		rate := float64(totalRecords) / duration.Seconds()
		fmt.Printf("⚡ 迁移速率: %.2f 记录/秒\n", rate)
	}

	if len(stats.Errors) > 0 {
		fmt.Printf("❌ 错误数量: %d\n", len(stats.Errors))
		fmt.Println("\n错误详情:")
		for i, err := range stats.Errors {
			if i < 10 { // 只显示前10个错误
				fmt.Printf("  %d. %s\n", i+1, err)
			}
		}
		if len(stats.Errors) > 10 {
			fmt.Printf("  ... 还有 %d 个错误未显示\n", len(stats.Errors)-10)
		}
	} else {
		fmt.Println("✅ 迁移过程无错误")
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🎉 数据库迁移完成！")
	fmt.Println("💡 建议: 迁移完成后请验证数据完整性并进行性能测试")
	fmt.Println(strings.Repeat("=", 60))
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// 添加命令行参数
	migrateCmd.Flags().StringP("type", "t", "full", "迁移类型: schema(仅结构), data(仅数据), full(完整迁移)")
	migrateCmd.Flags().IntP("batch-size", "b", 1000, "批处理大小 (推荐: 500-2000)")
	migrateCmd.Flags().BoolP("skip-existing", "s", false, "跳过已存在的记录")
}
