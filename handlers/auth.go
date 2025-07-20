package handlers

import (
	"log"
	"net/http"

	"claude/config"
	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 注册登录相关请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse 注册登录相关响应结构
type AuthResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Token   string    `json:"token,omitempty"`
	User    *UserData `json:"user,omitempty"`
}

type UserData struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

type UserInfoResponse struct {
	User *UserData `json:"user"`
}

// HandleRegister 用户注册
func HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	err := database.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, AuthResponse{
			Success: false,
			Message: "用户名或邮箱已存在",
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "密码加密失败",
		})
		return
	}

	// 创建用户
	hashedPasswordStr := string(hashedPassword)
	user := models.User{
		Username:              req.Username,
		Email:                 req.Email,
		Password:              &hashedPasswordStr,
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

	log.Printf("用户注册成功: user_id=%d, device_id=%s, ip=%s", user.ID, device.ID, device.IP)

	// 处理新用户注册套餐赠送
	if err := utils.ProcessRegistrationPlanGift(user.ID, "default"); err != nil {
		log.Printf("新用户套餐赠送失败: user_id=%d, error=%v", user.ID, err)
		// 套餐赠送失败不影响注册，继续处理
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

// HandleLogin 用户登录
func HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "请求格式错误：" + err.Error(),
		})
		return
	}

	// 查找用户
	var user models.User
	err := database.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "邮箱或密码错误",
		})
		return
	}

	// 验证密码
	if user.Password == nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "该账户未设置密码，请使用邮箱验证码登录",
		})
		return
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "邮箱或密码错误",
		})
		return
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

	log.Printf("用户登录成功: user_id=%d, device_id=%s, ip=%s", user.ID, device.ID, device.IP)

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

// HandleGetUserInfo 获取用户信息
func HandleGetUserInfo(c *gin.Context) {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 查询用户信息
	var user models.User
	err = database.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, UserInfoResponse{
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			IsAdmin:  user.IsAdmin,
		},
	})
}

// HandleLogout 用户登出
func HandleLogout(c *gin.Context) {
	// 获取当前设备ID
	deviceID, exists := c.Get("deviceID")
	if exists {
		userID := c.GetUint("userID")
		deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
		
		// 下线当前设备
		err := deviceManager.RevokeDevice(userID, deviceID.(string))
		if err != nil {
			log.Printf("登出时下线设备失败: user_id=%d, device_id=%s, error=%v", userID, deviceID, err)
		} else {
			log.Printf("用户登出成功: user_id=%d, device_id=%s", userID, deviceID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}
