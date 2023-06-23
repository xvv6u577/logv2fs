package model

import "time"

type CurrentNode struct {
	Status             string         `json:"status" validate:"required,eq=active|eq=inactive"` // status: "active", "inactive"
	Domain             string         `json:"domain"`
	NodeAtCurrentYear  NodeAtPeriod   `json:"node_at_current_year"`
	NodeAtCurrentMonth NodeAtPeriod   `json:"node_at_current_month"`
	NodeAtCurrentDay   NodeAtPeriod   `json:"node_at_current_day"`
	NodeByYear         []NodeAtPeriod `json:"node_by_year"`
	NodeByMonth        []NodeAtPeriod `json:"node_by_month"`
	NodeByDay          []NodeAtPeriod `json:"node_by_day"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type NodeAtPeriod struct {
	Period              string           `json:"period"`
	Amount              int64            `json:"amount"`
	UserTrafficAtPeriod map[string]int64 `json:"user_traffic_at_period"`
}

type TrafficByNodeInDB struct {
	CreatedAt         time.Time `json:"created_at"`
	Total             int64     `json:"total"`
	Domain            string    `json:"domain"`
	TrafficByUserList []Traffic `json:"traffic_by_user" bson:"traffic_by_user"`
}

type GlobalVariable struct {
	Name       string            `json:"name" bson:"name" validate:"required,min=2,max=100"`
	DomainList map[string]string `json:"domain_list" bson:"domain_list"`
}
