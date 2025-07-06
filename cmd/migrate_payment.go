package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// migratePaymentCmd 创建费用记录相关表的命令
var migratePaymentCmd = &cobra.Command{
	Use:   "payment",
	Short: "创建费用记录相关的数据表",
	Long:  `创建费用记录表（payment_records）、每日分摊表（daily_payment_allocations）和统计表（payment_statistics）`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("开始创建费用记录相关表...")

		// PostgreSQL
		if database.IsUsingPostgres() {
			if err := initializePostgresPaymentTables(); err != nil {
				log.Fatalf("初始化PostgreSQL费用记录表失败: %v", err)
			}
		} else {
			// MongoDB
			if err := initializeMongoPaymentCollections(); err != nil {
				log.Fatalf("初始化MongoDB费用记录集合失败: %v", err)
			}
		}

		log.Println("费用记录相关表创建完成！")

		// 显示使用说明
		fmt.Println("\n使用说明：")
		fmt.Println("1. 管理员可以在 /paymentinput 页面为用户添加缴费记录")
		fmt.Println("2. 管理员可以在 /paymentstatistics 页面查看费用统计")
		fmt.Println("3. 在用户管理页面可以查看每个用户的缴费历史")
		fmt.Println("4. 系统会自动创建每日费用分摊记录用于统计分析")
	},
}

func init() {
	migrateCmd.AddCommand(migratePaymentCmd)
}

func initializePostgresPaymentTables() error {
	db := database.GetPostgresDB()

	// 开始事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	log.Println("正在创建 payment_records 表...")
	// 自动创建 payment_records 表
	err := tx.AutoMigrate(&model.PaymentRecordPG{})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to migrate payment_records table: %v", err)
	}

	log.Println("正在创建 daily_payment_allocations 表...")
	// 自动创建 daily_payment_allocations 表
	err = tx.AutoMigrate(&model.DailyPaymentAllocationPG{})
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to migrate daily_payment_allocations table: %v", err)
	}

	log.Println("正在创建 payment_records 表的索引...")
	// 创建 payment_records 表的索引
	paymentRecordsIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_payment_records_user_email ON payment_records(user_email_as_id)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_start_date ON payment_records(start_date)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_end_date ON payment_records(end_date)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_amount ON payment_records(amount)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_created_at ON payment_records(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_operator ON payment_records(operator_email)",
	}

	for _, index := range paymentRecordsIndexes {
		if err := tx.Exec(index).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create payment_records index: %v", err)
		}
	}

	log.Println("正在创建 daily_payment_allocations 表的索引...")
	// 创建 daily_payment_allocations 表的索引
	dailyAllocationIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_daily_allocations_payment_record ON daily_payment_allocations(payment_record_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_allocations_user_email ON daily_payment_allocations(user_email_as_id)",
		"CREATE INDEX IF NOT EXISTS idx_daily_allocations_date ON daily_payment_allocations(date)",
		"CREATE INDEX IF NOT EXISTS idx_daily_allocations_date_string ON daily_payment_allocations(date_string)",
		"CREATE INDEX IF NOT EXISTS idx_daily_allocations_user_date ON daily_payment_allocations(user_email_as_id, date)",
		"CREATE INDEX IF NOT EXISTS idx_daily_allocations_date_range ON daily_payment_allocations(date_string, user_email_as_id)",
	}

	for _, index := range dailyAllocationIndexes {
		if err := tx.Exec(index).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create daily_payment_allocations index: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	fmt.Println("PostgreSQL payment tables initialized successfully")
	fmt.Println("  - payment_records table created with indexes")
	fmt.Println("  - daily_payment_allocations table created with indexes")
	return nil
}

func initializeMongoPaymentCollections() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 初始化 payment_records 集合
	log.Println("正在初始化 payment_records 集合...")
	paymentRecordsCollection := database.OpenCollection(database.Client, "payment_records")

	// 创建 payment_records 索引
	paymentRecordsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_email_as_id", Value: 1}},
			Options: options.Index().SetName("idx_user_email"),
		},
		{
			Keys:    bson.D{{Key: "start_date", Value: -1}},
			Options: options.Index().SetName("idx_start_date"),
		},
		{
			Keys:    bson.D{{Key: "end_date", Value: -1}},
			Options: options.Index().SetName("idx_end_date"),
		},
		{
			Keys:    bson.D{{Key: "amount", Value: -1}},
			Options: options.Index().SetName("idx_amount"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
		{
			Keys:    bson.D{{Key: "operator_email", Value: 1}},
			Options: options.Index().SetName("idx_operator_email"),
		},
		{
			Keys: bson.D{
				{Key: "user_email_as_id", Value: 1},
				{Key: "start_date", Value: -1},
			},
			Options: options.Index().SetName("idx_user_start_date"),
		},
	}

	_, err := paymentRecordsCollection.Indexes().CreateMany(ctx, paymentRecordsIndexes)
	if err != nil {
		return fmt.Errorf("failed to create payment_records indexes: %v", err)
	}

	// 初始化 daily_payment_allocations 集合
	log.Println("正在初始化 daily_payment_allocations 集合...")
	dailyAllocationCollection := database.OpenCollection(database.Client, "daily_payment_allocations")

	// 创建 daily_payment_allocations 索引
	dailyAllocationIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "payment_record_id", Value: 1}},
			Options: options.Index().SetName("idx_payment_record_id"),
		},
		{
			Keys:    bson.D{{Key: "user_email_as_id", Value: 1}},
			Options: options.Index().SetName("idx_user_email"),
		},
		{
			Keys:    bson.D{{Key: "date", Value: -1}},
			Options: options.Index().SetName("idx_date"),
		},
		{
			Keys:    bson.D{{Key: "date_string", Value: -1}},
			Options: options.Index().SetName("idx_date_string"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "user_email_as_id", Value: 1},
				{Key: "date", Value: -1},
			},
			Options: options.Index().SetName("idx_user_date"),
		},
		{
			Keys: bson.D{
				{Key: "date_string", Value: -1},
				{Key: "user_email_as_id", Value: 1},
			},
			Options: options.Index().SetName("idx_date_user"),
		},
		{
			Keys: bson.D{
				{Key: "user_email_as_id", Value: 1},
				{Key: "service_start_date", Value: -1},
				{Key: "service_end_date", Value: -1},
			},
			Options: options.Index().SetName("idx_user_service_dates"),
		},
	}

	_, err = dailyAllocationCollection.Indexes().CreateMany(ctx, dailyAllocationIndexes)
	if err != nil {
		return fmt.Errorf("failed to create daily_payment_allocations indexes: %v", err)
	}

	fmt.Println("MongoDB payment collections initialized successfully")
	fmt.Println("  - payment_records collection created with indexes")
	fmt.Println("  - daily_payment_allocations collection created with indexes")
	return nil
}
