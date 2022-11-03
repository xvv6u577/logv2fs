package yaml

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/caster8013/logv2rayfullstack/database"
	"github.com/caster8013/logv2rayfullstack/model"
	sanitize "github.com/caster8013/logv2rayfullstack/sanitize"
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

	err = GenerateOne(user)
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
		GenerateOne(*user)
	}

	log.Println("generate all user yaml file Done!")
	return nil
}

func GenerateOne(user User) error {

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
	user_email := sanitize.SanitizeStr(user.Email)
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
