package main

import (
	"context"
	"log"

	"github.com/caster8013/logv2rayfullstack/types"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cronInstance *cron.Cron
var collection *mongo.Collection
var ctx = context.TODO()

type User = types.User

func init() {

	cronInstance = cron.New()
	cronInstance.Start()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Panic(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Panic(err)
	}

	collection = client.Database("logV2rayTrafficDB").Collection("USERS")
}

func AddDBUserProperty() error {

	filter := bson.M{}
	update := bson.M{"$set": bson.M{"status": "plain"}}
	// update := bson.M{"$set": bson.M{"CreatedAt": time.Now(), "UpdatedAt": time.Now()}}
	_, err := collection.UpdateMany(ctx, filter, update)

	return err
}

func EmptyUsersInfoInDB() error {
	filter := bson.M{}
	_, error := collection.DeleteMany(ctx, filter)

	return error
}

func DeleteUsersDBs() error {

	users := []string{
		"yalin", "zhouyijun", "robmakelin", "jet", "xiaolan", "zikai", "zhu", "jiangbo", "chenyuanyuan", "wangning",
		"zhangxiaoxu", "liangying", "deliang", "jeff", "johnathonbai", "yumei", "7g", "jojo", "daibin", "deena", "xuyang", "alphaemma",
		"xiaohe", "bsclks", "joy", "sarah", "guowanyue", "baofeng", "jonah", "yuxiaofang", "cuixiaoli", "wangyakun", "pty", "wupeng", "xiangwei", "changhua",
		"weihongwei", "zhihu", "lujixiawu", "hepengfei", "mengchch", "21cpiaomu", "cuiyang", "bscdavid", "wangling", "21clsj", "anchagu", "bjbfl", "maylee",
		"frankw", "pansir", "yizhu", "huohuo", "chunxia", "caster", "yutou", "camel", "rongfan", "cannan", "wuqiong", "huidi", "zhaorui", "yanyong",
		"lijiaxin", "yongming", "jspotter", "haotian", "wrong", "sisi", "linbo", "bscalbert", "21caiqing", "shanshan", "bqgeorge",
	}

	for _, name := range users {

		clientOptions := options.Client().ApplyURI(mongoURI)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Panic(err)
		}

		collection = client.Database("logV2rayTrafficDB").Collection(name)

		collection.Drop(ctx)
	}

	return nil
}

// func UpdateDBUserByName(user *User) error {

// }

func CreateUserByName(user *User) error {

	var result bson.M
	filter := bson.D{primitive.E{Key: "email", Value: &user.Email}}

	err := collection.FindOneAndReplace(ctx, filter, user).Decode(&result)
	if err == mongo.ErrNoDocuments {
		_, err = collection.InsertOne(ctx, user)
		if err != nil {
			return err
		}
	}

	cliOpt := options.Client().ApplyURI(mongoURI)
	cli, err := mongo.Connect(ctx, cliOpt)
	if err != nil {
		log.Panic(err)
	}

	err = cli.Database("logV2rayTrafficDB").CreateCollection(ctx, user.Email)
	return err
}

func DeleteUserByName(name string) error {
	filter := bson.D{primitive.E{Key: "email", Value: name}}
	update := bson.M{"$set": bson.M{"status": "deleted"}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func GetAllUsersInfo() ([]*User, error) {
	filter := bson.D{{}}
	return FilterUsers(filter)
}

func GetUserByName(name string) ([]*User, error) {
	filter := bson.D{
		primitive.E{Key: "email", Value: name},
	}

	return FilterUsers(filter)
}

func FilterUsers(filter interface{}) ([]*User, error) {

	var users []*User

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
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

	// once exhausted, close the cursor
	cursor.Close(ctx)

	if len(users) == 0 {
		return users, mongo.ErrNoDocuments
	}

	return users, nil

}
