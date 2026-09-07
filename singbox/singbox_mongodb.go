package singbox

import (
	"context"
	"log"

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
		for _, user := range userTrafficLogsArr {
			if user.Status != "plain" {
				continue
			}

			// sing-box 1.12 stores protocol-specific fields in Inbound.Options.
			for inboundIndex := range opt.Inbounds {
				switch inboundOptions := opt.Inbounds[inboundIndex].Options.(type) {
				case *option.VLESSInboundOptions:
					inboundOptions.Users = append(inboundOptions.Users, option.VLESSUser{
						Name: user.Email_As_Id + "-reality",
						UUID: user.UUID,
						Flow: "xtls-rprx-vision",
					})
				case *option.Hysteria2InboundOptions:
					inboundOptions.Users = append(inboundOptions.Users, option.Hysteria2User{
						Name:     user.Email_As_Id + "-hysteria2",
						Password: user.User_id,
					})
				case *option.AnyTLSInboundOptions:
					inboundOptions.Users = append(inboundOptions.Users, option.AnyTLSUser{
						Name:     user.Email_As_Id + "-anytls",
						Password: user.User_id,
					})
				}
			}

			if opt.Experimental != nil && opt.Experimental.V2RayAPI != nil && opt.Experimental.V2RayAPI.Stats != nil {
				opt.Experimental.V2RayAPI.Stats.Users = append(
					opt.Experimental.V2RayAPI.Stats.Users,
					user.Email_As_Id+"-reality",
					user.Email_As_Id+"-hysteria2",
					user.Email_As_Id+"-anytls",
				)
			}
		}
		log.Printf("成功从 MongoDB 加载了 %d 个用户的配置", len(userTrafficLogsArr))
	} else {
		log.Println("MongoDB 中没有找到活跃用户")
	}

	return opt, nil
}
