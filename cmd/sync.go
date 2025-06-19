package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// SyncStats 同步统计信息
type SyncStats struct {
	StartTime             time.Time `json:"start_time"`
	EndTime               time.Time `json:"end_time"`
	MongoToPostgresCount  int64     `json:"mongo_to_postgres_count"`
	PostgresToMongoCount  int64     `json:"postgres_to_mongo_count"`
	SkippedConflictsCount int64     `json:"skipped_conflicts_count"`
	Errors                []string  `json:"errors"`
	Mode                  string    `json:"mode"`
}

// syncCmd 双向同步命令
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "双向同步MongoDB和PostgreSQL中的user_traffic_logs数据",
	Long: `这个命令实现MongoDB和PostgreSQL之间user_traffic_logs表的双向同步功能。

功能特性:
- 以 email_as_id 为索引进行数据对比
- 自动检测并同步缺失的用户记录
- 支持三种同步模式：单向和双向
- 冲突处理：跳过已存在的记录，避免数据覆盖
- 性能优化：支持批处理，适合大数据量场景
- 轻量化同步：跳过时间序列数据，只同步核心用户信息

同步模式:
- mongo-to-postgres: 只从MongoDB同步到PostgreSQL
- postgres-to-mongo: 只从PostgreSQL同步到MongoDB  
- bidirectional: 双向同步（默认模式）

安全特性:
- 只添加缺失记录，不修改现有数据
- 详细的进度监控和错误报告
- 支持断点续传和重复执行

使用示例:
  # 双向同步（推荐）
  ./logv2fs sync

  # 只从MongoDB同步到PostgreSQL
  ./logv2fs sync --mode=mongo-to-postgres

  # 只从PostgreSQL同步到MongoDB
  ./logv2fs sync --mode=postgres-to-mongo

  # 自定义批量大小
  ./logv2fs sync --batch-size=200
`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		mode, _ := cmd.Flags().GetString("mode")
		batchSize, _ := cmd.Flags().GetInt("batch-size")

		log.Printf("🔄 开始执行数据库同步，模式: %s", mode)

		// 初始化统计信息
		stats := &SyncStats{
			StartTime: time.Now(),
			Mode:      mode,
			Errors:    []string{},
		}

		// 验证数据库连接
		if err := validateDatabaseConnections(); err != nil {
			log.Fatalf("❌ 数据库连接验证失败: %v", err)
		}

		// 执行同步
		switch mode {
		case "mongo-to-postgres":
			err := syncMongoToPostgres(batchSize, stats)
			if err != nil {
				log.Fatalf("❌ MongoDB到PostgreSQL同步失败: %v", err)
			}
		case "postgres-to-mongo":
			err := syncPostgresToMongo(batchSize, stats)
			if err != nil {
				log.Fatalf("❌ PostgreSQL到MongoDB同步失败: %v", err)
			}
		case "bidirectional":
			// 双向同步：先MongoDB到PostgreSQL，再PostgreSQL到MongoDB
			err := syncMongoToPostgres(batchSize, stats)
			if err != nil {
				log.Fatalf("❌ MongoDB到PostgreSQL同步失败: %v", err)
			}

			err = syncPostgresToMongo(batchSize, stats)
			if err != nil {
				log.Fatalf("❌ PostgreSQL到MongoDB同步失败: %v", err)
			}
		default:
			log.Fatalf("❌ 不支持的同步模式: %s", mode)
		}

		// 输出同步统计信息
		stats.EndTime = time.Now()
		printSyncStats(stats)
	},
}

// validateDatabaseConnections 验证数据库连接
func validateDatabaseConnections() error {
	// 验证MongoDB连接
	mongoClient := database.Client
	if mongoClient == nil {
		return fmt.Errorf("MongoDB连接未初始化")
	}

	// 测试MongoDB连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := mongoClient.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("MongoDB连接测试失败: %v", err)
	}

	// 验证PostgreSQL连接
	postgresDB := database.GetPostgresDB()
	if postgresDB == nil {
		return fmt.Errorf("PostgreSQL连接未初始化")
	}

	// 测试PostgreSQL连接
	sqlDB, err := postgresDB.DB()
	if err != nil {
		return fmt.Errorf("获取PostgreSQL底层连接失败: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("PostgreSQL连接测试失败: %v", err)
	}

	log.Println("✅ 数据库连接验证成功")
	return nil
}

// syncMongoToPostgres 从MongoDB同步到PostgreSQL
func syncMongoToPostgres(batchSize int, stats *SyncStats) error {
	log.Println("📤 开始从MongoDB同步到PostgreSQL...")

	// 获取数据库连接
	mongoClient := database.Client
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	collection := database.OpenCollection(mongoClient, "USER_TRAFFIC_LOGS")

	// 获取PostgreSQL中已存在的email_as_id列表
	existingEmails, err := getExistingEmailsFromPostgres(postgresDB)
	if err != nil {
		return fmt.Errorf("获取PostgreSQL已存在邮箱列表失败: %v", err)
	}

	log.Printf("📊 PostgreSQL中已存在 %d 个用户记录", len(existingEmails))

	// 分批查询MongoDB中不在PostgreSQL中的记录
	var processed int64 = 0
	var synced int64 = 0
	var skipped int64 = 0

	// 创建查询条件：email_as_id不在PostgreSQL的列表中
	filter := bson.M{
		"email_as_id": bson.M{"$nin": existingEmails},
	}

	// 计算需要同步的总数
	totalCount, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("计算需要同步的记录数失败: %v", err)
	}

	log.Printf("📈 发现 %d 个需要从MongoDB同步到PostgreSQL的记录", totalCount)

	if totalCount == 0 {
		log.Println("✅ MongoDB中没有新的用户记录需要同步到PostgreSQL")
		return nil
	}

	// 分批处理
	for skip := int64(0); skip < totalCount; skip += int64(batchSize) {
		// 设置查询选项
		findOptions := options.Find()
		findOptions.SetSkip(skip)
		findOptions.SetLimit(int64(batchSize))

		// 查询一批数据
		cursor, err := collection.Find(context.Background(), filter, findOptions)
		if err != nil {
			return fmt.Errorf("查询MongoDB数据失败: %v", err)
		}

		// 处理这批数据
		var mongoUserLogs []model.UserTrafficLogs
		if err := cursor.All(context.Background(), &mongoUserLogs); err != nil {
			cursor.Close(context.Background())
			return fmt.Errorf("解析MongoDB数据失败: %v", err)
		}
		cursor.Close(context.Background())

		// 转换并保存到PostgreSQL
		for _, mongoUserLog := range mongoUserLogs {
			processed++

			// 转换数据结构（轻量化，跳过时间序列数据）
			pgUserLog, err := convertUserTrafficLogsToPostgresLight(mongoUserLog)
			if err != nil {
				log.Printf("⚠️  转换用户记录失败 [%s]: %v", mongoUserLog.Email_As_Id, err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("转换用户记录失败 [%s]: %v", mongoUserLog.Email_As_Id, err))
				continue
			}

			// 保存到PostgreSQL
			if err := postgresDB.Create(&pgUserLog).Error; err != nil {
				// 检查是否是重复键错误（可能在批处理间隙有新增记录）
				if isUniqueConstraintError(err) {
					skipped++
					log.Printf("⏭️  跳过重复记录: %s", mongoUserLog.Email_As_Id)
					continue
				}

				log.Printf("⚠️  保存用户记录到PostgreSQL失败 [%s]: %v", mongoUserLog.Email_As_Id, err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("保存用户记录失败 [%s]: %v", mongoUserLog.Email_As_Id, err))
				continue
			}

			synced++
		}

		// 打印进度
		if processed%int64(batchSize*5) == 0 || processed == totalCount {
			log.Printf("📈 MongoDB→PostgreSQL 同步进度: %d/%d (已同步: %d, 已跳过: %d)",
				processed, totalCount, synced, skipped)
		}
	}

	stats.MongoToPostgresCount = synced
	log.Printf("✅ MongoDB→PostgreSQL 同步完成: 共处理 %d 条记录，成功同步 %d 条，跳过 %d 条",
		processed, synced, skipped)

	return nil
}

// syncPostgresToMongo 从PostgreSQL同步到MongoDB
func syncPostgresToMongo(batchSize int, stats *SyncStats) error {
	log.Println("📤 开始从PostgreSQL同步到MongoDB...")

	// 获取数据库连接
	mongoClient := database.Client
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	collection := database.OpenCollection(mongoClient, "USER_TRAFFIC_LOGS")

	// 获取MongoDB中已存在的email_as_id列表
	existingEmails, err := getExistingEmailsFromMongo(collection)
	if err != nil {
		return fmt.Errorf("获取MongoDB已存在邮箱列表失败: %v", err)
	}

	log.Printf("📊 MongoDB中已存在 %d 个用户记录", len(existingEmails))

	// 分批查询PostgreSQL中不在MongoDB中的记录
	var processed int64 = 0
	var synced int64 = 0
	var skipped int64 = 0

	// 查询PostgreSQL中email_as_id不在MongoDB列表中的记录数
	var totalCount int64
	query := postgresDB.Model(&model.UserTrafficLogsPG{})
	if len(existingEmails) > 0 {
		query = query.Where("email_as_id NOT IN ?", existingEmails)
	}

	err = query.Count(&totalCount).Error
	if err != nil {
		return fmt.Errorf("计算需要同步的记录数失败: %v", err)
	}

	log.Printf("📈 发现 %d 个需要从PostgreSQL同步到MongoDB的记录", totalCount)

	if totalCount == 0 {
		log.Println("✅ PostgreSQL中没有新的用户记录需要同步到MongoDB")
		return nil
	}

	// 分批处理
	for offset := int64(0); offset < totalCount; offset += int64(batchSize) {
		var pgUserLogs []model.UserTrafficLogsPG

		// 查询一批数据
		query := postgresDB.Model(&model.UserTrafficLogsPG{})
		if len(existingEmails) > 0 {
			query = query.Where("email_as_id NOT IN ?", existingEmails)
		}

		err := query.Offset(int(offset)).Limit(batchSize).Find(&pgUserLogs).Error
		if err != nil {
			return fmt.Errorf("查询PostgreSQL数据失败: %v", err)
		}

		// 转换并保存到MongoDB
		for _, pgUserLog := range pgUserLogs {
			processed++

			// 转换数据结构（轻量化，跳过时间序列数据）
			mongoUserLog, err := convertUserTrafficLogsToMongoLight(pgUserLog)
			if err != nil {
				log.Printf("⚠️  转换用户记录失败 [%s]: %v", pgUserLog.EmailAsId, err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("转换用户记录失败 [%s]: %v", pgUserLog.EmailAsId, err))
				continue
			}

			// 保存到MongoDB
			_, err = collection.InsertOne(context.Background(), mongoUserLog)
			if err != nil {
				// 检查是否是重复键错误
				if isDuplicateKeyError(err) {
					skipped++
					log.Printf("⏭️  跳过重复记录: %s", pgUserLog.EmailAsId)
					continue
				}

				log.Printf("⚠️  保存用户记录到MongoDB失败 [%s]: %v", pgUserLog.EmailAsId, err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("保存用户记录失败 [%s]: %v", pgUserLog.EmailAsId, err))
				continue
			}

			synced++
		}

		// 打印进度
		if processed%int64(batchSize*5) == 0 || processed == totalCount {
			log.Printf("📈 PostgreSQL→MongoDB 同步进度: %d/%d (已同步: %d, 已跳过: %d)",
				processed, totalCount, synced, skipped)
		}
	}

	stats.PostgresToMongoCount = synced
	log.Printf("✅ PostgreSQL→MongoDB 同步完成: 共处理 %d 条记录，成功同步 %d 条，跳过 %d 条",
		processed, synced, skipped)

	return nil
}

// getExistingEmailsFromPostgres 获取PostgreSQL中已存在的email_as_id列表
func getExistingEmailsFromPostgres(db *gorm.DB) ([]string, error) {
	var emails []string
	err := db.Model(&model.UserTrafficLogsPG{}).Pluck("email_as_id", &emails).Error
	if err != nil {
		return nil, err
	}
	return emails, nil
}

// getExistingEmailsFromMongo 获取MongoDB中已存在的email_as_id列表
func getExistingEmailsFromMongo(collection *mongo.Collection) ([]string, error) {
	ctx := context.Background()

	// 使用distinct获取所有不重复的email_as_id
	emails, err := collection.Distinct(ctx, "email_as_id", bson.M{})
	if err != nil {
		return nil, err
	}

	// 转换为字符串切片
	var emailStrings []string
	for _, email := range emails {
		if emailStr, ok := email.(string); ok {
			emailStrings = append(emailStrings, emailStr)
		}
	}

	return emailStrings, nil
}

// convertUserTrafficLogsToPostgresLight 轻量化转换：MongoDB → PostgreSQL（跳过时间序列数据）
func convertUserTrafficLogsToPostgresLight(mongoUserLog model.UserTrafficLogs) (model.UserTrafficLogsPG, error) {
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
		// 注意：故意跳过时间序列数据的转换
		// HourlyLogs, DailyLogs, MonthlyLogs, YearlyLogs 保持为空
	}

	return pgUserLog, nil
}

// convertUserTrafficLogsToMongoLight 轻量化转换：PostgreSQL → MongoDB（跳过时间序列数据）
func convertUserTrafficLogsToMongoLight(pgUserLog model.UserTrafficLogsPG) (model.UserTrafficLogs, error) {
	mongoUserLog := model.UserTrafficLogs{
		Email_As_Id:   pgUserLog.EmailAsId,
		Password:      pgUserLog.Password,
		UUID:          pgUserLog.UUID,
		Role:          pgUserLog.Role,
		Status:        pgUserLog.Status,
		Name:          pgUserLog.Name,
		Token:         pgUserLog.Token,
		Refresh_token: pgUserLog.RefreshToken,
		User_id:       pgUserLog.UserID,
		Used:          pgUserLog.Used,
		Credit:        pgUserLog.Credit,
		CreatedAt:     pgUserLog.CreatedAt,
		UpdatedAt:     pgUserLog.UpdatedAt,
		// 注意：故意跳过时间序列数据的转换
		// HourlyLogs, DailyLogs, MonthlyLogs, YearlyLogs 保持为空
	}

	return mongoUserLog, nil
}

// isUniqueConstraintError 检查是否是唯一约束错误
func isUniqueConstraintError(err error) bool {
	// PostgreSQL唯一约束错误通常包含"duplicate key"或"UNIQUE constraint"
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "UNIQUE constraint") ||
		strings.Contains(errStr, "uniqueIndex")
}

// isDuplicateKeyError 检查是否是MongoDB重复键错误
func isDuplicateKeyError(err error) bool {
	// MongoDB重复键错误通常包含"duplicate key"或错误代码11000
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "E11000") ||
		strings.Contains(errStr, "duplicate")
}

// printSyncStats 输出同步统计信息
func printSyncStats(stats *SyncStats) {
	duration := stats.EndTime.Sub(stats.StartTime)

	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📊 数据库同步统计报告")
	log.Println(strings.Repeat("=", 60))
	log.Printf("🕐 开始时间: %s", stats.StartTime.Format("2006-01-02 15:04:05"))
	log.Printf("🕑 结束时间: %s", stats.EndTime.Format("2006-01-02 15:04:05"))
	log.Printf("⏱️  总耗时: %v", duration)
	log.Printf("🔄 同步模式: %s", stats.Mode)
	log.Println(strings.Repeat("-", 60))
	log.Printf("📤 MongoDB → PostgreSQL: %d 条记录", stats.MongoToPostgresCount)
	log.Printf("📥 PostgreSQL → MongoDB: %d 条记录", stats.PostgresToMongoCount)
	log.Printf("📊 总同步记录数: %d 条", stats.MongoToPostgresCount+stats.PostgresToMongoCount)
	log.Printf("⏭️  跳过冲突记录: %d 条", stats.SkippedConflictsCount)

	if len(stats.Errors) > 0 {
		log.Println(strings.Repeat("-", 60))
		log.Printf("⚠️  错误数量: %d", len(stats.Errors))
		for i, err := range stats.Errors {
			if i < 10 { // 只显示前10个错误
				log.Printf("   %d. %s", i+1, err)
			}
		}
		if len(stats.Errors) > 10 {
			log.Printf("   ... 以及其他 %d 个错误", len(stats.Errors)-10)
		}
	}

	log.Println(strings.Repeat("=", 60))

	if len(stats.Errors) == 0 {
		log.Println("✅ 同步完成，无错误发生")
	} else {
		log.Println("⚠️  同步完成，但发生了一些错误，请检查上述错误信息")
	}
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// 添加命令行参数
	syncCmd.Flags().StringP("mode", "m", "bidirectional", "同步模式: mongo-to-postgres, postgres-to-mongo, bidirectional")
	syncCmd.Flags().IntP("batch-size", "b", 100, "批处理大小，适合调整以优化性能")
}
