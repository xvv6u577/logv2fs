/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// UserExport 导出用户的结构体
type UserExport struct {
	EmailAsId string `json:"email_as_id" bson:"email_as_id"`
	Name      string `json:"name" bson:"name"`
}

// exportuserasjsonCmd represents the exportuserasjson command
var exportuserasjsonCmd = &cobra.Command{
	Use:   "exportuserasjson",
	Short: "导出活跃用户的邮箱ID和姓名到JSON文件",
	Long: `从MongoDB的USER_TRAFFIC_LOGS集合中导出活跃用户（status=plain）的数据。
只包含email_as_id和name字段，如果name为空则使用email_as_id代替。
导出结果保存为JSON数组格式到users_in_db.json文件中。

示例:
  logv2fs exportuserasjson

前提条件:
  - MongoDB服务必须运行在 mongodb://localhost:27017 (或.env中配置的地址)
  - 数据库名称: logV2rayTrafficDB
  - 集合名称: USER_TRAFFIC_LOGS`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := exportUsersAsJSON(); err != nil {
			log.Fatalf("❌ 导出用户数据失败: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportuserasjsonCmd)
}

// exportUsersAsJSON 实现导出用户数据的核心逻辑
func exportUsersAsJSON() error {
	log.Println("🚀 开始导出活跃用户数据...")

	// 检查数据库连接
	log.Println("🔗 检查MongoDB连接...")
	if database.Client == nil {
		return fmt.Errorf("MongoDB客户端未初始化，请检查.env文件中的mongoURI配置")
	}

	// 测试数据库连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := database.Client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("无法连接到MongoDB，请确保MongoDB服务正在运行: %v", err)
	}
	log.Println("✅ MongoDB连接成功")

	// 获取MongoDB集合
	collection := database.GetCollection(model.UserTrafficLogs{})

	// 设置查询上下文超时
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 查询条件：只查询活跃用户
	filter := bson.M{"status": "plain"}

	// 投影：只选择需要的字段
	projection := bson.M{
		"email_as_id": 1,
		"name":        1,
		"_id":         0, // 排除_id字段
	}

	// 查询选项
	opts := options.Find().SetProjection(projection)

	// 执行查询
	log.Println("📡 正在查询MongoDB数据...")
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return fmt.Errorf("查询用户数据失败: %v", err)
	}
	defer cursor.Close(ctx)

	// 解析查询结果
	var users []UserExport
	for cursor.Next(ctx) {
		var user UserExport
		if err := cursor.Decode(&user); err != nil {
			log.Printf("⚠️  解析用户数据失败: %v", err)
			continue
		}

		// 处理name为空的情况：使用email_as_id代替
		if user.Name == "" {
			user.Name = user.EmailAsId
		}

		users = append(users, user)
	}

	// 检查游标错误
	if err := cursor.Err(); err != nil {
		return fmt.Errorf("遍历查询结果时出错: %v", err)
	}

	log.Printf("✅ 成功查询到 %d 个活跃用户", len(users))

	// 将数据转换为JSON
	log.Println("🔄 正在生成JSON数据...")
	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("转换JSON失败: %v", err)
	}

	// 保存到文件
	filename := "users_in_db.json"
	log.Printf("💾 正在保存到文件: %s", filename)

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	log.Printf("🎉 导出完成！")
	log.Printf("📁 文件位置: %s", filename)
	log.Printf("📊 导出用户数量: %d", len(users))

	// 显示前几个用户作为预览
	if len(users) > 0 {
		log.Println("📋 数据预览:")
		previewCount := 3
		if len(users) < previewCount {
			previewCount = len(users)
		}

		for i := 0; i < previewCount; i++ {
			log.Printf("  - %s (名称: %s)", users[i].EmailAsId, users[i].Name)
		}

		if len(users) > previewCount {
			log.Printf("  ... 还有 %d 个用户", len(users)-previewCount)
		}
	}

	return nil
}
