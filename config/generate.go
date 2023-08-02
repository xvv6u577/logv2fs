package yaml

import (
	"io/ioutil"
	"log"
	"os"

	helper "github.com/caster8013/logv2rayfullstack/helpers"
	"github.com/caster8013/logv2rayfullstack/model"
	"gopkg.in/yaml.v2"
)

type (
	User         = model.User
	YamlTemplate = model.YamlTemplate
	Proxies      = model.Proxies
	Headers      = model.Headers
	WsOpts       = model.WsOpts
	ProxyGroups  = model.ProxyGroups
	Domain       = model.Domain
)

type UserInstance struct {
	*User
}

func RemoveOne(email string) error {
	// check if file exist, if exist, remove it
	filePath := helper.CurrentPath() + "/yaml/results/" + email + ".yaml"
	if _, err := os.Stat(filePath); err == nil {
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Remove: %v\n", err)
			return err
		}
	}
	return nil
}

func (u UserInstance) GenerateYAML(nodes []Domain) error {

	var noVlessNodes []Domain
	for _, item := range nodes {
		if item.Type != "vless" {
			noVlessNodes = append(noVlessNodes, item)
		}
	}

	yamlFile, err := os.ReadFile(helper.CurrentPath() + "/yaml/template.yaml")
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

	for _, node := range noVlessNodes {

		if !u.NodeInUseStatus[node.Domain] || !node.EnableSubcription {
			continue
		}

		yamlTemplate.Proxies = append(yamlTemplate.Proxies, Proxies{
			Name:           node.Remark,
			Server:         node.Domain,
			Port:           443,
			Type:           "vmess",
			UUID:           u.UUID,
			AlterID:        4,
			Cipher:         "none",
			TLS:            true,
			SkipCertVerify: false,
			Network:        "ws",
			SNI:            node.SNI,
			UDP:            false,
			WsOpts: WsOpts{
				Path:    "/" + u.Path,
				Headers: Headers{Host: node.SNI},
			},
		})

		for index, value := range yamlTemplate.ProxyGroups {
			if value.Name == "manual-select" || value.Name == "auto-select" || value.Name == "fallback" {
				yamlTemplate.ProxyGroups[index].Proxies = append(yamlTemplate.ProxyGroups[index].Proxies, node.Remark)
			}
		}

		if node.Type == "vmess" && node.EnableChatgpt {
			for index, value := range yamlTemplate.ProxyGroups {
				if value.Name == "chatGPT" || value.Name == "gpt-auto" {
					yamlTemplate.ProxyGroups[index].Proxies = append(yamlTemplate.ProxyGroups[index].Proxies, node.Remark)
				}
			}
		}
	}

	newYaml, err := yaml.Marshal(&yamlTemplate)
	if err != nil {
		log.Printf("Marshal: %v", err)
		return err
	}

	// if directory not exist, create it
	if _, err := os.Stat(helper.CurrentPath() + "/yaml/results"); os.IsNotExist(err) {
		os.Mkdir(helper.CurrentPath()+"/yaml/results", os.ModePerm)
	}
	err = ioutil.WriteFile(helper.CurrentPath()+"/yaml/results/"+u.Email+".yaml", newYaml, 0644)
	if err != nil {
		log.Printf("WriteFile: %v", err)
		return err
	}

	return nil
}
