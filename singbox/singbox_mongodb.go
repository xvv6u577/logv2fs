package singbox

import (
	"context"
	"log"
	"sync"

	"github.com/sagernet/sing-box/option"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	UserTrafficLogs = model.UserTrafficLogs
)

// UpdateOptionsFromMongoDB 从 MongoDB 数据库更新配置（公共函数）
func UpdateOptionsFromMongoDB(opt option.Options) (option.Options, error) {
	log.Println("使用 MongoDB 从数据库更新 sing-box 配置...")

	var projections = bson.D{
		{Key: "email_as_id", Value: 1},
		{Key: "status", Value: 1},
		{Key: "uuid", Value: 1},
		{Key: "user_id", Value: 1},
	}

	cur, err := database.GetCollection(model.UserTrafficLogs{}).Find(context.Background(), bson.D{}, options.Find().SetProjection(projections))
	if err != nil {
		log.Printf("error getting all users portion info: %v\n", err)
		return opt, err
	}

	var userTrafficLogsArr []*UserTrafficLogs
	if err = cur.All(context.Background(), &userTrafficLogsArr); err != nil {
		log.Printf("error getting all users portion info: %v\n", err)
		return opt, err
	}

	if len(userTrafficLogsArr) > 0 {
		var wg sync.WaitGroup

		for _, user := range userTrafficLogsArr {
			if user.Status == "plain" {
				wg.Add(1)
				go func(user UserTrafficLogs) {
					defer wg.Done()

					// add VlessUser and Hysteria2User to opt.Inbounds
					for inbound := range opt.Inbounds {

						var usersToAppend = []string{user.Email_As_Id + "-reality", user.Email_As_Id + "-hysteria2"}
						opt.Experimental.V2RayAPI.Stats.Users = append(opt.Experimental.V2RayAPI.Stats.Users, usersToAppend...)

						if opt.Inbounds[inbound].Type == "vless" {
							opt.Inbounds[inbound].VLESSOptions.Users = append(opt.Inbounds[inbound].VLESSOptions.Users, option.VLESSUser{
								Name: user.Email_As_Id + "-reality",
								UUID: user.UUID,
								Flow: "xtls-rprx-vision",
							})
						}

						if opt.Inbounds[inbound].Type == "hysteria2" {
							opt.Inbounds[inbound].Hysteria2Options.Users = append(opt.Inbounds[inbound].Hysteria2Options.Users, option.Hysteria2User{
								Name:     user.Email_As_Id + "-hysteria2",
								Password: user.User_id,
							})
						}
					}

				}(*user)
			}
		}

		wg.Wait()
		log.Printf("成功从 MongoDB 加载了 %d 个用户的配置", len(userTrafficLogsArr))
	} else {
		log.Println("MongoDB 中没有找到活跃用户")
	}

	return opt, nil
}
