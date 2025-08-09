package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/xvv6u577/logv2fs/database"
)

// SupabaseListener PostgreSQL LISTEN/NOTIFY 监听器
type SupabaseListener struct {
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSupabaseListener 创建新的 PostgreSQL 监听器
func NewSupabaseListener() *SupabaseListener {
	ctx, cancel := context.WithCancel(context.Background())

	// 获取 PostgreSQL 数据库连接
	pgDB := database.GetPostgresDB()
	if pgDB == nil {
		log.Println("PostgreSQL 数据库连接不可用")
		return &SupabaseListener{
			db:     nil,
			ctx:    ctx,
			cancel: cancel,
		}
	}

	// 获取底层 sql.DB 连接
	sqlDB, err := pgDB.DB()
	if err != nil {
		log.Printf("获取 PostgreSQL 底层连接失败: %v", err)
		return &SupabaseListener{
			db:     nil,
			ctx:    ctx,
			cancel: cancel,
		}
	}

	return &SupabaseListener{
		db:     sqlDB,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动 PostgreSQL LISTEN/NOTIFY 监听
func (l *SupabaseListener) Start() {
	if l.db == nil {
		log.Println("PostgreSQL 数据库未初始化，跳过 LISTEN/NOTIFY 监听")
		return
	}

	log.Println("启动 PostgreSQL LISTEN/NOTIFY 监听器...")

	// 监听用户表
	go l.watchTable("user_traffic_update", "user_traffic_update")

	// 监听流量日志表
	go l.watchTable("node_traffic_update", "node_traffic_update")

	// 监听订阅节点表
	go l.watchTable("subscription_node_update", "subscription_node_update")

	// 监听缴费记录表
	go l.watchTable("payment_update", "payment_update")
}

// Stop 停止监听
func (l *SupabaseListener) Stop() {
	log.Println("停止 PostgreSQL LISTEN/NOTIFY 监听器...")
	l.cancel()
}

// watchTable 监听指定通知通道
func (l *SupabaseListener) watchTable(channelName, messageType string) {
	// 创建 LISTEN 连接
	listener := pq.NewListener("", 10*time.Second, time.Minute, nil)
	defer listener.Close()

	// 监听通知通道
	err := listener.Listen(channelName)
	if err != nil {
		log.Printf("监听 PostgreSQL 通道失败 [%s]: %v", channelName, err)
		return
	}

	log.Printf("开始监听 PostgreSQL 通道: %s", channelName)

	// 监听通知
	for {
		select {
		case notification := <-listener.Notify:
			if notification != nil {
				l.handleNotification(notification, messageType)
			}
		case <-l.ctx.Done():
			log.Printf("停止监听 PostgreSQL 通道: %s", channelName)
			return
		}
	}
}

// handleNotification 处理 PostgreSQL 通知
func (l *SupabaseListener) handleNotification(notification *pq.Notification, messageType string) {
	if notification.Extra == "" {
		log.Printf("收到空通知: %s", notification.Channel)
		return
	}

	// 解析通知数据
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(notification.Extra), &eventData); err != nil {
		log.Printf("解析 PostgreSQL 通知失败: %v", err)
		return
	}

	// 提取事件信息
	event, ok := eventData["event"].(string)
	if !ok {
		log.Printf("无法获取事件类型")
		return
	}

	// 提取记录数据
	recordData := eventData["data"]

	// 创建 WebSocket 消息
	msg := Message{
		Type:       messageType,
		Action:     l.mapEventType(event),
		Collection: eventData["table"].(string),
		Data:       recordData,
		Timestamp:  time.Now(),
	}

	// 根据消息类型决定广播范围
	switch messageType {
	case "user_traffic_update":
		// 用户流量更新向所有客户端广播
		GlobalHub.BroadcastMessage(msg)
	case "node_traffic_update":
		// 节点流量更新向所有客户端广播
		GlobalHub.BroadcastMessage(msg)
	case "subscription_node_update":
		// 订阅节点更新向所有客户端广播
		GlobalHub.BroadcastMessage(msg)
	case "payment_update":
		// 缴费记录更新向所有客户端广播
		GlobalHub.BroadcastMessage(msg)
	default:
		// 默认向所有客户端广播
		GlobalHub.BroadcastMessage(msg)
	}

	log.Printf("PostgreSQL 通知事件: %s - %s", messageType, event)
}

// mapEventType 映射 PostgreSQL 事件类型到我们的操作类型
func (l *SupabaseListener) mapEventType(eventType string) string {
	switch eventType {
	case "INSERT":
		return "insert"
	case "UPDATE":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return eventType
	}
}

// 全局 PostgreSQL 监听器实例
var SupabaseListenerInstance *SupabaseListener

// InitSupabaseListener 初始化 PostgreSQL 监听器
func InitSupabaseListener() {
	SupabaseListenerInstance = NewSupabaseListener()
	SupabaseListenerInstance.Start()
}

// StopSupabaseListener 停止 PostgreSQL 监听器
func StopSupabaseListener() {
	if SupabaseListenerInstance != nil {
		SupabaseListenerInstance.Stop()
	}
}
