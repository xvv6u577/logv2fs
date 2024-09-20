package database

import (
	"context"
	"log"
	"time"

	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	User = model.User
)

func GetUserByName(name string, projections bson.D) (User, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user User
	filter := bson.D{
		primitive.E{Key: "email", Value: helper.SanitizeStr(name)},
	}
	opts := options.FindOne().SetProjection(projections)

	err := OpenCollection(Client, "USERS").FindOne(ctx, filter, opts).Decode(&user)
	if err != nil {
		log.Printf("error occured while finding user %s", helper.SanitizeStr(name))
		return user, err
	}

	return user, nil
}
