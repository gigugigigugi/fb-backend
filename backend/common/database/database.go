package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 为了兼容当前已写好的基于 database.DB 的代码，这里我们可以临时保留 DB（但打上不推荐使用的注记），
// 在此演示：新代码请统一注入 domain Repo
var DB *gorm.DB

// Init 初始化数据库仓储并返回 *gorm.DB
func Init(dsn string) *gorm.DB {
	return initPostgres(dsn)
}

func initPostgres(dsn string) *gorm.DB {
	var err error

	newLogger := gormlogger.Default.LogMode(gormlogger.Info)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database object: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Postgres Database connection initialized successfully")
	return DB
}
