package yaml

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/caster8013/logv2rayfullstack/database"
	helper "github.com/caster8013/logv2rayfullstack/helpers"
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

func CurrentPath() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func GenerateOneByQuery(email string) error {

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
	user, err := database.GetUserByName(email, projections)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	err = GenerateOneYAML(user)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return err
	}

	log.Printf("generate %v yaml config Done!", user.Email)
	return nil
}

func RemoveOne(email string) error {
	// check if file exist, if exist, remove it
	filePath := CurrentPath() + "/yaml/results/" + email + ".yaml"
	if _, err := os.Stat(filePath); err == nil {
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Remove: %v\n", err)
			return err
		}
	}
	return nil
}

func GenerateAllClashxConfig() error {
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
		log.Printf("Error: %v\n", err)
		return err
	}

	for _, user := range allUsers {
		GenerateOneYAML(*user)
	}

	log.Println("generate all user yaml file Done!")
	return nil
}

func GenerateOneYAML(user User) error {

	var exchangeMap = map[string]string{}
	for k, v := range user.NodeGlobalList {
		exchangeMap[v] = k
	}

	yamlFile, err := os.ReadFile(CurrentPath() + "/yaml/template.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
		return err
	}

	var yamlTemplate = YamlTemplate{}
	err = yaml.Unmarshal(yamlFile, &yamlTemplate)
	if err != nil {
		log.Printf("Unmarshal: %v", err)
		return err
	}

	for node, status := range user.NodeInUseStatus {

		if status {
			yamlTemplate.Proxies = append(yamlTemplate.Proxies, Proxies{
				Name:           exchangeMap[node],
				Server:         node,
				Port:           443,
				Type:           "vmess",
				UUID:           user.UUID,
				AlterID:        4,
				Cipher:         "none",
				TLS:            true,
				SkipCertVerify: false,
				Network:        "ws",
				WsOpts: WsOpts{
					Path:    "/" + user.Path,
					Headers: Headers{Host: node},
				},
			})

			for index, value := range yamlTemplate.ProxyGroups {
				if value.Name == "manual-select" || value.Name == "auto-select" || value.Name == "fallback" {
					yamlTemplate.ProxyGroups[index].Proxies = append(yamlTemplate.ProxyGroups[index].Proxies, exchangeMap[node])
				}
			}
		}
	}
	user_email := helper.SanitizeStr(user.Email)
	log.Printf("%v generated yaml!\n", user_email)

	newYaml, err := yaml.Marshal(&yamlTemplate)
	if err != nil {
		log.Printf("Marshal: %v", err)
		return err
	}

	// if directory not exist, create it
	if _, err := os.Stat(CurrentPath() + "/yaml/results"); os.IsNotExist(err) {
		os.Mkdir(CurrentPath()+"/yaml/results", os.ModePerm)
	}
	err = ioutil.WriteFile(CurrentPath()+"/yaml/results/"+user_email+".yaml", newYaml, 0644)
	if err != nil {
		log.Printf("WriteFile: %v", err)
		return err
	}

	return nil
}
