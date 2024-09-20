package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

type UserTrafficLogs struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	Email_As_Id   string             `json:"email_as_id" bson:"email_as_id"`
	Password      string             `json:"password" validate:"required,min=6"`
	UUID          string             `json:"uuid" bson:"uuid"`
	Role          string             `json:"role" bson:"role" validate:"required,eq=admin|eq=normal"`                 // role: "admin", "normal"
	Status        string             `json:"status" bson:"status" validate:"required,eq=plain|eq=deleted|eq=overdue"` // status: "plain", "deleted", "overdue"
	Name          string             `json:"name" bson:"name"`
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
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	Total     int64              `json:"total" bson:"total"`
	Domain    string             `json:"domain" bson:"domain"`
	Email     string             `json:"email" bson:"email"`
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
