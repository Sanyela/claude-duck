package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID                    uint           `gorm:"primarykey" json:"id"`
	Email                 string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"email"`
	Username              string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"username"`
	Password              *string        `gorm:"" json:"-"`                                   // 密码不在JSON中返回，可为空
	IsAdmin               bool           `gorm:"default:false" json:"is_admin"`              // 是否是管理员
	IsDisabled            bool           `gorm:"default:false" json:"is_disabled"`           // 是否被禁用
	DegradationGuaranteed int            `gorm:"default:0" json:"degradation_guaranteed"`    // 10条内保证不降级的数量
	DegradationSource     string         `gorm:"default:'system'" json:"degradation_source"` // system/admin/subscription
	DegradationLocked     bool           `gorm:"default:false" json:"degradation_locked"`    // 是否锁定，不被套餐覆盖
	DegradationCounter    int64          `gorm:"default:0" json:"degradation_counter"`       // 当前计数器
	FreeModelUsageCount   int64          `gorm:"default:0" json:"free_model_usage_count"`    // 免费模型使用次数
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

// DeviceCode 设备码模型
type DeviceCode struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Code      string    `gorm:"type:varchar(191);uniqueIndex;not null" json:"code"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Used      bool      `gorm:"default:false" json:"used"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Announcement 公告模型
type Announcement struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Type        string         `gorm:"not null" json:"type"` // info, warning, error, success
	Title       string         `gorm:"not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"` // 支持HTML
	Language    string         `gorm:"default:'en'" json:"language"` // 语言标识
	Active      bool           `gorm:"default:true" json:"active"`   // 是否启用
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// AccessToken 访问令牌模型（可选，用于追踪token状态）
type AccessToken struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Token     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// SubscriptionPlan 订阅计划模型
type SubscriptionPlan struct {
	ID                    uint           `gorm:"primarykey" json:"id"`
	Title                 string         `gorm:"not null" json:"title"`        // 标题
	Description           string         `gorm:"type:text" json:"description"` // 描述
	PointAmount           int64          `gorm:"not null" json:"point_amount"` // 套餐包含的积分数
	Price                 float64        `gorm:"not null" json:"price"`        // 套餐价格
	Currency              string         `gorm:"default:'USD'" json:"currency"`
	ValidityDays          int            `gorm:"not null" json:"validity_days"`             // 有效期（天数）
	DegradationGuaranteed int            `gorm:"default:0" json:"degradation_guaranteed"`   // 10条内保证不降级的数量
	DailyCheckinPoints    int64          `gorm:"default:0" json:"daily_checkin_points"`     // 每日签到奖励积分（最低值）
	DailyCheckinPointsMax int64          `gorm:"default:0" json:"daily_checkin_points_max"` // 每日签到奖励积分（最高值）
	DailyMaxPoints        int64          `gorm:"default:0" json:"daily_max_points"`         // 每日最大使用积分数量，0表示无限制
	
	// 自动补给配置
	AutoRefillEnabled   bool  `gorm:"default:false" json:"auto_refill_enabled"`     // 是否启用自动补给
	AutoRefillThreshold int64 `gorm:"default:0" json:"auto_refill_threshold"`       // 自动补给阈值，积分低于此值时触发
	AutoRefillAmount    int64 `gorm:"default:0" json:"auto_refill_amount"`          // 每次补给的积分数量
	
	Features              string         `gorm:"type:text" json:"features"`                 // JSON string array
	Active                bool           `gorm:"default:true" json:"active"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

// Subscription 用户订阅模型
type Subscription struct {
	ID                 uint             `gorm:"primarykey" json:"id"`
	UserID             uint             `gorm:"not null;index" json:"user_id"`
	User               User             `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	SubscriptionPlanID uint             `gorm:"not null" json:"subscription_plan_id"`
	Plan               SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID;references:ID" json:"plan,omitempty"`

	// 状态和时间
	Status      string    `gorm:"not null;index" json:"status"`       // active, expired
	ActivatedAt time.Time `gorm:"not null;index" json:"activated_at"` // 激活时间
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`   // 过期时间

	// 积分统计
	TotalPoints     int64 `gorm:"not null;default:0" json:"total_points"`     // 订阅总积分
	UsedPoints      int64 `gorm:"not null;default:0" json:"used_points"`      // 已使用积分
	AvailablePoints int64 `gorm:"not null;default:0" json:"available_points"` // 可用积分

	// 每日积分限制
	DailyMaxPoints int64 `gorm:"default:0" json:"daily_max_points"` // 每日最大使用积分数量，0表示无限制

	// 来源和支付信息
	SourceType string `gorm:"not null" json:"source_type"`          // activation_code, payment, admin_grant
	SourceID   string `gorm:"type:varchar(191)" json:"source_id"`   // 来源ID（激活码ID/支付ID等）
	InvoiceURL string `gorm:"type:varchar(500)" json:"invoice_url"` // 发票链接

	// 其他信息
	CancelAtPeriodEnd bool           `gorm:"default:false" json:"cancel_at_period_end"`
	CreatedAt         time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// ActivationCode 激活码
type ActivationCode struct {
	ID                 uint             `gorm:"primarykey" json:"id"`
	Code               string           `gorm:"type:varchar(191);uniqueIndex;not null" json:"code"`
	SubscriptionPlanID uint             `gorm:"not null" json:"subscription_plan_id"` // 关联订阅计划ID
	Plan               SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID" json:"plan,omitempty"`
	Status             string           `gorm:"default:'unused'" json:"status"` // unused, used, expired
	UsedByUserID       *uint            `json:"used_by_user_id"`
	UsedBy             *User            `gorm:"foreignKey:UsedByUserID" json:"used_by,omitempty"`
	UsedAt             *time.Time       `json:"used_at"`
	BatchNumber        string           `gorm:"type:varchar(191)" json:"batch_number"` // 批次号
	CreatedAt          time.Time        `json:"created_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ConfigKey   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"config_key"`
	ConfigValue string    `gorm:"type:text;not null" json:"config_value"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APITransaction 统一的API请求和积分使用记录
type APITransaction struct {
	ID     uint `gorm:"primarykey" json:"id"`
	UserID uint `gorm:"not null;index" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// API请求基本信息
	MessageID   string `gorm:"type:varchar(191);index" json:"message_id"` // Claude返回的message_id
	RequestID   string `gorm:"type:varchar(191);index" json:"request_id"` // 请求唯一ID
	Model       string `gorm:"not null;index" json:"model"`               // 使用的模型
	RequestType string `gorm:"default:'api'" json:"request_type"`         // api/stream 请求类型

	// Token使用情况
	InputTokens              int `gorm:"not null" json:"input_tokens"`                 // 输入tokens (prompt_tokens)
	OutputTokens             int `gorm:"not null" json:"output_tokens"`                // 输出tokens (completion_tokens)
	CacheCreationInputTokens int `gorm:"default:0" json:"cache_creation_input_tokens"` // 缓存创建输入tokens
	CacheReadInputTokens     int `gorm:"default:0" json:"cache_read_input_tokens"`     // 缓存读取输入tokens

	// 计费相关
	InputMultiplier  float64 `gorm:"not null" json:"input_multiplier"`    // 输入token倍率 (原prompt_multiplier)
	OutputMultiplier float64 `gorm:"not null" json:"output_multiplier"`   // 输出token倍率 (原completion_multiplier)
	CacheMultiplier  float64 `gorm:"default:1.0" json:"cache_multiplier"` // 缓存token倍率
	PointsUsed       int64   `gorm:"not null" json:"points_used"`         // 消耗的积分

	// 请求详情
	IP          string    `gorm:"type:varchar(45)" json:"ip"`             // 客户端IP
	UID         string    `gorm:"type:varchar(191)" json:"uid"`           // 用户唯一标识
	Username    string    `gorm:"type:varchar(191)" json:"username"`      // 用户名
	Status      string    `gorm:"not null;index" json:"status"`           // success/failed/billing_failed
	Error       string    `gorm:"type:text" json:"error,omitempty"`       // 错误信息（如果有）
	Duration    int       `gorm:"not null" json:"duration"`               // 请求耗时（毫秒）
	ServiceTier string    `gorm:"default:'standard'" json:"service_tier"` // 服务等级
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
}

// DailyCheckin 每日签到记录
type DailyCheckin struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CheckinDate string    `gorm:"type:varchar(10);not null;index" json:"checkin_date"` // 签到日期 YYYY-MM-DD
	Points      int64     `gorm:"not null" json:"points"`                              // 获得的积分
	CreatedAt   time.Time `gorm:"index" json:"created_at"`

	// 复合唯一索引：一个用户每天只能签到一次
}

// 添加表名方法
func (DailyCheckin) TableName() string {
	return "daily_checkins"
}

// DailyPointsUsage 每日积分使用记录
type DailyPointsUsage struct {
	ID             uint         `gorm:"primarykey" json:"id"`
	UserID         uint         `gorm:"not null;index" json:"user_id"`
	User           User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SubscriptionID uint         `gorm:"not null;index" json:"subscription_id"`
	Subscription   Subscription `gorm:"foreignKey:SubscriptionID" json:"subscription,omitempty"`
	UsageDate      string       `gorm:"type:varchar(10);not null;index" json:"usage_date"` // 使用日期 YYYY-MM-DD
	PointsUsed     int64        `gorm:"not null;default:0" json:"points_used"`             // 当日已使用积分
	CreatedAt      time.Time    `gorm:"index" json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	// 复合唯一索引：一个用户的一个订阅每天一条记录
}

// 添加表名方法
func (DailyPointsUsage) TableName() string {
	return "daily_points_usage"
}

// GiftRecord 卡密赠送记录
type GiftRecord struct {
	ID                 uint             `gorm:"primarykey" json:"id"`
	FromAdminID        *uint            `gorm:"index" json:"from_admin_id"`             // 赠送的管理员ID，NULL表示系统赠送
	FromAdmin          User             `gorm:"foreignKey:FromAdminID" json:"from_admin,omitempty"`
	ToUserID           uint             `gorm:"not null;index" json:"to_user_id"` // 接收的用户ID
	ToUser             User             `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
	SubscriptionPlanID uint             `gorm:"not null" json:"subscription_plan_id"` // 赠送的订阅计划ID
	Plan               SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID" json:"plan,omitempty"`

	// 赠送内容
	PointsAmount   int64  `gorm:"not null" json:"points_amount"`     // 赠送的积分数量
	ValidityDays   int    `gorm:"not null" json:"validity_days"`     // 有效天数
	DailyMaxPoints int64  `gorm:"default:0" json:"daily_max_points"` // 每日最大使用积分数量，0表示无限制
	Reason         string `gorm:"type:varchar(500)" json:"reason"`   // 赠送原因

	// 状态和结果
	Status         string `gorm:"default:'pending';index" json:"status"` // pending, completed, failed
	SubscriptionID *uint  `json:"subscription_id"`                       // 生成的订阅ID（成功时）
	ErrorMessage   string `gorm:"type:text" json:"error_message"`        // 失败原因（失败时）

	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 添加表名方法
func (GiftRecord) TableName() string {
	return "gift_records"
}

// UserWallet 用户钱包模型
type UserWallet struct {
	UserID uint `gorm:"primarykey" json:"user_id"` // 用户ID作为主键

	// 积分相关
	TotalPoints     int64 `gorm:"not null;default:0" json:"total_points"`     // 总积分 (历史累计充值)
	AvailablePoints int64 `gorm:"not null;default:0" json:"available_points"` // 可用积分
	UsedPoints      int64 `gorm:"not null;default:0" json:"used_points"`      // 已使用积分
	
	// 累计token计费相关
	AccumulatedTokens int64 `gorm:"not null;default:0" json:"accumulated_tokens"` // 累计加权token数量

	// 当前生效的订阅属性 (来自最新激活的套餐)
	DailyMaxPoints        int64 `gorm:"default:0" json:"daily_max_points"`       // 每日最大使用积分，0表示无限制
	DegradationGuaranteed int   `gorm:"default:0" json:"degradation_guaranteed"` // 保证不降级数量

	// 签到相关 (来自当前套餐)
	DailyCheckinPoints    int64 `gorm:"default:0" json:"daily_checkin_points"`     // 每日签到积分(最低)
	DailyCheckinPointsMax int64 `gorm:"default:0" json:"daily_checkin_points_max"` // 每日签到积分(最高)

	// 自动补给配置 (来自当前套餐)
	AutoRefillEnabled   bool  `gorm:"default:false" json:"auto_refill_enabled"`   // 是否启用自动补给
	AutoRefillThreshold int64 `gorm:"default:0" json:"auto_refill_threshold"`     // 自动补给阈值
	AutoRefillAmount    int64 `gorm:"default:0" json:"auto_refill_amount"`        // 每次补给积分数量
	LastAutoRefillTime  *time.Time `gorm:"" json:"last_auto_refill_time"`         // 最后一次自动补给时间

	// 钱包状态
	WalletExpiresAt time.Time `gorm:"not null" json:"wallet_expires_at"`       // 钱包过期时间 (最晚的订阅过期时间)
	Status          string    `gorm:"not null;default:'active'" json:"status"` // active, expired

	// 统计信息
	LastCheckinDate string `gorm:"type:varchar(10)" json:"last_checkin_date"` // 最后签到日期 YYYY-MM-DD

	// 时间戳
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联用户
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// 添加表名方法
func (UserWallet) TableName() string {
	return "user_wallets"
}

// RedemptionRecord 兑换记录表 - 替代原订阅表，记录所有兑换历史
type RedemptionRecord struct {
	ID     uint `gorm:"primarykey" json:"id"`
	UserID uint `gorm:"not null;index" json:"user_id"` // 用户ID

	// 兑换来源
	SourceType string `gorm:"not null;index" json:"source_type"`  // activation_code, admin_gift, daily_checkin, payment
	SourceID   string `gorm:"type:varchar(191)" json:"source_id"` // 来源标识

	// 兑换内容
	PointsAmount int64 `gorm:"not null" json:"points_amount"` // 兑换的积分数量 (可为负数，表示扣减)
	ValidityDays int   `gorm:"not null" json:"validity_days"` // 有效期天数

	// 套餐属性 (如果是套餐兑换)
	SubscriptionPlanID    *uint `json:"subscription_plan_id"`                    // 关联的订阅计划ID (可为空)
	DailyMaxPoints        int64 `gorm:"default:0" json:"daily_max_points"`       // 每日限制
	DegradationGuaranteed int   `gorm:"default:0" json:"degradation_guaranteed"` // 降级保证
	DailyCheckinPoints    int64 `gorm:"default:0" json:"daily_checkin_points"`   // 签到积分范围
	DailyCheckinPointsMax int64 `gorm:"default:0" json:"daily_checkin_points_max"`
	
	// 自动补给属性
	AutoRefillEnabled   bool  `gorm:"default:false" json:"auto_refill_enabled"`   // 自动补给开关
	AutoRefillThreshold int64 `gorm:"default:0" json:"auto_refill_threshold"`     // 补给阈值
	AutoRefillAmount    int64 `gorm:"default:0" json:"auto_refill_amount"`        // 补给数量

	// 记录信息
	ActivatedAt time.Time `gorm:"not null;index" json:"activated_at"` // 激活时间
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`   // 过期时间
	Reason      string    `gorm:"type:varchar(500)" json:"reason"`    // 兑换原因/描述

	// 关联信息
	BatchNumber string `gorm:"type:varchar(191)" json:"batch_number"` // 批次号 (激活码相关)
	InvoiceURL  string `gorm:"type:varchar(500)" json:"invoice_url"`  // 发票链接 (支付相关)
	AdminUserID *uint  `json:"admin_user_id"`                         // 操作管理员ID (admin_gift相关)

	// 时间戳
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	User             User              `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	SubscriptionPlan *SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID;references:ID" json:"plan,omitempty"`
	AdminUser        *User             `gorm:"foreignKey:AdminUserID;references:ID" json:"admin_user,omitempty"`
}

// 添加表名方法
func (RedemptionRecord) TableName() string {
	return "redemption_records"
}

// UserDailyUsage 用户每日使用记录 - 简化版，不再按订阅分组
type UserDailyUsage struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	UserID     uint   `gorm:"not null;index" json:"user_id"`                     // 用户ID
	UsageDate  string `gorm:"type:varchar(10);not null;index" json:"usage_date"` // 使用日期 YYYY-MM-DD
	PointsUsed int64  `gorm:"not null;default:0" json:"points_used"`             // 当日已使用积分

	// 时间戳
	CreatedAt time.Time `gorm:"index" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联用户
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// 添加表名方法
func (UserDailyUsage) TableName() string {
	return "user_daily_usage"
}

// OAuthAccount OAuth第三方账号关联表
type OAuthAccount struct {
	ID       uint `gorm:"primarykey" json:"id"`
	UserID   uint `gorm:"not null;index" json:"user_id"`           // 关联的本地用户ID
	User     User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// OAuth基本信息
	Provider    string `gorm:"not null;index" json:"provider"`     // oauth提供商：linux_do, github, google等
	ProviderUID string `gorm:"not null;index" json:"provider_uid"` // 第三方用户唯一ID
	Email       string `gorm:"type:varchar(191)" json:"email"`     // 第三方邮箱

	// Linux Do Connect 特有字段
	Username        string `gorm:"type:varchar(191)" json:"username"`         // 论坛用户名
	Name            string `gorm:"type:varchar(191)" json:"name"`              // 论坛用户昵称
	AvatarTemplate  string `gorm:"type:varchar(500)" json:"avatar_template"`  // 头像模板URL
	Active          bool   `gorm:"default:true" json:"active"`                 // 账号活跃状态
	TrustLevel      int    `gorm:"default:0" json:"trust_level"`               // 信任等级（0-4）
	Silenced        bool   `gorm:"default:false" json:"silenced"`              // 禁言状态
	ExternalIDs     string `gorm:"type:text" json:"external_ids"`              // 外部ID关联信息（JSON）
	APIKey          string `gorm:"type:varchar(255)" json:"api_key"`           // API访问密钥

	// OAuth令牌信息
	AccessToken  string     `gorm:"type:text" json:"access_token"`   // 访问令牌
	RefreshToken string     `gorm:"type:text" json:"refresh_token"`  // 刷新令牌  
	TokenExpiry  *time.Time `json:"token_expiry"`                    // 令牌过期时间

	// 同步信息
	LastSyncAt  *time.Time `json:"last_sync_at"`                    // 最后同步时间
	SyncEnabled bool       `gorm:"default:true" json:"sync_enabled"` // 是否启用数据同步

	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 添加表名方法
func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}
