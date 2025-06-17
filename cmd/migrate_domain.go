package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/xvv6u577/logv2fs/database"
	"github.com/xvv6u577/logv2fs/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

// migrateDomainData 迁移Domain数据的具体实现
func migrateDomainData(batchSize int, skipExisting bool, stats *model.MigrationStats) error {
	log.Println("🔗 开始迁移Domain数据...")

	// 获取数据库连接
	mongoClient := database.Client
	postgresDB := database.GetPostgresDB()

	// 获取MongoDB集合
	collection := database.OpenCollection(mongoClient, "GLOBAL") // 假设集合名为domains

	// 计算总数
	totalCount, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return fmt.Errorf("获取Domain总数失败: %v", err)
	}

	log.Printf("📊 发现 %d 个Domain记录需要迁移", totalCount)

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
			return fmt.Errorf("查询Domain数据失败: %v", err)
		}

		// 处理这批数据
		var mongoDomains []model.Domain
		if err := cursor.All(context.Background(), &mongoDomains); err != nil {
			cursor.Close(context.Background())
			return fmt.Errorf("解析Domain数据失败: %v", err)
		}
		cursor.Close(context.Background())

		// 转换并保存到PostgreSQL
		for _, mongoDomain := range mongoDomains {
			processed++

			// 检查是否跳过已存在的记录
			if skipExisting {
				var existingCount int64
				err := postgresDB.Model(&model.DomainPG{}).
					Where("domain = ?", mongoDomain.Domain).
					Count(&existingCount).Error
				if err != nil {
					log.Printf("⚠️  检查Domain重复失败: %v", err)
					stats.Errors = append(stats.Errors, fmt.Sprintf("检查Domain重复失败: %v", err))
					continue
				}

				if existingCount > 0 {
					skipped++
					continue
				}
			}

			// 转换数据结构
			pgDomain := convertDomainToPG(mongoDomain)

			// 保存到PostgreSQL
			if err := postgresDB.Create(&pgDomain).Error; err != nil {
				log.Printf("⚠️  保存Domain失败: %v", err)
				stats.Errors = append(stats.Errors, fmt.Sprintf("保存Domain失败: %v", err))
				continue
			}

			migrated++
		}

		// 打印进度
		if processed%int64(batchSize*5) == 0 || processed == totalCount {
			log.Printf("📈 Domain迁移进度: %d/%d (已迁移: %d, 已跳过: %d)",
				processed, totalCount, migrated, skipped)
		}
	}

	stats.DomainRecordsMigrated = migrated
	log.Printf("✅ Domain迁移完成: 共处理 %d 条记录，成功迁移 %d 条，跳过 %d 条",
		processed, migrated, skipped)

	return nil
}

// convertDomainToPG 将MongoDB的Domain转换为PostgreSQL的DomainPG
func convertDomainToPG(mongoDomain model.Domain) model.DomainPG {
	return model.DomainPG{
		ID:           uuid.New(),
		Type:         mongoDomain.Type,
		Remark:       mongoDomain.Remark,
		Domain:       mongoDomain.Domain,
		IP:           mongoDomain.IP,
		SNI:          mongoDomain.SNI,
		UUID:         mongoDomain.UUID,
		Path:         mongoDomain.PATH,
		ServerPort:   mongoDomain.SERVER_PORT,
		Password:     mongoDomain.PASSWORD,
		PublicKey:    mongoDomain.PUBLIC_KEY,
		ShortID:      mongoDomain.SHORT_ID,
		EnableOpenai: mongoDomain.EnableOpenai,
	}
}

// validateDomainData 验证Domain数据的完整性
func validateDomainData(domain *model.DomainPG) error {
	if domain.Domain == "" {
		return fmt.Errorf("域名不能为空")
	}

	if domain.Type == "" {
		domain.Type = "work" // 设置默认类型
	}

	// 验证类型是否合法
	validTypes := map[string]bool{
		"work":      true,
		"vmesstls":  true,
		"vmessws":   true,
		"reality":   true,
		"hysteria2": true,
		"vlessCDN":  true,
	}

	if !validTypes[domain.Type] {
		return fmt.Errorf("无效的域名类型: %s", domain.Type)
	}

	return nil
}

// createDomainIndexes 为Domain表创建额外的索引
func createDomainIndexes(db *gorm.DB) error {
	log.Println("🔍 为Domain表创建索引...")

	// 为domain字段创建唯一索引（如果还没有）
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_domains_domain_unique ON domains(domain)").Error; err != nil {
		return fmt.Errorf("创建domain唯一索引失败: %v", err)
	}

	// 为type字段创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_domains_type ON domains(type)").Error; err != nil {
		return fmt.Errorf("创建type索引失败: %v", err)
	}

	// 为enable_openai字段创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_domains_enable_openai ON domains(enable_openai)").Error; err != nil {
		return fmt.Errorf("创建enable_openai索引失败: %v", err)
	}

	// 为创建时间创建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_domains_created_at ON domains(created_at)").Error; err != nil {
		return fmt.Errorf("创建created_at索引失败: %v", err)
	}

	log.Println("✅ Domain索引创建完成")
	return nil
}
