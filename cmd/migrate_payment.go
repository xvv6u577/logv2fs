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
	Long:  `创建费用记录表（payment_records）和统计表（payment_statistics）`,
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
	},
}

func init() {
	migrateCmd.AddCommand(migratePaymentCmd)
}

func initializePostgresPaymentTables() error {
	db := database.GetPostgresDB()

	// 自动创建表
	err := db.AutoMigrate(&model.PaymentRecordPG{})
	if err != nil {
		return fmt.Errorf("failed to migrate payment_records table: %v", err)
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_payment_records_user_email ON payment_records(user_email_as_id)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_start_date ON payment_records(start_date)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_end_date ON payment_records(end_date)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_amount ON payment_records(amount)",
		"CREATE INDEX IF NOT EXISTS idx_payment_records_created_at ON payment_records(created_at)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	fmt.Println("PostgreSQL payment tables initialized successfully")
	return nil
}

func initializeMongoPaymentCollections() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := database.OpenCollection(database.Client, "payment_records")

	// 创建索引
	indexes := []mongo.IndexModel{
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
			Keys: bson.D{
				{Key: "user_email_as_id", Value: 1},
				{Key: "start_date", Value: -1},
			},
			Options: options.Index().SetName("idx_user_start_date"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	fmt.Println("MongoDB payment collections initialized successfully")
	return nil
}
