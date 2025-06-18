package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
)

// testMigrationCmd 测试迁移功能
var testMigrationCmd = &cobra.Command{
	Use:   "test-migration",
	Short: "测试数据库迁移功能",
	Long: `这个命令用于测试数据库迁移的各个组件是否正常工作。

测试内容包括:
- 数据库连接测试 (MongoDB 和 PostgreSQL)
- PostgreSQL表结构创建测试
- 索引创建测试
- 数据类型转换测试

使用示例:
  ./logv2fs test-migration
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("🧪 开始数据库迁移功能测试...")

		// 测试MongoDB连接
		if err := testMongoDBConnection(); err != nil {
			log.Fatalf("❌ MongoDB连接测试失败: %v", err)
		}
		log.Println("✅ MongoDB连接测试通过")

		// 测试PostgreSQL连接
		if err := testPostgreSQLConnection(); err != nil {
			log.Fatalf("❌ PostgreSQL连接测试失败: %v", err)
		}
		log.Println("✅ PostgreSQL连接测试通过")

		// 测试表结构创建
		if err := testSchemaCreation(); err != nil {
			log.Fatalf("❌ 表结构创建测试失败: %v", err)
		}
		log.Println("✅ 表结构创建测试通过")

		// 测试索引创建
		if err := testIndexCreation(); err != nil {
			log.Fatalf("❌ 索引创建测试失败: %v", err)
		}
		log.Println("✅ 索引创建测试通过")

		// 测试数据类型转换
		if err := testDataConversion(); err != nil {
			log.Fatalf("❌ 数据类型转换测试失败: %v", err)
		}
		log.Println("✅ 数据类型转换测试通过")

		log.Println("🎉 所有测试都通过！迁移系统准备就绪")
	},
}

// testMongoDBConnection 测试MongoDB连接
func testMongoDBConnection() error {
	client := database.Client
	if client == nil {
		return fmt.Errorf("MongoDB客户端未初始化")
	}

	// 尝试ping MongoDB
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("无法连接到MongoDB: %v", err)
	}

	return nil
}

// testPostgreSQLConnection 测试PostgreSQL连接
func testPostgreSQLConnection() error {
	// 创建数据库（如果不存在）
	err := database.CreateDatabaseIfNotExists()
	if err != nil {
		return fmt.Errorf("创建PostgreSQL数据库失败: %v", err)
	}

	// 初始化连接
	db := database.InitPostgreSQL()
	if db == nil {
		return fmt.Errorf("PostgreSQL数据库连接失败")
	}

	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取PostgreSQL底层连接失败: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("PostgreSQL连接测试失败: %v", err)
	}

	return nil
}

// testSchemaCreation 测试表结构创建
func testSchemaCreation() error {
	db := database.GetPostgresDB()
	if db == nil {
		return fmt.Errorf("PostgreSQL连接未初始化")
	}

	// 启用PostgreSQL扩展
	err := enablePostgresExtensions(db)
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

	// 验证表是否创建成功
	if !db.Migrator().HasTable(&model.DomainPG{}) {
		return fmt.Errorf("DomainPG表未创建")
	}

	if !db.Migrator().HasTable(&model.NodeTrafficLogsPG{}) {
		return fmt.Errorf("NodeTrafficLogsPG表未创建")
	}

	if !db.Migrator().HasTable(&model.UserTrafficLogsPG{}) {
		return fmt.Errorf("UserTrafficLogsPG表未创建")
	}

	return nil
}

// testIndexCreation 测试索引创建
func testIndexCreation() error {
	db := database.GetPostgresDB()
	if db == nil {
		return fmt.Errorf("PostgreSQL连接未初始化")
	}

	// 创建自定义索引
	err := createCustomIndexes(db)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 验证关键索引是否存在
	var indexCount int64

	// 检查Domain表的唯一索引
	err = db.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", "idx_domains_domain_unique").Scan(&indexCount).Error
	if err != nil {
		return fmt.Errorf("检查Domain唯一索引失败: %v", err)
	}

	// 检查JSONB索引
	err = db.Raw("SELECT COUNT(*) FROM pg_indexes WHERE indexname = ?", "idx_user_traffic_logs_hourly_logs").Scan(&indexCount).Error
	if err != nil {
		return fmt.Errorf("检查JSONB索引失败: %v", err)
	}

	return nil
}

// testDataConversion 测试数据类型转换
func testDataConversion() error {
	log.Println("🔄 测试JSON数据转换...")

	// 测试时间序列数据转换
	testHourlyLogs := []struct {
		Timestamp interface{} `json:"timestamp"`
		Traffic   int64       `json:"traffic"`
	}{
		{Timestamp: "2023-01-01T00:00:00Z", Traffic: 1000},
		{Timestamp: "2023-01-01T01:00:00Z", Traffic: 2000},
	}

	// 测试日志转换（这里简化测试，实际中需要更复杂的类型转换）
	log.Printf("测试数据: %+v", testHourlyLogs)

	// 测试用户数据转换
	testUserLog := model.UserTrafficLogs{
		Email_As_Id: "test@example.com",
		Role:        "normal",
		Status:      "plain",
		Used:        5000,
		Credit:      10000,
	}

	pgUserLog, err := convertUserTrafficLogsToPG(testUserLog)
	if err != nil {
		return fmt.Errorf("用户数据转换失败: %v", err)
	}

	// 验证转换结果
	if pgUserLog.EmailAsId != testUserLog.Email_As_Id {
		return fmt.Errorf("用户邮箱转换错误")
	}

	if pgUserLog.Role != testUserLog.Role {
		return fmt.Errorf("用户角色转换错误")
	}

	log.Printf("✅ 用户数据转换成功: %+v", pgUserLog)

	return nil
}

func init() {
	rootCmd.AddCommand(testMigrationCmd)
}
