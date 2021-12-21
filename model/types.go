package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id"`
	Email    string             `json:"email" bson:"email" validate:"required,min=4,max=100"`
	UUID     string             `json:"uuid" bson:"uuid"`
	Path     string             `json:"path" bson:"path" validate:"required,eq=ray|eq=cas|eq=kay"`
	Name     string             `json:"name" bson:"name"`
	Password string             `json:"password" bson:"password" validate:"required,min=6"`
	// role: "admin", "normal"
	Role string `json:"role" bson:"role" validate:"required,eq=admin|eq=normal"`
	// status: "plain", "deleted", "overdue"
	Status        string    `json:"status" bson:"status" validate:"required,eq=plain|eq=deleted|eq=overdue"`
	Token         *string   `json:"token"`
	Refresh_token *string   `json:"refresh_token"`
	User_id       string    `json:"user_id"`
	Usedtraffic   int64     `json:"used" bson:"used"`
	Credittraffic int64     `json:"credit" bson:"credit"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
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
