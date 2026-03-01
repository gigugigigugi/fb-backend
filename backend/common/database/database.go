package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// 为了兼容当前已写好的基于 database.DB 的代码，这里我们可以临时保留 DB（但打上不推荐使用的注记），
// 或者你也完全可以将现有业务模块中用到 database.DB 的地方也重构掉（通过依赖注入传入 Repository）。
// 在此演示：新代码请统一使用 GlobalRepo
var DB *gorm.DB

// Init 初始化数据库仓储
func Init(dsn string) Repository {
	return initPostgres(dsn)
}

func initPostgres(dsn string) Repository {
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
	return NewPostgresRepository(DB)
}
