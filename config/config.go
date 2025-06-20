package config

import (
	"os"
	"strconv"
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
}

var AppConfig *Config

func LoadConfig() {
	AppConfig = &Config{
		// 数据库配置
		DBHost:     getEnv("DB_HOST", "111.180.197.234"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "claudecode"),
		DBPassword: getEnv("DB_PASSWORD", "dhEEjzESJLnndSDh"),
		DBName:     getEnv("DB_NAME", "claudecode"),

		// OAuth配置
		ClientID:     getEnv("OAUTH_CLIENT_ID", "c35a52681f1fa87a6a11f69d26990326"),
		ClientSecret: getEnv("OAUTH_CLIENT_SECRET", "2935467f5e0e1d383a51a467c9680091dc29015291245dbb6b440adcaf9e1011"),

		// JWT配置
		JWTSecret: getEnv("JWT_SECRET_KEY", "claude-code-jwt-secret-key-change-in-production"),

		// Token过期时间
		AccessTokenExpireHours:  getEnvAsInt("ACCESS_TOKEN_EXPIRE_HOURS", 24),
		DeviceCodeExpireMinutes: getEnvAsInt("DEVICE_CODE_EXPIRE_MINUTES", 15),
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
