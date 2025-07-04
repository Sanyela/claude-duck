package middleware

import (
	"claude/database"
	"claude/models"
)

// isUserAdmin 检查用户是否为管理员
func isUserAdmin(userID uint) bool {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}
	return user.IsAdmin
}