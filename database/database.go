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
		&models.APITransaction{},
		&models.ActivationCode{},
		&models.SystemConfig{},
		&models.DailyCheckin{},
		&models.DailyPointsUsage{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 初始化默认系统配置
	initDefaultConfigs()

	// 确保签到表的唯一索引
	ensureCheckinTableIndexes()

	// 确保每日积分使用表的唯一索引
	ensureDailyPointsUsageIndexes()

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
		{
			ConfigKey:   "daily_checkin_enabled",
			ConfigValue: "true",
			Description: "是否启用每日签到功能",
		},
		{
			ConfigKey:   "daily_checkin_points",
			ConfigValue: "10",
			Description: "每日签到奖励积分数量",
		},
		{
			ConfigKey:   "daily_checkin_validity_days",
			ConfigValue: "1",
			Description: "签到奖励积分有效期（天）",
		},
		{
			ConfigKey:   "daily_checkin_multi_subscription_strategy",
			ConfigValue: "highest",
			Description: "多个订阅时的签到积分策略（highest=最高，lowest=最低）",
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

// ensureCheckinTableIndexes 确保签到表的唯一索引
func ensureCheckinTableIndexes() {
	// 先检查索引是否存在
	var indexCount int64
	checkSQL := `SELECT COUNT(*) FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = 'daily_checkins' 
		AND index_name = 'idx_daily_checkins_user_date'`

	if err := DB.Raw(checkSQL).Scan(&indexCount).Error; err != nil {
		log.Printf("检查签到表索引失败: %v", err)
		return
	}

	if indexCount == 0 {
		// 创建复合唯一索引：一个用户每天只能签到一次
		indexSQL := `CREATE UNIQUE INDEX idx_daily_checkins_user_date ON daily_checkins (user_id, checkin_date)`
		if err := DB.Exec(indexSQL).Error; err != nil {
			log.Printf("创建签到表唯一索引失败: %v", err)
		} else {
			log.Println("✅ 签到表唯一索引创建完成")
		}
	} else {
		log.Println("✅ 签到表唯一索引已存在")
	}
}

// ensureDailyPointsUsageIndexes 确保每日积分使用表的唯一索引
func ensureDailyPointsUsageIndexes() {
	// 先检查索引是否存在
	var indexCount int64
	checkSQL := `SELECT COUNT(*) FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = 'daily_points_usage' 
		AND index_name = 'idx_daily_points_usage_user_date'`

	if err := DB.Raw(checkSQL).Scan(&indexCount).Error; err != nil {
		log.Printf("检查每日积分使用表索引失败: %v", err)
		return
	}

	if indexCount == 0 {
		// 创建复合唯一索引：一个用户每天只能使用一次积分
		indexSQL := `CREATE UNIQUE INDEX idx_daily_points_usage_user_date ON daily_points_usage (user_id, usage_date)`
		if err := DB.Exec(indexSQL).Error; err != nil {
			log.Printf("创建每日积分使用表唯一索引失败: %v", err)
		} else {
			log.Println("✅ 每日积分使用表唯一索引创建完成")
		}
	} else {
		log.Println("✅ 每日积分使用表唯一索引已存在")
	}
}
