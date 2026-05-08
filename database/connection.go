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

	return client
}

// function to get the MongoDB client
func GetMongoDBClient() *mongo.Client {
	if Client == nil {
		Client = DBinstance()
	}
	return Client
}

// Client Database instance
// var Client *mongo.Client = DBinstance()

// OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = GetMongoDBClient().Database("logV2rayTrafficDB").Collection(collectionName)

	return collection
}

// OpenCollectionByModel 通过模型获取MongoDB集合，使用模型的CollectionName方法
func OpenCollectionByModel(client *mongo.Client, model CollectionNamer) *mongo.Collection {
	collectionName := model.CollectionName()
	return GetMongoDBClient().Database("logV2rayTrafficDB").Collection(collectionName)
}

// GetCollection 获取指定模型的MongoDB集合的便捷方法
func GetCollection(model CollectionNamer) *mongo.Collection {
	return OpenCollectionByModel(GetMongoDBClient(), model)
}
