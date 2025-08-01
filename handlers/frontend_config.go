package handlers

import (
	"net/http"

	"claude/config"

	"github.com/gin-gonic/gin"
)

// HandleFrontendConfig 提供前端配置信息
func HandleFrontendConfig(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	c.JSON(http.StatusOK, gin.H{
		"appName":        config.AppConfig.AppName,
		"apiUrl":         "https://api.duckcode.top", // 直接指向API域名
		"installCommand": config.AppConfig.InstallCommand,
		"docsUrl":        config.AppConfig.DocsURL,
		"claudeUrl":      config.AppConfig.ClaudeURL,
	})
}