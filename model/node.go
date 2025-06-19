package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NodeTrafficLogs struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id"`
	Domain_As_Id string             `json:"domain_as_id" bson:"domain_as_id"`
	Remark       string             `json:"remark" bson:"remark"`
	Status       string             `json:"status" bson:"status" validate:"required,eq=active|eq=inactive"` // status: "active", "inactive"
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
	HourlyLogs   []struct {
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

type NodeAtPeriod struct {
	Period              string           `json:"period" bson:"period" validate:"required,min=2,max=100"`
	Amount              int64            `json:"amount" bson:"amount"`
	UserTrafficAtPeriod map[string]int64 `json:"user_traffic_at_period" bson:"user_traffic_at_period"`
}

// Domain type: "work", "vmesstls", "vmessws", "reality", "hysteria2", "vlessCDN"
type SubscriptionNode struct {
	Type         string `json:"type" bason:"type"`
	Remark       string `json:"remark" bson:"remark"`
	Domain       string `json:"domain" bson:"domain" validate:"required,min=2,max=100"`
	IP           string `json:"ip" bason:"ip"`
	SNI          string `json:"sni" bson:"sni"`
	UUID         string `json:"uuid" bson:"uuid"`
	PATH         string `json:"path" bson:"path"`
	SERVER_PORT  string `json:"server_port" bson:"server_port"`
	PASSWORD     string `json:"password" bson:"password"`
	PUBLIC_KEY   string `json:"public_key" bson:"public_key"`
	SHORT_ID     string `json:"short_id" bson:"short_id"`
	EnableOpenai bool   `json:"enable_openai" bson:"enable_openai"`
}

type ExpiryCheckDomainInfo struct {
	Domain       string `json:"domain" bson:"domain"`
	Remark       string `json:"remark" bson:"remark"`
	ExpiredDate  string `json:"expired_date" bson:"expired_date"`
	DaysToExpire int    `json:"days_to_expire" bson:"days_to_expire"`
}
