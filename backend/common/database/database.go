package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Init 初始化数据库仓储并返回 *gorm.DB
func Init(dsn string) *gorm.DB {
	return initPostgres(dsn)
}

func initPostgres(dsn string) *gorm.DB {
	newLogger := gormlogger.Default.LogMode(gormlogger.Info)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database object: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Postgres Database connection initialized successfully")
	return db
}
