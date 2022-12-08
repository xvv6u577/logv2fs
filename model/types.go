package model

import (
	"encoding/json"
	"strings"
	"time"

	b64 "encoding/base64"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	NodeGlobalList     map[string]string  `json:"node_global_list" bson:"node_global_list"`
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
}

type Node struct {
	Version string `default:"2" json:"v"`
	Domain  string `json:"add"`
	UUID    string `json:"id"`
	Path    string `json:"path"`
	Remark  string `json:"ps"`
	Port    string `default:"443" json:"port"`
	Aid     string `default:"64" json:"aid"`
	Net     string `default:"ws" json:"net"`
	Type    string `json:"type" `
	Host    string `default:"none" json:"host"`
	Tls     string `default:"tls" json:"tls"`
	Method  string `default:"none" json:"method"`
}

// type YamlTemplate struct {
// 	Port               int           `yaml:"port"`
// 	SocksPort          int           `yaml:"socks-port"`
// 	AllowLan           bool          `yaml:"allow-lan"`
// 	Mode               string        `yaml:"mode"`
// 	LogLevel           string        `yaml:"log-level"`
// 	ExternalController string        `yaml:"external-controller"`
// 	Proxies            []Proxies     `yaml:"proxies"`
// 	ProxyGroups        []ProxyGroups `yaml:"proxy-groups"`
// 	Rules              []string      `yaml:"rules"`
// }

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

func (u *User) ProduceSuburl() {
	var node Node
	subscription := ""

	for index, item := range u.NodeInUseStatus {
		if item {
			node = Node{
				Domain:  index,
				Path:    "/" + u.Path,
				UUID:    u.UUID,
				Remark:  strings.Split(index, ".")[0],
				Version: "2",
				Port:    "443",
				Aid:     "4",
				Net:     "ws",
				Type:    "none",
				Tls:     "tls",
				Method:  "none",
			}

			jsonedNode, _ := json.Marshal(node)

			if len(subscription) == 0 {
				subscription = "vmess://" + b64.StdEncoding.EncodeToString(jsonedNode)
			} else {
				subscription = subscription + "\n" + "vmess://" + b64.StdEncoding.EncodeToString(jsonedNode)
			}
		}
	}

	u.Suburl = b64.StdEncoding.EncodeToString([]byte(subscription))
}

func (u *User) DeleteNodeInUse(domain string) {
	u.NodeInUseStatus[domain] = false
	u.ProduceSuburl()
}

func (u *User) AddNodeInUse(domain string) {
	u.NodeInUseStatus[domain] = true
	u.ProduceSuburl()
}

func (u *User) ProduceNodeInUse(submittedNodes map[string]string) {

	var updatedNodes = map[string]bool{}
	var submittedNodesReverse = map[string]string{}
	for k, v := range submittedNodes {
		submittedNodesReverse[v] = k
	}

	if u.Status == "plain" {
		// if node in u.NodeInUseStatus and in submittedNodesReverse, keep it
		for node, status := range u.NodeInUseStatus {
			if _, ok := submittedNodesReverse[node]; ok {
				updatedNodes[node] = status
			}
		}
		// if node in submittedNodesReverse not in updatedNodes, add it
		for node := range submittedNodesReverse {
			if _, ok := updatedNodes[node]; !ok {
				updatedNodes[node] = true
			}
		}
	} else {
		for node := range submittedNodesReverse {
			updatedNodes[node] = false
		}
	}
	u.NodeInUseStatus = updatedNodes
	u.ProduceSuburl()
}
