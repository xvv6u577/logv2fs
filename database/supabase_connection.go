package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	supabase "github.com/lengzuo/supa"
)

var SupabaseClient *supabase.Client

// InitSupabase 初始化 Supabase 客户端
func InitSupabase() (*supabase.Client, error) {
	// 加载环境变量
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("获取工作目录失败: %v", err)
	}

	if err := godotenv.Load(pwd + "/.env"); err != nil {
		log.Printf("警告: 无法加载.env文件: %v", err)
	}

	// 从环境变量获取 Supabase 配置
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		return nil, fmt.Errorf("环境变量 SUPABASE_URL 或 SUPABASE_KEY 未设置")
	}

	// 创建 Supabase 客户端配置
	config := supabase.Config{
		ApiKey:     supabaseKey,
		ProjectRef: extractProjectRef(supabaseURL),
		Debug:      true,
	}

	// 创建 Supabase 客户端
	client, err := supabase.New(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Supabase 客户端失败: %v", err)
	}

	log.Println("Supabase 客户端连接成功")
	SupabaseClient = client

	return client, nil
}

// extractProjectRef 从 Supabase URL 中提取项目引用
func extractProjectRef(supabaseURL string) string {
	// 从 URL 中提取项目引用
	// 例如：https://abc123.supabase.co -> abc123
	if len(supabaseURL) > 8 && supabaseURL[:8] == "https://" {
		url := supabaseURL[8:]
		if dotIndex := findString(url, "."); dotIndex != -1 {
			return url[:dotIndex]
		}
	}
	return ""
}

// findString 查找字符串中的子字符串位置
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetSupabaseClient 获取 Supabase 客户端实例
func GetSupabaseClient() *supabase.Client {
	if SupabaseClient == nil {
		client, err := InitSupabase()
		if err != nil {
			log.Printf("初始化 Supabase 客户端失败: %v", err)
			return nil
		}
		SupabaseClient = client
	}
	return SupabaseClient
}

// CloseSupabase 关闭 Supabase 连接
func CloseSupabase() {
	if SupabaseClient != nil {
		log.Println("Supabase 客户端连接已关闭")
		SupabaseClient = nil
	}
}
