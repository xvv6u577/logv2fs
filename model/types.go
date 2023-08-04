package model

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	b64 "encoding/base64"

	helper "github.com/caster8013/logv2rayfullstack/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/yaml.v2"
)

// NodeGlobalList     map[string]string  `json:"node_global_list" bson:"node_global_list"`
type User struct {
	ID                 primitive.ObjectID `bson:"_id"`
	Email              string             `json:"email" bson:"email" validate:"required,min=2,max=100"`
	Password           string             `json:"password" validate:"required,min=6"`
	Path               string             `json:"path" bson:"path" validate:"required,eq=ray|eq=cas|eq=kay"`
	UUID               string             `json:"uuid" bson:"uuid"`
	Role               string             `json:"role" bson:"role" validate:"required,eq=admin|eq=normal"`                 // role: "admin", "normal"
	Status             string             `json:"status" bson:"status" validate:"required,eq=plain|eq=deleted|eq=overdue"` // status: "plain", "deleted", "overdue"
	Name               string             `json:"name" bson:"name"`
	Token              *string            `json:"token"`
	Refresh_token      *string            `json:"refresh_token"`
	User_id            string             `json:"user_id" bson:"user_id"`
	Usedtraffic        int64              `json:"used" bson:"used"`
	Credittraffic      int64              `json:"credit" bson:"credit"`
	NodeInUseStatus    map[string]bool    `json:"node_in_use_status" bson:"node_in_use_status"`
	Suburl             string             `json:"suburl"`
	CreatedAt          time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" bson:"updated_at"`
	UsedByCurrentYear  TrafficAtPeriod    `json:"used_by_current_year" bson:"used_by_current_year"`
	UsedByCurrentMonth TrafficAtPeriod    `json:"used_by_current_month" bson:"used_by_current_month"`
	UsedByCurrentDay   TrafficAtPeriod    `json:"used_by_current_day" bson:"used_by_current_day"`
	TrafficByYear      []TrafficAtPeriod  `json:"traffic_by_year" bson:"traffic_by_year"`
	TrafficByMonth     []TrafficAtPeriod  `json:"traffic_by_month" bson:"traffic_by_month"`
	TrafficByDay       []TrafficAtPeriod  `json:"traffic_by_day" bson:"traffic_by_day"`
}

type TrafficAtPeriod struct {
	Period       string           `json:"period" bson:"period"`
	Amount       int64            `json:"amount" bson:"amount"`
	UsedByDomain map[string]int64 `json:"used_by_domain" bson:"used_by_domain"`
}

type Traffic struct {
	Name  string `json:"name" bson:"name"`
	Total int64  `json:"total" bson:"total"`
}

type TrafficInDB struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	Total     int64     `json:"total" bson:"total"`
	Domain    string    `json:"domain" bson:"domain"`
	Email     string    `json:"email" bson:"email"`
}

type Node struct {
	Version     string `default:"2" json:"v"`
	Remark      string `json:"ps"`
	Domain      string `json:"add"`
	Port        string `default:"443" json:"port"`
	UUID        string `json:"id"`
	Aid         string `default:"4" json:"aid"`
	Security    string `default:"auto" json:"scy"`
	Net         string `default:"ws" json:"net"`
	Type        string `default:"none" json:"type" `
	Host        string `json:"host"`
	Path        string `json:"path"`
	Tls         string `default:"tls" json:"tls"`
	SNI         string `json:"sni"`
	Alpn        string `default:"h2" json:"alpn"`
	FingerPrint string `default:"chrome" json:"fp"`
}

type YamlTemplate struct {
	Port               int           `default:"7890" yaml:"mixed-port"`
	AllowLan           bool          `yaml:"allow-lan"`
	BindAddress        string        `yaml:"bind-address"`
	Mode               string        `yaml:"mode"`
	LogLevel           string        `yaml:"log-level"`
	ExternalController string        `yaml:"external-controller"`
	Dns                Dns           `yaml:"dns"`
	Proxies            []Proxies     `yaml:"proxies"`
	ProxyGroups        []ProxyGroups `yaml:"proxy-groups"`
	Rules              []string      `yaml:"rules"`
}

type Dns struct {
	Enable            bool           `yaml:"enable"`
	Ipv6              bool           `yaml:"ipv6"`
	DefaultNameserver []string       `yaml:"default-nameserver"`
	EnhancedMode      string         `yaml:"enhanced-mode"`
	FakeIpRange       string         `yaml:"fake-ip-range"`
	UseHosts          bool           `yaml:"use-hosts"`
	NameServer        []string       `yaml:"nameserver"`
	Fallback          []string       `yaml:"fallback"`
	FallbackFilter    FallbackFilter `yaml:"fallback-filter"`
}

type FallbackFilter struct {
	Geoip  bool     `yaml:"geoip"`
	Ipcidr []string `yaml:"ipcidr"`
}

type Headers struct {
	Host string `yaml:"Host"`
}
type WsOpts struct {
	Path    string  `yaml:"path"`
	Headers Headers `yaml:"headers"`
}
type Proxies struct {
	Name           string `yaml:"name"`
	Server         string `yaml:"server"`
	Port           int    `yaml:"port"`
	Type           string `yaml:"type"`
	UUID           string `yaml:"uuid"`
	AlterID        int    `yaml:"alterId"`
	Cipher         string `yaml:"cipher"`
	TLS            bool   `yaml:"tls"`
	SkipCertVerify bool   `yaml:"skip-cert-verify"`
	SNI            string `yaml:"sni"`
	UDP            bool   `yaml:"udp"`
	Network        string `yaml:"network"`
	WsOpts         WsOpts `yaml:"ws-opts"`
}
type ProxyGroups struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Proxies  []string `yaml:"proxies"`
	URL      string   `yaml:"url,omitempty"`
	Interval int      `yaml:"interval,omitempty"`
}

func (u *User) ProduceSuburl(activeGlobalNodes []Domain) {

	if u.Status != "plain" {
		u.Suburl = ""
		return
	}

	subscription := ""
	for _, item := range activeGlobalNodes {

		if item.Domain == "localhost" || (item.Type == "vmess" && !u.NodeInUseStatus[item.Domain]) {
			continue
		}

		var node Node

		switch item.Type {
		case "vmess", "vmessCDN":
			node = Node{
				Domain:  item.Domain,
				Path:    "/" + u.Path,
				UUID:    u.UUID,
				Remark:  item.Remark,
				Version: "2",
				Port:    "443",
				Aid:     "4",
				Net:     "ws",
				Type:    "none",
				Tls:     "tls",
				Host:    item.SNI,
				SNI:     item.SNI,
			}
			jsonedNode, _ := json.Marshal(node)
			if len(subscription) == 0 {
				subscription = "vmess://" + b64.StdEncoding.EncodeToString(jsonedNode)
			} else {
				subscription = subscription + "\n" + "vmess://" + b64.StdEncoding.EncodeToString(jsonedNode)
			}

		case "vlessCDN":
			node = Node{
				Domain:  item.Domain,
				Path:    item.PATH,
				UUID:    item.UUID,
				Remark:  item.Remark,
				Version: "2",
				Port:    "443",
				Aid:     "4",
				Net:     "ws",
				Type:    "none",
				Tls:     "tls",
				Host:    item.SNI,
				SNI:     item.SNI,
			}
			jsonedNode, _ := json.Marshal(node)
			if len(subscription) == 0 {
				subscription = "vless://" + b64.StdEncoding.EncodeToString(jsonedNode)
			} else {
				subscription = subscription + "\n" + "vless://" + b64.StdEncoding.EncodeToString(jsonedNode)
			}

		}

	}

	u.Suburl = b64.StdEncoding.EncodeToString([]byte(subscription))
}

func (u *User) DeleteNodeInUse(domain string) {
	u.NodeInUseStatus[domain] = false
}

func (u *User) AddNodeInUse(domain string) {
	u.NodeInUseStatus[domain] = true
}

func (u *User) UpdateNodeStatusInUse(activeGlobalNodes []Domain) {

	var updatedNodes = map[string]bool{}
	var simplifiedNodes = map[string]string{}
	for _, node := range activeGlobalNodes {
		if node.Type == "vmess" {
			simplifiedNodes[node.Domain] = node.Remark
		}
	}

	if u.Status == "plain" {
		// if node in u.NodeInUseStatus and in simplifiedNodes, keep it
		for node, status := range u.NodeInUseStatus {
			if _, ok := simplifiedNodes[node]; ok {
				updatedNodes[node] = status
			}
		}
		// if node in simplifiedNodes not in updatedNodes, add it
		for node := range simplifiedNodes {
			if _, ok := updatedNodes[node]; !ok {
				updatedNodes[node] = true
			}
		}
	} else {
		for node := range simplifiedNodes {
			updatedNodes[node] = false
		}
	}

	u.NodeInUseStatus = updatedNodes
}

func (u *User) GenerateYAML(nodes []Domain) error {

	var yamlTemplate = YamlTemplate{}
	yamlFile, err := os.ReadFile(helper.CurrentPath() + "/config/template.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err #%v ", err)
		return err
	}

	if u.Status == "plain" {
		var noVlessNodes []Domain
		for _, item := range nodes {
			if item.Type != "vlessCDN" {
				noVlessNodes = append(noVlessNodes, item)
			}
		}

		err = yaml.Unmarshal(yamlFile, &yamlTemplate)
		if err != nil {
			log.Printf("Unmarshal: %v", err)
			return err
		}

		for _, node := range noVlessNodes {

			if node.Domain == "localhost" || (node.Type == "vmess" && !u.NodeInUseStatus[node.Domain]) || !node.EnableSubcription {
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
	}

	newYaml, err := yaml.Marshal(&yamlTemplate)
	if err != nil {
		log.Printf("Marshal: %v", err)
		return err
	}

	// if directory not exist, create it
	if _, err := os.Stat(helper.CurrentPath() + "/config/results"); os.IsNotExist(err) {
		os.Mkdir(helper.CurrentPath()+"/config/results", os.ModePerm)
	}
	err = ioutil.WriteFile(helper.CurrentPath()+"/config/results/"+u.Email+".yaml", newYaml, 0644)
	if err != nil {
		log.Printf("WriteFile: %v", err)
		return err
	}

	return nil
}
