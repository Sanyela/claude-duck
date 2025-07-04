package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// 请求结构体
type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=register login"`
}

type RegisterWithCodeRequest struct {
	Username string  `json:"username" binding:"required,min=5,max=20"`
	Email    string  `json:"email" binding:"required,email"`
	Password *string `json:"password,omitempty"` // 密码现在是可选的
	Code     string  `json:"code" binding:"required,len=6"`
}

type LoginWithCodeRequest struct {
	EmailOrUsername string `json:"email_or_username" binding:"required"`
	Password        string `json:"password" binding:"required"`
	Code            string `json:"code" binding:"required,len=6"`
}

// 新增：邮箱验证码一键登录/注册请求
type EmailOnlyAuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Username string `json:"username,omitempty"` // 可选，仅在注册时需要
}

// 邮箱检查请求结构体
type CheckEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// 邮箱检查响应结构体
type CheckEmailResponse struct {
	Success    bool   `json:"success"`
	UserExists bool   `json:"user_exists"`
	ActionType string `json:"action_type"` // "login" or "register"
	Message    string `json:"message"`
}

// 外部Redis客户端引用
var redisClientForAuth *redis.Client

// 初始化Redis客户端用于认证
func InitAuthRedisClient() {
	redisClientForAuth = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
		Password: config.AppConfig.RedisPassword,
		DB:       1, // 使用DB1专门存储验证码
	})
}

// HandleSendVerificationCode 发送验证码
func HandleSendVerificationCode(c *gin.Context) {
	var req SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查邮箱域名是否允许
	if !utils.IsAllowedEmailDomain(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的邮箱域名，仅支持: " + strings.Join(config.AppConfig.AllowedEmailDomains, ", "),
		})
		return
	}

	// 检查邮箱是否已存在（注册时）或不存在（登录时）
	var existingUser models.User
	userExists := database.DB.Where("email = ?", req.Email).First(&existingUser).Error == nil

	if req.Type == "register" && userExists {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "邮箱已被注册",
		})
		return
	}

	if req.Type == "login" && !userExists {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "邮箱未注册",
		})
		return
	}

	// 检查验证码发送频率限制（1分钟内只能发送一次）
	ctx := context.Background()
	rateLimitKey := fmt.Sprintf("email_rate_limit:%s", req.Email)
	exists, err := redisClientForAuth.Exists(ctx, rateLimitKey).Result()
	if err == nil && exists > 0 {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "发送过于频繁，请稍后再试",
		})
		return
	}

	// 生成验证码
	code := utils.GenerateVerificationCode()

	// 存储验证码到Redis
	verificationKey := fmt.Sprintf("email_verification:%s:%s", req.Email, req.Type)
	expireDuration := time.Duration(config.AppConfig.VerificationCodeExpireMinutes) * time.Minute

	err = redisClientForAuth.Set(ctx, verificationKey, code, expireDuration).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "验证码存储失败",
		})
		return
	}

	// 设置发送频率限制（1分钟）
	redisClientForAuth.Set(ctx, rateLimitKey, "1", time.Minute)

	// 发送邮件
	err = utils.SendVerificationEmail(req.Email, code, req.Type)
	if err != nil {
		// 如果发送失败，删除Redis中的验证码
		redisClientForAuth.Del(ctx, verificationKey)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "验证码发送失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "验证码已发送，请查收邮件",
	})
}

// HandleRegisterWithCode 带验证码的注册
func HandleRegisterWithCode(c *gin.Context) {
	var req RegisterWithCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查邮箱域名是否允许
	if !utils.IsAllowedEmailDomain(req.Email) {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "不支持的邮箱域名",
		})
		return
	}

	// 验证验证码
	ctx := context.Background()
	verificationKey := fmt.Sprintf("email_verification:%s:register", req.Email)
	storedCode, err := redisClientForAuth.Get(ctx, verificationKey).Result()
	if err != nil || storedCode != req.Code {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "验证码错误或已过期",
		})
		return
	}

	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	err = database.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, AuthResponse{
			Success: false,
			Message: "用户名或邮箱已存在",
		})
		return
	}

	// 创建用户
	user := models.User{
		Username:              req.Username,
		Email:                 req.Email,
		DegradationGuaranteed: config.AppConfig.DefaultDegradationGuaranteed,
		DegradationSource:     "system",
		DegradationLocked:     false,
		DegradationCounter:    0,
	}

	// 如果提供了密码，则加密存储
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "密码加密失败",
			})
			return
		}
		hashedPasswordStr := string(hashedPassword)
		user.Password = &hashedPasswordStr
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "创建用户失败",
		})
		return
	}

	// 删除已使用的验证码
	redisClientForAuth.Del(ctx, verificationKey)

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "生成访问令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "注册成功",
		Token:   token,
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// HandleLoginWithCode 带验证码的登录（支持用户名或邮箱）
func HandleLoginWithCode(c *gin.Context) {
	var req LoginWithCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 查找用户（支持用户名或邮箱）
	var user models.User
	err := database.DB.Where("email = ? OR username = ?", req.EmailOrUsername, req.EmailOrUsername).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "用户不存在",
		})
		return
	}

	// 验证密码
	if user.Password == nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "该账户未设置密码，无法使用密码验证",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "密码错误",
		})
		return
	}

	// 验证验证码
	ctx := context.Background()
	verificationKey := fmt.Sprintf("email_verification:%s:login", user.Email)
	storedCode, err := redisClientForAuth.Get(ctx, verificationKey).Result()
	if err != nil || storedCode != req.Code {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "验证码错误或已过期",
		})
		return
	}

	// 删除已使用的验证码
	redisClientForAuth.Del(ctx, verificationKey)

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "生成访问令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// HandleEmailOnlyAuth 邮箱验证码一键登录/注册
func HandleEmailOnlyAuth(c *gin.Context) {
	var req EmailOnlyAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查邮箱域名是否允许
	if !utils.IsAllowedEmailDomain(req.Email) {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "不支持的邮箱域名",
		})
		return
	}

	// 验证验证码
	ctx := context.Background()

	// 先尝试登录验证码
	loginVerificationKey := fmt.Sprintf("email_verification:%s:login", req.Email)
	storedCode, err := redisClientForAuth.Get(ctx, loginVerificationKey).Result()

	isLogin := false
	if err == nil && storedCode == req.Code {
		isLogin = true
	} else {
		// 尝试注册验证码
		registerVerificationKey := fmt.Sprintf("email_verification:%s:register", req.Email)
		storedCode, err = redisClientForAuth.Get(ctx, registerVerificationKey).Result()
		if err != nil || storedCode != req.Code {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "验证码错误或已过期",
			})
			return
		}
	}

	// 查找用户是否存在
	var user models.User
	userExists := database.DB.Where("email = ?", req.Email).First(&user).Error == nil

	if isLogin {
		// 登录流程
		if !userExists {
			c.JSON(http.StatusNotFound, AuthResponse{
				Success: false,
				Message: "账户不存在",
			})
			return
		}

		// 删除已使用的验证码
		redisClientForAuth.Del(ctx, loginVerificationKey)
	} else {
		// 注册流程
		if userExists {
			c.JSON(http.StatusConflict, AuthResponse{
				Success: false,
				Message: "该邮箱已被注册，请使用登录验证码",
			})
			return
		}

		// 如果是注册，需要用户名
		if req.Username == "" {
			c.JSON(http.StatusBadRequest, AuthResponse{
				Success: false,
				Message: "注册需要提供用户名",
			})
			return
		}

		// 检查用户名是否已存在
		var existingUser models.User
		if database.DB.Where("username = ?", req.Username).First(&existingUser).Error == nil {
			c.JSON(http.StatusConflict, AuthResponse{
				Success: false,
				Message: "用户名已存在",
			})
			return
		}

		// 创建新用户（无密码）
		user = models.User{
			Username:              req.Username,
			Email:                 req.Email,
			Password:              nil, // 无密码
			DegradationGuaranteed: config.AppConfig.DefaultDegradationGuaranteed,
			DegradationSource:     "system",
			DegradationLocked:     false,
			DegradationCounter:    0,
		}

		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "创建用户失败",
			})
			return
		}

		// 删除已使用的验证码
		registerVerificationKey := fmt.Sprintf("email_verification:%s:register", req.Email)
		redisClientForAuth.Del(ctx, registerVerificationKey)
	}

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "生成访问令牌失败",
		})
		return
	}

	// 注册设备到Redis
	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	device, err := deviceManager.RegisterDevice(
		user.ID,
		token,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"web",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "设备注册失败",
		})
		return
	}

	log.Printf("用户邮箱验证码%s成功: user_id=%d, device_id=%s, ip=%s",
		map[bool]string{true: "登录", false: "注册"}[isLogin],
		user.ID, device.ID, device.IP)

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: map[bool]string{true: "登录成功", false: "注册成功"}[isLogin],
		Token:   token,
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// HandleCheckEmail 检查邮箱是否已注册
func HandleCheckEmail(c *gin.Context) {
	var req CheckEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CheckEmailResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查邮箱域名是否允许
	if !utils.IsAllowedEmailDomain(req.Email) {
		c.JSON(http.StatusBadRequest, CheckEmailResponse{
			Success: false,
			Message: "不支持的邮箱域名，仅支持: " + strings.Join(config.AppConfig.AllowedEmailDomains, ", "),
		})
		return
	}

	// 查询用户是否存在
	var existingUser models.User
	userExists := database.DB.Where("email = ?", req.Email).First(&existingUser).Error == nil

	actionType := "register"
	message := "新用户，将进行注册流程"
	
	if userExists {
		actionType = "login"
		message = "用户已存在，将进行登录流程"
	}

	c.JSON(http.StatusOK, CheckEmailResponse{
		Success:    true,
		UserExists: userExists,
		ActionType: actionType,
		Message:    message,
	})
}
