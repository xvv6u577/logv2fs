package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/yaml.v2"
)

type (
	User         = model.User
	YamlTemplate = model.YamlTemplate
	Proxies      = model.Proxies
	Headers      = model.Headers
	WsOpts       = model.WsOpts
	ProxyGroups  = model.ProxyGroups
)

func GenerateAllClashxConfig() {

	yamlFile, err := os.ReadFile("./yaml/template.yaml")
	if err != nil {
		fmt.Printf("yamlFile.Get err #%v ", err)
	}

	var projections = bson.D{
		{Key: "used_by_current_year", Value: 0},
		{Key: "used_by_current_month", Value: 0},
		{Key: "used_by_current_day", Value: 0},
		{Key: "traffic_by_year", Value: 0},
		{Key: "traffic_by_month", Value: 0},
		{Key: "traffic_by_day", Value: 0},
		{Key: "password", Value: 0},
		{Key: "refresh_token", Value: 0},
		{Key: "token", Value: 0},
		{Key: "suburl", Value: 0},
	}
	allUsers, err := database.GetPartialInfosForAllUsers(projections)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	for _, user := range allUsers {

		var yamlTemplate = YamlTemplate{}
		err = yaml.Unmarshal(yamlFile, &yamlTemplate)
		if err != nil {
			fmt.Printf("Unmarshal: %v", err)
		}

		for node, status := range user.NodeInUseStatus {
			nodeFlag := strings.Split(node, ".")[0]
			if status {
				yamlTemplate.Proxies = append(yamlTemplate.Proxies, Proxies{
					Name:           nodeFlag,
					Server:         node,
					Port:           443,
					Type:           "vmess",
					UUID:           user.UUID,
					AlterID:        64,
					Cipher:         "auto",
					TLS:            true,
					SkipCertVerify: false,
					Network:        "ws",
					WsOpts: WsOpts{
						Path:    "/" + user.Path,
						Headers: Headers{Host: node},
					},
				})

				for index, value := range yamlTemplate.ProxyGroups {
					if value.Name == "manual-select" || value.Name == "auto-select" {
						yamlTemplate.ProxyGroups[index].Proxies = append(yamlTemplate.ProxyGroups[index].Proxies, nodeFlag)
					}
				}
			}
		}
		fmt.Printf("%v\n", user.Name)

		newYaml, err := yaml.Marshal(&yamlTemplate)
		if err != nil {
			fmt.Printf("Marshal: %v", err)
		}

		err = ioutil.WriteFile("./yaml/results/"+user.Email+".yaml", newYaml, 0644)
		if err != nil {
			fmt.Printf("WriteFile: %v", err)
		}
	}

	fmt.Println("generate all user yaml file Done!")

}
