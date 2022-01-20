package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	Email         string             `json:"email" bson:"email" validate:"required,min=2,max=100"`
	Password      string             `json:"password" validate:"required,min=6"`
	Path          string             `json:"path" bson:"path" validate:"required,eq=ray|eq=cas|eq=kay"`
	UUID          string             `json:"uuid" bson:"uuid"`
	Role          string             `json:"role" bson:"role" validate:"required,eq=admin|eq=normal"`                 // role: "admin", "normal"
	Status        string             `json:"status" bson:"status" validate:"required,eq=plain|eq=deleted|eq=overdue"` // status: "plain", "deleted", "overdue"
	Name          string             `json:"name" bson:"name"`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	User_id       string             `json:"user_id"`
	Usedtraffic   int64              `json:"used" bson:"used"`
	Credittraffic int64              `json:"credit" bson:"credit"`
	NodeInUse     *map[string]string `json:"nodeinuse" bson:"nodeinuse"`
	Suburl        string             `json:"suburl"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

type Traffic struct {
	Name string `json:"name" bson:"name"`
	// Uplink uint64 `json:"uplink" bson:"uplink"`
	// Downlink uint64 `json:"downlink" bson:"downlink"`
	Total int64 `json:"total" bson:"total"`
}

type TrafficInDB struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	Total     int64     `json:"total" bson:"total"`
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
