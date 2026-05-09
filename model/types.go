package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserTrafficLogs 是用户主文档；按周期粒度的流量日志已拆出到 UserTrafficPeriod 集合，
// 通过 email_as_id 关联。读接口会用 $lookup 将 daily_logs / monthly_logs / yearly_logs
// 重新拼回原嵌套数组形态以保持前端兼容（仅作为响应字段，不会持久化到本集合）。
type UserTrafficLogs struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	Email_As_Id   string             `json:"email_as_id" bson:"email_as_id"`
	Password      string             `json:"password" validate:"required,min=6"`
	UUID          string             `json:"uuid" bson:"uuid"`
	Role          string             `json:"role" bson:"role" validate:"required,eq=admin|eq=normal"`                 // role: "admin", "normal"
	Status        string             `json:"status" bson:"status" validate:"required,eq=plain|eq=deleted|eq=overdue"` // status: "plain", "deleted", "overdue"
	Name          string             `json:"name" bson:"name"`
	Remark        string             `json:"remark" bson:"remark"` // 用户备注
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	User_id       string             `json:"user_id" bson:"user_id"`
	Used          int64              `json:"used" bson:"used"`
	Credit        int64              `json:"credit" bson:"credit"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`

	// 以下三个字段不会写入 USER_TRAFFIC_LOGS 集合，仅在读接口通过 $lookup
	// 从 user_traffic_periods 集合关联出来，保持原前端 JSON 结构兼容。
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
func (UserTrafficLogs) CollectionName() string {
	return "USER_TRAFFIC_LOGS"
}

// UserTrafficPeriod 是从 USER_TRAFFIC_LOGS 拆出的周期级流量记录。
// 唯一键为 (email_as_id, kind, period)，由启动时的索引保障。
//   - kind 取值: "daily" | "monthly" | "yearly"
//   - period 格式: daily=yyyymmdd, monthly=yyyymm, yearly=yyyy
type UserTrafficPeriod struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	EmailAsId string             `json:"email_as_id" bson:"email_as_id"`
	Kind      string             `json:"kind" bson:"kind"`
	Period    string             `json:"period" bson:"period"`
	Traffic   int64              `json:"traffic" bson:"traffic"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// CollectionName 返回 MongoDB 集合名称
func (UserTrafficPeriod) CollectionName() string {
	return "user_traffic_periods"
}

// CustomDate 自定义日期模型
type CustomDate struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	DomainAsId string             `json:"domain_as_id" bson:"domain_as_id"`
	CustomDate string             `json:"custom_date" bson:"custom_date"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// CollectionName 返回MongoDB集合名称
func (CustomDate) CollectionName() string {
	return "CUSTOM_DATES"
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
