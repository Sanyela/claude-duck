package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"claude/database"
	"claude/models"
)

// GenerateDeviceCode 生成设备授权码
func GenerateDeviceCode() string {
	// 生成8位随机字符串，格式如：ABCD-1234
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	var result strings.Builder
	for i := 0; i < 8; i++ {
		if i == 4 {
			result.WriteString("-")
		}
		result.WriteByte(charset[rand.Intn(len(charset))])
	}
	
	return result.String()
}

// StoreDeviceCode 存储设备码到数据库
func StoreDeviceCode(userID uint) (string, error) {
	// 生成唯一的设备码
	var code string
	var exists bool
	
	// 确保生成的码是唯一的
	for {
		code = GenerateDeviceCode()
		
		var existingCode models.DeviceCode
		err := database.DB.Where("code = ?", code).First(&existingCode).Error
		if err != nil {
			// 如果没有找到记录，说明这个码是唯一的
			exists = false
			break
		} else {
			// 如果找到了记录，继续生成新的码
			exists = true
			continue
		}
	}
	
	if exists {
		return "", fmt.Errorf("failed to generate unique device code")
	}

	// 设置过期时间
	expiresAt := time.Now().Add(time.Duration(15) * time.Minute) // 15分钟过期

	// 创建设备码记录
	deviceCode := models.DeviceCode{
		Code:      code,
		UserID:    userID,
		Used:      false,
		ExpiresAt: expiresAt,
	}

	// 保存到数据库
	if err := database.DB.Create(&deviceCode).Error; err != nil {
		return "", fmt.Errorf("failed to store device code: %w", err)
	}

	return code, nil
}

// ValidateDeviceCode 验证设备码并返回用户信息
func ValidateDeviceCode(code string) (*models.User, error) {
	var deviceCode models.DeviceCode
	
	// 查找设备码并预加载用户信息
	err := database.DB.Preload("User").Where("code = ?", code).First(&deviceCode).Error
	if err != nil {
		return nil, fmt.Errorf("device code not found")
	}

	// 检查是否已经使用
	if deviceCode.Used {
		return nil, fmt.Errorf("device code already used")
	}

	// 检查是否过期
	if time.Now().After(deviceCode.ExpiresAt) {
		return nil, fmt.Errorf("device code expired")
	}

	// 标记为已使用
	deviceCode.Used = true
	if err := database.DB.Save(&deviceCode).Error; err != nil {
		return nil, fmt.Errorf("failed to mark device code as used: %w", err)
	}

	return &deviceCode.User, nil
}

// CleanExpiredDeviceCodes 清理过期的设备码
func CleanExpiredDeviceCodes() error {
	result := database.DB.Where("expires_at < ?", time.Now()).Delete(&models.DeviceCode{})
	if result.Error != nil {
		return fmt.Errorf("failed to clean expired device codes: %w", result.Error)
	}
	
	if result.RowsAffected > 0 {
		fmt.Printf("Cleaned %d expired device codes\n", result.RowsAffected)
	}
	
	return nil
} 