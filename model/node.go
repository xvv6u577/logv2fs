package model

import "time"

type CurrentNode struct {
	Status             string         `json:"status" bson:"status" validate:"required,eq=active|eq=inactive"` // status: "active", "inactive"
	Domain             string         `json:"domain" bson:"domain" validate:"required,min=2,max=100"`
	IP                 string         `json:"ip" bson:"ip"`
	Remark             string         `json:"remark" bson:"remark"`
	NodeAtCurrentYear  NodeAtPeriod   `json:"node_at_current_year" bson:"node_at_current_year"`
	NodeAtCurrentMonth NodeAtPeriod   `json:"node_at_current_month" bson:"node_at_current_month"`
	NodeAtCurrentDay   NodeAtPeriod   `json:"node_at_current_day" bson:"node_at_current_day"`
	NodeByYear         []NodeAtPeriod `json:"node_by_year" bson:"node_by_year"`
	NodeByMonth        []NodeAtPeriod `json:"node_by_month" bson:"node_by_month"`
	NodeByDay          []NodeAtPeriod `json:"node_by_day" bson:"node_by_day"`
	CreatedAt          time.Time      `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at" bson:"updated_at"`
}

type NodeAtPeriod struct {
	Period              string           `json:"period" bson:"period" validate:"required,min=2,max=100"`
	Amount              int64            `json:"amount" bson:"amount"`
	UserTrafficAtPeriod map[string]int64 `json:"user_traffic_at_period" bson:"user_traffic_at_period"`
}

// var nodeGlobalList = NodeGlobalList{"domain": "remark"}
type GlobalVariable struct {
	Name                  string   `json:"name" bson:"name" validate:"required,min=2,max=100"`
	WorkRelatedDomainList []Domain `json:"work_related_domain_list" bson:"work_related_domain_list"`
	ActiveGlobalNodes     []Domain `json:"active_global_nodes" bson:"active_global_nodes"`
}

// Domain type: "work", "vmess", "vmessCDN", "vlessCDN"
type Domain struct {
	Type              string `json:"type" bason:"type"`
	Domain            string `json:"domain" bson:"domain" validate:"required,min=2,max=100"`
	IP                string `json:"ip" bason:"ip"`
	Remark            string `json:"remark" bson:"remark"`
	EnableSubcription bool   `json:"enable_subscription" bson:"enable_subscription"`
	EnableChatgpt     bool   `json:"enable_chatgpt" bson:"enable_chatgpt"`
	SNI               string `json:"sni" bson:"sni"`
	UUID              string `json:"uuid" bson:"uuid"`
	PATH              string `json:"path" bson:"path"`
}

type DomainInfo struct {
	Domain       string `json:"domain"`
	Remark       string `json:"remark"`
	ExpiredDate  string `json:"expired_date"`
	DaysToExpire int    `json:"days_to_expire"`
}
