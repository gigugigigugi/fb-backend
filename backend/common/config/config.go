package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig 全局配置结构体
type AppConfig struct {
	Env  string // dev, prod, test
	Port string // 8080

	DB  DBConfig
	JWT JWTConfig
}

type DBConfig struct {
	DSN string // 数据库连接字符串
}

type JWTConfig struct {
	Secret string // 签名密钥
	Exp    int    // 过期时间(小时)
}

// App 全局配置实例，其他包直接用 config.App.DB.DSN 访问
var App *AppConfig

// Load 初始化配置
func Load() {
	// 1. 尝试加载 .env 文件，并处理错误
	env := getEnv("GIN_MODE", "dev")
	if err := godotenv.Load(); err != nil && env == "dev" {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	App = &AppConfig{
		Env:  env,
		Port: getEnv("PORT", "8080"),
		DB: DBConfig{
			DSN: getEnv("DB_DSN", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default-secret-do-not-use-in-prod"),
			Exp:    getEnvInt("JWT_EXP_HOURS", 72),
		},
	}

	// 2. 关键配置检查
	if App.DB.DSN == "" && App.Env != "test" {
		log.Println("Warning: DB_DSN is empty")
	}
}

// getEnv 获取环境变量，如果为空则返回默认值
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
