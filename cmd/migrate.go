/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/controllers"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
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
	
迁移的表包括:
- ExpiryCheckDomains: 域名证书过期信息
- SubscriptionNodes: 订阅节点配置
- NodeTrafficLogs: 节点流量日志
- UserTrafficLogs: 用户流量日志
- PaymentRecords: 缴费记录 (每日费用分摊记录将自动生成)
- PaymentStatistics: 费用统计信息
	
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
		&model.NodeTrafficLogsPG{},
		&model.UserTrafficLogsPG{},
		&model.ExpiryCheckDomainInfoPG{},
		&model.SubscriptionNodePG{},
		&model.PaymentRecordPG{},          // 新增：缴费记录表
		&model.DailyPaymentAllocationPG{}, // 新增：每日费用分摊表
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

	// 迁移ExpiryCheckDomains
	err := migrateExpiryCheckDomainsData(batchSize, skipExisting, stats)
	if err != nil {
		return fmt.Errorf("ExpiryCheckDomains迁移失败: %v", err)
	}

	// 迁移SubscriptionNodes
	err = migrateSubscriptionNodesData(batchSize, skipExisting, stats)
	if err != nil {
		return fmt.Errorf("SubscriptionNodes迁移失败: %v", err)
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

	// 迁移PaymentRecords
	err = migratePaymentRecordsData(batchSize, skipExisting, stats)
	if err != nil {
		return fmt.Errorf("PaymentRecords迁移失败: %v", err)
	}

	log.Println("✅ 数据迁移完成")
	return nil
}

// migrateExpiryCheckDomainsData 迁移ExpiryCheckDomains数据
func migrateExpiryCheckDomainsData(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	log.Println("🔄 开始迁移ExpiryCheckDomains数据...")

	// 获取数据库连接
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	expiryCheckDomainCol := database.GetCollection(model.ExpiryCheckDomainInfo{})

	// 查询所有记录
	ctx := context.Background()
	cursor, err := expiryCheckDomainCol.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查询MongoDB ExpiryCheckDomains失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 迁移数据
	var migratedCount int64
	for cursor.Next(ctx) {
		// 解析MongoDB记录
		var mongoDomain model.ExpiryCheckDomainInfo
		if err := cursor.Decode(&mongoDomain); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("解析MongoDB ExpiryCheckDomain失败: %v", err))
			continue
		}

		// 检查PostgreSQL中是否已存在该记录
		var existingCount int64
		if err := postgresDB.Model(&model.ExpiryCheckDomainInfoPG{}).Where("domain = ?", mongoDomain.Domain).Count(&existingCount).Error; err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("检查PostgreSQL ExpiryCheckDomain是否存在失败: %v", err))
			continue
		}

		if existingCount > 0 && skipExisting {
			log.Printf("跳过已存在的ExpiryCheckDomain记录: %s", mongoDomain.Domain)
			continue
		}

		// 创建PostgreSQL记录
		pgDomain := model.ExpiryCheckDomainInfoPG{
			ID:           uuid.New(),
			Domain:       mongoDomain.Domain,
			Remark:       mongoDomain.Remark,
			ExpiredDate:  mongoDomain.ExpiredDate,
			DaysToExpire: mongoDomain.DaysToExpire,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// 插入或更新记录
		if existingCount > 0 {
			if err := postgresDB.Model(&model.ExpiryCheckDomainInfoPG{}).Where("domain = ?", mongoDomain.Domain).Updates(map[string]interface{}{
				"remark":         pgDomain.Remark,
				"expired_date":   pgDomain.ExpiredDate,
				"days_to_expire": pgDomain.DaysToExpire,
				"updated_at":     pgDomain.UpdatedAt,
			}).Error; err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("更新PostgreSQL ExpiryCheckDomain失败: %v", err))
				continue
			}
		} else {
			if err := postgresDB.Create(&pgDomain).Error; err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("插入PostgreSQL ExpiryCheckDomain失败: %v", err))
				continue
			}
		}

		migratedCount++
		if migratedCount%int64(batchSize) == 0 {
			log.Printf("已迁移 %d 条ExpiryCheckDomain记录", migratedCount)
		}
	}

	// 更新统计信息
	stats.DomainRecordsMigrated += migratedCount
	log.Printf("✅ ExpiryCheckDomains迁移完成，共迁移 %d 条记录", migratedCount)
	return nil
}

// migrateSubscriptionNodesData 迁移SubscriptionNodes数据
func migrateSubscriptionNodesData(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	log.Println("🔄 开始迁移SubscriptionNodes数据...")

	// 获取数据库连接
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	subNodesCol := database.GetCollection(model.SubscriptionNode{})

	// 查询所有记录
	ctx := context.Background()
	cursor, err := subNodesCol.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查询MongoDB SubscriptionNodes失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 迁移数据
	var migratedCount int64
	for cursor.Next(ctx) {
		// 解析MongoDB记录
		var mongoNode model.SubscriptionNode
		if err := cursor.Decode(&mongoNode); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("解析MongoDB SubscriptionNode失败: %v", err))
			continue
		}

		// 检查PostgreSQL中是否已存在该记录
		var existingCount int64
		if err := postgresDB.Model(&model.SubscriptionNodePG{}).Where("remark = ?", mongoNode.Remark).Count(&existingCount).Error; err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("检查PostgreSQL SubscriptionNode是否存在失败: %v", err))
			continue
		}

		if existingCount > 0 && skipExisting {
			log.Printf("跳过已存在的SubscriptionNode记录: %s", mongoNode.Remark)
			continue
		}

		// 创建PostgreSQL记录
		pgNode := model.SubscriptionNodePG{
			ID:           uuid.New(),
			Type:         mongoNode.Type,
			Remark:       mongoNode.Remark,
			Domain:       mongoNode.Domain,
			IP:           mongoNode.IP,
			SNI:          mongoNode.SNI,
			UUID:         mongoNode.UUID,
			Path:         mongoNode.PATH,
			ServerPort:   mongoNode.SERVER_PORT,
			Password:     mongoNode.PASSWORD,
			PublicKey:    mongoNode.PUBLIC_KEY,
			ShortID:      mongoNode.SHORT_ID,
			EnableOpenai: mongoNode.EnableOpenai,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// 插入或更新记录
		if existingCount > 0 {
			if err := postgresDB.Model(&model.SubscriptionNodePG{}).Where("remark = ?", mongoNode.Remark).Updates(map[string]interface{}{
				"type":          pgNode.Type,
				"domain":        pgNode.Domain,
				"ip":            pgNode.IP,
				"sni":           pgNode.SNI,
				"uuid":          pgNode.UUID,
				"path":          pgNode.Path,
				"server_port":   pgNode.ServerPort,
				"password":      pgNode.Password,
				"public_key":    pgNode.PublicKey,
				"short_id":      pgNode.ShortID,
				"enable_openai": pgNode.EnableOpenai,
				"updated_at":    pgNode.UpdatedAt,
			}).Error; err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("更新PostgreSQL SubscriptionNode失败: %v", err))
				continue
			}
		} else {
			if err := postgresDB.Create(&pgNode).Error; err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("插入PostgreSQL SubscriptionNode失败: %v", err))
				continue
			}
		}

		migratedCount++
		if migratedCount%int64(batchSize) == 0 {
			log.Printf("已迁移 %d 条SubscriptionNode记录", migratedCount)
		}
	}

	// 更新统计信息
	stats.SubscriptionNodesMigrated += migratedCount
	log.Printf("✅ SubscriptionNodes迁移完成，共迁移 %d 条记录", migratedCount)
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

	// 创建ExpiryCheckDomains表的索引
	err := createExpiryCheckDomainsIndexes(db)
	if err != nil {
		return fmt.Errorf("创建ExpiryCheckDomains索引失败: %v", err)
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

	// 创建Payment相关表的索引
	err = createPaymentIndexes(db)
	if err != nil {
		return fmt.Errorf("创建Payment索引失败: %v", err)
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
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_daily_logs ON user_traffic_logs USING GIN (daily_logs)",
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_monthly_logs ON user_traffic_logs USING GIN (monthly_logs)",
		"CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_yearly_logs ON user_traffic_logs USING GIN (yearly_logs)",

		// NodeTrafficLogs的JSONB索引
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

// createExpiryCheckDomainsIndexes 为ExpiryCheckDomains表创建索引
func createExpiryCheckDomainsIndexes(db *gorm.DB) error {
	log.Println("🔍 创建ExpiryCheckDomains索引...")

	expiryCheckDomainsIndexes := []string{
		// 创建时间索引
		"CREATE INDEX IF NOT EXISTS idx_expiry_check_domains_created_at ON expiry_check_domains(created_at)",
		// 更新时间索引
		"CREATE INDEX IF NOT EXISTS idx_expiry_check_domains_updated_at ON expiry_check_domains(updated_at)",
	}

	for _, indexSQL := range expiryCheckDomainsIndexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️  创建ExpiryCheckDomains索引失败: %s, 错误: %v", indexSQL, err)
		}
	}

	log.Println("✅ ExpiryCheckDomains索引创建完成")
	return nil
}

// createPaymentIndexes 为Payment相关表创建索引
func createPaymentIndexes(db *gorm.DB) error {
	log.Println("💰 创建Payment相关索引...")

	paymentIndexes := []string{
		// PaymentRecords表的索引
		"CREATE INDEX IF NOT EXISTS idx_payment_records_user_email ON payment_records(user_email_as_id)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_start_date ON payment_records(start_date)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_end_date ON payment_records(end_date)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_created_at ON payment_records(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_operator ON payment_records(operator_email)",
		// 联合索引，用于查询用户某个时间段的缴费记录
		"CREATE INDEX IF NOT EXISTS idx_payment_records_user_date_range ON payment_records(user_email_as_id, start_date, end_date)",

		// DailyPaymentAllocations表的索引
		"CREATE INDEX IF NOT EXISTS idx_daily_payment_allocations_user_email ON daily_payment_allocations(user_email_as_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_payment_allocations_date ON daily_payment_allocations(date)",
		"CREATE INDEX IF NOT EXISTS idx_daily_payment_allocations_date_string ON daily_payment_allocations(date_string)",
		"CREATE INDEX IF NOT EXISTS idx_daily_payment_allocations_payment_record_id ON daily_payment_allocations(payment_record_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_payment_allocations_created_at ON daily_payment_allocations(created_at)",
		// 联合索引，用于快速查询用户在某日期的费用分摊
		"CREATE INDEX IF NOT EXISTS idx_daily_payment_allocations_user_date ON daily_payment_allocations(user_email_as_id, date_string)",

		// PaymentStatistics表的索引
		"CREATE INDEX IF NOT EXISTS idx_payment_statistics_stat_type ON payment_statistics(stat_type)",
		"CREATE INDEX IF NOT EXISTS idx_payment_statistics_stat_date ON payment_statistics(stat_date)",
		"CREATE INDEX IF NOT EXISTS idx_payment_statistics_created_at ON payment_statistics(created_at)",
		// 联合索引，用于快速查询特定类型和日期的统计数据
		"CREATE INDEX IF NOT EXISTS idx_payment_statistics_type_date ON payment_statistics(stat_type, stat_date)",
	}

	for _, indexSQL := range paymentIndexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("⚠️  创建Payment索引失败: %s, 错误: %v", indexSQL, err)
		}
	}

	log.Println("✅ Payment索引创建完成")
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
	fmt.Printf("📡 订阅节点: %d\n", stats.SubscriptionNodesMigrated)
	fmt.Printf("💰 缴费记录: %d\n", stats.PaymentRecordsMigrated)
	fmt.Printf("📊 每日费用分摊: %d (自动生成)\n", stats.DailyPaymentAllocationsMigrated)

	totalRecords := stats.DomainRecordsMigrated + stats.NodeRecordsMigrated + stats.UserRecordsMigrated + stats.SubscriptionNodesMigrated + stats.PaymentRecordsMigrated + stats.DailyPaymentAllocationsMigrated
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

// migratePaymentRecordsData 迁移PaymentRecords数据
func migratePaymentRecordsData(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	log.Println("🔄 开始迁移PaymentRecords数据...")

	// 获取数据库连接
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	paymentCol := database.GetCollection(model.PaymentRecord{})

	// 查询所有记录
	ctx := context.Background()
	cursor, err := paymentCol.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("查询MongoDB PaymentRecords失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 迁移数据
	var migratedCount int64
	for cursor.Next(ctx) {
		// 解析MongoDB记录
		var mongoPayment model.PaymentRecord
		if err := cursor.Decode(&mongoPayment); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("解析MongoDB PaymentRecord失败: %v", err))
			continue
		}

		// 检查PostgreSQL中是否已存在该记录
		var existingCount int64
		if err := postgresDB.Model(&model.PaymentRecordPG{}).Where("user_email_as_id = ? AND start_date = ? AND end_date = ?",
			mongoPayment.UserEmailAsId, mongoPayment.StartDate, mongoPayment.EndDate).Count(&existingCount).Error; err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("检查PostgreSQL PaymentRecord是否存在失败: %v", err))
			continue
		}

		if existingCount > 0 && skipExisting {
			log.Printf("跳过已存在的PaymentRecord记录: %s (%s - %s)",
				mongoPayment.UserEmailAsId, mongoPayment.StartDate.Format("2006-01-02"), mongoPayment.EndDate.Format("2006-01-02"))
			continue
		}

		// 创建PostgreSQL记录
		pgPayment := model.PaymentRecordPG{
			ID:            uuid.New(),
			UserEmailAsId: mongoPayment.UserEmailAsId,
			UserName:      mongoPayment.UserName,
			Amount:        mongoPayment.Amount,
			StartDate:     mongoPayment.StartDate,
			EndDate:       mongoPayment.EndDate,
			DailyAmount:   mongoPayment.DailyAmount,
			ServiceDays:   mongoPayment.ServiceDays,
			Remark:        mongoPayment.Remark,
			OperatorEmail: mongoPayment.OperatorEmail,
			OperatorName:  mongoPayment.OperatorName,
			CreatedAt:     mongoPayment.CreatedAt,
			UpdatedAt:     mongoPayment.UpdatedAt,
		}

		// 开始事务处理
		tx := postgresDB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 插入或更新记录
		if existingCount > 0 {
			if err := tx.Model(&model.PaymentRecordPG{}).Where("user_email_as_id = ? AND start_date = ? AND end_date = ?",
				mongoPayment.UserEmailAsId, mongoPayment.StartDate, mongoPayment.EndDate).Updates(map[string]interface{}{
				"user_name":      pgPayment.UserName,
				"amount":         pgPayment.Amount,
				"daily_amount":   pgPayment.DailyAmount,
				"service_days":   pgPayment.ServiceDays,
				"remark":         pgPayment.Remark,
				"operator_email": pgPayment.OperatorEmail,
				"operator_name":  pgPayment.OperatorName,
				"updated_at":     time.Now(),
			}).Error; err != nil {
				tx.Rollback()
				stats.Errors = append(stats.Errors, fmt.Sprintf("更新PostgreSQL PaymentRecord失败: %v", err))
				continue
			}

			// 获取已存在记录的ID
			var existingRecord model.PaymentRecordPG
			if err := tx.Where("user_email_as_id = ? AND start_date = ? AND end_date = ?",
				mongoPayment.UserEmailAsId, mongoPayment.StartDate, mongoPayment.EndDate).First(&existingRecord).Error; err != nil {
				tx.Rollback()
				stats.Errors = append(stats.Errors, fmt.Sprintf("查找已更新的PaymentRecord失败: %v", err))
				continue
			}

			// 删除原有的每日分摊记录
			if err := tx.Where("payment_record_id = ?", existingRecord.ID).Delete(&model.DailyPaymentAllocationPG{}).Error; err != nil {
				tx.Rollback()
				stats.Errors = append(stats.Errors, fmt.Sprintf("删除原有每日分摊记录失败: %v", err))
				continue
			}

			// 重新创建每日分摊记录
			if err := controllers.CreateDailyAllocationsPG(tx, existingRecord.ID, existingRecord); err != nil {
				tx.Rollback()
				stats.Errors = append(stats.Errors, fmt.Sprintf("重新创建每日分摊记录失败: %v", err))
				continue
			}

			// 统计生成的每日分摊记录数量
			stats.DailyPaymentAllocationsMigrated += int64(existingRecord.ServiceDays)
		} else {
			if err := tx.Create(&pgPayment).Error; err != nil {
				tx.Rollback()
				stats.Errors = append(stats.Errors, fmt.Sprintf("插入PostgreSQL PaymentRecord失败: %v", err))
				continue
			}

			// 创建每日分摊记录
			if err := controllers.CreateDailyAllocationsPG(tx, pgPayment.ID, pgPayment); err != nil {
				tx.Rollback()
				stats.Errors = append(stats.Errors, fmt.Sprintf("创建每日分摊记录失败: %v", err))
				continue
			}

			// 统计生成的每日分摊记录数量
			stats.DailyPaymentAllocationsMigrated += int64(pgPayment.ServiceDays)
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("提交事务失败: %v", err))
			continue
		}

		migratedCount++
		if migratedCount%int64(batchSize) == 0 {
			log.Printf("已迁移 %d 条PaymentRecord记录", migratedCount)
		}
	}

	// 更新统计信息
	stats.PaymentRecordsMigrated += migratedCount
	log.Printf("✅ PaymentRecords迁移完成，共迁移 %d 条记录", migratedCount)
	return nil
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// 添加命令行参数
	migrateCmd.Flags().StringP("type", "t", "full", "迁移类型: schema(仅结构), data(仅数据), full(完整迁移)")
	migrateCmd.Flags().IntP("batch-size", "b", 1000, "批处理大小 (推荐: 500-2000)")
	migrateCmd.Flags().BoolP("skip-existing", "s", false, "跳过已存在的记录")
}
