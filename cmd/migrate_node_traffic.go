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
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// migrateNodeTrafficLogsDataImpl NodeTrafficLogs数据迁移的具体实现
func migrateNodeTrafficLogsDataImpl(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	log.Println("🌐 开始迁移NodeTrafficLogs数据...")

	// 获取数据库连接
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	collection := database.GetCollection(model.NodeTrafficLogs{})

	// 计算总数
	totalCount, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return fmt.Errorf("获取NodeTrafficLogs总数失败: %v", err)
	}

	log.Printf("📊 发现 %d 个NodeTrafficLogs记录需要迁移", totalCount)

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
			return fmt.Errorf("查询NodeTrafficLogs数据失败: %v", err)
		}

		// 处理这批数据
		var mongoNodeLogs []model.NodeTrafficLogs
		if err := cursor.All(context.Background(), &mongoNodeLogs); err != nil {
			cursor.Close(context.Background())
			return fmt.Errorf("解析NodeTrafficLogs数据失败: %v", err)
		}
		cursor.Close(context.Background())

		// 转换并保存到PostgreSQL
		for _, mongoNodeLog := range mongoNodeLogs {
			processed++

			// 检查是否跳过已存在的记录
			if skipExisting {
				var existingCount int64
				err := postgresDB.Model(&model.NodeTrafficLogsPG{}).
					Where("domain_as_id = ?", mongoNodeLog.Domain_As_Id).
					Count(&existingCount).Error
				if err != nil {
					log.Printf("⚠️  检查NodeTrafficLogs重复失败: %v", err)
					stats.Errors = append(stats.Errors, fmt.Sprintf("检查NodeTrafficLogs重复失败: %v", err))
					continue
				}

				if existingCount > 0 {
					skipped++
					continue
				}
			}

			// 转换数据结构
			pgNodeLog, err := convertNodeTrafficLogsToPG(mongoNodeLog, postgresDB)
			if err != nil {
				log.Printf("⚠️  转换NodeTrafficLogs失败: %v", err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("转换NodeTrafficLogs失败: %v", err))
				continue
			}

			// 数据验证
			if err := validateNodeTrafficLogsData(&pgNodeLog); err != nil {
				log.Printf("⚠️  NodeTrafficLogs数据验证失败: %v", err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("NodeTrafficLogs数据验证失败: %v", err))
				continue
			}

			// 保存到PostgreSQL
			if err := postgresDB.Create(&pgNodeLog).Error; err != nil {
				log.Printf("⚠️  保存NodeTrafficLogs失败: %v", err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("保存NodeTrafficLogs失败: %v", err))
				continue
			}

			migrated++
		}

		// 打印进度
		if processed%int64(batchSize*5) == 0 || processed == totalCount {
			log.Printf("📈 NodeTrafficLogs迁移进度: %d/%d (已迁移: %d, 已跳过: %d)",
				processed, totalCount, migrated, skipped)
		}
	}

	stats.NodeRecordsMigrated = migrated
	log.Printf("✅ NodeTrafficLogs迁移完成: 共处理 %d 条记录，成功迁移 %d 条，跳过 %d 条",
		processed, migrated, skipped)

	return nil
}

// convertNodeTrafficLogsToPG 将MongoDB的NodeTrafficLogs转换为PostgreSQL的NodeTrafficLogsPG
func convertNodeTrafficLogsToPG(mongoNodeLog model.NodeTrafficLogs, db *gorm.DB) (model.NodeTrafficLogsPG, error) {
	pgNodeLog := model.NodeTrafficLogsPG{
		ID:         uuid.New(),
		DomainAsId: mongoNodeLog.Domain_As_Id,
		Remark:     mongoNodeLog.Remark,
		Status:     mongoNodeLog.Status,
		CreatedAt:  mongoNodeLog.CreatedAt,
		UpdatedAt:  mongoNodeLog.UpdatedAt,
	}

	// 转换时间序列数据为JSONB
	var err error

	// 转换HourlyLogs
	if len(mongoNodeLog.HourlyLogs) > 0 {
		hourlyJSON, err := convertNodeHourlyLogsToJSON(mongoNodeLog.HourlyLogs)
		if err != nil {
			return pgNodeLog, fmt.Errorf("转换HourlyLogs失败: %v", err)
		}
		pgNodeLog.HourlyLogs = hourlyJSON
	}

	// 转换DailyLogs
	if len(mongoNodeLog.DailyLogs) > 0 {
		dailyJSON, err := convertNodeDailyLogsToJSON(mongoNodeLog.DailyLogs)
		if err != nil {
			return pgNodeLog, fmt.Errorf("转换DailyLogs失败: %v", err)
		}
		pgNodeLog.DailyLogs = dailyJSON
	}

	// 转换MonthlyLogs
	if len(mongoNodeLog.MonthlyLogs) > 0 {
		monthlyJSON, err := convertNodeMonthlyLogsToJSON(mongoNodeLog.MonthlyLogs)
		if err != nil {
			return pgNodeLog, fmt.Errorf("转换MonthlyLogs失败: %v", err)
		}
		pgNodeLog.MonthlyLogs = monthlyJSON
	}

	// 转换YearlyLogs
	if len(mongoNodeLog.YearlyLogs) > 0 {
		yearlyJSON, err := convertNodeYearlyLogsToJSON(mongoNodeLog.YearlyLogs)
		if err != nil {
			return pgNodeLog, fmt.Errorf("转换YearlyLogs失败: %v", err)
		}
		pgNodeLog.YearlyLogs = yearlyJSON
	}

	// 尝试关联Domain
	err = linkDomainToNodeTrafficLogs(&pgNodeLog, db)
	if err != nil {
		log.Printf("⚠️  关联Domain失败: %v", err)
		// 不返回错误，继续处理，只是没有外键关联
	}

	return pgNodeLog, nil
}

// linkDomainToNodeTrafficLogs 验证NodeTrafficLogs关联的域名是否存在
func linkDomainToNodeTrafficLogs(pgNodeLog *model.NodeTrafficLogsPG, db *gorm.DB) error {
	// 根据domain_as_id查找对应的SubscriptionNode
	var node model.SubscriptionNodePG
	err := db.Where("domain = ?", pgNodeLog.DomainAsId).First(&node).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 对应的SubscriptionNode不存在，但仍然可以创建记录
			log.Printf("⚠️  找不到匹配的SubscriptionNode: %s", pgNodeLog.DomainAsId)
			return nil
		}
		return fmt.Errorf("查找SubscriptionNode失败: %v", err)
	}

	// 找到了匹配的SubscriptionNode，记录成功
	log.Printf("✅ 找到匹配的SubscriptionNode: %s", node.Remark)
	return nil
}

// convertNodeHourlyLogsToJSON 转换节点小时级别日志为JSON
func convertNodeHourlyLogsToJSON(hourlyLogs []struct {
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
		return nil, fmt.Errorf("序列化节点小时日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// convertNodeDailyLogsToJSON 转换节点日级别日志为JSON
func convertNodeDailyLogsToJSON(dailyLogs []struct {
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
		return nil, fmt.Errorf("序列化节点日级日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// convertNodeMonthlyLogsToJSON 转换节点月级别日志为JSON
func convertNodeMonthlyLogsToJSON(monthlyLogs []struct {
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
		return nil, fmt.Errorf("序列化节点月级日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// convertNodeYearlyLogsToJSON 转换节点年级别日志为JSON
func convertNodeYearlyLogsToJSON(yearlyLogs []struct {
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
		return nil, fmt.Errorf("序列化节点年级日志失败: %v", err)
	}

	return datatypes.JSON(jsonData), nil
}

// validateNodeTrafficLogsData 验证NodeTrafficLogs数据的完整性
func validateNodeTrafficLogsData(nodeLog *model.NodeTrafficLogsPG) error {
	if nodeLog.DomainAsId == "" {
		return fmt.Errorf("DomainAsId不能为空")
	}

	// 验证状态字段
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
	}

	if !validStatuses[nodeLog.Status] {
		// 如果状态无效，设置默认值
		nodeLog.Status = "active"
	}

	return nil
}

// createNodeTrafficLogsIndexes 为NodeTrafficLogs表创建额外的索引
func createNodeTrafficLogsIndexes(db *gorm.DB) error {
	log.Println("🔍 为NodeTrafficLogs表创建索引...")

	// 为domain_as_id字段创建唯一索引
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_node_traffic_logs_domain_as_id_unique ON node_traffic_logs(domain_as_id)").Error; err != nil {
		return fmt.Errorf("创建domain_as_id唯一索引失败: %v", err)
	}

	// 为status字段创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_status ON node_traffic_logs(status)").Error; err != nil {
		return fmt.Errorf("创建status索引失败: %v", err)
	}

	// domain_id字段已不存在，不再创建相关索引

	// 复合索引 - status + created_at
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_node_traffic_logs_status_created_at ON node_traffic_logs(status, created_at)").Error; err != nil {
		return fmt.Errorf("创建复合索引失败: %v", err)
	}

	log.Println("✅ NodeTrafficLogs索引创建完成")
	return nil
}
