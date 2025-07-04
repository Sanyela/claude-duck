package middleware

import (
	"net/http"
	"strings"

	"claude/database"
	"claude/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件 - 基于Redis设备验证
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

		// 先验证JWT格式
		claims, err := utils.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌格式无效"})
			c.Abort()
			return
		}

		// 通过Redis验证设备
		deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
		device, err := deviceManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证令牌无效或设备已下线",
				"code":  "DEVICE_OFFLINE",
			})
			c.Abort()
			return
		}

		// 验证用户ID匹配
		if device.UserID != claims.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌用户不匹配"})
			c.Abort()
			return
		}

		// 将用户和设备信息存储到上下文中
		c.Set("userID", claims.UserID)
		c.Set("deviceID", device.ID)
		c.Set("deviceInfo", device)
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
					// 验证Redis中的设备
					deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
					device, err := deviceManager.ValidateToken(token)
					if err == nil && device.UserID == claims.UserID {
						c.Set("userID", claims.UserID)
						c.Set("deviceID", device.ID)
						c.Set("deviceInfo", device)
					}
				}
			}
		}
		c.Next()
	}
}

// AdminAuth 管理员认证中间件 - 先验证Redis再验证管理员权限
func AdminAuth() gin.HandlerFunc {
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

		// 先验证JWT格式
		claims, err := utils.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌格式无效"})
			c.Abort()
			return
		}

		// 通过Redis验证设备
		deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
		device, err := deviceManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证令牌无效或设备已下线",
				"code":  "DEVICE_OFFLINE",
			})
			c.Abort()
			return
		}

		// 验证用户ID匹配
		if device.UserID != claims.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌用户不匹配"})
			c.Abort()
			return
		}

		// 验证管理员权限
		if !isUserAdmin(claims.UserID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		// 将用户和设备信息存储到上下文中
		c.Set("userID", claims.UserID)
		c.Set("deviceID", device.ID)
		c.Set("deviceInfo", device)
		c.Next()
	}
}
