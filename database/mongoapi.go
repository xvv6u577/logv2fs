package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/caster8013/logv2rayfullstack/model"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User = model.User

var validate = validator.New()

func AddDBUserProperty() error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{}
	update := bson.M{"$set": bson.M{"status": "plain"}}
	_, err := OpenCollection(Client, "USERS").UpdateMany(ctx, filter, update)

	return err
}

func DelUsersInfo() error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{}
	_, error := OpenCollection(Client, "USERS").DeleteMany(ctx, filter)
	if error != nil {
		fmt.Printf("%v\n", error)
		return error
	}

	return nil
}

func DelUsersTable() error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	users, err := GetAllUsersInfo()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}

	for _, ele := range users {
		OpenCollection(Client, ele.Email).Drop(ctx)
	}

	return nil
}

func DeleteUserByName(email string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.D{primitive.E{Key: "email", Value: email}}
	_, error := OpenCollection(Client, "USERS").DeleteOne(ctx, filter)
	if error != nil {
		log.Printf("error occured while deleting user %s", email)
		return error
	}

	Client.Database("logV2rayTrafficDB").Collection(email).Drop(ctx)

	return nil
}

func CreateUserByName(user *User) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	validationErr := validate.Struct(user)
	if validationErr != nil {
		log.Printf("error occured while validating user %s", user.Email)
		return validationErr
	}

	count, err := OpenCollection(Client, "USERS").CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		log.Printf("error occured while counting user %v", user.Email)
		return err
	}

	if count > 0 {
		log.Printf("user %s already exists", user.Email)
		return errors.New("this email already exists")
	}

	_, err = OpenCollection(Client, "USERS").InsertOne(ctx, user)
	if err != nil {
		log.Printf("error occured while inserting user %v", user.Email)
		return err
	}

	err = Client.Database("logV2rayTrafficDB").CreateCollection(ctx, user.Email)
	return err
}

func UpdateUserStatusByName(name string, status string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.D{primitive.E{Key: "email", Value: name}}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err := OpenCollection(Client, "USERS").UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("error occured while updating user %s", name)
		return err
	}

	return nil
}

func GetAllUsersInfo() ([]*User, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var users []*User
	var filter = bson.D{{}}
	var projections = bson.D{
		{Key: "_id", Value: 0},
		{Key: "token", Value: 0},
		{Key: "password", Value: 0},
		{Key: "refresh_token", Value: 0},
		{Key: "user_id", Value: 0},
	}

	cursor, err := OpenCollection(Client, "USERS").Find(ctx, filter, options.Find().SetProjection(projections))
	if err != nil {
		log.Printf("error occured while finding users")
		return users, err
	}

	for cursor.Next(ctx) {
		var t User
		err := cursor.Decode(&t)
		if err != nil {
			return users, err
		}

		users = append(users, &t)
	}

	if err := cursor.Err(); err != nil {
		return users, err
	}
	cursor.Close(ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil
}

func GetUserByName(name string, projections bson.D) (User, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user User
	filter := bson.D{
		primitive.E{Key: "email", Value: name},
	}
	opts := options.FindOne().SetProjection(projections)

	err := OpenCollection(Client, "USERS").FindOne(ctx, filter, opts).Decode(&user)
	if err != nil {
		log.Printf("error occured while finding user %s", name)
		return user, err
	}

	return user, nil
}
