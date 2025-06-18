package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var PostgresDB *gorm.DB

// InitPostgreSQL 初始化PostgreSQL连接
func InitPostgreSQL() *gorm.DB {
	// 加载环境变量
	pwd, err := os.Getwd()
	if err != nil {
		log.Panic("获取工作目录失败: ", err)
	}

	if err := godotenv.Load(pwd + "/.env"); err != nil {
		log.Printf("警告: 无法加载.env文件: %v", err)
	}

	// 从环境变量获取PostgreSQL连接URI
	postgresURI := os.Getenv("postgresURI")
	if postgresURI == "" {
		log.Fatal("环境变量 postgresURI 未设置")
	}

	// 连接PostgreSQL
	db, err := gorm.Open(postgres.Open(postgresURI), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 开启SQL日志
		// 在迁移过程中禁用外键约束，避免迁移时的复杂性
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

	// 配置连接池
	sqlDB.SetMaxIdleConns(10)   // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)  // 最大打开连接数
	sqlDB.SetConnMaxLifetime(0) // 连接最大生存时间

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

// getEnvOrDefault 获取环境变量，如果不存在则使用默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateDatabaseIfNotExists 创建数据库（如果不存在）
func CreateDatabaseIfNotExists() error {
	// 连接到postgres数据库来创建目标数据库
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	username := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "")
	sslmode := getEnvOrDefault("POSTGRES_SSLMODE", "disable")
	dbname := getEnvOrDefault("POSTGRES_DB", "logv2fs")

	// 先连接到postgres数据库
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		host, username, password, port, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("连接到postgres数据库失败: %v", err)
	}

	// 检查目标数据库是否存在，如果不存在则创建
	var count int64
	err = db.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", dbname).Scan(&count).Error
	if err != nil {
		return fmt.Errorf("检查数据库是否存在失败: %v", err)
	}

	if count == 0 {
		// 创建数据库
		err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname)).Error
		if err != nil {
			return fmt.Errorf("创建数据库失败: %v", err)
		}
		log.Printf("数据库 %s 创建成功", dbname)
	} else {
		log.Printf("数据库 %s 已存在", dbname)
	}

	// 关闭连接
	sqlDB, _ := db.DB()
	sqlDB.Close()

	return nil
}
