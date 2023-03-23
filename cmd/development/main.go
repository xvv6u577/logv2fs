package main

import (
	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	"go.mongodb.org/mongo-driver/bson"
)

var (
// userCollection *mongo.Collection = database.OpenCollection(database.Client, "USERS")
// err            error
)

type (
	User = model.User
)

func main() {
	// var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	// defer cancel()

	var projections = bson.D{
		{Key: "_id", Value: 0},
		{Key: "token", Value: 0},
		{Key: "password", Value: 0},
		{Key: "refresh_token", Value: 0},
		{Key: "used_by_current_day", Value: 0},
		{Key: "used_by_current_month", Value: 0},
		{Key: "used_by_current_year", Value: 0},
		{Key: "traffic_by_day", Value: 0},
		{Key: "traffic_by_month", Value: 0},
		{Key: "traffic_by_year", Value: 0},
		{Key: "suburl", Value: 0},
	}
	allUsersInDB, _ := database.GetPartialInfosForAllUsers(projections)

	for _, user := range allUsersInDB {

		println("Updated user: " + user.Email)
	}
}
