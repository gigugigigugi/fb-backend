package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AppConfig 是全局配置结构体。
type AppConfig struct {
	Env        string // debug / release / test
	Port       string // 监听端口
	AuthBypass bool   // 是否绕过鉴权（仅开发期）

	DB           DBConfig
	JWT          JWTConfig
	Verification VerificationConfig
}

// DBConfig 是数据库配置。
type DBConfig struct {
	DSN string // PostgreSQL DSN
}

// JWTConfig 是 JWT 配置。
type JWTConfig struct {
	Secret string // JWT 签名密钥
	Exp    int    // 过期小时数
}

// VerificationConfig 是验证码发送配置。
type VerificationConfig struct {
	Provider       string // mock / real
	HTTPTimeoutSec int
	EmailEnabled   bool // 是否启用邮箱验证码发送与验证流程。
	SMSEnabled     bool // 是否启用短信验证码发送与验证流程。
	Email          VerificationEmailConfig
	SMS            VerificationSMSConfig
}

// VerificationEmailConfig 是邮箱发送配置。
type VerificationEmailConfig struct {
	APIURL string
	APIKey string
	From   string
}

// VerificationSMSConfig 是短信发送配置。
type VerificationSMSConfig struct {
	APIURL string
	APIKey string
	Sender string
}

// App 是全局配置实例。
var App *AppConfig

// Load 从环境变量加载配置。
func Load() {
	env := getEnv("GIN_MODE", "debug")
	if err := godotenv.Load(); err != nil && env == "debug" {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	App = &AppConfig{
		Env:        env,
		Port:       getEnv("PORT", "8080"),
		AuthBypass: getEnvBool("AUTH_BYPASS", false),
		DB: DBConfig{
			DSN: getEnv("DB_DSN", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default-secret-do-not-use-in-prod"),
			Exp:    getEnvInt("JWT_EXP_HOURS", 72),
		},
		Verification: VerificationConfig{
			Provider:       getEnv("VERIFICATION_PROVIDER", "mock"),
			HTTPTimeoutSec: getEnvInt("VERIFICATION_HTTP_TIMEOUT_SEC", 8),
			// 为了控制成本，默认关闭 email/sms 验证流程；需要时可在环境变量显式开启。
			EmailEnabled: getEnvBool("VERIFICATION_EMAIL_ENABLED", false),
			SMSEnabled:   getEnvBool("VERIFICATION_SMS_ENABLED", false),
			Email: VerificationEmailConfig{
				APIURL: getEnv("VERIFICATION_EMAIL_API_URL", ""),
				APIKey: getEnv("VERIFICATION_EMAIL_API_KEY", ""),
				From:   getEnv("VERIFICATION_EMAIL_FROM", ""),
			},
			SMS: VerificationSMSConfig{
				APIURL: getEnv("VERIFICATION_SMS_API_URL", ""),
				APIKey: getEnv("VERIFICATION_SMS_API_KEY", ""),
				Sender: getEnv("VERIFICATION_SMS_SENDER", ""),
			},
		},
	}

	if App.DB.DSN == "" && App.Env != "test" {
		log.Println("Warning: DB_DSN is empty")
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseBool(strValue); err == nil {
		return value
	}
	return fallback
}
