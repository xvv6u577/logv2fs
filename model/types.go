package model

import (
	"time"

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
	NodeInUse          *map[string]string `json:"nodeinuse" bson:"nodeinuse"`
	NodeGlobal         *map[string]string `json:"nodeglobal,omitempty" bson:"nodeglobal"`
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
}

func (u *User) ProduceSuburl() string {
	return "https://" + u.Email + ".ray.io"
}
