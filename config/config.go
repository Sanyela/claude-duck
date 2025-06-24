package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
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
	DefaultTokensPerPoint       int
	DefaultRoundUpEnabled       bool

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
}

var AppConfig *Config

func LoadConfig() {
	AppConfig = &Config{
		// 数据库配置
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "claudecode"),
		DBPassword: getEnv("DB_PASSWORD", "dhEEjzESJLnndSDh"),
		DBName:     getEnv("DB_NAME", "claudecode"),

		// OAuth配置
		ClientID:     getEnv("OAUTH_CLIENT_ID", "Claude Duck"),
		ClientSecret: getEnv("OAUTH_CLIENT_SECRET", "peh63yltlhiue7yv5qs193b3bm9tc02w04acaup5tub6nzrylk9r6gkrvgkzssur"),

		// JWT配置
		JWTSecret: getEnv("JWT_SECRET_KEY", "$xhE6D0gFtYa4sey'Ooy#.LBK*1/9lwfJNuzC3qvkHrdbT7mAMX2j+RQVnUIcZ8i'"),

		// Token过期时间
		AccessTokenExpireHours:  getEnvAsInt("ACCESS_TOKEN_EXPIRE_HOURS", 24),
		DeviceCodeExpireMinutes: getEnvAsInt("DEVICE_CODE_EXPIRE_MINUTES", 15),

		// New API配置
		NewAPIEndpoint: getEnv("NEW_API_ENDPOINT", "http://152.53.82.23:2999"),
		NewAPIKey:      getEnv("NEW_API_KEY", "sk-ijk47MsAmnmJ7sgb0I8Dx6OVXswFBm5Y760tvwpNv3Te0ptp"),

		// 积分系统默认配置
		DefaultPromptMultiplier:     getEnvAsFloat("DEFAULT_PROMPT_MULTIPLIER", 5.0),
		DefaultCompletionMultiplier: getEnvAsFloat("DEFAULT_COMPLETION_MULTIPLIER", 10.0),
		DefaultTokensPerPoint:       getEnvAsInt("DEFAULT_TOKENS_PER_POINT", 10000),
		DefaultRoundUpEnabled:       getEnvAsBool("DEFAULT_ROUND_UP_ENABLED", true),

		// 服务降级配置
		DegradationAPIKey:            getEnv("DEGRADATION_API_KEY", "sk-ijk47MsAmnmJ7sgb0I8Dx6OVXswFBm5Y760tvwpNv3Te0ptp"),
		DefaultDegradationGuaranteed: getEnvAsInt("DEFAULT_DEGRADATION_GUARANTEED", 0),

		// Redis配置
		RedisHost:     getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "EtK67tbzP6kabzhZ"),

		// SMTP邮件配置
		SMTPHost:     getEnv("SMTP_HOST", "smtpdm.aliyun.com"),
		SMTPPort:     getEnv("SMTP_PORT", "25"),
		SMTPUser:     getEnv("SMTP_USER", "no-reply@mail.claude-duck.com"),
		SMTPPassword: getEnv("SMTP_PASSWORD", "ASDasd123456"),
		SMTPFrom:     getEnv("SMTP_FROM", "no-reply@mail.claude-duck.com"),

		// 允许注册的邮箱域名
		AllowedEmailDomains: getEnvAsSlice("ALLOWED_EMAIL_DOMAINS", []string{
			"qq.com",
			"outlook.com",
			"google.com",
			"foxmail.com",
			"163.com",
			"cloxl.com",
			"52ai.org",
		}),

		// 验证码配置
		VerificationCodeExpireMinutes: getEnvAsInt("VERIFICATION_CODE_EXPIRE_MINUTES", 10),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
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
