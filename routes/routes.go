package routes

import (
	"claude/handlers"

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

	// 认证相关路由（无需token验证）
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handlers.HandleRegister)
		auth.POST("/login", handlers.HandleLogin)
		auth.POST("/logout", handlers.HandleLogout)
		auth.GET("/user", handlers.HandleGetUserInfo) // 需要token验证
	}

	// SSO相关路由
	sso := r.Group("/api/sso")
	{
		sso.POST("/authorize", handlers.HandleAuthorize)
		sso.POST("/verify-token", handlers.HandleVerifyToken)
		sso.POST("/verify-code", handlers.HandleVerifyCode)
	}

	// API路由（需要认证）
	api := r.Group("/api")
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
	}

	// 静态文件服务 - 提供前端构建的静态资源
	r.Static("/assets", "./ui/dist/assets")
	r.StaticFile("/favicon.ico", "./ui/dist/favicon.ico")

	// SPA fallback - 所有未匹配的路由都返回前端应用
	r.NoRoute(func(c *gin.Context) {
		// 如果请求的是API路径但不是sso/authorize，返回404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" && c.Request.URL.Path != "/api/sso/authorize" {
			c.JSON(404, gin.H{
				"error": "API endpoint not found",
			})
			return
		}

		// 否则返回前端应用`
		c.File("./ui/dist/index.html")
	})
}
