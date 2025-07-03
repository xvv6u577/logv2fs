package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	paymentRecordsCol *mongo.Collection = database.OpenCollection(database.Client, "payment_records")
)

// AddPaymentRecord 添加缴费记录
func AddPaymentRecord() gin.HandlerFunc {
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
		paymentRecord := model.PaymentRecord{
			ID:            primitive.NewObjectID(),
			UserEmailAsId: req.UserEmailAsId,
			UserName:      getUserNameByEmail(req.UserEmailAsId), // 获取被充值用户名
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

		// 插入缴费记录
		_, err = paymentRecordsCol.InsertOne(context.Background(), paymentRecord)
		if err != nil {
			log.Printf("添加缴费记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加缴费记录失败"})
			return
		}

		// 创建每日分摊记录
		if err := createDailyAllocations(paymentRecord.ID, paymentRecord); err != nil {
			log.Printf("创建每日分摊记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建每日分摊记录失败"})
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

// GetUserPayments 获取用户缴费记录
func GetUserPayments() gin.HandlerFunc {
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

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// 查询该用户的所有缴费记录
		cursor, err := paymentRecordsCol.Find(ctx, bson.M{"user_email_as_id": userEmail}, options.Find().SetSort(bson.D{{"payment_date", -1}}))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询缴费记录失败"})
			log.Printf("Query payment records error: %v", err)
			return
		}
		defer cursor.Close(ctx)

		var payments []model.PaymentRecord
		if err = cursor.All(ctx, &payments); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析缴费记录失败"})
			log.Printf("Decode payment records error: %v", err)
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

// GetPaymentStatistics 获取费用统计
func GetPaymentStatistics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查权限
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取查询参数
		statType := c.DefaultQuery("type", "daily") // daily, monthly, yearly, overall
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		// 解析日期范围
		var startDate, endDate time.Time
		var err error

		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
				return
			}
		} else {
			// 默认为30天前
			startDate = time.Now().AddDate(0, 0, -30)
		}

		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式错误"})
				return
			}
			// 设置为当天的最后一秒
			endDate = endDate.Add(24*time.Hour - time.Second)
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

		// 基于每日分摊记录进行统计
		collection := database.OpenCollection(database.Client, "daily_payment_allocations")

		switch statType {
		case "daily":
			dailyStats, err := getDailyStats(collection, startDateTime, endDateTime)
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
			monthlyStats, err := getMonthlyStats(collection, startDateTime, endDateTime)
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
			yearlyStats, err := getYearlyStats(collection, startDateTime, endDateTime)
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
			dailyStats, _ := getDailyStats(collection, startDateTime, endDateTime)
			monthlyStats, _ := getMonthlyStats(collection, startDateTime, endDateTime)
			yearlyStats, _ := getYearlyStats(collection, startDateTime, endDateTime)

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

// 获取每日统计
func getDailyStats(collection *mongo.Collection, startDate, endDate time.Time) ([]model.DailyPaymentStats, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":           "$date_string",
				"total_amount":  bson.M{"$sum": "$allocated_amount"},
				"payment_count": bson.M{"$sum": 1},
				"users":         bson.M{"$addToSet": "$user_email_as_id"},
			},
		},
		{
			"$project": bson.M{
				"date":          "$_id",
				"total_amount":  1,
				"payment_count": 1,
				"user_count":    bson.M{"$size": "$users"},
			},
		},
		{
			"$sort": bson.M{"date": 1},
		},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []model.DailyPaymentStats
	for cursor.Next(context.Background()) {
		var result struct {
			Date         string  `bson:"date"`
			TotalAmount  float64 `bson:"total_amount"`
			PaymentCount int64   `bson:"payment_count"`
			UserCount    int64   `bson:"user_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		results = append(results, model.DailyPaymentStats{
			Date:         result.Date,
			TotalAmount:  result.TotalAmount,
			PaymentCount: result.PaymentCount,
			UserCount:    result.UserCount,
		})
	}

	return results, nil
}

// 获取每月统计
func getMonthlyStats(collection *mongo.Collection, startDate, endDate time.Time) ([]model.MonthlyPaymentStats, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y%m",
						"date":   "$date",
					},
				},
				"total_amount":  bson.M{"$sum": "$allocated_amount"},
				"payment_count": bson.M{"$sum": 1},
				"users":         bson.M{"$addToSet": "$user_email_as_id"},
			},
		},
		{
			"$project": bson.M{
				"month":         "$_id",
				"total_amount":  1,
				"payment_count": 1,
				"user_count":    bson.M{"$size": "$users"},
			},
		},
		{
			"$sort": bson.M{"month": 1},
		},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []model.MonthlyPaymentStats
	for cursor.Next(context.Background()) {
		var result struct {
			Month        string  `bson:"month"`
			TotalAmount  float64 `bson:"total_amount"`
			PaymentCount int64   `bson:"payment_count"`
			UserCount    int64   `bson:"user_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		results = append(results, model.MonthlyPaymentStats{
			Month:        result.Month,
			TotalAmount:  result.TotalAmount,
			PaymentCount: result.PaymentCount,
			UserCount:    result.UserCount,
		})
	}

	return results, nil
}

// 获取每年统计
func getYearlyStats(collection *mongo.Collection, startDate, endDate time.Time) ([]model.YearlyPaymentStats, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"date": bson.M{
					"$gte": startDate,
					"$lte": endDate,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y",
						"date":   "$date",
					},
				},
				"total_amount":  bson.M{"$sum": "$allocated_amount"},
				"payment_count": bson.M{"$sum": 1},
				"users":         bson.M{"$addToSet": "$user_email_as_id"},
			},
		},
		{
			"$project": bson.M{
				"year":          "$_id",
				"total_amount":  1,
				"payment_count": 1,
				"user_count":    bson.M{"$size": "$users"},
			},
		},
		{
			"$sort": bson.M{"year": 1},
		},
	}

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []model.YearlyPaymentStats
	for cursor.Next(context.Background()) {
		var result struct {
			Year         string  `bson:"year"`
			TotalAmount  float64 `bson:"total_amount"`
			PaymentCount int64   `bson:"payment_count"`
			UserCount    int64   `bson:"user_count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		results = append(results, model.YearlyPaymentStats{
			Year:         result.Year,
			TotalAmount:  result.TotalAmount,
			PaymentCount: result.PaymentCount,
			UserCount:    result.UserCount,
		})
	}

	return results, nil
}

// GetPaymentRecords 获取缴费记录列表
func GetPaymentRecords() gin.HandlerFunc {
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

		collection := database.OpenCollection(database.Client, "payment_records")

		// 构建查询条件
		filter := bson.M{}
		if userEmail != "" {
			filter["user_email_as_id"] = userEmail
		}

		// 计算总数
		total, err := collection.CountDocuments(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取记录总数失败"})
			return
		}

		// 查询数据
		findOptions := options.Find()
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64((page - 1) * limit))
		findOptions.SetSort(bson.D{{"created_at", -1}})

		cursor, err := collection.Find(context.Background(), filter, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询缴费记录失败"})
			return
		}
		defer cursor.Close(context.Background())

		var records []model.PaymentRecord
		if err = cursor.All(context.Background(), &records); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "解析缴费记录失败"})
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

// DeletePaymentRecord 删除缴费记录
func DeletePaymentRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 权限检查 - 只有管理员可以删除缴费记录
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		paymentId := c.Param("id")
		if paymentId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缴费记录ID不能为空"})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// 转换ID
		objID, err := primitive.ObjectIDFromHex(paymentId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的缴费记录ID"})
			return
		}

		// 先查询记录是否存在
		var payment model.PaymentRecord
		err = paymentRecordsCol.FindOne(ctx, bson.M{"_id": objID}).Decode(&payment)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "缴费记录不存在"})
			return
		}

		// 删除每日分摊记录
		allocationCollection := database.OpenCollection(database.Client, "daily_payment_allocations")
		_, err = allocationCollection.DeleteMany(ctx, bson.M{"payment_record_id": objID})
		if err != nil {
			log.Printf("删除每日分摊记录失败: %v", err)
		}

		// 删除缴费记录
		result, err := paymentRecordsCol.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除缴费记录失败"})
			log.Printf("Delete payment record error: %v", err)
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "缴费记录不存在"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("成功删除用户 %s 的缴费记录，金额：%.2f", payment.UserName, payment.Amount),
		})
	}
}

// UpdatePaymentRecord 更新缴费记录 - MongoDB版本
func UpdatePaymentRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 权限检查 - 只有管理员可以更新缴费记录
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		recordId := c.Param("id")
		if recordId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "记录ID不能为空"})
			return
		}

		// 转换ID
		objectId, err := primitive.ObjectIDFromHex(recordId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
			return
		}

		// 查找原记录
		var existingRecord model.PaymentRecord
		err = paymentRecordsCol.FindOne(ctx, bson.M{"_id": objectId}).Decode(&existingRecord)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "缴费记录不存在"})
			return
		}

		// 绑定更新数据
		var updateData model.PaymentRecord
		if err := c.BindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// 构建更新字段
		update := bson.M{"updated_at": time.Now()}

		// 只更新提供的字段
		if updateData.Amount > 0 {
			update["amount"] = updateData.Amount
		}

		if !updateData.StartDate.IsZero() {
			update["start_date"] = updateData.StartDate
		}

		if !updateData.EndDate.IsZero() {
			update["end_date"] = updateData.EndDate
		}

		if updateData.Remark != "" {
			update["remark"] = updateData.Remark
		}

		// 验证日期逻辑
		startDate := existingRecord.StartDate
		endDate := existingRecord.EndDate

		if !updateData.StartDate.IsZero() {
			startDate = updateData.StartDate
		}
		if !updateData.EndDate.IsZero() {
			endDate = updateData.EndDate
		}

		if endDate.Before(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
			return
		}

		// 更新记录
		result, err := paymentRecordsCol.UpdateOne(
			ctx,
			bson.M{"_id": objectId},
			bson.M{"$set": update},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			log.Printf("Update payment record error: %v", err)
			return
		}

		if result.ModifiedCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "没有记录被更新"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "缴费记录更新成功",
		})
	}
}

// 辅助函数：安全获取float64值
func getFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

// 辅助函数：安全获取int64值
func getInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}

// 获取用户名
func getUserNameByEmail(email string) string {
	// 从users集合查询用户名
	userCollection := database.OpenCollection(database.Client, "users")
	var user struct {
		Username string `bson:"username"`
	}

	err := userCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return email // 如果找不到用户名，返回邮箱
	}

	return user.Username
}

// 创建每日分摊记录
func createDailyAllocations(paymentRecordID primitive.ObjectID, payment model.PaymentRecord) error {
	collection := database.OpenCollection(database.Client, "daily_payment_allocations")

	// 生成从开始日期到结束日期的每日分摊记录
	current := payment.StartDate
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
			CreatedAt:        time.Now(),
		}

		_, err := collection.InsertOne(context.Background(), allocation)
		if err != nil {
			return err
		}

		current = current.AddDate(0, 0, 1) // 增加一天
	}

	return nil
}
