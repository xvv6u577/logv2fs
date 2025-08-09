package websocket

import (
	"context"
	"log"
	"time"

	"github.com/xvv6u577/logv2fs/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBListener MongoDB 变更监听器
type MongoDBListener struct {
	client *mongo.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMongoDBListener 创建新的 MongoDB 监听器
func NewMongoDBListener() *MongoDBListener {
	ctx, cancel := context.WithCancel(context.Background())
	return &MongoDBListener{
		client: database.Client,
		ctx:    ctx,
		cancel: cancel,
	}
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
	l.cancel()
}

// watchCollection 监听指定集合的变更
func (l *MongoDBListener) watchCollection(collectionName, messageType string) {
	collection := l.client.Database("logV2rayTrafficDB").Collection(collectionName)

	// 创建 Change Stream 选项
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{
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

	log.Printf("MongoDB 变更事件: %s - %s", messageType, operationType)
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
