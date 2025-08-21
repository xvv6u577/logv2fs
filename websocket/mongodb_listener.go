package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/xvv6u577/logv2fs/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventBatch 事件批次结构
type EventBatch struct {
	Messages []Message `json:"messages"`
	Count    int       `json:"count"`
}

// MongoDBListener MongoDB 变更监听器
type MongoDBListener struct {
	client         *mongo.Client
	ctx            context.Context
	cancel         context.CancelFunc
	eventQueue     []Message
	eventMutex     sync.RWMutex
	batchTicker    *time.Ticker
	lastBatchTime  time.Time
}

// NewMongoDBListener 创建新的 MongoDB 监听器
func NewMongoDBListener() *MongoDBListener {
	ctx, cancel := context.WithCancel(context.Background())
	listener := &MongoDBListener{
		client:        database.Client,
		ctx:           ctx,
		cancel:        cancel,
		eventQueue:    make([]Message, 0),
		batchTicker:   time.NewTicker(5 * time.Second),
		lastBatchTime: time.Now(),
	}
	
	// 启动批量处理协程
	go listener.processBatchEvents()
	
	return listener
}

// Start 启动 MongoDB 变更监听
func (l *MongoDBListener) Start() {
	log.Println("启动 MongoDB Change Streams 监听器...")

	// 监听用户集合
	go l.watchCollection("USER_TRAFFIC_LOGS", "user_traffic_update")

	// 监听流量日志集合
	go l.watchCollection("NODE_TRAFFIC_LOGS", "node_traffic_update")

	// 监听订阅节点集合（如果存在）
	go l.watchCollection("subscription_nodes", "subscription_node_update")

	// 监听其他相关集合
	go l.watchCollection("payment_records", "payment_update")
}

// Stop 停止监听
func (l *MongoDBListener) Stop() {
	log.Println("停止 MongoDB Change Streams 监听器...")
	if l.batchTicker != nil {
		l.batchTicker.Stop()
	}
	
	// 发送剩余的事件
	l.flushPendingEvents()
	l.cancel()
}

// watchCollection 监听指定集合的变更
func (l *MongoDBListener) watchCollection(collectionName, messageType string) {
	collection := l.client.Database("logV2rayTrafficDB").Collection(collectionName)

	// 创建 Change Stream 选项
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"operationType": bson.M{
				"$in": []string{"insert", "update", "delete", "replace"},
			},
		}}},
	}

	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	// 创建 Change Stream
	changeStream, err := collection.Watch(l.ctx, pipeline, opts)
	if err != nil {
		log.Printf("创建 Change Stream 失败 [%s]: %v", collectionName, err)
		return
	}
	defer changeStream.Close(l.ctx)

	log.Printf("开始监听集合: %s", collectionName)

	// 监听变更事件
	for changeStream.Next(l.ctx) {
		var changeEvent bson.M
		if err := changeStream.Decode(&changeEvent); err != nil {
			log.Printf("解码变更事件失败: %v", err)
			continue
		}

		// 处理变更事件
		l.handleChangeEvent(changeEvent, messageType)
	}

	if err := changeStream.Err(); err != nil {
		log.Printf("Change Stream 错误 [%s]: %v", collectionName, err)
	}
}

// handleChangeEvent 处理变更事件
func (l *MongoDBListener) handleChangeEvent(changeEvent bson.M, messageType string) {
	operationType, ok := changeEvent["operationType"].(string)
	if !ok {
		return
	}

	// 提取文档数据
	var documentData interface{}

	switch operationType {
	case "insert", "replace":
		if fullDocument, exists := changeEvent["fullDocument"]; exists {
			documentData = fullDocument
		}
	case "update":
		if fullDocument, exists := changeEvent["fullDocument"]; exists {
			documentData = fullDocument
		} else if updateDescription, exists := changeEvent["updateDescription"]; exists {
			// 对于更新操作，如果没有完整文档，使用更新描述
			documentData = bson.M{
				"documentKey":       changeEvent["documentKey"],
				"updateDescription": updateDescription,
			}
		}
	case "delete":
		if documentKey, exists := changeEvent["documentKey"]; exists {
			documentData = bson.M{
				"documentKey": documentKey,
			}
		}
	}

	// 创建 WebSocket 消息
	msg := Message{
		Type:       messageType,
		Action:     operationType,
		Collection: changeEvent["ns"].(bson.M)["coll"].(string),
		Data:       documentData,
		Timestamp:  time.Now(),
	}

	// 将事件添加到队列中，而不是立即广播
	l.addEventToQueue(msg)

	log.Printf("MongoDB 变更事件已加入队列: %s - %s", messageType, operationType)
}

// HandleChangeEvent 公开方法用于测试
func (l *MongoDBListener) HandleChangeEvent(changeEvent bson.M, messageType string) {
	l.handleChangeEvent(changeEvent, messageType)
}

// addEventToQueue 将事件添加到队列中
func (l *MongoDBListener) addEventToQueue(msg Message) {
	l.eventMutex.Lock()
	defer l.eventMutex.Unlock()
	
	l.eventQueue = append(l.eventQueue, msg)
}

// processBatchEvents 批量处理事件
func (l *MongoDBListener) processBatchEvents() {
	for {
		select {
		case <-l.batchTicker.C:
			l.flushPendingEvents()
		case <-l.ctx.Done():
			return
		}
	}
}

// flushPendingEvents 发送待处理的事件
func (l *MongoDBListener) flushPendingEvents() {
	l.eventMutex.Lock()
	defer l.eventMutex.Unlock()
	
	if len(l.eventQueue) == 0 {
		return
	}
	
	// 创建批量消息
	batchMsg := Message{
		Type:      "batch_update",
		Action:    "batch",
		Data: EventBatch{
			Messages: l.eventQueue,
			Count:    len(l.eventQueue),
		},
		Timestamp: time.Now(),
	}
	
	// 广播批量消息
	GlobalHub.BroadcastMessage(batchMsg)
	
	log.Printf("批量发送 %d 个事件", len(l.eventQueue))
	
	// 清空队列
	l.eventQueue = l.eventQueue[:0]
	l.lastBatchTime = time.Now()
}

// 全局 MongoDB 监听器实例
var MongoDBListenerInstance *MongoDBListener

// InitMongoDBListener 初始化 MongoDB 监听器
func InitMongoDBListener() {
	MongoDBListenerInstance = NewMongoDBListener()
	MongoDBListenerInstance.Start()
}

// StopMongoDBListener 停止 MongoDB 监听器
func StopMongoDBListener() {
	if MongoDBListenerInstance != nil {
		MongoDBListenerInstance.Stop()
	}
}
