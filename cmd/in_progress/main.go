package main

import (
	"context"
	"time"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userCollection *mongo.Collection = database.OpenCollection(database.Client, "USERS")
)

type (
	// User = model.User
	GlobalVariable = model.GlobalVariable
)

func main() {

	var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var global = GlobalVariable{
		Name:       "GLOBAL",
		DomainList: map[string]string{},
	}

	// insert into database
	_, err := userCollection.InsertOne(ctx, global)
	if err != nil {
		panic(err)
	}

}
