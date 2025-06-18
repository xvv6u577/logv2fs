package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

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
	MongoDB := os.Getenv("mongoURI")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Panic(err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Panic(err)
	}

	// fmt.Println("MongoDB successfully connected and pinged.")
	log.Println("MongoDB successfully connected and pinged.")

	return client
}

// Client Database instance
var Client *mongo.Client = DBinstance()

// OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("logV2rayTrafficDB").Collection(collectionName)

	return collection
}

// GetDB 获取数据库连接，优先返回PostgreSQL连接，如果不可用则返回MongoDB连接
func GetDB() interface{} {
	// 尝试获取PostgreSQL连接
	pgDB := GetPostgresDB()
	if pgDB != nil {
		return pgDB
	}

	// 如果PostgreSQL不可用，返回MongoDB连接
	return Client
}

// IsUsingPostgres 检查是否使用PostgreSQL
func IsUsingPostgres() bool {
	// 从环境变量中读取配置
	usePostgres := os.Getenv("USE_POSTGRES")
	return usePostgres == "true" || usePostgres == "1" || usePostgres == "yes"
}
