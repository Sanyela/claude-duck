package handlers

import (
	"net/http"
	"strings"

	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
)

// AnnouncementsResponse 公告响应结构
type AnnouncementsResponse struct {
	Announcements []models.Announcement `json:"announcements"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleAnnouncements 获取公告处理器
func HandleAnnouncements(c *gin.Context) {
	// 从请求头获取访问令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authorization header required",
		})
		return
	}

	// 验证Bearer token格式
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid authorization header format",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// 验证访问令牌
	_, err := utils.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid or expired token",
		})
		return
	}

	// 获取语言设置
	language := c.GetHeader("Accept-Language")
	if language == "" {
		language = "en" // 默认英语
	}

	// 处理语言优先级（取第一个语言）
	if strings.Contains(language, ",") {
		language = strings.Split(language, ",")[0]
	}
	if strings.Contains(language, ";") {
		language = strings.Split(language, ";")[0]
	}
	language = strings.TrimSpace(language)

	// 查询活跃的公告
	var announcements []models.Announcement
	query := database.DB.Where("active = ?", true)

	// 按语言过滤，如果没有对应语言的公告，回退到英语
	var result []models.Announcement
	if err := query.Where("language = ?", language).Find(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to fetch announcements",
		})
		return
	}

	// 如果没有找到对应语言的公告，尝试英语
	if len(result) == 0 && language != "en" {
		if err := query.Where("language = ?", "en").Find(&result).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to fetch announcements",
			})
			return
		}
	}

	announcements = result

	// 返回公告列表
	c.JSON(http.StatusOK, AnnouncementsResponse{
		Announcements: announcements,
	})
} 