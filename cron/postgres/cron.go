package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/robfig/cron"
	box "github.com/sagernet/sing-box"
	"github.com/xvv6u577/logv2fs/database/postgres"
	postgres_pkg "github.com/xvv6u577/logv2fs/pkg/postgres"
)

var (
	currentDomainPG = os.Getenv("CURRENT_DOMAIN")
)

// UserTrafficRequest 定义调用 upsert_user_traffic_log 函数的请求参数
type UserTrafficRequest struct {
	Email     string    `json:"p_email"`
	Timestamp time.Time `json:"p_timestamp"`
	Traffic   int64     `json:"p_traffic"`
}

// NodeTrafficRequest 定义调用 upsert_node_traffic_log 函数的请求参数
type NodeTrafficRequest struct {
	Domain    string    `json:"p_domain"`
	Timestamp time.Time `json:"p_timestamp"`
	Traffic   int64     `json:"p_traffic"`
}

// LogUserTrafficPG PostgreSQL版本的用户流量记录函数
func LogUserTrafficPG(email string, timestamp time.Time, traffic int64) error {
	supaClient := postgres.GetSupabaseClient()
	if supaClient == nil {
		log.Printf("supabase 客户端初始化失败")
		return fmt.Errorf("supabase 客户端初始化失败")
	}

	ctx := context.Background()
	userRequest := UserTrafficRequest{
		Email:     email,
		Timestamp: timestamp,
		Traffic:   traffic,
	}

	rpcBuilder := supaClient.DB.RPC("upsert_user_traffic_log", userRequest)
	err := rpcBuilder.Execute(ctx, nil)
	if err != nil {
		log.Printf("用户流量记录 RPC 调用失败: %v", err)
		return err
	}

	log.Printf("用户流量记录成功 - 用户: %s, 流量: %d, 时间: %s",
		email, traffic, timestamp.Format("2006-01-02 15:04:05"))
	return nil
}

// LogNodeTrafficPG PostgreSQL版本的节点流量记录函数
func LogNodeTrafficPG(domain string, timestamp time.Time, traffic int64) error {
	supaClient := postgres.GetSupabaseClient()
	if supaClient == nil {
		log.Printf("Supabase 客户端初始化失败")
		return fmt.Errorf("supabase 客户端初始化失败")
	}

	ctx := context.Background()
	nodeRequest := NodeTrafficRequest{
		Domain:    domain,
		Timestamp: timestamp,
		Traffic:   traffic,
	}

	rpcBuilder := supaClient.DB.RPC("upsert_node_traffic_log", nodeRequest)
	err := rpcBuilder.Execute(ctx, nil)
	if err != nil {
		log.Printf("节点流量记录 RPC 调用失败: %v", err)
		return err
	}

	log.Printf("节点流量记录成功 - 节点: %s, 流量: %d, 时间: %s",
		domain, traffic, timestamp.Format("2006-01-02 15:04:05"))
	return nil
}

// Cron_loggingJobsPG PostgreSQL版本的定时任务
func Cron_loggingJobsPG(c *cron.Cron, instance *box.Box) {

	// cron job by 15 mins - PostgreSQL版本
	c.AddFunc("0 */15 * * * *", func() {

		timesteamp := time.Now().Local()
		usageData, err := postgres_pkg.UsageDataOfAll(instance)
		if err != nil {
			log.Printf("获取使用数据时出错: %v\n", err)
			return
		}

		if len(usageData) == 0 {
			log.Printf("没有流量数据需要记录: %v", timesteamp.Format("20060102 15:04:05"))
			return
		}

		log.Printf("使用PostgreSQL记录流量数据...")
		// 使用 Supabase RPC 调用方式记录流量
		for _, perUser := range usageData {

			// 记录用户流量
			if err := LogUserTrafficPG(perUser.Name, timesteamp, perUser.Total); err != nil {
				log.Printf("PostgreSQL用户流量记录失败: %v\n", err)
			}

			// 记录节点流量
			if err := LogNodeTrafficPG(currentDomainPG, timesteamp, perUser.Total); err != nil {
				log.Printf("PostgreSQL节点流量记录失败: %v\n", err)
			}
		}
		log.Printf("PostgreSQL流量记录完成: %v 用户=%d", timesteamp.Format("20060102 15:04:05"), len(usageData))

	})

}
