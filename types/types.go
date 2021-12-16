package types

import "time"

type User struct {
	Email    string `json:"email" bson:"email"`
	ID       string `json:"uuid" bson:"uuid"`
	Path     string `json:"path" bson:"path"`
	Name     string `json:"name" bson:"name"`
	Password string `json:"password" bson:"password"`
	// status: "plain", "deleted", "overdue"
	Status        string    `json:"status" bson:"status"`
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
