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
	Password              string         `gorm:"not null" json:"-"`                          // 密码不在JSON中返回
	IsAdmin               bool           `gorm:"default:false" json:"is_admin"`              // 是否是管理员
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
	ValidityDays          int            `gorm:"not null" json:"validity_days"`           // 有效期（天数）
	DegradationGuaranteed int            `gorm:"default:0" json:"degradation_guaranteed"` // 10条内保证不降级的数量
	Features              string         `gorm:"type:text" json:"features"`               // JSON string array
	Active                bool           `gorm:"default:true" json:"active"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

// Subscription 用户订阅模型
type Subscription struct {
	ID                 uint             `gorm:"primarykey" json:"id"`
	UserID             uint             `gorm:"not null" json:"user_id"`
	User               User             `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	SubscriptionPlanID uint             `gorm:"not null" json:"subscription_plan_id"`
	Plan               SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID;references:ID" json:"plan,omitempty"`
	ExternalID         string           `gorm:"type:varchar(191);uniqueIndex" json:"external_id"` // 外部支付系统ID
	Status             string           `gorm:"not null" json:"status"`                           // active, canceled, past_due
	CurrentPeriodEnd   time.Time        `gorm:"not null" json:"current_period_end"`
	CancelAtPeriodEnd  bool             `gorm:"default:false" json:"cancel_at_period_end"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`
}

// PaymentHistory 支付历史模型
type PaymentHistory struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PlanName    string         `gorm:"not null" json:"plan_name"`
	Amount      float64        `gorm:"not null" json:"amount"`
	Currency    string         `gorm:"not null" json:"currency"`
	Status      string         `gorm:"not null" json:"status"` // paid, failed
	InvoiceURL  string         `json:"invoice_url"`
	PaymentDate time.Time      `gorm:"not null" json:"payment_date"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// PointBalance 用户积分余额模型
type PointBalance struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserID          uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TotalPoints     int64     `gorm:"default:0" json:"total_points"`     // 总积分
	UsedPoints      int64     `gorm:"default:0" json:"used_points"`      // 已使用积分
	AvailablePoints int64     `gorm:"default:0" json:"available_points"` // 可用积分
	UpdatedAt       time.Time `json:"updated_at"`
}

// PointPool 积分池模型
type PointPool struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserID          uint      `gorm:"not null;index:idx_user_expires" json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SourceType      string    `gorm:"not null" json:"source_type"`                       // activation_code/admin_grant/subscription
	SourceID        string    `gorm:"type:varchar(191)" json:"source_id"`                // 来源ID（激活码ID/订阅ID等）
	PointsTotal     int64     `gorm:"not null" json:"points_total"`                      // 充值积分总量
	PointsRemaining int64     `gorm:"not null" json:"points_remaining"`                  // 剩余积分
	ExpiresAt       time.Time `gorm:"not null;index:idx_user_expires" json:"expires_at"` // 过期时间
	CreatedAt       time.Time `json:"created_at"`
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
	IP          string `gorm:"type:varchar(45)" json:"ip"`             // 用户IP
	UID         string `gorm:"type:varchar(191)" json:"uid"`           // 用户标识
	Username    string `gorm:"type:varchar(191)" json:"username"`      // 用户名
	Status      string `gorm:"not null;index" json:"status"`           // success, failed
	Error       string `gorm:"type:text" json:"error"`                 // 错误信息
	Duration    int    `json:"duration"`                               // 请求耗时（毫秒）
	ServiceTier string `gorm:"default:'standard'" json:"service_tier"` // Claude服务等级
	StopReason  string `json:"stop_reason"`                            // 停止原因

	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
