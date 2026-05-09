package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentRecord 缴费记录（费用领域内唯一的事实表）
//
// 设计取向：
//   - 这里存的是「不可压缩的事实」：谁、付了多少钱、覆盖哪段服务时间。
//   - 任何按天/按月/按年的统计都视为该事实的派生视图，
//     在 GetPaymentStatistics 等查询接口里实时聚合，不再单独物化。
//   - DailyAmount / ServiceDays 是冗余但便利的派生字段，
//     必须在 Add/Update 时与 Amount、StartDate、EndDate 保持一致。
type PaymentRecord struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	UserEmailAsId string             `json:"user_email_as_id" bson:"user_email_as_id" validate:"required"` // 关联的用户邮箱
	UserName      string             `json:"user_name" bson:"user_name"`                                   // 用户名（冗余存储，方便查询）
	Amount        float64            `json:"amount" bson:"amount" validate:"required,min=0"`               // 缴费金额
	StartDate     time.Time          `json:"start_date" bson:"start_date"`                                 // 服务开始日期
	EndDate       time.Time          `json:"end_date" bson:"end_date"`                                     // 服务结束日期
	DailyAmount   float64            `json:"daily_amount" bson:"daily_amount"`                             // 每日分摊金额（派生字段：Amount / ServiceDays）
	ServiceDays   int                `json:"service_days" bson:"service_days"`                             // 服务天数（派生字段，含首尾两端）
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

// PaymentStatistics 费用统计结构（响应载荷）
//
// 注：日级粒度已不再支持，前端只展示月/年。
type PaymentStatistics struct {
	TotalAmount  float64               `json:"total_amount"`  // 总金额（区间内分摊到的金额合计）
	PaymentCount int64                 `json:"payment_count"` // 涉及到的缴费记录笔数（去重后）
	DateRange    string                `json:"date_range"`    // 日期范围描述
	StartDate    time.Time             `json:"start_date"`    // 开始日期
	EndDate      time.Time             `json:"end_date"`      // 结束日期
	MonthlyStats []MonthlyPaymentStats `json:"monthly_stats"` // 每月统计
	YearlyStats  []YearlyPaymentStats  `json:"yearly_stats"`  // 每年统计
}

// MonthlyPaymentStats 每月缴费统计
//
// PaymentCount / UserCount 都是在「该月内」去重后的计数：
// 一笔横跨多月的缴费，在它涉及的每一个月里各计 1 次。
type MonthlyPaymentStats struct {
	Month        string  `json:"month"`         // 月份 (YYYYMM)
	TotalAmount  float64 `json:"total_amount"`  // 当月分摊到的总金额
	PaymentCount int64   `json:"payment_count"` // 当月涉及的缴费记录数（去重）
	UserCount    int64   `json:"user_count"`    // 当月涉及的不同用户数
}

// YearlyPaymentStats 每年缴费统计
type YearlyPaymentStats struct {
	Year         string  `json:"year"`          // 年份 (YYYY)
	TotalAmount  float64 `json:"total_amount"`  // 当年分摊到的总金额
	PaymentCount int64   `json:"payment_count"` // 当年涉及的缴费记录数（去重）
	UserCount    int64   `json:"user_count"`    // 当年涉及的不同用户数
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
