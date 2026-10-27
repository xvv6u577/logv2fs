package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// 定义集合名称接口
type CollectionNamer interface {
	CollectionName() string
}

var Client *mongo.Client

func getMongoDBURI() string {
	return os.Getenv("mongoURI")
}

// DBinstance func
func DBinstance() *mongo.Client {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pwd, err := os.Getwd()
	if err != nil {
		log.Panic("Panic: ", err)
	}

	if err := godotenv.Load(pwd + "/.env"); err != nil {
		log.Panicf("Error loading .env file: %v", err)
	}
	MongoDB := getMongoDBURI()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Panic(err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Panic(err)
	}

	log.Println("MongoDB successfully connected and pinged.")

	// 首次连接成功后，立即确保关键集合的唯一索引存在。
	// 拆出周期表后这一步是数据一致性的硬性前提（防止重复 (owner, kind, period)）。
	ensureCoreIndexes(ctx, client)

	return client
}

// ensureCoreIndexes 在 MongoDB 客户端首次建立后保证核心索引存在。
//
// 当前涵盖：
//   - user_traffic_periods: 唯一复合索引 (email_as_id, kind, period)
//   - node_traffic_periods: 唯一复合索引 (domain_as_id, kind, period)
//
// 索引已存在时 CreateOne 是幂等的；任何失败都只打日志而不中断启动，
// 以便在权限受限的运行环境下也能降级跑起来。
func ensureCoreIndexes(ctx context.Context, client *mongo.Client) {
	db := client.Database("logV2rayTrafficDB")

	// 周期表的 (owner, kind, period) 唯一索引
	uniqueOwnerJobs := []struct {
		collection string
		ownerKey   string
	}{
		{"user_traffic_periods", "email_as_id"},
		{"node_traffic_periods", "domain_as_id"},
	}

	for _, job := range uniqueOwnerJobs {
		coll := db.Collection(job.collection)
		idx := mongo.IndexModel{
			Keys: bson.D{
				{Key: job.ownerKey, Value: 1},
				{Key: "kind", Value: 1},
				{Key: "period", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("uniq_owner_kind_period"),
		}
		if _, err := coll.Indexes().CreateOne(ctx, idx); err != nil {
			log.Printf("ensureCoreIndexes 警告: 为 %s 创建唯一索引失败: %v", job.collection, err)
			continue
		}
		log.Printf("ensureCoreIndexes: %s 唯一索引就绪 (%s, kind, period)", job.collection, job.ownerKey)
	}

	// 用户别名多键索引：订阅/登录按旧 email_as_id 查找
	userCol := db.Collection("USER_TRAFFIC_LOGS")
	aliasIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "email_as_id_aliases", Value: 1}},
		Options: options.Index().SetName("idx_email_as_id_aliases"),
	}
	if _, err := userCol.Indexes().CreateOne(ctx, aliasIdx); err != nil {
		log.Printf("ensureCoreIndexes 警告: email_as_id_aliases 索引失败: %v", err)
	} else {
		log.Printf("ensureCoreIndexes: USER_TRAFFIC_LOGS.email_as_id_aliases 索引就绪")
	}
}

// function to get the MongoDB client
func GetMongoDBClient() *mongo.Client {
	if Client == nil {
		Client = DBinstance()
	}
	return Client
}

// GetCollection 通过模型的 CollectionName 获取 MongoDB 集合。
func GetCollection(model CollectionNamer) *mongo.Collection {
	return GetMongoDBClient().
		Database("logV2rayTrafficDB").
		Collection(model.CollectionName())
}
