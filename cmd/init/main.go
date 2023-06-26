package main

import (
	"context"
	"fmt"
	"time"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// trafficCollection *mongo.Collection = database.OpenCollection(database.Client, "TRAFFIC")
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "USERS")
)

type (
	User = model.User
	// TrafficInDB = model.TrafficInDB
)

func main() {

	var ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var projections = bson.D{
		{Key: "email", Value: 1},
		{Key: "node_in_use_status", Value: 1},
	}
	users, err := database.GetPartialInfosForAllUsers(projections)
	if err != nil {
		panic(err)
	}

	// set all node status to true except w8.undervineayrd.com and localhost. then update the database.
	for _, user := range users {
		for node, _ := range user.NodeInUseStatus {
			if node == "w8.undervineyard.com" || node == "localhost" {
				continue
			}
			user.NodeInUseStatus[node] = true
		}

		fmt.Printf("user: %v\n", user.NodeInUseStatus)

		var filter = bson.D{
			{Key: "email", Value: user.Email},
		}

		var update = bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "node_in_use_status", Value: user.NodeInUseStatus},
			}},
		}

		updateResult, err := userCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Matched %v document(s) and updated %v document(s)\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	}

}
