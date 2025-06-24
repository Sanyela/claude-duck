package handlers

import (
	"net/http"
	"strings"

	"claude/database"
	"claude/models"

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
	// 获取查询参数
	language := c.Query("language")
	if language == "" {
		// 从请求头获取语言设置
		language = c.GetHeader("Accept-Language")
		if language == "" {
			language = "zh" // 默认中文
		}

		// 处理语言优先级（取第一个语言）
		if strings.Contains(language, ",") {
			language = strings.Split(language, ",")[0]
		}
		if strings.Contains(language, ";") {
			language = strings.Split(language, ";")[0]
		}
		language = strings.TrimSpace(language)
	}

	active := c.Query("active")

	// 构建查询
	query := database.DB.Model(&models.Announcement{})

	// 按活跃状态过滤
	if active != "" {
		query = query.Where("active = ?", active == "true")
	} else {
		// 默认只返回活跃的公告
		query = query.Where("active = ?", true)
	}

	// 按语言过滤
	query = query.Where("language = ?", language)

	// 查询公告
	var announcements []models.Announcement
	if err := query.Order("created_at DESC").Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to fetch announcements",
		})
		return
	}

	// 如果没有找到对应语言的公告且语言不是中文，尝试中文
	if len(announcements) == 0 && language != "zh" {
		query = database.DB.Where("active = ?", true).Where("language = ?", "zh")
		if err := query.Order("created_at DESC").Find(&announcements).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Failed to fetch announcements",
			})
			return
		}
	}

	// 返回公告列表
	c.JSON(http.StatusOK, announcements)
}
