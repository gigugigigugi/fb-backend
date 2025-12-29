package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 全局数据库连接实例
var DB *gorm.DB

// Init 初始化 Gorm 的 Postgres 连接
func Init(dsn string) {
	var err error

	// 使用 Gorm 自带的默认 Logger 并设置为 Info 级别，以便在开发过程中查看 SQL 语句
	newLogger := gormlogger.Default.LogMode(gormlogger.Info)

	// 连接数据库
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 获取底层的 sql.DB 对象来配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database object: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection initialized successfully")
}
