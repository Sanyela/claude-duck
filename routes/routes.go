package routes

import (
	"claude/handlers"
	"claude/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine) {
	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "claude-code-api",
		})
	})

	// 无需认证的公共API
	public := r.Group("/api")
	{
		// Bing每日图片API
		public.GET("/bing", handlers.GetBingDailyImage)
		// 公告API
		public.GET("/announcements", handlers.HandleAnnouncements)
	}

	// 认证相关路由
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.HandleRegister)
		auth.POST("/login", handlers.HandleLogin)
		auth.POST("/logout", middleware.JWTAuth(), handlers.HandleLogout)   // 登出需要token验证
		auth.GET("/user", middleware.JWTAuth(), handlers.HandleGetUserInfo) // 需要token验证

		// 邮箱验证码相关路由
		auth.POST("/check-email", handlers.HandleCheckEmail) // 新增：检查邮箱是否已注册
		auth.POST("/send-verification-code", handlers.HandleSendVerificationCode)
		auth.POST("/register-with-code", handlers.HandleRegisterWithCode)
		auth.POST("/login-with-code", handlers.HandleLoginWithCode)
		auth.POST("/email-auth", handlers.HandleEmailOnlyAuth) // 邮箱验证码一键登录/注册
	}

	// OAuth相关路由
	oauth := r.Group("/api/oauth")
	{
		// Linux Do OAuth - 只保留必要的API路由
		linuxdo := oauth.Group("/linux-do")
		{
			linuxdo.GET("/config", handlers.HandleLinuxDoConfig)       // 获取配置状态
			linuxdo.GET("/authorize", handlers.HandleLinuxDoAuthorize) // 生成授权URL
		}
	}

	// OAuth辅助路由（无需认证）
	oauthHelper := r.Group("/api/oauth/linuxdo")
	{
		oauthHelper.GET("/temp-user", handlers.HandleGetTemporaryLinuxDoUser)          // 获取临时用户信息
		oauthHelper.POST("/complete-registration", handlers.HandleCompleteLinuxDoRegistration) // 完成注册
	}

	// Linux Do固定回调路径 - 直接在根路由处理
	r.GET("/oauth/linuxdo", handlers.HandleLinuxDoCallback)

	// SSO相关路由
	sso := r.Group("/api/sso")
	{
		// authorize需要用户登录（JWT认证）
		sso.POST("/authorize", middleware.JWTAuth(), handlers.HandleAuthorize)
		// 以下两个端点使用客户端凭据认证，不需要用户JWT
		sso.POST("/verify-token", handlers.HandleVerifyToken)
		sso.POST("/verify-code", handlers.HandleVerifyCode)
	}

	// API路由（需要认证）
	api := r.Group("/api")
	api.Use(middleware.JWTAuth()) // 应用JWT认证中间件到整个组
	{
		// 订阅相关路由
		api.GET("/subscription/active", handlers.HandleGetActiveSubscription)
		api.GET("/subscription/history", handlers.HandleGetSubscriptionHistory)
		api.POST("/subscription/redeem/preview", handlers.HandleRedeemCouponPreview) // 预检查接口
		api.POST("/subscription/redeem", handlers.HandleRedeemCoupon)

		// 积分相关路由
		api.GET("/credits/balance", handlers.HandleGetCreditBalance)
		api.GET("/credits/model-costs", handlers.HandleGetModelCosts)
		api.GET("/credits/history", handlers.HandleGetCreditUsageHistory)
		api.GET("/credits/pricing-table", handlers.HandleGetPricingTable)
		api.GET("/credits/daily-usage", handlers.HandleGetDailyUsage)

		// 签到相关路由
		api.GET("/checkin/status", handlers.HandleGetCheckinStatus)
		api.POST("/checkin", handlers.HandleDailyCheckin)

		// Claude API 代理路由
		api.POST("/claude", handlers.HandleClaudeProxy)
		api.Any("/claude/*path", handlers.HandleClaudeProxy) // 支持所有方法和子路径

		// 设备管理路由
		devices := api.Group("/devices")
		{
			devices.GET("", handlers.GetDevices)                     // 获取设备列表
			devices.DELETE("/:deviceId", handlers.RevokeDevice)      // 下线指定设备
			devices.DELETE("", handlers.RevokeAllDevices)            // 下线其他设备
			devices.DELETE("/force", handlers.RevokeAllDevicesForce) // 强制下线所有设备
			devices.GET("/stats", handlers.GetDeviceStats)           // 获取设备统计
		}
	}

	// 管理员路由（需要认证 + 管理员权限）
	admin := r.Group("/api/admin")
	admin.Use(middleware.AdminAuth()) // AdminAuth已经包含了JWT验证
	{
		// 数据看板
		admin.GET("/dashboard", handlers.HandleAdminDashboard)

		// 用户管理
		admin.GET("/users", handlers.HandleAdminGetUsers)
		admin.PUT("/users/:id", handlers.HandleAdminUpdateUser)
		admin.DELETE("/users/:id", handlers.HandleAdminDeleteUser)
		admin.PUT("/users/:id/status", handlers.HandleAdminToggleUserStatus)
		admin.GET("/users/:id/subscriptions", handlers.HandleAdminGetUserSubscriptions)
		admin.PUT("/users/:id/subscriptions/:subscription_id/limit", handlers.HandleAdminUpdateUserSubscriptionLimit)
		admin.POST("/users/:id/gift", handlers.HandleAdminGiftSubscription)

		// 赠送记录管理
		admin.GET("/gift-records", handlers.HandleAdminGetGiftRecords)

		// 系统配置管理
		admin.GET("/system-configs", handlers.HandleAdminGetSystemConfigs)
		admin.PUT("/system-config", handlers.HandleAdminUpdateSystemConfig)

		// 订阅计划管理
		admin.GET("/subscription-plans", handlers.HandleAdminGetSubscriptionPlans)
		admin.POST("/subscription-plans", handlers.HandleAdminCreateSubscriptionPlan)
		admin.PUT("/subscription-plans/:id", handlers.HandleAdminUpdateSubscriptionPlan)
		admin.DELETE("/subscription-plans/:id", handlers.HandleAdminDeleteSubscriptionPlan)

		// 激活码管理
		admin.GET("/activation-codes", handlers.HandleAdminGetActivationCodes)
		admin.POST("/activation-codes", handlers.HandleAdminCreateActivationCodes)
		admin.DELETE("/activation-codes/:id", handlers.HandleAdminDeleteActivationCode)
		admin.GET("/activation-codes/:id/daily-limit", handlers.HandleGetActivationCodeDailyLimit)
		admin.PUT("/activation-codes/:id/daily-limit", handlers.HandleUpdateActivationCodeDailyLimit)

		// 公告管理
		admin.GET("/announcements", handlers.HandleAdminGetAnnouncements)
		admin.POST("/announcements", handlers.HandleAdminCreateAnnouncement)
		admin.PUT("/announcements/:id", handlers.HandleAdminUpdateAnnouncement)
		admin.DELETE("/announcements/:id", handlers.HandleAdminDeleteAnnouncement)
	}

	// 静态文件服务 - 提供前端构建的静态资源
	r.Static("/assets", "./ui/dist/assets")
	r.Static("/_next", "./ui/dist/_next") // Next.js 静态文件
	r.StaticFile("/favicon.ico", "./ui/dist/favicon.ico")
	r.StaticFile("/icon.png", "./ui/dist/icon.png")

	// SPA fallback - 所有未匹配的路由都返回前端应用
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果请求的是API路径，返回404
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, gin.H{
				"error": "API endpoint not found",
			})
			return
		}

		// 如果请求的是静态文件路径但文件不存在，返回404
		if len(path) >= 6 && path[:6] == "/_next" {
			c.JSON(404, gin.H{
				"error": "Static file not found",
			})
			return
		}

		// 否则返回前端应用
		c.File("./ui/dist/index.html")
	})
}
