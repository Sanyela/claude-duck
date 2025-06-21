package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Email     string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"email"`
	Username  string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"not null" json:"-"` // 密码不在JSON中返回
	IsAdmin   bool           `gorm:"default:false" json:"is_admin"` // 是否是管理员
	GroupID   *uint          `json:"group_id"` // 用户分组ID
	Group     *UserGroup     `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
	Type        string         `gorm:"not null" json:"type"`        // info, warning, error, success
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
	ID            uint           `gorm:"primarykey" json:"id"`
	PlanID        string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"plan_id"`
	Name          string         `gorm:"not null" json:"name"`
	CreditAmount  float64        `gorm:"not null" json:"credit_amount"` // 套餐包含的额度（美元）
	Price         float64        `gorm:"not null" json:"price"` // 套餐价格
	Currency      string         `gorm:"default:'USD'" json:"currency"`
	ValidityDays  int            `gorm:"not null" json:"validity_days"` // 有效期（天数）
	GroupID       *uint          `json:"group_id"` // 关联的用户分组
	Group         *UserGroup     `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	Features      string         `gorm:"type:text" json:"features"` // JSON string array
	Active        bool           `gorm:"default:true" json:"active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// Subscription 用户订阅模型
type Subscription struct {
	ID                uint             `gorm:"primarykey" json:"id"`
	UserID            uint             `gorm:"not null" json:"user_id"`
	User              User             `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	SubscriptionPlanID uint            `gorm:"not null" json:"subscription_plan_id"`
	Plan              SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID;references:ID" json:"plan,omitempty"`
	ExternalID        string           `gorm:"type:varchar(191);uniqueIndex" json:"external_id"` // 外部支付系统ID
	Status            string           `gorm:"not null" json:"status"` // active, canceled, past_due
	CurrentPeriodEnd  time.Time        `gorm:"not null" json:"current_period_end"`
	CancelAtPeriodEnd bool             `gorm:"default:false" json:"cancel_at_period_end"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"-"`
}

// PaymentHistory 支付历史模型
type PaymentHistory struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"not null" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PlanName     string         `gorm:"not null" json:"plan_name"`
	Amount       float64        `gorm:"not null" json:"amount"`
	Currency     string         `gorm:"not null" json:"currency"`
	Status       string         `gorm:"not null" json:"status"` // paid, failed
	InvoiceURL   string         `json:"invoice_url"`
	PaymentDate  time.Time      `gorm:"not null" json:"payment_date"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreditBalance 用户额度余额模型
type CreditBalance struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserID          uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User            User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TotalAmount     float64   `gorm:"default:0" json:"total_amount"` // 总额度（美元）
	UsedAmount      float64   `gorm:"default:0" json:"used_amount"` // 已使用额度（美元）
	AvailableAmount float64   `gorm:"default:0" json:"available_amount"` // 可用额度（美元）
	UpdatedAt       time.Time `json:"updated_at"`
}

// ModelCost 模型成本配置
type ModelCost struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	ModelID              string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"model_id"`
	ModelName            string         `gorm:"not null" json:"model_name"`
	InputPricePerK       float64        `gorm:"not null" json:"input_price_per_k"` // 输入价格（每1K tokens）
	OutputPricePerK      float64        `gorm:"not null" json:"output_price_per_k"` // 输出价格（每1K tokens）
	ModelMultiplier      float64        `gorm:"not null" json:"model_multiplier"` // 模型倍率
	CompletionMultiplier float64        `gorm:"not null" json:"completion_multiplier"` // 补全倍率
	Status               string         `gorm:"not null" json:"status"` // available, unavailable, limited
	Description          string         `json:"description"`
	Active               bool           `gorm:"default:true" json:"active"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreditUsageHistory 额度使用历史
type CreditUsageHistory struct {
	ID                   uint      `gorm:"primarykey" json:"id"`
	UserID               uint      `gorm:"not null" json:"user_id"`
	User                 User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RequestID            string    `gorm:"type:varchar(191)" json:"request_id"` // API请求ID
	Description          string    `gorm:"not null" json:"description"`
	Amount               float64   `gorm:"not null" json:"amount"` // 金额（美元），正数表示充值，负数表示消费
	InputTokens          int       `json:"input_tokens"` // 输入tokens
	OutputTokens         int       `json:"output_tokens"` // 输出tokens
	ModelName            string    `json:"model_name"` // 使用的模型
	GroupMultiplier      float64   `json:"group_multiplier"` // 分组倍率
	ModelMultiplier      float64   `json:"model_multiplier"` // 模型倍率
	CompletionMultiplier float64   `json:"completion_multiplier"` // 补全倍率
	CalculationDetails   string    `gorm:"type:text" json:"calculation_details"` // 计算详情（JSON）
	CreatedAt            time.Time `json:"created_at"`
}

// UserGroup 用户分组
type UserGroup struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	Name            string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"name"`
	GroupMultiplier float64        `gorm:"not null;default:1" json:"group_multiplier"` // 分组倍率
	Description     string         `gorm:"type:text" json:"description"`
	Active          bool           `gorm:"default:true" json:"active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// APIChannel API渠道配置
type APIChannel struct {
	ID                  uint           `gorm:"primarykey" json:"id"`
	Name                string         `gorm:"type:varchar(191);not null" json:"name"`
	BaseURL             string         `gorm:"not null" json:"base_url"`
	APIKey              string         `gorm:"not null" json:"api_key"`
	Weight              int            `gorm:"default:1" json:"weight"` // 权重（用于负载均衡）
	Status              string         `gorm:"default:'active'" json:"status"` // active, inactive, error
	HealthCheckURL      string         `json:"health_check_url"`
	LastHealthCheckTime *time.Time     `json:"last_health_check_time"`
	ResponseTimeMs      int            `json:"response_time_ms"` // 平均响应时间（毫秒）
	Active              bool           `gorm:"default:true" json:"active"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// ActivationCode 激活码
type ActivationCode struct {
	ID               uint             `gorm:"primarykey" json:"id"`
	Code             string           `gorm:"type:varchar(191);uniqueIndex;not null" json:"code"`
	Type             string           `gorm:"not null" json:"type"` // plan, credit
	SubscriptionPlanID *uint          `json:"subscription_plan_id"` // 关联套餐ID（type=plan时）
	Plan             *SubscriptionPlan `gorm:"foreignKey:SubscriptionPlanID" json:"plan,omitempty"`
	CreditAmount     float64          `json:"credit_amount"` // 额度数量（type=credit时）
	Status           string           `gorm:"default:'unused'" json:"status"` // unused, used, expired
	UsedByUserID     *uint            `json:"used_by_user_id"`
	UsedBy           *User            `gorm:"foreignKey:UsedByUserID" json:"used_by,omitempty"`
	UsedAt           *time.Time       `json:"used_at"`
	ExpiresAt        *time.Time       `json:"expires_at"`
	BatchNumber      string           `gorm:"type:varchar(191)" json:"batch_number"` // 批次号
	CreatedAt        time.Time        `json:"created_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`
}

// APIRequest API请求记录
type APIRequest struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	UserID         uint           `gorm:"not null" json:"user_id"`
	User           User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	MessageID      string         `gorm:"type:varchar(191)" json:"message_id"` // Claude的message_id
	ChannelID      uint           `gorm:"not null" json:"channel_id"`
	Channel        APIChannel     `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Model          string         `gorm:"not null" json:"model"`
	InputTokens    int            `gorm:"not null" json:"input_tokens"`
	OutputTokens   int            `gorm:"not null" json:"output_tokens"`
	TotalCost      float64        `gorm:"not null" json:"total_cost"` // 计算后的费用（美元）
	Status         string         `gorm:"not null" json:"status"` // success, failed
	ErrorMessage   string         `gorm:"type:text" json:"error_message"`
	RequestTime    time.Time      `gorm:"not null" json:"request_time"`
	ResponseTime   time.Time      `json:"response_time"`
	DurationMs     int            `json:"duration_ms"` // 请求耗时（毫秒）
	CreatedAt      time.Time      `json:"created_at"`
}

// StreamingSession 流式会话
type StreamingSession struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	RequestID         uint      `gorm:"uniqueIndex;not null" json:"request_id"`
	Request           APIRequest `gorm:"foreignKey:RequestID" json:"request,omitempty"`
	UserID            uint      `gorm:"not null" json:"user_id"`
	AccumulatedInput  int       `gorm:"default:0" json:"accumulated_input"` // 累计输入tokens
	AccumulatedOutput int       `gorm:"default:0" json:"accumulated_output"` // 累计输出tokens
	Status            string    `gorm:"default:'active'" json:"status"` // active, completed, error
	StartTime         time.Time `gorm:"not null" json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// BillingRule 计费规则配置
type BillingRule struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	Name           string         `gorm:"type:varchar(191);not null" json:"name"`
	BasePrice      float64        `gorm:"not null;default:0.002" json:"base_price"` // 基准价格
	Formula        string         `gorm:"type:text" json:"formula"` // 计算公式配置（JSON）
	Description    string         `gorm:"type:text" json:"description"`
	EffectiveFrom  time.Time      `gorm:"not null" json:"effective_from"`
	EffectiveTo    *time.Time     `json:"effective_to"`
	Active         bool           `gorm:"default:true" json:"active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}