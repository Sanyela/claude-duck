package handlers

import (
	"net/http"

	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 注册登录相关请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
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
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
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

	// 为新用户创建默认积分余额
	creditBalance := models.CreditBalance{
		UserID:              user.ID,
		Available:           1000, // 默认1000积分
		Total:               1000,
		RechargeRatePerHour: 0,
		CanRequestReset:     true,
	}
	database.DB.Create(&creditBalance)

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
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
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

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		User: &UserData{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
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
		},
	})
}

// HandleLogout 用户登出（可选，主要是客户端清除token）
func HandleLogout(c *gin.Context) {
	// 这里可以将token加入黑名单，或者其他登出逻辑
	// 目前主要依赖客户端清除token
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}
