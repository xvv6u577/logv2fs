package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// migrateUserTrafficLogsDataImpl UserTrafficLogs数据迁移的具体实现
func migrateUserTrafficLogsDataImpl(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	log.Println("👥 开始迁移UserTrafficLogs数据...")

	// 获取数据库连接
	mongoClient := database.Client
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	collection := database.OpenCollection(mongoClient, "USER_TRAFFIC_LOGS") // 假设用户集合名为users

	// 计算总数
	totalCount, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return fmt.Errorf("获取UserTrafficLogs总数失败: %v", err)
	}

	log.Printf("📊 发现 %d 个UserTrafficLogs记录需要迁移", totalCount)

	// 分批处理
	var processed int64 = 0
	var migrated int64 = 0
	var skipped int64 = 0

	for skip := int64(0); skip < totalCount; skip += int64(batchSize) {
		// 设置查询选项
		findOptions := options.Find()
		findOptions.SetSkip(skip)
		findOptions.SetLimit(int64(batchSize))

		// 查询一批数据
		cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
		if err != nil {
			return fmt.Errorf("查询UserTrafficLogs数据失败: %v", err)
		}

		// 处理这批数据
		var mongoUserLogs []model.UserTrafficLogs
		if err := cursor.All(context.Background(), &mongoUserLogs); err != nil {
			cursor.Close(context.Background())
			return fmt.Errorf("解析UserTrafficLogs数据失败: %v", err)
		}
		cursor.Close(context.Background())

		// 转换并保存到PostgreSQL
		for _, mongoUserLog := range mongoUserLogs {
			processed++

			// 检查是否跳过已存在的记录
			if skipExisting {
				var existingCount int64
				err := postgresDB.Model(&model.UserTrafficLogsPG{}).
					Where("email_as_id = ?", mongoUserLog.Email_As_Id).
					Count(&existingCount).Error
				if err != nil {
					log.Printf("⚠️  检查UserTrafficLogs重复失败: %v", err)
					stats.Errors = append(stats.Errors, fmt.Sprintf("检查UserTrafficLogs重复失败: %v", err))
					continue
				}

				if existingCount > 0 {
					skipped++
					continue
				}
			}

			// 转换数据结构
			pgUserLog, err := convertUserTrafficLogsToPG(mongoUserLog)
			if err != nil {
				log.Printf("⚠️  转换UserTrafficLogs失败: %v", err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("转换UserTrafficLogs失败: %v", err))
				continue
			}

			// 数据验证
			if err := validateUserTrafficLogsData(&pgUserLog); err != nil {
				log.Printf("⚠️  UserTrafficLogs数据验证失败: %v", err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("UserTrafficLogs数据验证失败: %v", err))
				continue
			}

			// 保存到PostgreSQL
			if err := postgresDB.Create(&pgUserLog).Error; err != nil {
				log.Printf("⚠️  保存UserTrafficLogs失败: %v", err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("保存UserTrafficLogs失败: %v", err))
				continue
			}

			migrated++
		}

		// 打印进度
		if processed%int64(batchSize*5) == 0 || processed == totalCount {
			log.Printf("📈 UserTrafficLogs迁移进度: %d/%d (已迁移: %d, 已跳过: %d)",
				processed, totalCount, migrated, skipped)
		}
	}

	stats.UserRecordsMigrated = migrated
	log.Printf("✅ UserTrafficLogs迁移完成: 共处理 %d 条记录，成功迁移 %d 条，跳过 %d 条",
		processed, migrated, skipped)

	return nil
}

// convertUserTrafficLogsToPG 将MongoDB的UserTrafficLogs转换为PostgreSQL的UserTrafficLogsPG
func convertUserTrafficLogsToPG(mongoUserLog model.UserTrafficLogs) (model.UserTrafficLogsPG, error) {
	pgUserLog := model.UserTrafficLogsPG{
		ID:           uuid.New(),
		EmailAsId:    mongoUserLog.Email_As_Id,
		Password:     mongoUserLog.Password,
		UUID:         mongoUserLog.UUID,
		Role:         mongoUserLog.Role,
		Status:       mongoUserLog.Status,
		Name:         mongoUserLog.Name,
		Token:        mongoUserLog.Token,
		RefreshToken: mongoUserLog.Refresh_token,
		UserID:       mongoUserLog.User_id,
		Used:         mongoUserLog.Used,
		Credit:       mongoUserLog.Credit,
		CreatedAt:    mongoUserLog.CreatedAt,
		UpdatedAt:    mongoUserLog.UpdatedAt,
	}

	// 转换时间序列数据为JSONB

	// 转换HourlyLogs
	if len(mongoUserLog.HourlyLogs) > 0 {
		hourlyJSON, err := convertUserHourlyLogsToJSON(mongoUserLog.HourlyLogs)
		if err != nil {
			return pgUserLog, fmt.Errorf("转换用户HourlyLogs失败: %v", err)
		}
		pgUserLog.HourlyLogs = hourlyJSON
	}

	// 转换DailyLogs
	if len(mongoUserLog.DailyLogs) > 0 {
		dailyJSON, err := convertUserDailyLogsToJSON(mongoUserLog.DailyLogs)
		if err != nil {
			return pgUserLog, fmt.Errorf("转换用户DailyLogs失败: %v", err)
		}
		pgUserLog.DailyLogs = dailyJSON
	}

	// 转换MonthlyLogs
	if len(mongoUserLog.MonthlyLogs) > 0 {
		monthlyJSON, err := convertUserMonthlyLogsToJSON(mongoUserLog.MonthlyLogs)
		if err != nil {
			return pgUserLog, fmt.Errorf("转换用户MonthlyLogs失败: %v", err)
		}
		pgUserLog.MonthlyLogs = monthlyJSON
	}

	// 转换YearlyLogs
	if len(mongoUserLog.YearlyLogs) > 0 {
		yearlyJSON, err := convertUserYearlyLogsToJSON(mongoUserLog.YearlyLogs)
		if err != nil {
			return pgUserLog, fmt.Errorf("转换用户YearlyLogs失败: %v", err)
		}
		pgUserLog.YearlyLogs = yearlyJSON
	}

	return pgUserLog, nil
}

// convertUserHourlyLogsToJSON 转换用户小时级别日志为JSON
func convertUserHourlyLogsToJSON(hourlyLogs []struct {
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Traffic   int64     `json:"traffic" bson:"traffic"`
}) (datatypes.JSON, error) {
	var logs []model.TrafficLogEntry

	for _, log := range hourlyLogs {
		logs = append(logs, model.TrafficLogEntry{
			Timestamp: log.Timestamp,
			Traffic:   log.Traffic,
		})
	}

	jsonData, err := json.Marshal(logs)
	if err != nil {
		return nil, fmt.Errorf("序列化用户小时日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// convertUserDailyLogsToJSON 转换用户日级别日志为JSON
func convertUserDailyLogsToJSON(dailyLogs []struct {
	Date    string `json:"date" bson:"date"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}) (datatypes.JSON, error) {
	var logs []model.DailyLogEntry

	for _, log := range dailyLogs {
		logs = append(logs, model.DailyLogEntry{
			Date:    log.Date,
			Traffic: log.Traffic,
		})
	}

	jsonData, err := json.Marshal(logs)
	if err != nil {
		return nil, fmt.Errorf("序列化用户日级日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// convertUserMonthlyLogsToJSON 转换用户月级别日志为JSON
func convertUserMonthlyLogsToJSON(monthlyLogs []struct {
	Month   string `json:"month" bson:"month"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}) (datatypes.JSON, error) {
	var logs []model.MonthlyLogEntry

	for _, log := range monthlyLogs {
		logs = append(logs, model.MonthlyLogEntry{
			Month:   log.Month,
			Traffic: log.Traffic,
		})
	}

	jsonData, err := json.Marshal(logs)
	if err != nil {
		return nil, fmt.Errorf("序列化用户月级日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// convertUserYearlyLogsToJSON 转换用户年级别日志为JSON
func convertUserYearlyLogsToJSON(yearlyLogs []struct {
	Year    string `json:"year" bson:"year"`
	Traffic int64  `json:"traffic" bson:"traffic"`
}) (datatypes.JSON, error) {
	var logs []model.YearlyLogEntry

	for _, log := range yearlyLogs {
		logs = append(logs, model.YearlyLogEntry{
			Year:    log.Year,
			Traffic: log.Traffic,
		})
	}

	jsonData, err := json.Marshal(logs)
	if err != nil {
		return nil, fmt.Errorf("序列化用户年级日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// validateUserTrafficLogsData 验证UserTrafficLogs数据的完整性
func validateUserTrafficLogsData(userLog *model.UserTrafficLogsPG) error {
	if userLog.EmailAsId == "" {
		return fmt.Errorf("EmailAsId不能为空")
	}

	// 验证角色字段
	validRoles := map[string]bool{
		"admin":  true,
		"normal": true,
	}

	if !validRoles[userLog.Role] {
		// 如果角色无效，设置默认值
		userLog.Role = "normal"
	}

	// 验证状态字段
	validStatuses := map[string]bool{
		"plain":   true,
		"deleted": true,
		"overdue": true,
	}

	if !validStatuses[userLog.Status] {
		// 如果状态无效，设置默认值
		userLog.Status = "plain"
	}

	// 验证流量数据
	if userLog.Used < 0 {
		userLog.Used = 0
	}

	if userLog.Credit < 0 {
		userLog.Credit = 0
	}

	return nil
}

// createUserTrafficLogsIndexes 为UserTrafficLogs表创建额外的索引
func createUserTrafficLogsIndexes(db *gorm.DB) error {
	log.Println("🔍 为UserTrafficLogs表创建索引...")

	// 为email_as_id字段创建唯一索引
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_traffic_logs_email_as_id_unique ON user_traffic_logs(email_as_id)").Error; err != nil {
		return fmt.Errorf("创建email_as_id唯一索引失败: %v", err)
	}

	// 为role字段创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_role ON user_traffic_logs(role)").Error; err != nil {
		return fmt.Errorf("创建role索引失败: %v", err)
	}

	// 为status字段创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_status ON user_traffic_logs(status)").Error; err != nil {
		return fmt.Errorf("创建status索引失败: %v", err)
	}

	// 为user_id字段创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_user_id ON user_traffic_logs(user_id)").Error; err != nil {
		return fmt.Errorf("创建user_id索引失败: %v", err)
	}

	// 复合索引 - role + status
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_role_status ON user_traffic_logs(role, status)").Error; err != nil {
		return fmt.Errorf("创建role_status复合索引失败: %v", err)
	}

	// 流量查询索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_used ON user_traffic_logs(used)").Error; err != nil {
		return fmt.Errorf("创建used索引失败: %v", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_user_traffic_logs_credit ON user_traffic_logs(credit)").Error; err != nil {
		return fmt.Errorf("创建credit索引失败: %v", err)
	}

	log.Println("✅ UserTrafficLogs索引创建完成")
	return nil
}

// parseTimestamp 解析时间戳，处理多种时间格式
func parseTimestamp(timestamp interface{}) (time.Time, error) {
	switch v := timestamp.(type) {
	case time.Time:
		return v, nil
	case primitive.DateTime:
		return v.Time(), nil
	case string:
		// 尝试解析RFC3339格式
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		// 尝试解析其他常见格式
		if t, err := time.Parse("2006-01-02T15:04:05Z", v); err == nil {
			return t, nil
		}
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t, nil
		}
		return time.Time{}, fmt.Errorf("无法解析时间格式: %s", v)
	case int64:
		// Unix时间戳
		return time.Unix(v, 0), nil
	case float64:
		// Unix时间戳 (浮点数)
		return time.Unix(int64(v), 0), nil
	default:
		return time.Time{}, fmt.Errorf("不支持的时间戳类型: %T", v)
	}
}
