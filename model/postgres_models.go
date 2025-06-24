package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// PostgreSQL版本的用户流量日志模型 - 混合设计
type UserTrafficLogsPG struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EmailAsId    string    `json:"email_as_id" gorm:"uniqueIndex;not null"`
	Password     string    `json:"password" gorm:"not null"`
	UUID         string    `json:"uuid" gorm:"index"`
	Role         string    `json:"role" gorm:"type:varchar(20);check:role IN ('admin','normal');not null"`
	Status       string    `json:"status" gorm:"type:varchar(20);check:status IN ('plain','deleted','overdue');not null"`
	Name         string    `json:"name"`
	Token        *string   `json:"token"`
	RefreshToken *string   `json:"refresh_token"`
	UserID       string    `json:"user_id" gorm:"index"`
	Used         int64     `json:"used" gorm:"default:0"`
	Credit       int64     `json:"credit" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 时间序列数据使用JSONB存储 - 这是混合设计的核心
	HourlyLogs  datatypes.JSON `json:"hourly_logs" gorm:"type:jsonb"`
	DailyLogs   datatypes.JSON `json:"daily_logs" gorm:"type:jsonb"`
	MonthlyLogs datatypes.JSON `json:"monthly_logs" gorm:"type:jsonb"`
	YearlyLogs  datatypes.JSON `json:"yearly_logs" gorm:"type:jsonb"`
}

// 为PostgreSQL表设置表名
func (UserTrafficLogsPG) TableName() string {
	return "user_traffic_logs"
}

// PostgreSQL版本的节点流量日志模型 - 混合设计
type NodeTrafficLogsPG struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DomainAsId string    `json:"domain_as_id" gorm:"uniqueIndex;not null"`
	Remark     string    `json:"remark"`
	Status     string    `json:"status" gorm:"type:varchar(20);check:status IN ('active','inactive');not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 时间序列数据使用JSONB存储
	HourlyLogs  datatypes.JSON `json:"hourly_logs" gorm:"type:jsonb"`
	DailyLogs   datatypes.JSON `json:"daily_logs" gorm:"type:jsonb"`
	MonthlyLogs datatypes.JSON `json:"monthly_logs" gorm:"type:jsonb"`
	YearlyLogs  datatypes.JSON `json:"yearly_logs" gorm:"type:jsonb"`
}

// 为PostgreSQL表设置表名
func (NodeTrafficLogsPG) TableName() string {
	return "node_traffic_logs"
}

// 时间序列数据的结构定义 - 用于JSONB字段
type TrafficLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Traffic   int64     `json:"traffic"`
}

type DailyLogEntry struct {
	Date    string `json:"date"`
	Traffic int64  `json:"traffic"`
}

type MonthlyLogEntry struct {
	Month   string `json:"month"`
	Traffic int64  `json:"traffic"`
}

type YearlyLogEntry struct {
	Year    string `json:"year"`
	Traffic int64  `json:"traffic"`
}

// JSONB字段的辅助类型
type HourlyLogs []TrafficLogEntry
type DailyLogs []DailyLogEntry
type MonthlyLogs []MonthlyLogEntry
type YearlyLogs []YearlyLogEntry

// 实现database/sql/driver.Valuer接口，用于写入数据库
func (h HourlyLogs) Value() (driver.Value, error) {
	return json.Marshal(h)
}

func (d DailyLogs) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (m MonthlyLogs) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (y YearlyLogs) Value() (driver.Value, error) {
	return json.Marshal(y)
}

// 实现sql.Scanner接口，用于从数据库读取数据
func (h *HourlyLogs) Scan(value interface{}) error {
	if value == nil {
		*h = HourlyLogs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, h)
	case string:
		return json.Unmarshal([]byte(v), h)
	}
	return nil
}

func (d *DailyLogs) Scan(value interface{}) error {
	if value == nil {
		*d = DailyLogs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, d)
	case string:
		return json.Unmarshal([]byte(v), d)
	}
	return nil
}

func (m *MonthlyLogs) Scan(value interface{}) error {
	if value == nil {
		*m = MonthlyLogs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	}
	return nil
}

func (y *YearlyLogs) Scan(value interface{}) error {
	if value == nil {
		*y = YearlyLogs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, y)
	case string:
		return json.Unmarshal([]byte(v), y)
	}
	return nil
}

// 数据迁移辅助结构
type MigrationStats struct {
	UserRecordsMigrated       int64     `json:"user_records_migrated"`
	NodeRecordsMigrated       int64     `json:"node_records_migrated"`
	DomainRecordsMigrated     int64     `json:"domain_records_migrated"`
	SubscriptionNodesMigrated int64     `json:"subscription_nodes_migrated"`
	StartTime                 time.Time `json:"start_time"`
	EndTime                   time.Time `json:"end_time"`
	Errors                    []string  `json:"errors"`
}

// PostgreSQL版本的域名证书过期信息模型
type ExpiryCheckDomainInfoPG struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Domain       string    `json:"domain" gorm:"not null;index"`
	Remark       string    `json:"remark"`
	ExpiredDate  string    `json:"expired_date"`
	DaysToExpire int       `json:"days_to_expire"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// 为PostgreSQL表设置表名
func (ExpiryCheckDomainInfoPG) TableName() string {
	return "expiry_check_domains"
}

// PostgreSQL版本的订阅节点模型
type SubscriptionNodePG struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Type         string    `json:"type" gorm:"type:varchar(50);check:type IN ('reality','hysteria2','vlessCDN');not null"`
	Remark       string    `json:"remark" gorm:"uniqueIndex;not null"`
	Domain       string    `json:"domain" gorm:"not null;index"`
	IP           string    `json:"ip" gorm:"type:inet"`
	SNI          string    `json:"sni"`
	UUID         string    `json:"uuid" gorm:"index"`
	Path         string    `json:"path"`
	ServerPort   string    `json:"server_port"`
	Password     string    `json:"password"`
	PublicKey    string    `json:"public_key"`
	ShortID      string    `json:"short_id"`
	EnableOpenai bool      `json:"enable_openai" gorm:"default:false"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// 为PostgreSQL表设置表名
func (SubscriptionNodePG) TableName() string {
	return "subscription_nodes"
}
