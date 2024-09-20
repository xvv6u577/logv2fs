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
)

var (
	CURRENT_DOMAIN     = os.Getenv("CURRENT_DOMAIN")
	userTrafficLogsCol = database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")
)

func UpdateOptionsFromDB(opt option.Options) (option.Options, error) {

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
