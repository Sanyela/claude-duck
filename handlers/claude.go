package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// CLAUDE_API_URL Claude API 代理目标地址
	CLAUDE_API_URL = "http://152.53.82.23:2999/v1/messages"
	// CLAUDE_API_KEY API Key
	CLAUDE_API_KEY = "sk-BxYNfpirLM4E4TI7k1Cu1WoqOVTpMzyl6B2GNeYngdX9J5VD"
)

// HandleClaudeProxy 处理 Claude API 代理请求
func HandleClaudeProxy(c *gin.Context) {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// 创建代理请求
	req, err := http.NewRequest(c.Request.Method, CLAUDE_API_URL, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create proxy request",
		})
		return
	}

	// 复制原始请求头
	for key, values := range c.Request.Header {
		// 跳过 Host 和 Authorization 头
		if strings.ToLower(key) == "host" || strings.ToLower(key) == "authorization" {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// 设置正确的 Content-Type
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置 Claude API Key
	req.Header.Set("x-api-key", CLAUDE_API_KEY)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Failed to contact upstream API",
		})
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 设置状态码
	c.Status(resp.StatusCode)

	// 直接复制响应体
	io.Copy(c.Writer, resp.Body)
}
