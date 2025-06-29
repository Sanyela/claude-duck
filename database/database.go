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

	// 检查是否需要执行订阅表架构重构迁移
	if needsSubscriptionRefactor() {
		log.Println("检测到需要执行订阅表架构重构迁移...")
		if err := MigrateSubscriptionRefactor(); err != nil {
			log.Printf("订阅表架构重构迁移失败: %v", err)
			log.Println("如果这是首次运行，可以忽略此错误")
		} else {
			// 标记迁移已完成
			markSubscriptionRefactorComplete()
		}
	}

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
			ConfigKey:   "cache_multiplier",
			ConfigValue: "0.2",
			Description: "缓存token倍率",
		},
		{
			ConfigKey:   "token_pricing_table",
			ConfigValue: `{"0":2,"7680":3,"15360":4,"23040":5,"30720":6,"38400":7,"46080":8,"53760":9,"61440":10,"69120":11,"76800":12,"84480":13,"92160":14,"99840":15,"107520":16,"115200":17,"122880":18,"130560":19,"138240":20,"145920":21,"153600":22,"161280":23,"168960":24,"176640":25,"184320":25,"192000":25,"200000":25}`,
			Description: "基于总token的阶梯积分扣费表",
		},
		{
			ConfigKey:   "free_models_list",
			ConfigValue: `["claude-3-5-haiku-20241022"]`,
			Description: "免费模型列表，JSON数组格式",
		},
		{
			ConfigKey:   "new_api_endpoint",
			ConfigValue: config.AppConfig.NewAPIEndpoint,
			Description: "New API 端点地址",
		},
		{
			ConfigKey:   "new_api_key",
			ConfigValue: config.AppConfig.NewAPIKey,
			Description: "New API 密钥",
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

// needsSubscriptionRefactor 检查是否需要执行订阅表重构迁移
func needsSubscriptionRefactor() bool {
	// 检查是否存在迁移标记
	var config models.SystemConfig
	err := DB.Where("config_key = ?", "subscription_refactor_completed").First(&config).Error
	if err == nil && config.ConfigValue == "true" {
		return false // 迁移已完成
	}

	// 检查是否存在旧的积分池表
	var tableExists bool
	err = DB.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'point_pools')").Scan(&tableExists).Error
	if err != nil || !tableExists {
		return false // 表不存在，不需要迁移
	}

	// 检查积分池表是否有数据
	var count int64
	err = DB.Raw("SELECT COUNT(*) FROM point_pools").Scan(&count).Error
	if err != nil || count == 0 {
		return false // 没有数据需要迁移
	}

	// 检查订阅表是否已经有新字段
	var columnExists bool
	err = DB.Raw("SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'subscriptions' AND column_name = 'activated_at')").Scan(&columnExists).Error
	if err == nil && columnExists {
		return false // 新字段已存在，可能已迁移
	}

	log.Printf("检测到 %d 条积分池数据需要迁移", count)
	return true
}

// markSubscriptionRefactorComplete 标记订阅表重构迁移已完成
func markSubscriptionRefactorComplete() {
	config := models.SystemConfig{
		ConfigKey:   "subscription_refactor_completed",
		ConfigValue: "true",
		Description: "订阅表架构重构迁移完成标记",
	}

	// 使用 UPSERT 语义，如果存在就更新，不存在就创建
	var existingConfig models.SystemConfig
	err := DB.Where("config_key = ?", config.ConfigKey).First(&existingConfig).Error
	if err != nil {
		// 记录不存在，创建新记录
		if err := DB.Create(&config).Error; err != nil {
			log.Printf("创建迁移完成标记失败: %v", err)
		}
	} else {
		// 记录存在，更新现有记录
		if err := DB.Model(&existingConfig).Updates(config).Error; err != nil {
			log.Printf("更新迁移完成标记失败: %v", err)
		}
	}
}
