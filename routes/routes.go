package routes

import (
	"claude/handlers"
	"claude/middleware"
	"net/http"

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

	// OAuth授权页面重定向（兼容旧的URL）
	r.GET("/api/sso/authorize", func(c *gin.Context) {
		// 将请求重定向到正确的前端OAuth页面
		redirectURL := "/oauth/authorize?" + c.Request.URL.RawQuery
		c.Redirect(http.StatusMovedPermanently, redirectURL)
	})

	// 认证相关路由（无需token验证）
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.HandleRegister)
		auth.POST("/login", handlers.HandleLogin)
		auth.POST("/logout", handlers.HandleLogout)
		auth.GET("/user", middleware.JWTAuth(), handlers.HandleGetUserInfo) // 需要token验证
	}

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
		api.GET("/announcements", handlers.HandleAnnouncements)

		// 订阅相关路由
		api.GET("/subscription/active", handlers.HandleGetActiveSubscription)
		api.GET("/subscription/history", handlers.HandleGetSubscriptionHistory)
		api.POST("/subscription/redeem", handlers.HandleRedeemCoupon)

		// 积分相关路由
		api.GET("/credits/balance", handlers.HandleGetCreditBalance)
		api.GET("/credits/model-costs", handlers.HandleGetModelCosts)
		api.GET("/credits/history", handlers.HandleGetCreditUsageHistory)
		api.POST("/credits/request-reset", handlers.HandleRequestCreditReset)

		// Claude API 代理路由
		api.POST("/claude", handlers.HandleClaudeProxy)
		api.Any("/claude/*path", handlers.HandleClaudeProxy) // 支持所有方法和子路径
	}

	// 管理员路由（需要认证 + 管理员权限）
	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTAuth())
	admin.Use(middleware.AdminAuth())
	{
		// 用户管理
		admin.GET("/users", handlers.HandleAdminGetUsers)
		admin.PUT("/users/:id", handlers.HandleAdminUpdateUser)
		admin.DELETE("/users/:id", handlers.HandleAdminDeleteUser)

		// 用户分组管理
		admin.GET("/user-groups", handlers.HandleAdminGetUserGroups)
		admin.POST("/user-groups", handlers.HandleAdminCreateUserGroup)
		admin.PUT("/user-groups/:id", handlers.HandleAdminUpdateUserGroup)
		admin.DELETE("/user-groups/:id", handlers.HandleAdminDeleteUserGroup)

		// API渠道管理
		admin.GET("/api-channels", handlers.HandleAdminGetAPIChannels)
		admin.POST("/api-channels", handlers.HandleAdminCreateAPIChannel)
		admin.PUT("/api-channels/:id", handlers.HandleAdminUpdateAPIChannel)
		admin.DELETE("/api-channels/:id", handlers.HandleAdminDeleteAPIChannel)

		// 模型成本配置管理
		admin.GET("/model-costs", handlers.HandleAdminGetModelCosts)
		admin.POST("/model-costs", handlers.HandleAdminCreateModelCost)
		admin.PUT("/model-costs/:id", handlers.HandleAdminUpdateModelCost)
		admin.DELETE("/model-costs/:id", handlers.HandleAdminDeleteModelCost)

		// 激活码管理
		admin.GET("/activation-codes", handlers.HandleAdminGetActivationCodes)
		admin.POST("/activation-codes", handlers.HandleAdminCreateActivationCodes)
		admin.DELETE("/activation-codes/:id", handlers.HandleAdminDeleteActivationCode)

		// 计费规则管理
		admin.GET("/billing-rules", handlers.HandleAdminGetBillingRules)
		admin.POST("/billing-rules", handlers.HandleAdminCreateBillingRule)
		admin.PUT("/billing-rules/:id", handlers.HandleAdminUpdateBillingRule)
		admin.DELETE("/billing-rules/:id", handlers.HandleAdminDeleteBillingRule)
	}

	// 静态文件服务 - 提供前端构建的静态资源
	r.Static("/assets", "./ui/dist/assets")
	r.StaticFile("/favicon.ico", "./ui/dist/favicon.ico")

	// SPA fallback - 所有未匹配的路由都返回前端应用
	r.NoRoute(func(c *gin.Context) {
		// 如果请求的是API路径，返回404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{
				"error": "API endpoint not found",
			})
			return
		}

		// 否则返回前端应用
		c.File("./ui/dist/index.html")
	})
}
