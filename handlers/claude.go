package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	// 检查用户是否被禁用
	if user.IsDisabled {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "您的账户已被管理员禁用，无法使用API服务",
			"code":  "USER_DISABLED",
		})
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

	// 获取模型名称（原始请求模型）
	originalModel, _ := requestData["model"].(string)
	if originalModel == "" {
		originalModel = "unknown"
	}
	model := originalModel // 用于数据库记录的模型名

	// 获取系统配置
	var configs []models.SystemConfig
	database.DB.Find(&configs)
	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.ConfigKey] = cfg.ConfigValue
	}

	// 处理模型重定向
	modelRedirectConfig := configMap["model_redirect_map"]
	var actualModel string = originalModel // 实际发送给API的模型名
	if modelRedirectConfig != "" {
		var redirectMap map[string]string
		if err := json.Unmarshal([]byte(modelRedirectConfig), &redirectMap); err == nil {
			if redirectedModel, exists := redirectMap[originalModel]; exists {
				actualModel = redirectedModel
				// 修改请求体中的模型参数
				requestData["model"] = actualModel
				// 重新序列化请求体
				if newBody, err := json.Marshal(requestData); err == nil {
					body = newBody
				}
			}
		}
	}

	// 获取免费模型列表配置
	freeModelsConfig := configMap["free_models_list"]
	if freeModelsConfig == "" {
		freeModelsConfig = `["claude-3-5-haiku-20241022"]` // 默认免费模型
	}

	// 解析免费模型列表
	var freeModels []string
	isFreeModel := false
	if err := json.Unmarshal([]byte(freeModelsConfig), &freeModels); err == nil {
		if slices.Contains(freeModels, model) {
			isFreeModel = true
		}
	}

	// 如果不是免费模型，则需要检查积分
	if !isFreeModel {
		// 检查用户钱包是否有效和可用积分
		if !utils.IsWalletActive(userID) {
			available, _, _, err := utils.GetWalletBalance(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "检查积分余额失败",
					"code":  "CREDITS_CHECK_ERROR",
				})
				return
			}

			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":            "积分余额不足或已过期，请先充值",
				"code":             "INSUFFICIENT_CREDITS",
				"available_points": available,
			})
			return
		}
	}

	// 检查是否是流式请求
	isStream := false
	if stream, ok := requestData["stream"].(bool); ok {
		isStream = stream
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
		handleNonStreamResponse(c, resp, userID, user.Username, model, startTime, configMap, isFreeModel, requestData)
	} else {
		// 流式响应处理
		handleStreamResponse(c, resp, userID, user.Username, model, startTime, configMap, isFreeModel, requestData)
	}
}

// 处理非流式响应
func handleNonStreamResponse(c *gin.Context, resp *http.Response, userID uint, username, model string, startTime time.Time, configMap map[string]string, isFreeModel bool, requestData map[string]interface{}) {
	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response body",
		})
		return
	}

	// 特殊处理429状态码
	if resp.StatusCode == http.StatusTooManyRequests {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "我们的API服务正在历经高负载请求,请稍等一分钟后重试(此条消息可忽略)",
		})
		return
	}

	// 特殊处理400状态码 - 检查是否是没有可用token的错误
	if resp.StatusCode == http.StatusBadRequest && strings.Contains(string(responseBody), "没有可用token") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "号池暂无可用账号，请等待管理员添加账号...",
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
			// 记录成功的请求并扣费
			recordUsage(userID, username, model, claudeResp.ID,
				claudeResp.Usage.InputTokens,
				claudeResp.Usage.OutputTokens,
				claudeResp.Usage.CacheCreationInputTokens,
				claudeResp.Usage.CacheReadInputTokens,
				claudeResp.Usage.ServiceTier,
				"api", // 非流式请求
				c.ClientIP(), startTime, configMap, isFreeModel)

			// 记录完整的对话日志
			recordConversationLog(userID, username, c.ClientIP(), requestData, &claudeResp, nil, "api", isFreeModel, startTime)
		}
	} else {
		// 记录失败的请求但不扣费
		apiTransaction := models.APITransaction{
			UserID:      userID,
			RequestID:   fmt.Sprintf("req_%d_%d", userID, time.Now().UnixNano()),
			Model:       model,
			RequestType: "api",
			IP:          c.ClientIP(),
			UID:         fmt.Sprintf("%d", userID),
			Username:    username,
			Status:      "failed",
			Error:       fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody)),
			Duration:    int(time.Since(startTime).Milliseconds()),
			ServiceTier: "standard",
			CreatedAt:   time.Now(),
		}
		database.DB.Create(&apiTransaction)

		// 记录失败的对话日志
		recordConversationLog(userID, username, c.ClientIP(), requestData, nil, nil, "api", isFreeModel, startTime)
	}
}

// 处理流式响应
func handleStreamResponse(c *gin.Context, resp *http.Response, userID uint, username, model string, startTime time.Time, configMap map[string]string, isFreeModel bool, requestData map[string]interface{}) {
	// 特殊处理429状态码
	if resp.StatusCode == http.StatusTooManyRequests {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Status(http.StatusTooManyRequests)
		c.Writer.Write([]byte("data: {\"error\": \"我们的API服务正在历经高负载请求,请稍等一分钟后重试(此条消息可忽略)\"}\n\n"))
		c.Writer.Flush()
		return
	}

	// 特殊处理400状态码 - 检查是否是没有可用token的错误
	if resp.StatusCode == http.StatusBadRequest {
		// 读取响应体以检查错误内容
		tempBody, err := io.ReadAll(resp.Body)
		if err == nil && strings.Contains(string(tempBody), "没有可用token") {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Status(http.StatusBadRequest)
			c.Writer.Write([]byte("data: {\"error\": \"号池暂无可用账号，请等待管理员添加账号...\"}\n\n"))
			c.Writer.Flush()
			return
		}
		// 如果不是没有可用token的错误，需要重新创建body给后续处理
		resp.Body = io.NopCloser(bytes.NewReader(tempBody))
	}

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
	var streamError error
	var finalClaudeResp *ClaudeResponse

	// 创建读取器
	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				streamError = err
				fmt.Fprintf(c.Writer, "data: {\"error\": \"Stream reading error\"}\n\n")
			}
			break
		}

		// 写入原始数据到客户端
		c.Writer.Write(line)
		c.Writer.Flush()

		// 解析SSE数据
		if after, ok := bytes.CutPrefix(line, []byte("data: ")); ok {
			data := after
			data = bytes.TrimSpace(data)

			if len(data) > 0 && !bytes.Equal(data, []byte("[DONE]")) {
				var event ClaudeStreamEvent
				if err := json.Unmarshal(data, &event); err == nil {
					// 记录消息开始事件
					if event.Type == "message_start" && event.Message != nil {
						messageID = event.Message.ID
						totalInputTokens = event.Message.Usage.InputTokens
						totalOutputTokens = event.Message.Usage.OutputTokens
						// 保存完整的响应对象
						finalClaudeResp = event.Message
					}
					// 更新输出token数
					if event.Type == "message_delta" && event.Usage != nil {
						totalOutputTokens = event.Usage.OutputTokens
						// 更新响应对象的token数
						if finalClaudeResp != nil {
							finalClaudeResp.Usage.OutputTokens = totalOutputTokens
						}
					}
				}
			}
		}
	}

	// 记录流式请求的使用情况
	if messageID != "" && resp.StatusCode == http.StatusOK && streamError == nil {
		// 成功的流式请求
		recordUsage(userID, username, model, messageID,
			totalInputTokens, totalOutputTokens,
			0, 0, // 流式响应暂时没有缓存信息
			"standard", // 默认服务等级
			"stream",   // 流式请求
			c.ClientIP(), startTime, configMap, isFreeModel)

		// 记录成功的流式对话日志
		recordConversationLog(userID, username, c.ClientIP(), requestData, finalClaudeResp, nil, "stream", isFreeModel, startTime)
	} else if messageID != "" {
		// 失败的流式请求，记录但不扣费
		apiTransaction := models.APITransaction{
			UserID:       userID,
			MessageID:    messageID,
			RequestID:    messageID,
			Model:        model,
			RequestType:  "stream",
			InputTokens:  totalInputTokens,
			OutputTokens: totalOutputTokens,
			IP:           c.ClientIP(),
			UID:          fmt.Sprintf("%d", userID),
			Username:     username,
			Status:       "failed",
			Error:        fmt.Sprintf("HTTP %d or stream error", resp.StatusCode),
			Duration:     int(time.Since(startTime).Milliseconds()),
			ServiceTier:  "standard",
			CreatedAt:    time.Now(),
		}
		if streamError != nil {
			apiTransaction.Error = streamError.Error()
		}
		database.DB.Create(&apiTransaction)

		// 记录失败的流式对话日志
		recordConversationLog(userID, username, c.ClientIP(), requestData, finalClaudeResp, nil, "stream", isFreeModel, startTime)
	}
}

// 记录使用情况
func recordUsage(userID uint, username string, model string, messageID string, inputTokens int, outputTokens int, cacheCreationTokens int, cacheReadTokens int, serviceTier string, requestType string, ip string, startTime time.Time, configMap map[string]string, isFreeModel bool) {
	// 如果是免费模型，只增加使用次数，不扣积分，不记录API事务
	if isFreeModel {
		// 开始数据库事务
		tx := database.DB.Begin()

		// 更新用户免费模型使用次数
		err := tx.Model(&models.User{}).Where("id = ?", userID).
			UpdateColumn("free_model_usage_count", gorm.Expr("free_model_usage_count + ?", 1)).Error
		if err != nil {
			tx.Rollback()
			// 即使更新失败也不影响用户体验，只记录日志
			return
		}

		tx.Commit()
		return
	}

	// 获取倍率配置
	inputMultiplier, _ := strconv.ParseFloat(configMap["prompt_multiplier"], 64)
	if inputMultiplier == 0 {
		inputMultiplier = config.AppConfig.DefaultPromptMultiplier
	}

	outputMultiplier, _ := strconv.ParseFloat(configMap["completion_multiplier"], 64)
	if outputMultiplier == 0 {
		outputMultiplier = config.AppConfig.DefaultCompletionMultiplier
	}

	// 获取缓存倍率配置
	cacheMultiplier, _ := strconv.ParseFloat(configMap["cache_multiplier"], 64)
	if cacheMultiplier == 0 {
		cacheMultiplier = inputMultiplier // 如果没有配置，默认使用输入倍率
	}

	// 计算总缓存token（创建 + 读取）
	totalCacheTokens := cacheCreationTokens + cacheReadTokens

	// 按照新的计费公式计算总tokens
	// 缓存token * 缓存倍率 + 输入token * 输入倍率 + 输出token * 输出倍率
	cacheTokensWeighted := float64(totalCacheTokens) * cacheMultiplier
	inputTokensWeighted := float64(inputTokens) * inputMultiplier
	outputTokensWeighted := float64(outputTokens) * outputMultiplier

	totalWeightedTokens := cacheTokensWeighted + inputTokensWeighted + outputTokensWeighted

	// 应用模型倍率
	modelMultiplierConfig := configMap["model_multiplier_map"]
	modelMultiplier := 1.0 // 默认倍率为1
	if modelMultiplierConfig != "" {
		var multiplierMap map[string]float64
		if err := json.Unmarshal([]byte(modelMultiplierConfig), &multiplierMap); err == nil {
			if multiplier, exists := multiplierMap[model]; exists && multiplier > 0 {
				modelMultiplier = multiplier
			}
		}
	}
	
	// 应用模型倍率到总加权token
	finalWeightedTokens := totalWeightedTokens * modelMultiplier

	// 使用新的累计token计费逻辑
	err := utils.AccumulateTokensAndDeduct(userID, int64(finalWeightedTokens))

	// 开始数据库事务
	tx := database.DB.Begin()
	if err != nil {
		tx.Rollback()
		// 如果扣费失败，仍然记录API调用，但标记为失败
		apiTransaction := models.APITransaction{
			UserID:                   userID,
			MessageID:                messageID,
			RequestID:                messageID,
			Model:                    model,
			RequestType:              requestType,
			InputTokens:              inputTokens,
			OutputTokens:             outputTokens,
			CacheCreationInputTokens: cacheCreationTokens,
			CacheReadInputTokens:     cacheReadTokens,
			InputMultiplier:          inputMultiplier,
			OutputMultiplier:         outputMultiplier,
			CacheMultiplier:          cacheMultiplier,
			ModelMultiplier:          modelMultiplier,
			PointsUsed:               0, // 扣费失败时记录为0
			IP:                       ip,
			UID:                      fmt.Sprintf("%d", userID),
			Username:                 username,
			Status:                   "billing_failed",
			Error:                    err.Error(),
			Duration:                 int(time.Since(startTime).Milliseconds()),
			ServiceTier:              serviceTier,
			CreatedAt:                time.Now(),
		}
		database.DB.Create(&apiTransaction)
		return
	}

	// 创建成功的API事务记录
	apiTransaction := models.APITransaction{
		UserID:                   userID,
		MessageID:                messageID,
		RequestID:                messageID, // 使用messageID作为requestID
		Model:                    model,
		RequestType:              requestType, // "api" 或 "stream"
		InputTokens:              inputTokens,
		OutputTokens:             outputTokens,
		CacheCreationInputTokens: cacheCreationTokens,
		CacheReadInputTokens:     cacheReadTokens,
		InputMultiplier:          inputMultiplier,
		OutputMultiplier:         outputMultiplier,
		CacheMultiplier:          cacheMultiplier,
		ModelMultiplier:          modelMultiplier,
		PointsUsed:               0, // 累计token计费模式下，这里记录为0，实际扣费由AccumulateTokensAndDeduct处理
		IP:                       ip,
		UID:                      fmt.Sprintf("%d", userID),
		Username:                 username,
		Status:                   "success",
		Duration:                 int(time.Since(startTime).Milliseconds()),
		ServiceTier:              serviceTier,
		CreatedAt:                time.Now(),
	}

	if err := tx.Create(&apiTransaction).Error; err != nil {
		tx.Rollback()
		return
	}

	// 提交事务
	tx.Commit()
}

// recordConversationLog 记录完整的对话日志
func recordConversationLog(userID uint, username string, ip string, requestData map[string]interface{}, claudeResp *ClaudeResponse, apiTransactionID *uint, requestType string, isFreeModel bool, startTime time.Time) {
	// 解析请求数据
	model, _ := requestData["model"].(string)
	messages, _ := json.Marshal(requestData["messages"])
	systemPrompt, _ := requestData["system"].(string)
	tools, _ := json.Marshal(requestData["tools"])
	stopSequences, _ := json.Marshal(requestData["stop_sequences"])
	userInputJSON, _ := json.Marshal(requestData)

	// 解析参数
	var temperature *float64
	var maxTokens *int
	var topP *float64
	var topK *int

	if temp, ok := requestData["temperature"].(float64); ok {
		temperature = &temp
	}
	if tokens, ok := requestData["max_tokens"].(float64); ok {
		maxTokensInt := int(tokens)
		maxTokens = &maxTokensInt
	}
	if p, ok := requestData["top_p"].(float64); ok {
		topP = &p
	}
	if k, ok := requestData["top_k"].(float64); ok {
		topKInt := int(k)
		topK = &topKInt
	}

	// 准备响应数据
	var aiResponseJSON []byte
	var responseText string
	var messageID string
	var stopReason string
	var stopSequence string
	var inputTokens, outputTokens, cacheCreationTokens, cacheReadTokens int
	var serviceTier string

	if claudeResp != nil {
		aiResponseJSON, _ = json.Marshal(claudeResp)
		messageID = claudeResp.ID
		stopReason = claudeResp.StopReason
		if claudeResp.StopSequence != nil {
			stopSequenceBytes, _ := json.Marshal(claudeResp.StopSequence)
			stopSequence = string(stopSequenceBytes)
		}
		inputTokens = claudeResp.Usage.InputTokens
		outputTokens = claudeResp.Usage.OutputTokens
		cacheCreationTokens = claudeResp.Usage.CacheCreationInputTokens
		cacheReadTokens = claudeResp.Usage.CacheReadInputTokens
		serviceTier = claudeResp.Usage.ServiceTier

		// 提取文本响应
		for _, content := range claudeResp.Content {
			if content.Type == "text" {
				responseText = content.Text
				break
			}
		}
	}

	// 创建对话日志记录
	conversationLog := models.ConversationLog{
		UserID:           userID,
		APITransactionID: apiTransactionID,
		MessageID:        messageID,
		RequestID:        messageID,
		Model:            model,
		RequestType:      requestType,
		IP:               ip,
		Username:         username,
		UserInput:        string(userInputJSON),
		SystemPrompt:     systemPrompt,
		Messages:         string(messages),
		Tools:            string(tools),
		Temperature:      temperature,
		MaxTokens:        maxTokens,
		TopP:             topP,
		TopK:             topK,
		StopSequences:    string(stopSequences),
		AIResponse:       string(aiResponseJSON),
		ResponseText:     responseText,
		StopReason:       stopReason,
		StopSequence:     stopSequence,
		InputTokens:      inputTokens,
		OutputTokens:     outputTokens,
		CacheCreationInputTokens: cacheCreationTokens,
		CacheReadInputTokens:     cacheReadTokens,
		TotalTokens:              inputTokens + outputTokens,
		Duration:                 int(time.Since(startTime).Milliseconds()),
		ServiceTier:              serviceTier,
		Status:                   "success",
		IsFreeModel:              isFreeModel,
		CreatedAt:                time.Now(),
	}

	// 如果有API事务ID，获取计费信息
	if apiTransactionID != nil {
		var apiTx models.APITransaction
		if err := database.DB.Where("id = ?", *apiTransactionID).First(&apiTx).Error; err == nil {
			conversationLog.InputMultiplier = apiTx.InputMultiplier
			conversationLog.OutputMultiplier = apiTx.OutputMultiplier
			conversationLog.CacheMultiplier = apiTx.CacheMultiplier
			conversationLog.PointsUsed = apiTx.PointsUsed
		}
	}

	// 保存到数据库
	database.DB.Create(&conversationLog)
}
