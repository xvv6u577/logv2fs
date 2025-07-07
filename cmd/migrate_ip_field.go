package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
)

// migrateIPFieldCmd represents the migrate-ip-field command
var migrateIPFieldCmd = &cobra.Command{
	Use:   "migrate-ip-field",
	Short: "修复subscription_nodes表中的IP字段类型",
	Long: `将subscription_nodes表中的IP字段从inet类型更改为text类型，
以支持域名输入功能。这个迁移会保留所有现有数据。`,
	Run: func(cmd *cobra.Command, args []string) {
		migrateIPField()
	},
}

func init() {
	rootCmd.AddCommand(migrateIPFieldCmd)
}

func migrateIPField() {
	log.Printf("开始迁移 subscription_nodes 表中的 IP 字段...")

	// 连接PostgreSQL数据库
	db := database.GetPostgresDB()

	// 检查当前IP字段的类型
	var dataType string
	checkQuery := `SELECT data_type FROM information_schema.columns 
				   WHERE table_name = 'subscription_nodes' AND column_name = 'ip'`

	err := db.Raw(checkQuery).Scan(&dataType).Error
	if err != nil {
		log.Fatalf("检查IP字段类型失败: %v", err)
	}

	log.Printf("当前IP字段类型: %s", dataType)

	// 如果已经是text类型，则跳过迁移
	if dataType == "text" {
		log.Printf("✅ IP字段已经是text类型，无需迁移")
		return
	}

	// 如果不是inet类型，警告用户
	if dataType != "inet" {
		log.Printf("⚠️  警告：IP字段类型为 %s，不是预期的 inet 类型", dataType)
		log.Printf("继续执行迁移...")
	}

	// 开始事务
	tx := db.Begin()
	if tx.Error != nil {
		log.Fatalf("开始事务失败: %v", tx.Error)
	}

	// 第一步：创建临时列
	log.Printf("第一步：创建临时列...")
	if err := tx.Exec("ALTER TABLE subscription_nodes ADD COLUMN ip_temp TEXT").Error; err != nil {
		tx.Rollback()
		log.Fatalf("创建临时列失败: %v", err)
	}

	// 第二步：转换数据
	log.Printf("第二步：转换现有数据...")
	if err := tx.Exec("UPDATE subscription_nodes SET ip_temp = CAST(ip AS TEXT)").Error; err != nil {
		tx.Rollback()
		log.Fatalf("转换数据失败: %v", err)
	}

	// 第三步：删除原列
	log.Printf("第三步：删除原IP列...")
	if err := tx.Exec("ALTER TABLE subscription_nodes DROP COLUMN ip").Error; err != nil {
		tx.Rollback()
		log.Fatalf("删除原列失败: %v", err)
	}

	// 第四步：重命名临时列
	log.Printf("第四步：重命名临时列...")
	if err := tx.Exec("ALTER TABLE subscription_nodes RENAME COLUMN ip_temp TO ip").Error; err != nil {
		tx.Rollback()
		log.Fatalf("重命名列失败: %v", err)
	}

	// 第五步：创建索引
	log.Printf("第五步：创建索引...")
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_subscription_nodes_ip ON subscription_nodes (ip)").Error; err != nil {
		tx.Rollback()
		log.Fatalf("创建索引失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}

	// 验证更改
	log.Printf("验证更改...")
	var newDataType string
	if err := db.Raw(checkQuery).Scan(&newDataType).Error; err != nil {
		log.Fatalf("验证更改失败: %v", err)
	}

	log.Printf("✅ 迁移完成！")
	log.Printf("IP字段类型已从 %s 更改为 %s", dataType, newDataType)

	// 显示受影响的记录数
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM subscription_nodes").Scan(&count).Error; err != nil {
		log.Printf("⚠️  无法获取记录数: %v", err)
	} else {
		log.Printf("受影响的记录数: %d", count)
	}

	fmt.Printf("\n=== 迁移成功完成 ===\n")
	fmt.Printf("现在可以在节点配置中使用域名作为IP字段的值了。\n")
	fmt.Printf("例如：'visa.com' 或 'example.com'\n")
}
