package database

import (
	"fmt"
	"log"

	"claude/config"
	"claude/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	cfg := config.AppConfig

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// 连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的 sql.DB 对象来配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connected successfully")
	return nil
}

// Migrate 执行数据库迁移
func Migrate() error {
	err := DB.AutoMigrate(
		&models.User{},
		&models.DeviceCode{},
		&models.Announcement{},
		&models.AccessToken{},
		&models.SubscriptionPlan{},
		&models.Subscription{},
		&models.PaymentHistory{},
		&models.PointBalance{},
		&models.PointPool{},
		&models.APITransaction{},
		&models.ActivationCode{},
		&models.SystemConfig{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化默认系统配置
	initDefaultConfigs()

	log.Println("Database migration completed")
	return nil
}

// initDefaultConfigs 初始化默认系统配置
func initDefaultConfigs() {
	defaultConfigs := []models.SystemConfig{
		{
			ConfigKey:   "prompt_multiplier",
			ConfigValue: "5",
			Description: "提示token倍率",
		},
		{
			ConfigKey:   "completion_multiplier",
			ConfigValue: "10",
			Description: "补全token倍率",
		},
		{
			ConfigKey:   "tokens_per_point",
			ConfigValue: "10000",
			Description: "多少token等于1积分",
		},
		{
			ConfigKey:   "round_up_enabled",
			ConfigValue: "false",
			Description: "是否向上取整",
		},
		{
			ConfigKey:   "new_api_endpoint",
			ConfigValue: config.AppConfig.NewAPIEndpoint,
			Description: "New API接入点",
		},
		{
			ConfigKey:   "new_api_key",
			ConfigValue: config.AppConfig.NewAPIKey,
			Description: "New API密钥",
		},
		{
			ConfigKey:   "degradation_api_key",
			ConfigValue: config.AppConfig.DegradationAPIKey,
			Description: "服务降级API密钥",
		},
		{
			ConfigKey:   "default_degradation_guaranteed",
			ConfigValue: fmt.Sprintf("%d", config.AppConfig.DefaultDegradationGuaranteed),
			Description: "默认10条内保证不降级数量",
		},
	}

	for _, cfg := range defaultConfigs {
		var existing models.SystemConfig
		if err := DB.Where("config_key = ?", cfg.ConfigKey).First(&existing).Error; err != nil {
			// 配置不存在，创建新配置
			if err := DB.Create(&cfg).Error; err != nil {
				log.Printf("Failed to create default config %s: %v", cfg.ConfigKey, err)
			}
		}
	}
}
