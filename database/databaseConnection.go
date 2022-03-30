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

//DBinstance func
func DBinstance() *mongo.Client {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := godotenv.Load(".env"); err != nil {
		log.Panic("Error loading .env file")
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

//Client Database instance
var Client *mongo.Client = DBinstance()

//OpenCollection is a  function makes a connection with a collection in the database
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {

	var collection *mongo.Collection = client.Database("logV2rayTrafficDB").Collection(collectionName)

	return collection
}
