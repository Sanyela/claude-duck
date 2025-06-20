package middleware

import (
	"net/http"
	"strings"

	"claude/utils"
	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌格式错误"})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证token
		claims, err := utils.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌无效或已过期"})
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// OptionalJWTAuth 可选的JWT认证中间件（不强制要求认证）
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token := parts[1]
				claims, err := utils.ValidateAccessToken(token)
				if err == nil {
					c.Set("userID", claims.UserID)
				}
			}
		}
		c.Next()
	}
}
