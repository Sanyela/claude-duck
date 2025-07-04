package handlers

import (
	"net/http"
	"time"

	"claude/database"
	"claude/utils"

	"github.com/gin-gonic/gin"
)

// DeviceResponse 设备信息响应结构
type DeviceResponse struct {
	ID         string    `json:"id"`
	DeviceName string    `json:"device_name"`
	DeviceType string    `json:"device_type"`
	IP         string    `json:"ip"`
	Location   string    `json:"location"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsCurrent  bool      `json:"is_current"`
}

// GetDevices 获取用户所有设备
func GetDevices(c *gin.Context) {
	userID := c.GetUint("userID")
	currentDeviceID, _ := c.Get("deviceID")

	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	devices, err := deviceManager.GetUserDevices(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设备列表失败"})
		return
	}

	// 格式化返回数据
	var result []DeviceResponse
	for _, device := range devices {
		isCurrent := false
		if currentDeviceID != nil {
			isCurrent = device.ID == currentDeviceID.(string)
		}

		result = append(result, DeviceResponse{
			ID:         device.ID,
			DeviceName: device.DeviceName,
			DeviceType: device.DeviceType,
			IP:         device.IP,
			Location:   device.Location,
			Source:     device.Source,
			CreatedAt:  device.CreatedAt,
			LastActive: device.LastActive,
			ExpiresAt:  device.ExpiresAt,
			IsCurrent:  isCurrent,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"devices": result,
			"total":   len(result),
		},
	})
}

// RevokeDevice 下线指定设备
func RevokeDevice(c *gin.Context) {
	userID := c.GetUint("userID")
	deviceID := c.Param("deviceId")
	currentDeviceID, _ := c.Get("deviceID")

	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	err := deviceManager.RevokeDevice(userID, deviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 检查是否下线的是当前设备
	isCurrent := currentDeviceID != nil && deviceID == currentDeviceID.(string)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "设备已成功下线",
		"data": gin.H{
			"is_current": isCurrent,
		},
	})
}

// RevokeAllDevices 下线用户所有设备（除当前设备）
func RevokeAllDevices(c *gin.Context) {
	userID := c.GetUint("userID")
	currentDeviceID, _ := c.Get("deviceID")

	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)

	// 获取所有设备
	devices, err := deviceManager.GetUserDevices(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取设备列表失败",
		})
		return
	}

	// 逐个下线非当前设备
	var revokedCount int
	for _, device := range devices {
		if currentDeviceID != nil && device.ID == currentDeviceID.(string) {
			continue // 跳过当前设备
		}

		err := deviceManager.RevokeDevice(userID, device.ID)
		if err == nil {
			revokedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已成功下线其他设备",
		"data": gin.H{
			"revoked_count": revokedCount,
		},
	})
}

// RevokeAllDevicesForce 强制下线用户所有设备（包括当前设备）
func RevokeAllDevicesForce(c *gin.Context) {
	userID := c.GetUint("userID")

	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	err := deviceManager.RevokeAllUserDevices(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "下线所有设备失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "已强制下线所有设备",
	})
}

// GetDeviceStats 获取设备统计信息
func GetDeviceStats(c *gin.Context) {
	userID := c.GetUint("userID")

	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	devices, err := deviceManager.GetUserDevices(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设备信息失败"})
		return
	}

	// 统计信息
	stats := map[string]int{
		"total":   0,
		"web":     0,
		"sso":     0,
		"mobile":  0,
		"desktop": 0,
		"tablet":  0,
	}

	for _, device := range devices {
		stats["total"]++
		stats[device.Source]++
		stats[device.DeviceType]++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}