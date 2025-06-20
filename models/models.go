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
	ID           uint           `gorm:"primarykey" json:"id"`
	PlanID       string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"plan_id"`
	Name         string         `gorm:"not null" json:"name"`
	PricePerMonth float64       `gorm:"not null" json:"price_per_month"`
	Currency     string         `gorm:"default:'USD'" json:"currency"`
	Features     string         `gorm:"type:text" json:"features"` // JSON string array
	Active       bool           `gorm:"default:true" json:"active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
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

// CreditBalance 积分余额模型
type CreditBalance struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	UserID              uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User                User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Available           int       `gorm:"default:0" json:"available"`
	Total               int       `gorm:"default:0" json:"total"`
	RechargeRatePerHour int       `gorm:"default:0" json:"recharge_rate_per_hour"`
	CanRequestReset     bool      `gorm:"default:true" json:"can_request_reset"`
	NextResetTime       *time.Time `json:"next_reset_time"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// ModelCost 模型成本配置
type ModelCost struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	ModelID     string         `gorm:"type:varchar(191);uniqueIndex;not null" json:"model_id"`
	ModelName   string         `gorm:"not null" json:"model_name"`
	Status      string         `gorm:"not null" json:"status"` // available, unavailable, limited
	CostFactor  *float64       `json:"cost_factor"`
	Description string         `json:"description"`
	Active      bool           `gorm:"default:true" json:"active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// CreditUsageHistory 积分使用历史
type CreditUsageHistory struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Description  string    `gorm:"not null" json:"description"`
	Amount       int       `gorm:"not null" json:"amount"` // 正数表示增加，负数表示消费
	RelatedModel string    `json:"related_model"`
	CreatedAt    time.Time `json:"created_at"`
}