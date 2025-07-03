package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"gorm.io/gorm"
)

// AddPaymentRecordPG 添加缴费记录 - PostgreSQL版本
func AddPaymentRecordPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查权限
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var req struct {
			UserEmailAsId string  `json:"user_email_as_id" binding:"required"`
			Amount        float64 `json:"amount" binding:"required,min=0"`
			StartDate     string  `json:"start_date" binding:"required"`
			EndDate       string  `json:"end_date" binding:"required"`
			Remark        string  `json:"remark"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 解析日期
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期格式"})
			return
		}

		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期格式"})
			return
		}

		if endDate.Before(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
			return
		}

		// 计算服务天数（包含结束日期）
		serviceDays := int(endDate.Sub(startDate).Hours()/24) + 1
		dailyAmount := req.Amount / float64(serviceDays)

		// 获取用户信息
		userEmail, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证用户"})
			return
		}

		userName, _ := c.Get("username")

		// 安全转换字符串
		operatorEmail := ""
		if userEmail != nil {
			operatorEmail = userEmail.(string)
		}

		operatorName := ""
		if userName != nil {
			operatorName = userName.(string)
		}

		// 创建缴费记录
		paymentRecord := model.PaymentRecordPG{
			UserEmailAsId: req.UserEmailAsId,
			UserName:      getUserNameByEmailPG(req.UserEmailAsId), // 获取被充值用户名
			Amount:        req.Amount,
			StartDate:     startDate,
			EndDate:       endDate,
			DailyAmount:   dailyAmount,
			ServiceDays:   serviceDays,
			Remark:        req.Remark,
			OperatorEmail: operatorEmail,
			OperatorName:  operatorName,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// 开始事务
		tx := database.GetPostgresDB().Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 插入缴费记录
		if err := tx.Create(&paymentRecord).Error; err != nil {
			tx.Rollback()
			log.Printf("添加缴费记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加缴费记录失败"})
			return
		}

		// 创建每日分摊记录
		if err := createDailyAllocationsPG(tx, paymentRecord.ID, paymentRecord); err != nil {
			tx.Rollback()
			log.Printf("创建每日分摊记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建每日分摊记录失败"})
			return
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			log.Printf("提交事务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "缴费记录添加成功",
			"payment_id":   paymentRecord.ID,
			"service_days": serviceDays,
			"daily_amount": dailyAmount,
		})
	}
}

// 创建每日分摊记录 - PostgreSQL版本
func createDailyAllocationsPG(tx *gorm.DB, paymentRecordID uuid.UUID, payment model.PaymentRecordPG) error {
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
			CreatedAt:        time.Now(),
		}

		allocations = append(allocations, allocation)
		current = current.AddDate(0, 0, 1) // 增加一天
	}

	// 批量插入分摊记录
	if err := tx.CreateInBatches(allocations, 100).Error; err != nil {
		return err
	}

	return nil
}

// 获取用户名 - PostgreSQL版本
func getUserNameByEmailPG(email string) string {
	var user struct {
		Username string `gorm:"column:username"`
	}

	err := database.GetPostgresDB().Table("users").Select("username").Where("email = ?", email).First(&user).Error
	if err != nil {
		return email // 如果找不到用户名，返回邮箱
	}

	return user.Username
}

// GetPaymentStatisticsPG 获取费用统计 - PostgreSQL版本
func GetPaymentStatisticsPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查权限
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")
		statType := c.DefaultQuery("type", "daily") // daily, monthly, yearly, overall

		var startDate, endDate time.Time
		var err error

		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期格式"})
				return
			}
		} else {
			// 默认为30天前
			startDate = time.Now().AddDate(0, 0, -30)
		}

		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期格式"})
				return
			}
		} else {
			// 默认为今天
			endDate = time.Now()
		}

		// 设置查询范围：startDate 00:00:00 到 endDate 23:59:59
		startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

		stats := model.PaymentStatistics{
			StartDate: startDateTime,
			EndDate:   endDateTime,
			DateRange: fmt.Sprintf("%s 至 %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		}

		switch statType {
		case "daily":
			dailyStats, err := getDailyStatsPG(startDateTime, endDateTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取每日统计失败"})
				return
			}
			stats.DailyStats = dailyStats

			// 计算总计
			for _, daily := range dailyStats {
				stats.TotalAmount += daily.TotalAmount
				stats.PaymentCount += daily.PaymentCount
			}

		case "monthly":
			monthlyStats, err := getMonthlyStatsPG(startDateTime, endDateTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取每月统计失败"})
				return
			}
			stats.MonthlyStats = monthlyStats

			// 计算总计
			for _, monthly := range monthlyStats {
				stats.TotalAmount += monthly.TotalAmount
				stats.PaymentCount += monthly.PaymentCount
			}

		case "yearly":
			yearlyStats, err := getYearlyStatsPG(startDateTime, endDateTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "获取每年统计失败"})
				return
			}
			stats.YearlyStats = yearlyStats

			// 计算总计
			for _, yearly := range yearlyStats {
				stats.TotalAmount += yearly.TotalAmount
				stats.PaymentCount += yearly.PaymentCount
			}

		case "overall":
			// 获取所有类型的统计
			dailyStats, _ := getDailyStatsPG(startDateTime, endDateTime)
			monthlyStats, _ := getMonthlyStatsPG(startDateTime, endDateTime)
			yearlyStats, _ := getYearlyStatsPG(startDateTime, endDateTime)

			stats.DailyStats = dailyStats
			stats.MonthlyStats = monthlyStats
			stats.YearlyStats = yearlyStats

			// 基于日统计计算总计（避免重复计算）
			for _, daily := range dailyStats {
				stats.TotalAmount += daily.TotalAmount
				stats.PaymentCount += daily.PaymentCount
			}
		}

		c.JSON(http.StatusOK, stats)
	}
}

// 获取每日统计 - PostgreSQL版本
func getDailyStatsPG(startDate, endDate time.Time) ([]model.DailyPaymentStats, error) {
	var results []model.DailyPaymentStats

	query := `
		SELECT 
			date_string as date,
			SUM(allocated_amount) as total_amount,
			COUNT(*) as payment_count,
			COUNT(DISTINCT user_email_as_id) as user_count
		FROM daily_payment_allocations 
		WHERE date >= ? AND date <= ?
		GROUP BY date_string
		ORDER BY date_string
	`

	rows, err := database.GetPostgresDB().Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.DailyPaymentStats
		if err := rows.Scan(&stat.Date, &stat.TotalAmount, &stat.PaymentCount, &stat.UserCount); err != nil {
			continue
		}
		results = append(results, stat)
	}

	return results, nil
}

// 获取每月统计 - PostgreSQL版本
func getMonthlyStatsPG(startDate, endDate time.Time) ([]model.MonthlyPaymentStats, error) {
	var results []model.MonthlyPaymentStats

	query := `
		SELECT 
			TO_CHAR(date, 'YYYYMM') as month,
			SUM(allocated_amount) as total_amount,
			COUNT(*) as payment_count,
			COUNT(DISTINCT user_email_as_id) as user_count
		FROM daily_payment_allocations 
		WHERE date >= ? AND date <= ?
		GROUP BY TO_CHAR(date, 'YYYYMM')
		ORDER BY month
	`

	rows, err := database.GetPostgresDB().Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.MonthlyPaymentStats
		if err := rows.Scan(&stat.Month, &stat.TotalAmount, &stat.PaymentCount, &stat.UserCount); err != nil {
			continue
		}
		results = append(results, stat)
	}

	return results, nil
}

// 获取每年统计 - PostgreSQL版本
func getYearlyStatsPG(startDate, endDate time.Time) ([]model.YearlyPaymentStats, error) {
	var results []model.YearlyPaymentStats

	query := `
		SELECT 
			TO_CHAR(date, 'YYYY') as year,
			SUM(allocated_amount) as total_amount,
			COUNT(*) as payment_count,
			COUNT(DISTINCT user_email_as_id) as user_count
		FROM daily_payment_allocations 
		WHERE date >= ? AND date <= ?
		GROUP BY TO_CHAR(date, 'YYYY')
		ORDER BY year
	`

	rows, err := database.GetPostgresDB().Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat model.YearlyPaymentStats
		if err := rows.Scan(&stat.Year, &stat.TotalAmount, &stat.PaymentCount, &stat.UserCount); err != nil {
			continue
		}
		results = append(results, stat)
	}

	return results, nil
}

// GetUserPaymentsPG 获取用户缴费记录 - PostgreSQL版本
func GetUserPaymentsPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		userEmail := c.Param("email")
		if userEmail == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户邮箱不能为空"})
			return
		}

		// 权限检查 - 管理员可以查看所有人的，普通用户只能查看自己的
		if err := helper.MatchUserTypeAndName(c, userEmail); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 查询该用户的所有缴费记录
		var payments []model.PaymentRecordPG
		if err := database.GetPostgresDB().Where("user_email_as_id = ?", userEmail).Order("start_date DESC").Find(&payments).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询缴费记录失败"})
			log.Printf("Query payment records error: %v", err)
			return
		}

		// 计算总金额
		var totalAmount float64
		for _, p := range payments {
			totalAmount += p.Amount
		}

		c.JSON(http.StatusOK, gin.H{
			"payments":      payments,
			"total_amount":  totalAmount,
			"payment_count": len(payments),
		})
	}
}

// GetPaymentRecordsPG 获取缴费记录列表 - PostgreSQL版本
func GetPaymentRecordsPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		userEmail := c.Query("user_email")

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		// 构建查询
		query := database.GetPostgresDB().Model(&model.PaymentRecordPG{})
		if userEmail != "" {
			query = query.Where("user_email_as_id = ?", userEmail)
		}

		// 计算总数
		var total int64
		if err := query.Count(&total).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取记录总数失败"})
			return
		}

		// 查询数据
		var records []model.PaymentRecordPG
		if err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&records).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询缴费记录失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"records": records,
			"total":   total,
			"page":    page,
			"limit":   limit,
		})
	}
}

// DeletePaymentRecordPG 删除缴费记录 - PostgreSQL版本（同时删除相关的每日分摊记录）
func DeletePaymentRecordPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查权限
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		idStr := c.Param("id")
		paymentID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
			return
		}

		// 开始事务
		tx := database.GetPostgresDB().Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 删除每日分摊记录
		if err := tx.Where("payment_record_id = ?", paymentID).Delete(&model.DailyPaymentAllocationPG{}).Error; err != nil {
			tx.Rollback()
			log.Printf("删除每日分摊记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除每日分摊记录失败"})
			return
		}

		// 删除缴费记录
		result := tx.Delete(&model.PaymentRecordPG{}, paymentID)
		if result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除缴费记录失败"})
			return
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "缴费记录不存在"})
			return
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			log.Printf("提交事务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "缴费记录删除成功"})
	}
}

// UpdatePaymentRecordPG 更新缴费记录 - PostgreSQL版本（需要重新计算每日分摊）
func UpdatePaymentRecordPG() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查权限
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		idStr := c.Param("id")
		paymentID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
			return
		}

		var req struct {
			Amount    float64 `json:"amount" binding:"required,min=0"`
			StartDate string  `json:"start_date" binding:"required"`
			EndDate   string  `json:"end_date" binding:"required"`
			Remark    string  `json:"remark"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 解析日期
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的开始日期格式"})
			return
		}

		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的结束日期格式"})
			return
		}

		if endDate.Before(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
			return
		}

		// 计算新的服务天数和每日金额
		serviceDays := int(endDate.Sub(startDate).Hours()/24) + 1
		dailyAmount := req.Amount / float64(serviceDays)

		// 开始事务
		tx := database.GetPostgresDB().Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 先获取原记录
		var originalRecord model.PaymentRecordPG
		if err := tx.First(&originalRecord, paymentID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "缴费记录不存在"})
			return
		}

		// 更新缴费记录
		updates := map[string]interface{}{
			"amount":       req.Amount,
			"start_date":   startDate,
			"end_date":     endDate,
			"daily_amount": dailyAmount,
			"service_days": serviceDays,
			"remark":       req.Remark,
			"updated_at":   time.Now(),
		}

		if err := tx.Model(&originalRecord).Updates(updates).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新缴费记录失败"})
			return
		}

		// 删除旧的每日分摊记录
		if err := tx.Where("payment_record_id = ?", paymentID).Delete(&model.DailyPaymentAllocationPG{}).Error; err != nil {
			tx.Rollback()
			log.Printf("删除旧的每日分摊记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除旧的每日分摊记录失败"})
			return
		}

		// 创建新的每日分摊记录
		updatedRecord := originalRecord
		updatedRecord.Amount = req.Amount
		updatedRecord.StartDate = startDate
		updatedRecord.EndDate = endDate
		updatedRecord.DailyAmount = dailyAmount
		updatedRecord.ServiceDays = serviceDays
		updatedRecord.Remark = req.Remark

		if err := createDailyAllocationsPG(tx, paymentID, updatedRecord); err != nil {
			tx.Rollback()
			log.Printf("创建新的每日分摊记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新每日分摊记录失败"})
			return
		}

		// 提交事务
		if err := tx.Commit().Error; err != nil {
			log.Printf("提交事务失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "缴费记录更新成功",
			"service_days": serviceDays,
			"daily_amount": dailyAmount,
		})
	}
}
