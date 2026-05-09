package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xvv6u577/logv2fs/database"
	helper "github.com/xvv6u577/logv2fs/helpers"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 计费模块设计取向（与之前版本相比的核心变化）：
//
//   - 唯一事实表：payment_records。一笔缴费 = 一行，不再向 daily_payment_allocations
//     展开 N 行。collection size 直接缩小到原来的 ~1/服务天数。
//   - 月/年统计走「实时分摊」：把每条 PaymentRecord 的金额按服务期均摊到日，
//     再把日金额累加到对应月/年。整个过程发生在 Go 内存里，避免冗余物化。
//   - daily_amount / service_days 仍保留为 PaymentRecord 自带的派生字段，
//     在新增 / 更新 时由后端统一计算并写回，前端列表展示无需改动。

// computeServiceDaysAndDaily 把 (amount, start, end) 转换成派生字段 (days, daily)。
// 服务天数包含起止两端 —— 与前端 paymentRecords.js 的 calculateDays 对齐。
func computeServiceDaysAndDaily(amount float64, start, end time.Time) (int, float64) {
	if end.Before(start) {
		return 0, 0
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}
	return days, amount / float64(days)
}

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

		serviceDays, dailyAmount := computeServiceDaysAndDaily(req.Amount, startDate, endDate)

		// 获取用户信息
		userEmail, exists := c.Get("email")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证用户"})
			return
		}

		userName, _ := c.Get("username")

		operatorEmail := ""
		if userEmail != nil {
			operatorEmail = userEmail.(string)
		}

		operatorName := ""
		if userName != nil {
			operatorName = userName.(string)
		}

		paymentRecord := model.PaymentRecord{
			ID:            primitive.NewObjectID(),
			UserEmailAsId: req.UserEmailAsId,
			UserName:      getUserNameByEmail(req.UserEmailAsId),
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

		_, err = database.GetCollection(model.PaymentRecord{}).InsertOne(context.Background(), paymentRecord)
		if err != nil {
			log.Printf("添加缴费记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加缴费记录失败"})
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

		cursor, err := database.GetCollection(model.PaymentRecord{}).Find(ctx,
			bson.M{"user_email_as_id": userEmail},
			options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
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

// GetPaymentStatistics 获取费用统计（仅支持 monthly / yearly / overall）
//
// 算法概要：
//  1. 按 start_date <= Q_end AND end_date >= Q_start 拉出所有相关 PaymentRecord
//  2. 对每条 record，把 [max(S, Q_start), min(E, Q_end)] 这段交集按月切片，
//     每个月分到 daily_amount * 该月落入交集的天数
//  3. 顺手统计每个月/年涉及到的 PaymentRecord 数量与不同用户数（去重）
func GetPaymentStatistics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		statType := c.DefaultQuery("type", "monthly") // monthly | yearly | overall
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		var startDate, endDate time.Time
		var err error

		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
				return
			}
		} else {
			// 默认为 30 天前
			startDate = time.Now().AddDate(0, 0, -30)
		}

		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式错误"})
				return
			}
		} else {
			endDate = time.Now()
		}

		// 查询区间统一在 UTC 下取「日」边界，避免和 PaymentRecord 中带时区的时间戳错位
		startDateTime := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
		endDateTime := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.UTC)

		stats := model.PaymentStatistics{
			StartDate: startDateTime,
			EndDate:   endDateTime,
			DateRange: fmt.Sprintf("%s 至 %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		}

		records, err := findOverlappingPayments(startDateTime, endDateTime)
		if err != nil {
			log.Printf("查询缴费记录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询缴费记录失败"})
			return
		}

		monthly, yearly := aggregatePayments(records, startDateTime, endDateTime)

		switch statType {
		case "monthly":
			stats.MonthlyStats = bucketsToMonthly(monthly)
			stats.TotalAmount, stats.PaymentCount = totalsFromBuckets(monthly)
		case "yearly":
			stats.YearlyStats = bucketsToYearly(yearly)
			stats.TotalAmount, stats.PaymentCount = totalsFromBuckets(yearly)
		case "overall":
			stats.MonthlyStats = bucketsToMonthly(monthly)
			stats.YearlyStats = bucketsToYearly(yearly)
			// 用月桶推总数，避免年/月口径不一致
			stats.TotalAmount, stats.PaymentCount = totalsFromBuckets(monthly)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的统计类型，仅支持 monthly / yearly / overall"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// periodBucket 按月或按年聚合时的中间结构
type periodBucket struct {
	Amount   float64
	Payments map[primitive.ObjectID]struct{}
	Users    map[string]struct{}
}

func newBucket() *periodBucket {
	return &periodBucket{
		Payments: map[primitive.ObjectID]struct{}{},
		Users:    map[string]struct{}{},
	}
}

// findOverlappingPayments 拉出所有与 [qStart, qEnd] 有交集的缴费记录
func findOverlappingPayments(qStart, qEnd time.Time) ([]model.PaymentRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	filter := bson.M{
		"start_date": bson.M{"$lte": qEnd},
		"end_date":   bson.M{"$gte": qStart},
	}

	cursor, err := database.GetCollection(model.PaymentRecord{}).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []model.PaymentRecord
	if err := cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

// aggregatePayments 在内存里把 records 按月、按年同时分摊
//
// 月桶 key = "YYYYMM"，年桶 key = "YYYY"。同一笔缴费横跨 N 个月 / Y 年时，
// 在每个月 / 年里都计 1 次（PaymentCount / UserCount 通过 set 去重）。
func aggregatePayments(records []model.PaymentRecord, qStart, qEnd time.Time) (
	monthly map[string]*periodBucket, yearly map[string]*periodBucket,
) {
	monthly = map[string]*periodBucket{}
	yearly = map[string]*periodBucket{}

	for i := range records {
		rec := &records[i]
		allocateOne(rec, qStart, qEnd, monthly, yearly)
	}
	return
}

// allocateOne 把一条 PaymentRecord 按月切片分摊到 monthly / yearly 桶
func allocateOne(rec *model.PaymentRecord, qStart, qEnd time.Time,
	monthly, yearly map[string]*periodBucket) {

	// 计算和查询区间的交集
	iStart := maxTime(rec.StartDate, qStart)
	iEnd := minTime(rec.EndDate, qEnd)
	if iEnd.Before(iStart) {
		return
	}

	// 把交集两端规整到 UTC 日，避免时分秒误差
	s := time.Date(iStart.Year(), iStart.Month(), iStart.Day(), 0, 0, 0, 0, time.UTC)
	e := time.Date(iEnd.Year(), iEnd.Month(), iEnd.Day(), 0, 0, 0, 0, time.UTC)

	// 容错：DailyAmount 缺失时即时算（兼容历史脏数据）
	daily := rec.DailyAmount
	if daily == 0 && rec.ServiceDays > 0 {
		daily = rec.Amount / float64(rec.ServiceDays)
	}

	cur := time.Date(s.Year(), s.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(e.Year(), e.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !cur.After(endMonth) {
		monthStart := cur
		monthEnd := cur.AddDate(0, 1, -1) // 该月最后一天

		lo := monthStart
		if s.After(lo) {
			lo = s
		}
		hi := monthEnd
		if e.Before(hi) {
			hi = e
		}

		if !hi.Before(lo) {
			days := int(hi.Sub(lo).Hours()/24) + 1
			amount := daily * float64(days)
			mKey := cur.Format("200601")
			yKey := cur.Format("2006")

			mb := monthly[mKey]
			if mb == nil {
				mb = newBucket()
				monthly[mKey] = mb
			}
			mb.Amount += amount
			mb.Payments[rec.ID] = struct{}{}
			mb.Users[rec.UserEmailAsId] = struct{}{}

			yb := yearly[yKey]
			if yb == nil {
				yb = newBucket()
				yearly[yKey] = yb
			}
			yb.Amount += amount
			yb.Payments[rec.ID] = struct{}{}
			yb.Users[rec.UserEmailAsId] = struct{}{}
		}

		cur = cur.AddDate(0, 1, 0)
	}
}

func bucketsToMonthly(buckets map[string]*periodBucket) []model.MonthlyPaymentStats {
	keys := sortedKeys(buckets)
	out := make([]model.MonthlyPaymentStats, 0, len(keys))
	for _, k := range keys {
		b := buckets[k]
		out = append(out, model.MonthlyPaymentStats{
			Month:        k,
			TotalAmount:  b.Amount,
			PaymentCount: int64(len(b.Payments)),
			UserCount:    int64(len(b.Users)),
		})
	}
	return out
}

func bucketsToYearly(buckets map[string]*periodBucket) []model.YearlyPaymentStats {
	keys := sortedKeys(buckets)
	out := make([]model.YearlyPaymentStats, 0, len(keys))
	for _, k := range keys {
		b := buckets[k]
		out = append(out, model.YearlyPaymentStats{
			Year:         k,
			TotalAmount:  b.Amount,
			PaymentCount: int64(len(b.Payments)),
			UserCount:    int64(len(b.Users)),
		})
	}
	return out
}

// totalsFromBuckets 用某一组互斥桶（同一记录可能出现在多个桶里，但金额不重叠）
// 推算出总金额与去重后的缴费笔数。
func totalsFromBuckets(buckets map[string]*periodBucket) (float64, int64) {
	var total float64
	allPayments := map[primitive.ObjectID]struct{}{}
	for _, b := range buckets {
		total += b.Amount
		for id := range b.Payments {
			allPayments[id] = struct{}{}
		}
	}
	return total, int64(len(allPayments))
}

func sortedKeys(buckets map[string]*periodBucket) []string {
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// GetPaymentRecords 获取缴费记录列表（分页）
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

		collection := database.GetCollection(model.PaymentRecord{})

		filter := bson.M{}
		if userEmail != "" {
			filter["user_email_as_id"] = userEmail
		}

		total, err := collection.CountDocuments(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取记录总数失败"})
			return
		}

		findOptions := options.Find()
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64((page - 1) * limit))
		findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

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
//
// 改造后不再有 daily_payment_allocations 派生表需要联动清理。
func DeletePaymentRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		paymentId := c.Param("id")
		if paymentId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缴费记录ID不能为空"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		objID, err := primitive.ObjectIDFromHex(paymentId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的缴费记录ID"})
			return
		}

		var payment model.PaymentRecord
		err = database.GetCollection(model.PaymentRecord{}).FindOne(ctx, bson.M{"_id": objID}).Decode(&payment)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "缴费记录不存在"})
			return
		}

		result, err := database.GetCollection(model.PaymentRecord{}).DeleteOne(ctx, bson.M{"_id": objID})
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

// UpdatePaymentRecord 更新缴费记录
//
// 关键修复：amount / start_date / end_date 任一字段变化时，
// 必须同步重算 service_days 与 daily_amount，防止派生字段失同步。
func UpdatePaymentRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "admin"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		recordId := c.Param("id")
		if recordId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "记录ID不能为空"})
			return
		}

		objectId, err := primitive.ObjectIDFromHex(recordId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的记录ID"})
			return
		}

		var existingRecord model.PaymentRecord
		err = database.GetCollection(model.PaymentRecord{}).FindOne(ctx, bson.M{"_id": objectId}).Decode(&existingRecord)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "缴费记录不存在"})
			return
		}

		var updateData model.PaymentRecord
		if err := c.BindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			log.Printf("BindJSON error: %v", err)
			return
		}

		// 取「合并后」的最终值，再决定是否需要重算派生字段
		amount := existingRecord.Amount
		startDate := existingRecord.StartDate
		endDate := existingRecord.EndDate
		needRecalc := false

		if updateData.Amount > 0 && updateData.Amount != existingRecord.Amount {
			amount = updateData.Amount
			needRecalc = true
		}
		if !updateData.StartDate.IsZero() && !updateData.StartDate.Equal(existingRecord.StartDate) {
			startDate = updateData.StartDate
			needRecalc = true
		}
		if !updateData.EndDate.IsZero() && !updateData.EndDate.Equal(existingRecord.EndDate) {
			endDate = updateData.EndDate
			needRecalc = true
		}

		if endDate.Before(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
			return
		}

		update := bson.M{"updated_at": time.Now()}

		if needRecalc {
			serviceDays, dailyAmount := computeServiceDaysAndDaily(amount, startDate, endDate)
			update["amount"] = amount
			update["start_date"] = startDate
			update["end_date"] = endDate
			update["service_days"] = serviceDays
			update["daily_amount"] = dailyAmount
		}

		if updateData.Remark != "" && updateData.Remark != existingRecord.Remark {
			update["remark"] = updateData.Remark
		}

		// 只有 updated_at 一项时直接拒绝，避免「空更新」让用户误以为修改了字段
		if len(update) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "没有可更新的字段"})
			return
		}

		result, err := database.GetCollection(model.PaymentRecord{}).UpdateOne(
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

// 获取用户名（用户表里查 name；查不到则退化为邮箱）
func getUserNameByEmail(email string) string {
	userCollection := database.GetCollection(model.UserTrafficLogs{})
	var user struct {
		Name string `bson:"name"`
	}

	err := userCollection.FindOne(context.Background(), bson.M{"email_as_id": email}).Decode(&user)
	if err != nil {
		return email
	}

	return user.Name
}
