package database

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var PostgresDB *gorm.DB

// getLogLevel 根据 GIN_MODE 环境变量返回对应的日志级别
func getLogLevel() logger.LogLevel {
	ginMode := os.Getenv("GIN_MODE")

	switch ginMode {
	case "debug":
		return logger.Info // 开发环境，显示详细日志
	case "release":
		return logger.Error // 生产环境，只显示错误信息
	case "test":
		return logger.Silent // 测试环境，静默模式
	default:
		// 默认情况下使用 Info 级别
		return logger.Info
	}
}

// 从URI获取PostgreSQL连接参数
func getConnectionParamsFromURI(connectToSystemDB bool) (string, string, error) {
	// 加载环境变量
	pwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("获取工作目录失败: %v", err)
	}

	if err := godotenv.Load(pwd + "/.env"); err != nil {
		log.Printf("警告: 无法加载.env文件: %v", err)
	}

	// 从环境变量获取PostgreSQL连接URI
	postgresURI := os.Getenv("postgresURI")
	if postgresURI == "" {
		return "", "", fmt.Errorf("环境变量 postgresURI 未设置")
	}

	// 解析 postgresURI 来获取连接信息
	parsedURL, err := url.Parse(postgresURI)
	if err != nil {
		return "", "", fmt.Errorf("解析 postgresURI 失败: %v", err)
	}

	// 提取目标数据库名
	targetDBName := strings.TrimPrefix(parsedURL.Path, "/")
	if targetDBName == "" {
		targetDBName = "logv2fs" // 默认数据库名
	}

	// 如果需要连接到系统数据库，则修改路径
	if connectToSystemDB {
		parsedURL.Path = "/postgres"
	}

	// 构建DSN
	dsn := parsedURL.String()

	// 添加查询参数
	paramSeparator := "?"
	if strings.Contains(dsn, "?") {
		paramSeparator = "&"
	}
	dsn += paramSeparator + "default_query_exec_mode=simple_protocol"

	return dsn, targetDBName, nil
}

// InitPostgreSQL 初始化PostgreSQL连接
func InitPostgreSQL() *gorm.DB {
	// 获取连接参数
	dsn, _, err := getConnectionParamsFromURI(false)
	if err != nil {
		log.Fatal(err)
	}

	// 连接PostgreSQL - 使用动态日志级别
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(getLogLevel()),
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		log.Fatalf("连接PostgreSQL失败: %v", err)
	}

	// 获取底层sql.DB实例来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取底层数据库连接失败: %v", err)
	}

	// 更严格的连接池配置，减少连接复用带来的缓存问题
	sqlDB.SetMaxIdleConns(5)                   // 减少最大空闲连接数
	sqlDB.SetMaxOpenConns(50)                  // 减少最大打开连接数
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // 设置连接最大生存时间，定期刷新连接
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // 设置连接最大空闲时间

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("PostgreSQL连接测试失败: %v", err)
	}

	log.Println("PostgreSQL连接成功")

	PostgresDB = db

	return db
}

// GetPostgresDB 获取PostgreSQL数据库实例
func GetPostgresDB() *gorm.DB {
	if PostgresDB == nil {
		PostgresDB = InitPostgreSQL()
	}
	return PostgresDB
}

// ClosePostgreSQL 关闭PostgreSQL连接
func ClosePostgreSQL() {
	if PostgresDB != nil {
		sqlDB, err := PostgresDB.DB()
		if err != nil {
			log.Printf("获取底层数据库连接失败: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("关闭PostgreSQL连接失败: %v", err)
		} else {
			log.Println("PostgreSQL连接已关闭")
		}
	}
}

// CreateDatabaseIfNotExists 创建数据库（如果不存在）
func CreateDatabaseIfNotExists() error {
	// 获取连接参数
	postgresDSN, targetDBName, err := getConnectionParamsFromURI(true)
	if err != nil {
		return err
	}

	// 连接到 postgres 数据库
	db, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接到postgres数据库失败: %v", err)
	}

	// 检查目标数据库是否存在，如果不存在则创建
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", targetDBName).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("检查数据库是否存在失败: %v", err)
	}

	if count == 0 {
		// 创建数据库
		err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", targetDBName)).Error
		if err != nil {
			return fmt.Errorf("创建数据库失败: %v", err)
		}
		log.Printf("数据库 %s 创建成功", targetDBName)
	} else {
		log.Printf("数据库 %s 已存在", targetDBName)
	}

	// 关闭连接
	sqlDB, _ := db.DB()
	sqlDB.Close()

	return nil
}
