/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

// JSON数据结构定义
type PaymentJSON struct {
	ID      string `json:"id"`
	Comment string `json:"comment"`
	Payment struct {
		ReceivedDate string  `json:"received_date"`
		Amount       float64 `json:"amount"`
		PeriodStart  string  `json:"period_start"`
		PeriodEnd    string  `json:"period_end"`
		Payer        string  `json:"payer,omitempty"` // 可选字段
	} `json:"payment"`
}

var (
	dbType   string
	jsonFile string
)

// addpaymentrecordsCmd represents the addpaymentrecords command
var addpaymentrecordsCmd = &cobra.Command{
	Use:   "addpaymentrecords",
	Short: "从JSON文件导入付款记录到数据库",
	Long: `从JSON文件导入付款记录到MongoDB或PostgreSQL数据库。

该命令会：
1. 读取JSON文件中的付款记录
2. 验证用户是否存在
3. 计算服务天数和每日分摊金额
4. 创建payment_records和daily_payment_allocation记录

示例：
  # 导入到MongoDB
  ./logv2fs addpaymentrecords --db-type=mongodb --file=payment_records.json
  
  # 导入到PostgreSQL  
  ./logv2fs addpaymentrecords --db-type=postgres --file=payment_records.json`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := importPaymentRecords(); err != nil {
			log.Fatalf("导入付款记录失败: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addpaymentrecordsCmd)

	// 添加命令行参数
	addpaymentrecordsCmd.Flags().StringVar(&dbType, "db-type", "", "数据库类型 (mongodb 或 postgres)")
	addpaymentrecordsCmd.Flags().StringVar(&jsonFile, "file", "payment_records.json", "JSON文件路径")

	// 标记必需参数
	addpaymentrecordsCmd.MarkFlagRequired("db-type")
}

// 主要导入逻辑
func importPaymentRecords() error {
	// 验证数据库类型
	if dbType != "mongodb" && dbType != "postgres" {
		return fmt.Errorf("无效的数据库类型: %s，请选择 mongodb 或 postgres", dbType)
	}

	// 读取JSON文件
	log.Printf("正在读取JSON文件: %s", jsonFile)
	jsonData, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("读取JSON文件失败: %v", err)
	}

	// 解析JSON数据
	var payments []PaymentJSON
	if err := json.Unmarshal(jsonData, &payments); err != nil {
		return fmt.Errorf("解析JSON数据失败: %v", err)
	}

	log.Printf("找到 %d 条付款记录", len(payments))

	// 初始化数据库连接
	if err := initializeDatabase(); err != nil {
		return fmt.Errorf("初始化数据库连接失败: %v", err)
	}

	// 统计信息
	var successCount, skipCount int
	var skippedIDs []string

	// 处理每条记录
	for i, payment := range payments {
		log.Printf("正在处理第 %d/%d 条记录，ID: %s", i+1, len(payments), payment.ID)

		success, err := processPaymentRecord(payment)
		if err != nil {
			log.Printf("处理记录 %s 失败: %v", payment.ID, err)
			continue
		}

		if success {
			successCount++
			log.Printf("✓ 记录 %s 导入成功", payment.ID)
		} else {
			skipCount++
			skippedIDs = append(skippedIDs, payment.ID)
			log.Printf("⚠ 记录 %s 被跳过（用户不存在）", payment.ID)
		}
	}

	// 输出结果统计
	log.Printf("\n=== 导入完成 ===")
	log.Printf("成功导入: %d 条", successCount)
	log.Printf("跳过记录: %d 条", skipCount)

	if len(skippedIDs) > 0 {
		log.Printf("跳过的ID列表: %v", skippedIDs)
	}

	return nil
}

// 初始化数据库连接
func initializeDatabase() error {
	switch dbType {
	case "mongodb":
		// 延迟初始化MongoDB连接
		if database.Client == nil {
			log.Println("正在初始化MongoDB连接...")
			// 重新初始化MongoDB客户端
			client := database.DBinstance()
			if client == nil {
				return fmt.Errorf("MongoDB连接初始化失败")
			}
			database.Client = client
		}
		log.Println("使用MongoDB数据库")

	case "postgres":
		// 初始化PostgreSQL连接
		log.Println("正在初始化PostgreSQL连接...")
		db := database.GetPostgresDB()
		if db == nil {
			return fmt.Errorf("PostgreSQL连接初始化失败")
		}
		log.Println("使用PostgreSQL数据库")
	}

	return nil
}

// 处理单条付款记录
func processPaymentRecord(payment PaymentJSON) (bool, error) {
	// 检查用户是否存在并获取用户名
	userName, exists, err := checkUserExists(payment.ID)
	if err != nil {
		return false, fmt.Errorf("检查用户存在性失败: %v", err)
	}

	if !exists {
		return false, nil // 用户不存在，跳过但不报错
	}

	// 解析日期
	receivedDate, err := time.Parse("2006/01/02", payment.Payment.ReceivedDate)
	if err != nil {
		return false, fmt.Errorf("解析received_date失败: %v", err)
	}

	startDate, err := time.Parse("2006/01/02", payment.Payment.PeriodStart)
	if err != nil {
		return false, fmt.Errorf("解析period_start失败: %v", err)
	}

	endDate, err := time.Parse("2006/01/02", payment.Payment.PeriodEnd)
	if err != nil {
		return false, fmt.Errorf("解析period_end失败: %v", err)
	}

	// 计算服务天数和每日分摊金额
	serviceDays := int(endDate.Sub(startDate).Hours()/24) + 1
	dailyAmount := payment.Payment.Amount / float64(serviceDays)

	// 根据数据库类型执行不同的插入逻辑
	switch dbType {
	case "mongodb":
		return insertToMongoDB(payment, userName, receivedDate, startDate, endDate, serviceDays, dailyAmount)
	case "postgres":
		return insertToPostgreSQL(payment, userName, receivedDate, startDate, endDate, serviceDays, dailyAmount)
	}

	return false, fmt.Errorf("未知的数据库类型")
}

// 检查用户是否存在并获取用户名
func checkUserExists(userEmailAsId string) (string, bool, error) {
	switch dbType {
	case "mongodb":
		return checkUserExistsMongoDB(userEmailAsId)
	case "postgres":
		return checkUserExistsPostgreSQL(userEmailAsId)
	}
	return "", false, fmt.Errorf("未知的数据库类型")
}

// MongoDB - 检查用户是否存在
func checkUserExistsMongoDB(userEmailAsId string) (string, bool, error) {
	collection := database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")

	var user struct {
		Name string `bson:"name"`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"email_as_id": userEmailAsId}).Decode(&user)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return "", false, nil // 用户不存在
		}
		return "", false, err // 其他错误
	}

	userName := user.Name
	if userName == "" {
		userName = userEmailAsId // 如果没有名称，使用邮箱
	}

	return userName, true, nil
}

// PostgreSQL - 检查用户是否存在
func checkUserExistsPostgreSQL(userEmailAsId string) (string, bool, error) {
	var user struct {
		Name string `gorm:"column:name"`
	}

	err := database.GetPostgresDB().Table("user_traffic_logs").
		Select("name").
		Where("email_as_id = ?", userEmailAsId).
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil // 用户不存在
		}
		return "", false, err // 其他错误
	}

	userName := user.Name
	if userName == "" {
		userName = userEmailAsId // 如果没有名称，使用邮箱
	}

	return userName, true, nil
}

// MongoDB - 插入数据
func insertToMongoDB(payment PaymentJSON, userName string, receivedDate, startDate, endDate time.Time, serviceDays int, dailyAmount float64) (bool, error) {
	// 创建付款记录
	paymentRecord := model.PaymentRecord{
		ID:            primitive.NewObjectID(),
		UserEmailAsId: payment.ID,
		UserName:      userName,
		Amount:        payment.Payment.Amount,
		StartDate:     startDate,
		EndDate:       endDate,
		DailyAmount:   dailyAmount,
		ServiceDays:   serviceDays,
		Remark:        payment.Comment,
		OperatorEmail: "admin",
		OperatorName:  "admin",
		CreatedAt:     receivedDate,
		UpdatedAt:     time.Now(),
	}

	// 插入付款记录
	collection := database.OpenCollection(database.Client, "payment_records")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, paymentRecord)
	if err != nil {
		return false, fmt.Errorf("插入付款记录失败: %v", err)
	}

	// 创建每日分摊记录
	if err := createDailyAllocationsMongoDB(paymentRecord.ID, paymentRecord); err != nil {
		return false, fmt.Errorf("创建每日分摊记录失败: %v", err)
	}

	return true, nil
}

// PostgreSQL - 插入数据
func insertToPostgreSQL(payment PaymentJSON, userName string, receivedDate, startDate, endDate time.Time, serviceDays int, dailyAmount float64) (bool, error) {
	// 创建付款记录
	paymentRecord := model.PaymentRecordPG{
		UserEmailAsId: payment.ID,
		UserName:      userName,
		Amount:        payment.Payment.Amount,
		StartDate:     startDate,
		EndDate:       endDate,
		DailyAmount:   dailyAmount,
		ServiceDays:   serviceDays,
		Remark:        payment.Comment,
		OperatorEmail: "admin",
		OperatorName:  "admin",
		CreatedAt:     receivedDate,
		UpdatedAt:     time.Now(),
	}

	// 开始事务
	tx := database.GetPostgresDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 插入付款记录
	if err := tx.Create(&paymentRecord).Error; err != nil {
		tx.Rollback()
		return false, fmt.Errorf("插入付款记录失败: %v", err)
	}

	// 创建每日分摊记录
	if err := createDailyAllocationsPostgreSQL(tx, paymentRecord.ID, paymentRecord); err != nil {
		tx.Rollback()
		return false, fmt.Errorf("创建每日分摊记录失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return false, fmt.Errorf("提交事务失败: %v", err)
	}

	return true, nil
}

// MongoDB - 创建每日分摊记录
func createDailyAllocationsMongoDB(paymentRecordID primitive.ObjectID, payment model.PaymentRecord) error {
	collection := database.OpenCollection(database.Client, "daily_payment_allocations")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 生成从开始日期到结束日期的每日分摊记录
	current := payment.StartDate
	var allocations []interface{}

	for current.Before(payment.EndDate) || current.Equal(payment.EndDate) {
		allocation := model.DailyPaymentAllocation{
			ID:               primitive.NewObjectID(),
			PaymentRecordID:  paymentRecordID,
			UserEmailAsId:    payment.UserEmailAsId,
			UserName:         payment.UserName,
			Date:             current,
			DateString:       current.Format("20060102"),
			AllocatedAmount:  payment.DailyAmount,
			OriginalAmount:   payment.Amount,
			ServiceStartDate: payment.StartDate,
			ServiceEndDate:   payment.EndDate,
			CreatedAt:        payment.CreatedAt,
		}

		allocations = append(allocations, allocation)
		current = current.AddDate(0, 0, 1) // 增加一天
	}

	// 批量插入分摊记录
	if len(allocations) > 0 {
		_, err := collection.InsertMany(ctx, allocations)
		if err != nil {
			return err
		}
	}

	return nil
}

// PostgreSQL - 创建每日分摊记录
func createDailyAllocationsPostgreSQL(tx *gorm.DB, paymentRecordID uuid.UUID, payment model.PaymentRecordPG) error {
	// 生成从开始日期到结束日期的每日分摊记录
	current := payment.StartDate
	allocations := []model.DailyPaymentAllocationPG{}

	for current.Before(payment.EndDate) || current.Equal(payment.EndDate) {
		allocation := model.DailyPaymentAllocationPG{
			PaymentRecordID:  paymentRecordID,
			UserEmailAsId:    payment.UserEmailAsId,
			UserName:         payment.UserName,
			Date:             current,
			DateString:       current.Format("20060102"),
			AllocatedAmount:  payment.DailyAmount,
			OriginalAmount:   payment.Amount,
			ServiceStartDate: payment.StartDate,
			ServiceEndDate:   payment.EndDate,
			CreatedAt:        payment.CreatedAt,
		}

		allocations = append(allocations, allocation)
		current = current.AddDate(0, 0, 1) // 增加一天
	}

	// 批量插入分摊记录
	if len(allocations) > 0 {
		if err := tx.CreateInBatches(allocations, 100).Error; err != nil {
			return err
		}
	}

	return nil
}
