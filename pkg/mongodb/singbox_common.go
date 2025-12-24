package mongodb_pkg

import (
	"context"
	"log"
	"os"
	"regexp"
	"strings"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/experimental/v2rayapi"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
	"github.com/xvv6u577/logv2fs/model"
)

type (
	Traffic = model.Traffic
)

// UsageDataOfAll 获取所有用户的流量使用数据
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

		matches := compRegEx.FindAllStringSubmatch(stat.Name, -1)
		for _, n := range matches {

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

// InitOptionsFromConfig 从配置文件初始化选项
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
