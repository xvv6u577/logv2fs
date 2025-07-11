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

// JSON数据结构定义 - 支持多个付款记录
type PaymentJSON struct {
	ID       string        `json:"id"`
	Comment  string        `json:"remark"`   // 修改为使用remark字段
	Payments []PaymentInfo `json:"payments"` // 改为数组支持多个付款记录
}

// 单个付款信息结构
type PaymentInfo struct {
	ReceivedDate string  `json:"received_date"`
	Amount       float64 `json:"amount"`
	PeriodStart  string  `json:"period_start"`
	PeriodEnd    string  `json:"period_end"`
	Payer        string  `json:"payer,omitempty"` // 可选字段
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
3. 更新用户的备注信息
4. 检查重复付款记录
5. 计算服务天数和每日分摊金额
6. 创建payment_records和daily_payment_allocation记录

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
	var users []PaymentJSON
	if err := json.Unmarshal(jsonData, &users); err != nil {
		return fmt.Errorf("解析JSON数据失败: %v", err)
	}

	// 统计总的付款记录数量
	totalPayments := 0
	for _, user := range users {
		totalPayments += len(user.Payments)
	}

	log.Printf("找到 %d 个用户，共 %d 条付款记录", len(users), totalPayments)

	// 初始化数据库连接
	if err := initializeDatabase(); err != nil {
		return fmt.Errorf("初始化数据库连接失败: %v", err)
	}

	// 统计信息
	var (
		successCount       int      // 成功导入的付款记录数
		skipCount          int      // 跳过的付款记录数（用户不存在）
		duplicateCount     int      // 重复的付款记录数
		errorCount         int      // 错误的付款记录数
		skippedUsers       []string // 跳过的用户ID列表（用户不存在）
		duplicateRecords   []string // 重复的付款记录列表
		errorRecords       []string // 错误的付款记录列表
		remarkUpdatedCount int      // 更新备注的用户数量
	)

	// 处理每个用户的付款记录
	for userIndex, user := range users {
		log.Printf("正在处理用户 %d/%d，ID: %s，共 %d 条付款记录",
			userIndex+1, len(users), user.ID, len(user.Payments))

		// 检查用户是否存在
		userName, exists, err := checkUserExists(user.ID)
		if err != nil {
			log.Printf("❌ 检查用户 %s 存在性失败: %v", user.ID, err)
			errorCount += len(user.Payments)
			for i := range user.Payments {
				errorRecords = append(errorRecords, fmt.Sprintf("%s[%d]", user.ID, i+1))
			}
			continue
		}

		if !exists {
			log.Printf("⚠ 用户 %s 不存在，跳过该用户的所有 %d 条付款记录", user.ID, len(user.Payments))
			skipCount += len(user.Payments)
			skippedUsers = append(skippedUsers, user.ID)
			continue
		}

		// 更新用户备注信息
		if user.Comment != "" {
			if err := updateUserRemark(user.ID, user.Comment); err != nil {
				log.Printf("⚠ 更新用户 %s 备注失败: %v", user.ID, err)
			} else {
				log.Printf("✓ 已更新用户 %s 的备注信息", user.ID)
				remarkUpdatedCount++
			}
		}

		// 处理该用户的每条付款记录
		for paymentIndex, payment := range user.Payments {
			log.Printf("  正在处理付款记录 %d/%d (用户: %s)",
				paymentIndex+1, len(user.Payments), user.ID)

			result, err := processPaymentRecord(user.ID, userName, user.Comment, payment, paymentIndex+1)

			switch result {
			case "success":
				successCount++
				log.Printf("  ✓ 付款记录 %d 导入成功", paymentIndex+1)
			case "duplicate":
				duplicateCount++
				duplicateRecords = append(duplicateRecords, fmt.Sprintf("%s[%d]", user.ID, paymentIndex+1))
				log.Printf("  ⚠ 付款记录 %d 已存在，跳过", paymentIndex+1)
			case "error":
				errorCount++
				errorRecords = append(errorRecords, fmt.Sprintf("%s[%d]", user.ID, paymentIndex+1))
				log.Printf("  ❌ 付款记录 %d 处理失败: %v", paymentIndex+1, err)
			}
		}
	}

	// 输出结果统计
	log.Printf("\n=== 导入完成 ===")
	log.Printf("总用户数量: %d", len(users))
	log.Printf("总付款记录: %d", totalPayments)
	log.Printf("成功导入: %d 条", successCount)
	log.Printf("重复跳过: %d 条", duplicateCount)
	log.Printf("用户不存在跳过: %d 条", skipCount)
	log.Printf("错误失败: %d 条", errorCount)
	log.Printf("更新备注的用户: %d 个", remarkUpdatedCount)

	if len(skippedUsers) > 0 {
		log.Printf("不存在的用户ID: %v", skippedUsers)
	}

	if len(duplicateRecords) > 0 {
		log.Printf("重复的付款记录: %v", duplicateRecords)
	}

	if len(errorRecords) > 0 {
		log.Printf("错误的付款记录: %v", errorRecords)
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
func processPaymentRecord(userID, userName, comment string, payment PaymentInfo, paymentIndex int) (string, error) {
	// 解析日期
	receivedDate, err := time.Parse("2006/01/02", payment.ReceivedDate)
	if err != nil {
		return "error", fmt.Errorf("解析received_date失败: %v", err)
	}

	startDate, err := time.Parse("2006/01/02", payment.PeriodStart)
	if err != nil {
		return "error", fmt.Errorf("解析period_start失败: %v", err)
	}

	endDate, err := time.Parse("2006/01/02", payment.PeriodEnd)
	if err != nil {
		return "error", fmt.Errorf("解析period_end失败: %v", err)
	}

	// 检查重复付款记录
	isDuplicate, err := checkDuplicatePayment(userID, startDate, endDate)
	if err != nil {
		return "error", fmt.Errorf("检查重复付款记录失败: %v", err)
	}

	if isDuplicate {
		return "duplicate", nil // 重复记录，跳过但不报错
	}

	// 计算服务天数和每日分摊金额
	serviceDays := int(endDate.Sub(startDate).Hours()/24) + 1
	dailyAmount := payment.Amount / float64(serviceDays)

	// 根据数据库类型执行不同的插入逻辑
	switch dbType {
	case "mongodb":
		return insertToMongoDB(userID, userName, comment, receivedDate, startDate, endDate, serviceDays, dailyAmount, payment.Amount)
	case "postgres":
		return insertToPostgreSQL(userID, userName, comment, receivedDate, startDate, endDate, serviceDays, dailyAmount, payment.Amount)
	}

	return "error", fmt.Errorf("未知的数据库类型")
}

// 检查重复付款记录
func checkDuplicatePayment(userEmailAsId string, startDate, endDate time.Time) (bool, error) {
	switch dbType {
	case "mongodb":
		return checkDuplicatePaymentMongoDB(userEmailAsId, startDate, endDate)
	case "postgres":
		return checkDuplicatePaymentPostgreSQL(userEmailAsId, startDate, endDate)
	}
	return false, fmt.Errorf("未知的数据库类型")
}

// MongoDB - 检查重复付款记录
func checkDuplicatePaymentMongoDB(userEmailAsId string, startDate, endDate time.Time) (bool, error) {
	collection := database.OpenCollection(database.Client, "payment_records")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 查找相同用户ID和相同时间段的记录
	filter := bson.M{
		"user_email_as_id": userEmailAsId,
		"start_date":       startDate,
		"end_date":         endDate,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// PostgreSQL - 检查重复付款记录
func checkDuplicatePaymentPostgreSQL(userEmailAsId string, startDate, endDate time.Time) (bool, error) {
	var count int64

	err := database.GetPostgresDB().Table("payment_records").
		Where("user_email_as_id = ? AND start_date = ? AND end_date = ?", userEmailAsId, startDate, endDate).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
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
func insertToMongoDB(userID, userName, comment string, receivedDate, startDate, endDate time.Time, serviceDays int, dailyAmount, originalAmount float64) (string, error) {
	// 创建付款记录
	paymentRecord := model.PaymentRecord{
		ID:            primitive.NewObjectID(),
		UserEmailAsId: userID,
		UserName:      userName,
		Amount:        originalAmount,
		StartDate:     startDate,
		EndDate:       endDate,
		DailyAmount:   dailyAmount,
		ServiceDays:   serviceDays,
		Remark:        comment,
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
		return "error", fmt.Errorf("插入付款记录失败: %v", err)
	}

	// 创建每日分摊记录
	if err := createDailyAllocationsMongoDB(paymentRecord.ID, paymentRecord); err != nil {
		return "error", fmt.Errorf("创建每日分摊记录失败: %v", err)
	}

	return "success", nil
}

// PostgreSQL - 插入数据
func insertToPostgreSQL(userID, userName, comment string, receivedDate, startDate, endDate time.Time, serviceDays int, dailyAmount, originalAmount float64) (string, error) {
	// 创建付款记录
	paymentRecord := model.PaymentRecordPG{
		UserEmailAsId: userID,
		UserName:      userName,
		Amount:        originalAmount,
		StartDate:     startDate,
		EndDate:       endDate,
		DailyAmount:   dailyAmount,
		ServiceDays:   serviceDays,
		Remark:        comment,
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
		return "error", fmt.Errorf("插入付款记录失败: %v", err)
	}

	// 创建每日分摊记录
	if err := createDailyAllocationsPostgreSQL(tx, paymentRecord.ID, paymentRecord); err != nil {
		tx.Rollback()
		return "error", fmt.Errorf("创建每日分摊记录失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return "error", fmt.Errorf("提交事务失败: %v", err)
	}

	return "success", nil
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

// 更新用户备注信息
func updateUserRemark(userEmailAsId, newRemark string) error {
	switch dbType {
	case "mongodb":
		return updateUserRemarkMongoDB(userEmailAsId, newRemark)
	case "postgres":
		return updateUserRemarkPostgreSQL(userEmailAsId, newRemark)
	}
	return fmt.Errorf("未知的数据库类型")
}

// MongoDB - 更新用户备注
func updateUserRemarkMongoDB(userEmailAsId, newRemark string) error {
	collection := database.OpenCollection(database.Client, "USER_TRAFFIC_LOGS")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先获取现有备注
	var existingUser struct {
		Remark string `bson:"remark"`
	}

	err := collection.FindOne(ctx, bson.M{"email_as_id": userEmailAsId}).Decode(&existingUser)
	if err != nil {
		return fmt.Errorf("查找用户失败: %v", err)
	}

	// 合并备注信息
	var finalRemark string
	if existingUser.Remark != "" {
		// 如果现有备注不包含新备注，则追加
		if existingUser.Remark != newRemark {
			finalRemark = existingUser.Remark + "; " + newRemark
		} else {
			finalRemark = existingUser.Remark // 已存在相同备注，不变
		}
	} else {
		finalRemark = newRemark
	}

	// 更新备注
	filter := bson.M{"email_as_id": userEmailAsId}
	update := bson.M{
		"$set": bson.M{
			"remark":     finalRemark,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("更新用户备注失败: %v", err)
	}

	return nil
}

// PostgreSQL - 更新用户备注
func updateUserRemarkPostgreSQL(userEmailAsId, newRemark string) error {
	db := database.GetPostgresDB()

	// 先获取现有备注
	var existingRemark string
	err := db.Table("user_traffic_logs").
		Select("remark").
		Where("email_as_id = ?", userEmailAsId).
		Scan(&existingRemark).Error

	if err != nil {
		return fmt.Errorf("查找用户失败: %v", err)
	}

	// 合并备注信息
	var finalRemark string
	if existingRemark != "" {
		// 如果现有备注不包含新备注，则追加
		if existingRemark != newRemark {
			finalRemark = existingRemark + "; " + newRemark
		} else {
			finalRemark = existingRemark // 已存在相同备注，不变
		}
	} else {
		finalRemark = newRemark
	}

	// 更新备注
	err = db.Table("user_traffic_logs").
		Where("email_as_id = ?", userEmailAsId).
		Updates(map[string]interface{}{
			"remark":     finalRemark,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		return fmt.Errorf("更新用户备注失败: %v", err)
	}

	return nil
}
