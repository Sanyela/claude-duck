package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// 应用配置
	AppName string

	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// OAuth配置
	ClientID     string
	ClientSecret string

	// JWT配置
	JWTSecret string

	// Token过期时间（小时）
	AccessTokenExpireHours int
	// 设备码过期时间（分钟）
	DeviceCodeExpireMinutes int

	// New API配置
	NewAPIEndpoint string
	NewAPIKey      string

	// 积分系统默认配置
	DefaultPromptMultiplier     float64
	DefaultCompletionMultiplier float64

	// 服务降级配置
	DegradationAPIKey            string
	DefaultDegradationGuaranteed int

	// Redis配置
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// SMTP邮件配置
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// 允许注册的邮箱域名
	AllowedEmailDomains []string

	// 验证码配置
	VerificationCodeExpireMinutes int

	// Linux Do OAuth配置
	LinuxDoClientID     string
	LinuxDoClientSecret string
	LinuxDoBaseURL      string
	
	// 前端域名配置
	FrontendURL         string
}

var AppConfig *Config

func LoadConfig() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	AppConfig = &Config{
		// 应用配置
		AppName: getEnv("APP_NAME", "Duck Code"),

		// 数据库配置
		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", ""),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),

		// OAuth配置
		ClientID:     getEnv("OAUTH_CLIENT_ID", ""),
		ClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),

		// JWT配置
		JWTSecret: getEnv("JWT_SECRET_KEY", ""),

		// New API配置
		NewAPIEndpoint: getEnv("NEW_API_ENDPOINT", ""),
		NewAPIKey:      getEnv("NEW_API_KEY", ""),

		// 积分系统默认配置
		DefaultPromptMultiplier:     getEnvAsFloat("DEFAULT_PROMPT_MULTIPLIER", 5.0),
		DefaultCompletionMultiplier: getEnvAsFloat("DEFAULT_COMPLETION_MULTIPLIER", 10.0),

		// 服务降级配置
		DegradationAPIKey:            getEnv("DEGRADATION_API_KEY", ""),
		DefaultDegradationGuaranteed: getEnvAsInt("DEFAULT_DEGRADATION_GUARANTEED", 0),

		// Redis配置
		RedisHost:     getEnv("REDIS_HOST", ""),
		RedisPort:     getEnv("REDIS_PORT", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// SMTP邮件配置
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", ""),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),

		// 允许注册的邮箱域名
		AllowedEmailDomains: getEnvAsSlice("ALLOWED_EMAIL_DOMAINS", []string{}),

		// 验证码配置
		VerificationCodeExpireMinutes: getEnvAsInt("VERIFICATION_CODE_EXPIRE_MINUTES", 10),

		// Linux Do OAuth配置
		LinuxDoClientID:     getEnv("LINUX_DO_CLIENT_ID", ""),
		LinuxDoClientSecret: getEnv("LINUX_DO_CLIENT_SECRET", ""),
		LinuxDoBaseURL:      getEnv("LINUX_DO_BASE_URL", "https://connect.linux.do"),
		
		// 前端域名配置
		FrontendURL:         getEnv("FRONTEND_URL", "https://www.duckcode.top"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
