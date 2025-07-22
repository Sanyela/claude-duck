package database

import (
	"context"
	"fmt"
	"log"

	"claude/config"
	"claude/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var TokenRedisClient *redis.Client  // DB 0: Token映射
var UserRedisClient *redis.Client   // DB 1: 用户设备集合
var DeviceRedisClient *redis.Client // DB 2: 设备详情

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
		&models.GiftRecord{},
		&models.UserWallet{},
		&models.RedemptionRecord{},
		&models.UserDailyUsage{},
		&models.OAuthAccount{},
		&models.ConversationLog{},
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

	// 确保新架构表的索引
	ensureNewArchitectureIndexes()

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
			ConfigKey:   "token_threshold",
			ConfigValue: "5000",
			Description: "累计token计费阈值",
		},
		{
			ConfigKey:   "points_per_threshold",
			ConfigValue: "1",
			Description: "每阈值扣费积分数量",
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
		{
			ConfigKey:   "registration_plan_mapping",
			ConfigValue: `{"default": -1, "linux_do": -1, "github": -1, "google": -1}`,
			Description: "用户注册套餐映射，JSON格式：{\"注册方式\": 套餐ID}，default为普通注册，-1表示不赠送",
		},
		{
			ConfigKey:   "model_redirect_map",
			ConfigValue: `{}`,
			Description: "模型重定向映射，JSON格式：{\"原始模型\": \"目标模型\"}，空对象表示不重定向",
		},
		{
			ConfigKey:   "model_multiplier_map",
			ConfigValue: `{}`,
			Description: "模型倍率映射，JSON格式：{\"模型名\": 倍率}，在现有计费基础上乘以对应倍率，空对象表示不应用额外倍率",
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

// ensureNewArchitectureIndexes 确保新架构表的索引
func ensureNewArchitectureIndexes() {
	// 先检查索引是否存在
	var indexCount int64
	checkSQL := `SELECT COUNT(*) FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = 'user_wallets' 
		AND index_name = 'idx_user_wallets_user_id'`

	if err := DB.Raw(checkSQL).Scan(&indexCount).Error; err != nil {
		log.Printf("检查新架构表索引失败: %v", err)
		return
	}

	if indexCount == 0 {
		// 创建复合唯一索引：一个用户只能有一个钱包
		indexSQL := `CREATE UNIQUE INDEX idx_user_wallets_user_id ON user_wallets (user_id)`
		if err := DB.Exec(indexSQL).Error; err != nil {
			log.Printf("创建新架构表唯一索引失败: %v", err)
		} else {
			log.Println("✅ 新架构表唯一索引创建完成")
		}
	} else {
		log.Println("✅ 新架构表唯一索引已存在")
	}
}

// InitRedis 初始化Redis连接
func InitRedis() error {
	cfg := config.AppConfig

	// 初始化Token映射Redis客户端 (DB 0)
	TokenRedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0, // Token映射
	})

	// 初始化用户设备集合Redis客户端 (DB 3)
	UserRedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       3, // 用户设备集合 (避开DB 1邮箱验证和DB 2 Bing缓存)
	})

	// 初始化设备详情Redis客户端 (DB 4)
	DeviceRedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       4, // 设备详情
	})

	// 测试所有连接
	ctx := context.Background()
	
	if _, err := TokenRedisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to connect to Token Redis (DB 0): %w", err)
	}

	if _, err := UserRedisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to connect to User Redis (DB 3): %w", err)
	}

	if _, err := DeviceRedisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("failed to connect to Device Redis (DB 4): %w", err)
	}

	log.Println("All Redis clients connected successfully")
	log.Println("- DB 0: Token映射")
	log.Println("- DB 1: 邮箱验证码 (已存在)")
	log.Println("- DB 2: Bing图片缓存 (已存在)")
	log.Println("- DB 3: 用户设备集合")
	log.Println("- DB 4: 设备详情")
	return nil
}
