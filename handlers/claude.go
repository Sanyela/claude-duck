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

	// 获取系统配置
	var configs []models.SystemConfig
	database.DB.Find(&configs)
	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.ConfigKey] = cfg.ConfigValue
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
		for _, freeModelName := range freeModels {
			if model == freeModelName {
				isFreeModel = true
				break
			}
		}
	}

	// 如果不是免费模型，则需要检查积分
	if !isFreeModel {
		// 检查用户积分余额
		var pointBalance models.PointBalance
		err = database.DB.Where("user_id = ?", userID).First(&pointBalance).Error
		if err != nil {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error": "无积分余额信息，请先充值",
				"code":  "INSUFFICIENT_CREDITS",
			})
			return
		}

		// 检查是否有可用积分（只计算未过期的积分）
		var availablePoints int64
		err = database.DB.Model(&models.PointPool{}).
			Where("user_id = ? AND points_remaining > 0 AND expires_at > ?", userID, time.Now()).
			Select("COALESCE(SUM(points_remaining), 0)").
			Scan(&availablePoints).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "检查积分余额失败",
				"code":  "CREDITS_CHECK_ERROR",
			})
			return
		}

		if availablePoints <= 0 {
			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":            "积分余额不足或已过期，请先充值",
				"code":             "INSUFFICIENT_CREDITS",
				"available_points": availablePoints,
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
		handleNonStreamResponse(c, resp, userID, user.Username, model, startTime, configMap, isFreeModel)
	} else {
		// 流式响应处理
		handleStreamResponse(c, resp, userID, user.Username, model, startTime, configMap, isFreeModel)
	}
}

// 处理非流式响应
func handleNonStreamResponse(c *gin.Context, resp *http.Response, userID uint, username string, model string, startTime time.Time, configMap map[string]string, isFreeModel bool) {
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
	}
}

// 处理流式响应
func handleStreamResponse(c *gin.Context, resp *http.Response, userID uint, username string, model string, startTime time.Time, configMap map[string]string, isFreeModel bool) {
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
	if messageID != "" && resp.StatusCode == http.StatusOK && streamError == nil {
		// 成功的流式请求
		recordUsage(userID, username, model, messageID,
			totalInputTokens, totalOutputTokens,
			0, 0, // 流式响应暂时没有缓存信息
			"standard", // 默认服务等级
			"stream",   // 流式请求
			c.ClientIP(), startTime, configMap, isFreeModel)
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

	// 使用阶梯计费表计算积分
	tokenPricingTable := configMap["token_pricing_table"]
	if tokenPricingTable == "" {
		tokenPricingTable = `{"0":2,"7680":3,"15360":4,"23040":5,"30720":6,"38400":7,"46080":8,"53760":9,"61440":10,"69120":11,"76800":12,"84480":13,"92160":14,"99840":15,"107520":16,"115200":17,"122880":18,"130560":19,"138240":20,"145920":21,"153600":22,"161280":23,"168960":24,"176640":25,"184320":25,"192000":25,"200000":25}`
	}

	pointsUsed := utils.CalculatePointsByTokenTable(totalWeightedTokens, tokenPricingTable)

	// 确保至少收取0积分
	if pointsUsed < 0 {
		pointsUsed = 0
	}

	// 开始数据库事务
	tx := database.DB.Begin()

	// 扣除用户积分
	err := deductUserPoints(tx, userID, pointsUsed)
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
			PointsUsed:               pointsUsed,
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
		PointsUsed:               pointsUsed,
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

// deductUserPoints 扣除用户积分
func deductUserPoints(tx *gorm.DB, userID uint, pointsToDeduct int64) error {
	if pointsToDeduct <= 0 {
		return nil // 不需要扣费
	}

	// 获取用户积分余额
	var pointBalance models.PointBalance
	err := tx.Where("user_id = ?", userID).First(&pointBalance).Error
	if err != nil {
		return fmt.Errorf("获取用户积分余额失败: %v", err)
	}

	// 检查余额是否充足
	if pointBalance.AvailablePoints < pointsToDeduct {
		return fmt.Errorf("积分余额不足，需要 %d 积分，可用 %d 积分", pointsToDeduct, pointBalance.AvailablePoints)
	}

	// 按照FIFO原则从积分池中扣除积分
	var pointPools []models.PointPool
	err = tx.Where("user_id = ? AND points_remaining > 0 AND expires_at > ?", userID, time.Now()).
		Order("created_at ASC").
		Find(&pointPools).Error
	if err != nil {
		return fmt.Errorf("获取积分池失败: %v", err)
	}

	remainingToDeduct := pointsToDeduct
	for _, pool := range pointPools {
		if remainingToDeduct <= 0 {
			break
		}

		// 计算从当前池子扣除的积分
		deductFromPool := remainingToDeduct
		if deductFromPool > pool.PointsRemaining {
			deductFromPool = pool.PointsRemaining
		}

		// 更新积分池
		pool.PointsRemaining -= deductFromPool
		if err := tx.Save(&pool).Error; err != nil {
			return fmt.Errorf("更新积分池失败: %v", err)
		}

		remainingToDeduct -= deductFromPool
	}

	// 更新用户积分余额汇总
	pointBalance.UsedPoints += pointsToDeduct
	pointBalance.AvailablePoints -= pointsToDeduct
	pointBalance.UpdatedAt = time.Now()

	if err := tx.Save(&pointBalance).Error; err != nil {
		return fmt.Errorf("更新积分余额失败: %v", err)
	}

	return nil
}
