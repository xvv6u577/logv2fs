package thirdparty

import (
	"context"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/experimental/v2rayapi"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Traffic         = model.Traffic
	UserTrafficLogs = model.UserTrafficLogs
	// PostgreSQL 类型别名
	UserTrafficLogsPG = model.UserTrafficLogsPG
)

var (
	CURRENT_DOMAIN     = os.Getenv("CURRENT_DOMAIN")
	userTrafficLogsCol = database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")
)

// UpdateOptionsFromDB 根据配置从数据库更新 sing-box 选项
// 支持 MongoDB 和 PostgreSQL 两种数据库
func UpdateOptionsFromDB(opt option.Options) (option.Options, error) {
	// 根据环境变量决定使用哪种数据库
	if database.IsUsingPostgres() {
		log.Println("使用 PostgreSQL 从数据库更新 sing-box 配置...")
		return updateOptionsFromPostgreSQL(opt)
	} else {
		log.Println("使用 MongoDB 从数据库更新 sing-box 配置...")
		return updateOptionsFromMongoDB(opt)
	}
}

// updateOptionsFromPostgreSQL 从 PostgreSQL 数据库更新配置
func updateOptionsFromPostgreSQL(opt option.Options) (option.Options, error) {
	db := database.GetPostgresDB()
	if db == nil {
		log.Printf("PostgreSQL 数据库连接不可用，尝试回退到 MongoDB")
		return updateOptionsFromMongoDB(opt)
	}

	// 查询活跃用户的关键信息
	var pgUsers []UserTrafficLogsPG
	if err := db.Select("email_as_id, status, uuid, user_id").
		Where("status = ?", "plain").
		Find(&pgUsers).Error; err != nil {
		log.Printf("查询 PostgreSQL 用户信息时出错: %v\n", err)
		return opt, err
	}

	if len(pgUsers) > 0 {
		var wg sync.WaitGroup

		for _, user := range pgUsers {
			wg.Add(1)
			go func(user UserTrafficLogsPG) {
				defer wg.Done()

				// 为每个用户添加 VlessUser 和 Hysteria2User 到 opt.Inbounds
				for inbound := range opt.Inbounds {

					var usersToAppend = []string{user.EmailAsId + "-reality", user.EmailAsId + "-hysteria2"}
					opt.Experimental.V2RayAPI.Stats.Users = append(opt.Experimental.V2RayAPI.Stats.Users, usersToAppend...)

					if opt.Inbounds[inbound].Type == "vless" {
						opt.Inbounds[inbound].VLESSOptions.Users = append(opt.Inbounds[inbound].VLESSOptions.Users, option.VLESSUser{
							Name: user.EmailAsId + "-reality",
							UUID: user.UUID,
							Flow: "xtls-rprx-vision",
						})
					}

					if opt.Inbounds[inbound].Type == "hysteria2" {
						opt.Inbounds[inbound].Hysteria2Options.Users = append(opt.Inbounds[inbound].Hysteria2Options.Users, option.Hysteria2User{
							Name:     user.EmailAsId + "-hysteria2",
							Password: user.UserID,
						})
					}

					// 如果需要支持 vmess，可以取消注释以下代码
					// if opt.Inbounds[inbound].Type == "vmess" {
					// 	opt.Inbounds[inbound].VMessOptions.Users = append(opt.Inbounds[inbound].VMessOptions.Users, option.VMessUser{
					// 		Name:    user.EmailAsId + "-vmess",
					// 		UUID:    user.UUID,
					// 		AlterId: 0,
					// 	})
					// }
				}

			}(user)
		}

		wg.Wait()
		log.Printf("成功从 PostgreSQL 加载了 %d 个用户的配置", len(pgUsers))
	} else {
		log.Println("PostgreSQL 中没有找到活跃用户")
	}

	return opt, nil
}

// updateOptionsFromMongoDB 从 MongoDB 数据库更新配置（原有逻辑）
func updateOptionsFromMongoDB(opt option.Options) (option.Options, error) {
	var projections = bson.D{
		{Key: "email_as_id", Value: 1},
		{Key: "status", Value: 1},
		{Key: "uuid", Value: 1},
		{Key: "user_id", Value: 1},
	}

	cur, err := userTrafficLogsCol.Find(context.Background(), bson.D{}, options.Find().SetProjection(projections))
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

						// if opt.Inbounds[inbound].Type == "vmess" {
						// 	opt.Inbounds[inbound].VMessOptions.Users = append(opt.Inbounds[inbound].VMessOptions.Users, option.VMessUser{
						// 		Name:    user.Email_As_Id + "-vmess",
						// 		UUID:    user.UUID,
						// 		AlterId: 0,
						// 	})
						// }
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

func UsageDataOfAll(instance *box.Box) ([]Traffic, error) {

	statsService := instance.Router().V2RayServer().StatsService()

	regEx := `(?P<tag>[\w]+)>>>(?P<name>[-\w]+)>>>traffic>>>(?P<direction>[\w]+)`
	compRegEx := regexp.MustCompile(regEx)

	var loggingData = []Traffic{}
	var temp = map[string]int64{}

	response, err := statsService.(v2rayapi.StatsServiceServer).QueryStats(context.Background(),
		&v2rayapi.QueryStatsRequest{Reset_: true, Regexp: true, Patterns: []string{".*"}})
	if err != nil {
		log.Printf("%s", err)
		return nil, err
	}

	myStats := response.GetStat()

	for _, stat := range myStats {

		if stat.Value == 0 {
			continue
		}
		// log.Printf("%s: %d\n", stat.Name, stat.Value)

		matches := compRegEx.FindAllStringSubmatch(stat.Name, -1)
		for _, n := range matches {
			// log.Printf("%s: %s: %s: %s\n", n[0], n[1], n[2], n[3])

			if n[1] == "user" {
				parts := strings.Split(n[2], "-")
				if len(parts) > 0 {
					if value, ok := temp[parts[0]]; ok {
						temp[parts[0]] = value + stat.Value
					} else {
						temp[parts[0]] = stat.Value
					}
				}
			}

			// if n[2] == "proxy" {
			// 	if value, ok := temp["proxy"]; ok {
			// 		temp["proxy"] = value + stat.Value
			// 	} else {
			// 		temp["proxy"] = stat.Value
			// 	}
			// }
		}
	}

	for name, value := range temp {
		loggingData = append(loggingData, Traffic{
			Name:  name,
			Total: value,
		})
	}

	return loggingData, nil
}

func InitOptionsFromConfig(config string) (option.Options, error) {

	var options = option.Options{}

	configContent, err := os.ReadFile(config)
	if err != nil {
		log.Printf("error reading config file: %v\n", err)
		return options, err
	}

	options, err = json.UnmarshalExtended[option.Options]([]byte(configContent))
	if err != nil {
		log.Printf("error unmarshalling config file: %v\n", err)
		return options, err
	}

	return options, nil
}
