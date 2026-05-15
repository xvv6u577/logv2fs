package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NodeTrafficLogs 是节点主文档；按周期粒度的流量日志已拆出到 NodeTrafficPeriod 集合，
// 通过 domain_as_id 关联。读接口会用 $lookup 将 daily_logs / monthly_logs / yearly_logs
// 拼回原嵌套数组形态以兼容前端（仅作为响应字段，不会持久化到本集合）。
type NodeTrafficLogs struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id"`
	Domain_As_Id string             `json:"domain_as_id" bson:"domain_as_id"`
	Remark       string             `json:"remark" bson:"remark"`
	Status       string             `json:"status" bson:"status" validate:"required,eq=active|eq=inactive"` // status: "active", "inactive"
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`

	// 以下三个字段不会写入 NODE_TRAFFIC_LOGS 集合，仅在读接口通过 $lookup
	// 从 node_traffic_periods 集合关联出来，保持原前端 JSON 结构兼容。
	DailyLogs []struct {
		Date    string `json:"date" bson:"date"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	} `json:"daily_logs" bson:"daily_logs,omitempty"`
	MonthlyLogs []struct {
		Month   string `json:"month" bson:"month"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	} `json:"monthly_logs" bson:"monthly_logs,omitempty"`
	YearlyLogs []struct {
		Year    string `json:"year" bson:"year"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	} `json:"yearly_logs" bson:"yearly_logs,omitempty"`
}

// CollectionName 返回MongoDB集合名称
func (NodeTrafficLogs) CollectionName() string {
	return "NODE_TRAFFIC_LOGS"
}

// NodeTrafficPeriod 是从 NODE_TRAFFIC_LOGS 拆出的周期级流量记录。
// 唯一键为 (domain_as_id, kind, period)，由启动时的索引保障。
//   - kind 取值: "daily" | "monthly" | "yearly"
//   - period 格式: daily=yyyymmdd, monthly=yyyymm, yearly=yyyy
type NodeTrafficPeriod struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	DomainAsId string             `json:"domain_as_id" bson:"domain_as_id"`
	Kind       string             `json:"kind" bson:"kind"`
	Period     string             `json:"period" bson:"period"`
	Traffic    int64              `json:"traffic" bson:"traffic"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// CollectionName 返回 MongoDB 集合名称
func (NodeTrafficPeriod) CollectionName() string {
	return "node_traffic_periods"
}

type NodeAtPeriod struct {
	Period              string           `json:"period" bson:"period" validate:"required,min=2,max=100"`
	Amount              int64            `json:"amount" bson:"amount"`
	UserTrafficAtPeriod map[string]int64 `json:"user_traffic_at_period" bson:"user_traffic_at_period"`
}

// Domain type: "vmessws", "reality", "hysteria2", "vlessCDN"
type SubscriptionNode struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Type         string             `json:"type" bson:"type"`
	Remark       string             `json:"remark" bson:"remark"`
	Domain       string             `json:"domain" bson:"domain" validate:"required,min=2,max=100"`
	IP           string             `json:"ip" bson:"ip"`
	SNI          string             `json:"sni" bson:"sni"`
	UUID         string             `json:"uuid" bson:"uuid"`
	PATH         string             `json:"path" bson:"path"`
	SERVER_PORT  string             `json:"server_port" bson:"server_port"`
	PASSWORD     string             `json:"password" bson:"password"`
	PUBLIC_KEY   string             `json:"public_key" bson:"public_key"`
	SHORT_ID     string             `json:"short_id" bson:"short_id"`
	EnableOpenai bool               `json:"enable_openai" bson:"enable_openai"`
	Weight       int                `json:"weight" bson:"weight"` // 权重字段，用于节点排序，数值越小越靠前
	ControlPort  string             `json:"control_port" bson:"control_port"`
	Status       string             `json:"status" bson:"status"` // status: "active", "inactive"
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// CollectionName 返回MongoDB集合名称
func (SubscriptionNode) CollectionName() string {
	return "subscription_nodes"
}
