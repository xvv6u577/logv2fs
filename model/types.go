package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	HourlyLogs    []struct {
		Timestamp time.Time `json:"timestamp" bson:"timestamp"`
		Traffic   int64     `json:"traffic" bson:"traffic"`
	} `json:"hourly_logs" bson:"hourly_logs"`
	DailyLogs []struct {
		Date    string `json:"date" bson:"date"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	} `json:"daily_logs" bson:"daily_logs"`
	MonthlyLogs []struct {
		Month   string `json:"month" bson:"month"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	} `json:"monthly_logs" bson:"monthly_logs"`
	YearlyLogs []struct {
		Year    string `json:"year" bson:"year"`
		Traffic int64  `json:"traffic" bson:"traffic"`
	} `json:"yearly_logs" bson:"yearly_logs"`
}

// CollectionName 返回MongoDB集合名称
func (UserTrafficLogs) CollectionName() string {
	return "USER_TRAFFIC_LOGS"
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
