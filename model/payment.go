package model

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentRecord MongoDB版本的缴费记录
type PaymentRecord struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	UserEmailAsId string             `json:"user_email_as_id" bson:"user_email_as_id" validate:"required"` // 关联的用户邮箱
	UserName      string             `json:"user_name" bson:"user_name"`                                   // 用户名（冗余存储，方便查询）
	Amount        float64            `json:"amount" bson:"amount" validate:"required,min=0"`               // 缴费金额
	StartDate     time.Time          `json:"start_date" bson:"start_date"`                                 // 服务开始日期
	EndDate       time.Time          `json:"end_date" bson:"end_date"`                                     // 服务结束日期
	DailyAmount   float64            `json:"daily_amount" bson:"daily_amount"`                             // 每日分摊金额
	ServiceDays   int                `json:"service_days" bson:"service_days"`                             // 服务天数
	Remark        string             `json:"remark" bson:"remark"`                                         // 备注
	OperatorEmail string             `json:"operator_email" bson:"operator_email"`                         // 操作员邮箱（记录是谁录入的）
	OperatorName  string             `json:"operator_name" bson:"operator_name"`                           // 操作员名称
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`
}

// CollectionName 返回MongoDB集合名称
func (PaymentRecord) CollectionName() string {
	return "payment_records"
}

// PaymentRecordPG PostgreSQL版本的缴费记录
type PaymentRecordPG struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserEmailAsId string    `json:"user_email_as_id" gorm:"index;not null"`   // 关联的用户邮箱
	UserName      string    `json:"user_name"`                                // 用户名（冗余存储）
	Amount        float64   `json:"amount" gorm:"not null;check:amount >= 0"` // 缴费金额
	StartDate     time.Time `json:"start_date" gorm:"index"`                  // 服务开始日期
	EndDate       time.Time `json:"end_date" gorm:"index"`                    // 服务结束日期
	DailyAmount   float64   `json:"daily_amount"`                             // 每日分摊金额
	ServiceDays   int       `json:"service_days"`                             // 服务天数
	Remark        string    `json:"remark" gorm:"type:text"`                  // 备注
	OperatorEmail string    `json:"operator_email"`                           // 操作员邮箱
	OperatorName  string    `json:"operator_name"`                            // 操作员名称
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 设置表名
func (PaymentRecordPG) TableName() string {
	return "payment_records"
}

// DailyPaymentAllocation 每日费用分摊记录 - MongoDB版本
type DailyPaymentAllocation struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	PaymentRecordID  primitive.ObjectID `json:"payment_record_id" bson:"payment_record_id"`   // 关联的缴费记录ID
	UserEmailAsId    string             `json:"user_email_as_id" bson:"user_email_as_id"`     // 用户邮箱
	UserName         string             `json:"user_name" bson:"user_name"`                   // 用户名
	Date             time.Time          `json:"date" bson:"date"`                             // 分摊日期
	DateString       string             `json:"date_string" bson:"date_string"`               // 日期字符串 (YYYYMMDD)
	AllocatedAmount  float64            `json:"allocated_amount" bson:"allocated_amount"`     // 分摊金额
	OriginalAmount   float64            `json:"original_amount" bson:"original_amount"`       // 原始总金额
	ServiceStartDate time.Time          `json:"service_start_date" bson:"service_start_date"` // 服务开始日期
	ServiceEndDate   time.Time          `json:"service_end_date" bson:"service_end_date"`     // 服务结束日期
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
}

// CollectionName 返回MongoDB集合名称
func (DailyPaymentAllocation) CollectionName() string {
	return "daily_payment_allocations"
}

// DailyPaymentAllocationPG 每日费用分摊记录 - PostgreSQL版本
type DailyPaymentAllocationPG struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PaymentRecordID  uuid.UUID `json:"payment_record_id" gorm:"index;not null"` // 关联的缴费记录ID
	UserEmailAsId    string    `json:"user_email_as_id" gorm:"index;not null"`  // 用户邮箱
	UserName         string    `json:"user_name"`                               // 用户名
	Date             time.Time `json:"date" gorm:"index"`                       // 分摊日期
	DateString       string    `json:"date_string" gorm:"index"`                // 日期字符串 (YYYYMMDD)
	AllocatedAmount  float64   `json:"allocated_amount"`                        // 分摊金额
	OriginalAmount   float64   `json:"original_amount"`                         // 原始总金额
	ServiceStartDate time.Time `json:"service_start_date"`                      // 服务开始日期
	ServiceEndDate   time.Time `json:"service_end_date"`                        // 服务结束日期
	CreatedAt        time.Time `json:"created_at"`
}

func (DailyPaymentAllocationPG) TableName() string {
	return "daily_payment_allocations"
}

// PaymentStatistics 费用统计结构
type PaymentStatistics struct {
	TotalAmount  float64               `json:"total_amount"`  // 总金额
	PaymentCount int64                 `json:"payment_count"` // 缴费次数
	DateRange    string                `json:"date_range"`    // 日期范围描述
	StartDate    time.Time             `json:"start_date"`    // 开始日期
	EndDate      time.Time             `json:"end_date"`      // 结束日期
	DailyStats   []DailyPaymentStats   `json:"daily_stats"`   // 每日统计
	MonthlyStats []MonthlyPaymentStats `json:"monthly_stats"` // 每月统计
	YearlyStats  []YearlyPaymentStats  `json:"yearly_stats"`  // 每年统计
}

// DailyPaymentStats 每日缴费统计
type DailyPaymentStats struct {
	Date         string  `json:"date"`          // 日期 (YYYYMMDD)
	TotalAmount  float64 `json:"total_amount"`  // 当日总金额
	PaymentCount int64   `json:"payment_count"` // 当日缴费次数
	UserCount    int64   `json:"user_count"`    // 当日缴费用户数
}

// MonthlyPaymentStats 每月缴费统计
type MonthlyPaymentStats struct {
	Month        string  `json:"month"`         // 月份 (YYYYMM)
	TotalAmount  float64 `json:"total_amount"`  // 当月总金额
	PaymentCount int64   `json:"payment_count"` // 当月缴费次数
	UserCount    int64   `json:"user_count"`    // 当月缴费用户数
}

// YearlyPaymentStats 每年缴费统计
type YearlyPaymentStats struct {
	Year         string  `json:"year"`          // 年份 (YYYY)
	TotalAmount  float64 `json:"total_amount"`  // 当年总金额
	PaymentCount int64   `json:"payment_count"` // 当年缴费次数
	UserCount    int64   `json:"user_count"`    // 当年缴费用户数
}

// UserPaymentSummary 用户缴费汇总
type UserPaymentSummary struct {
	UserEmailAsId   string          `json:"user_email_as_id"`  // 用户邮箱
	UserName        string          `json:"user_name"`         // 用户名
	TotalAmount     float64         `json:"total_amount"`      // 总缴费金额
	PaymentCount    int64           `json:"payment_count"`     // 缴费次数
	LastPaymentDate time.Time       `json:"last_payment_date"` // 最后缴费日期
	PaymentHistory  []PaymentRecord `json:"payment_history"`   // 缴费历史记录
}

// PaymentStatisticsPG PostgreSQL版本的统计数据（用于存储预计算的统计结果）
type PaymentStatisticsPG struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StatType     string    `json:"stat_type" gorm:"type:varchar(20);index"` // daily, monthly, yearly
	StatDate     string    `json:"stat_date" gorm:"index"`                  // 统计日期
	TotalAmount  float64   `json:"total_amount"`                            // 总金额
	PaymentCount int64     `json:"payment_count"`                           // 缴费次数
	UserCount    int64     `json:"user_count"`                              // 缴费用户数
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (PaymentStatisticsPG) TableName() string {
	return "payment_statistics"
}
