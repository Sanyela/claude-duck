package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"

	"github.com/gin-gonic/gin"
)

// ClaudeResponse Claude API 响应结构
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason   string      `json:"stop_reason"`
	StopSequence interface{} `json:"stop_sequence"`
	Usage        struct {
		InputTokens              int    `json:"input_tokens"`
		CacheCreationInputTokens int    `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int    `json:"cache_read_input_tokens"`
		OutputTokens             int    `json:"output_tokens"`
		ServiceTier              string `json:"service_tier"`
	} `json:"usage"`
}

// Claude 流式响应结构
type ClaudeStreamEvent struct {
	Type    string          `json:"type"`
	Message *ClaudeResponse `json:"message,omitempty"`
	Delta   *struct {
		StopReason   string      `json:"stop_reason,omitempty"`
		StopSequence interface{} `json:"stop_sequence,omitempty"`
	} `json:"delta,omitempty"`
	Usage *struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// HandleClaudeProxy 处理 Claude API 代理请求
func HandleClaudeProxy(c *gin.Context) {
	// 获取用户信息
	userID, err := getUserIDFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	// 获取用户详细信息
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user info"})
		return
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	// 解析请求以获取模型信息
	var requestData map[string]interface{}
	if err := json.Unmarshal(body, &requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 获取模型名称
	model, _ := requestData["model"].(string)
	if model == "" {
		model = "unknown"
	}

	// 检查是否是流式请求
	isStream := false
	if stream, ok := requestData["stream"].(bool); ok {
		isStream = stream
	}

	// 获取系统配置
	var configs []models.SystemConfig
	database.DB.Find(&configs)
	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.ConfigKey] = cfg.ConfigValue
	}

	// 获取New API配置
	apiEndpoint := configMap["new_api_endpoint"]
	if apiEndpoint == "" {
		apiEndpoint = config.AppConfig.NewAPIEndpoint
	}
	apiKey := configMap["new_api_key"]
	if apiKey == "" {
		apiKey = config.AppConfig.NewAPIKey
	}

	// 创建代理请求
	targetURL := apiEndpoint + strings.TrimPrefix(c.Request.URL.Path, "/api/claude")
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(body))
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

	// 设置 API Key
	req.Header.Set("x-api-key", apiKey)

	// 记录开始时间
	startTime := time.Now()

	// 发送请求
	client := &http.Client{
		Timeout: 5 * time.Minute, // 设置5分钟超时
	}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Failed to contact upstream API",
		})
		return
	}
	defer resp.Body.Close()

	// 如果是非流式响应，直接处理
	if !isStream {
		handleNonStreamResponse(c, resp, userID, user.Username, model, startTime, configMap)
	} else {
		// 流式响应处理
		handleStreamResponse(c, resp, userID, user.Username, model, startTime, configMap)
	}
}

// 处理非流式响应
func handleNonStreamResponse(c *gin.Context, resp *http.Response, userID uint, username string, model string, startTime time.Time, configMap map[string]string) {
	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response body",
		})
		return
	}

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 设置状态码
	c.Status(resp.StatusCode)

	// 写入响应
	c.Writer.Write(responseBody)

	// 如果请求成功，解析响应并记录
	if resp.StatusCode == http.StatusOK {
		var claudeResp ClaudeResponse
		if err := json.Unmarshal(responseBody, &claudeResp); err == nil {
			// 记录请求历史
			recordUsage(userID, username, model, claudeResp.ID,
				claudeResp.Usage.InputTokens,
				claudeResp.Usage.OutputTokens,
				c.ClientIP(), startTime, configMap)
		}
	}
}

// 处理流式响应
func handleStreamResponse(c *gin.Context, resp *http.Response, userID uint, username string, model string, startTime time.Time, configMap map[string]string) {
	// 设置SSE相关头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// 设置状态码
	c.Status(resp.StatusCode)

	// 刷新头部
	c.Writer.Flush()

	var messageID string
	var totalInputTokens, totalOutputTokens int

	// 创建读取器
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(c.Writer, "data: {\"error\": \"Stream reading error\"}\n\n")
			}
			break
		}

		// 写入原始数据到客户端
		c.Writer.Write(line)
		c.Writer.Flush()

		// 解析SSE数据
		if bytes.HasPrefix(line, []byte("data: ")) {
			data := bytes.TrimPrefix(line, []byte("data: "))
			data = bytes.TrimSpace(data)

			if len(data) > 0 && !bytes.Equal(data, []byte("[DONE]")) {
				var event ClaudeStreamEvent
				if err := json.Unmarshal(data, &event); err == nil {
					// 记录消息开始事件
					if event.Type == "message_start" && event.Message != nil {
						messageID = event.Message.ID
						totalInputTokens = event.Message.Usage.InputTokens
						totalOutputTokens = event.Message.Usage.OutputTokens
					}
					// 更新输出token数
					if event.Type == "message_delta" && event.Usage != nil {
						totalOutputTokens = event.Usage.OutputTokens
					}
				}
			}
		}
	}

	// 记录流式请求的使用情况
	if messageID != "" {
		recordUsage(userID, username, model, messageID,
			totalInputTokens, totalOutputTokens,
			c.ClientIP(), startTime, configMap)
	}
}

// 记录使用情况
func recordUsage(userID uint, username string, model string, messageID string, inputTokens int, outputTokens int, ip string, startTime time.Time, configMap map[string]string) {
	// 获取倍率配置
	promptMultiplier, _ := strconv.ParseFloat(configMap["prompt_multiplier"], 64)
	if promptMultiplier == 0 {
		promptMultiplier = config.AppConfig.DefaultPromptMultiplier
	}

	completionMultiplier, _ := strconv.ParseFloat(configMap["completion_multiplier"], 64)
	if completionMultiplier == 0 {
		completionMultiplier = config.AppConfig.DefaultCompletionMultiplier
	}

	tokensPerPoint, _ := strconv.Atoi(configMap["tokens_per_point"])
	if tokensPerPoint == 0 {
		tokensPerPoint = config.AppConfig.DefaultTokensPerPoint
	}

	roundUpEnabled := configMap["round_up_enabled"] == "true"

	// 计算积分
	promptPoints := float64(inputTokens) * promptMultiplier / float64(tokensPerPoint)
	completionPoints := float64(outputTokens) * completionMultiplier / float64(tokensPerPoint)
	totalPoints := promptPoints + completionPoints

	var pointsUsed int64
	if roundUpEnabled && totalPoints != float64(int64(totalPoints)) {
		pointsUsed = int64(totalPoints) + 1
	} else {
		pointsUsed = int64(totalPoints)
	}

	// 记录到数据库
	duration := int(time.Since(startTime).Milliseconds())

	// 创建API请求记录
	apiRequest := models.APIRequest{
		UserID:           userID,
		MessageID:        messageID,
		Model:            model,
		PromptTokens:     inputTokens,
		CompletionTokens: outputTokens,
		PointsUsed:       pointsUsed,
		Status:           "success",
		Duration:         duration,
		CreatedAt:        time.Now(),
	}
	database.DB.Create(&apiRequest)

	// 创建积分使用历史记录
	usageHistory := models.PointUsageHistory{
		UserID:               userID,
		RequestID:            messageID,
		IP:                   ip,
		UID:                  fmt.Sprintf("%d", userID),
		Username:             username,
		Model:                model,
		PromptTokens:         inputTokens,
		CompletionTokens:     outputTokens,
		PromptMultiplier:     promptMultiplier,
		CompletionMultiplier: completionMultiplier,
		PointsUsed:           pointsUsed,
		IsRoundUp:            roundUpEnabled && totalPoints != float64(int64(totalPoints)),
		CreatedAt:            time.Now(),
	}
	database.DB.Create(&usageHistory)
}
