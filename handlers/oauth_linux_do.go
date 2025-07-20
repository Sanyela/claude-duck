package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"
	"claude/utils"

	"github.com/gin-gonic/gin"
)

// Linux Do用户信息结构
type LinuxDoUserInfo struct {
	ID             uint   `json:"id"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	AvatarTemplate string `json:"avatar_template"`
	Active         bool   `json:"active"`
	TrustLevel     int    `json:"trust_level"`
	Silenced       bool   `json:"silenced"`
	ExternalIDs    any    `json:"external_ids"` // 可能是object或null
	APIKey         string `json:"api_key"`
}

// OAuth2 Token响应结构
type LinuxDoTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// OAuth2错误响应结构
type LinuxDoErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// 生成OAuth2授权URL
func HandleLinuxDoAuthorize(c *gin.Context) {
	// 检查是否配置了Linux Do OAuth
	if !isLinuxDoConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Linux Do登录服务未配置",
		})
		return
	}

	// 生成state参数防止CSRF攻击
	state := generateRandomState()

	// 获取动态回调地址
	redirectURI := getLinuxDoRedirectURI(c)

	// 构建授权URL
	authURL := fmt.Sprintf("%s/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=user&state=%s",
		config.AppConfig.LinuxDoBaseURL,
		url.QueryEscape(config.AppConfig.LinuxDoClientID),
		url.QueryEscape(redirectURI),
		state,
	)

	// 将state存储到session或Redis中（这里简化处理，实际项目中应该存储）
	// TODO: 在生产环境中应该将state存储到Redis中，设置过期时间

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"auth_url": authURL,
		"state":    state,
	})
}

// 处理OAuth2回调
func HandleLinuxDoCallback(c *gin.Context) {
	// 检查是否配置了Linux Do OAuth
	if !isLinuxDoConfigured() {
		c.JSON(http.StatusServiceUnavailable, AuthResponse{
			Success: false,
			Message: "Linux Do登录服务未配置",
		})
		return
	}

	// 获取授权码和state
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// 检查是否有错误
	if errorParam != "" {
		errorDescription := c.Query("error_description")
		log.Printf("Linux Do OAuth错误: %s - %s", errorParam, errorDescription)
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "授权失败：" + errorDescription,
		})
		return
	}

	// 检查必要参数
	if code == "" {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "缺少授权码",
		})
		return
	}

	// TODO: 在生产环境中应该验证state参数
	if state == "" {
		log.Println("警告: 缺少state参数，可能存在CSRF攻击风险")
	}

	// 使用授权码获取访问令牌
	tokenResp, err := exchangeCodeForToken(code, c)
	if err != nil {
		log.Printf("获取访问令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "获取访问令牌失败",
		})
		return
	}

	// 使用访问令牌获取用户信息
	userInfo, err := fetchLinuxDoUserInfo(tokenResp.AccessToken)
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "获取用户信息失败",
		})
		return
	}

	// 处理用户登录或注册
	user, token, isNewUser, err := processLinuxDoLogin(userInfo, tokenResp)
	if err != nil {
		log.Printf("处理Linux Do登录失败: %v", err)
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "登录处理失败: " + err.Error(),
		})
		return
	}

	// 如果是新用户且user为nil，说明需要引导注册
	if isNewUser && user == nil {
		// 重定向到注册引导页面
		redirectURL := fmt.Sprintf("%s/register/oauth-complete?temp_token=%s",
			config.AppConfig.FrontendURL,
			url.QueryEscape(token), // 这里的token实际是临时token
		)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	// 注册设备到Redis
	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	device, err := deviceManager.RegisterDevice(
		user.ID,
		token,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"web",
	)
	if err != nil {
		log.Printf("设备注册失败: user_id=%d, error=%v", user.ID, err)
		// 设备注册失败不影响登录，继续处理
	} else {
		log.Printf("Linux Do用户%s成功: user_id=%d, device_id=%s, ip=%s",
			map[bool]string{true: "注册", false: "登录"}[isNewUser],
			user.ID, device.ID, device.IP)
	}

	// 成功后重定向到前端，携带token信息
	redirectURL := fmt.Sprintf("%s/?auth=success&token=%s&message=%s",
		config.AppConfig.FrontendURL,
		url.QueryEscape(token),
		url.QueryEscape(map[bool]string{true: "注册成功", false: "登录成功"}[isNewUser]),
	)

	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// 获取Linux Do配置状态
func HandleLinuxDoConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"available": isLinuxDoConfigured(),
	})
}

// 检查Linux Do是否已配置
func isLinuxDoConfigured() bool {
	return config.AppConfig.LinuxDoClientID != "" &&
		config.AppConfig.LinuxDoClientSecret != ""
}

// 获取回调地址 - Linux Do固定格式
func getLinuxDoRedirectURI(c *gin.Context) string {
	// Linux Do回调地址是固定格式，直接使用当前Host
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	
	return fmt.Sprintf("%s://%s/oauth/linuxdo", scheme, c.Request.Host)
}

// 生成随机state参数
func generateRandomState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// 使用授权码换取访问令牌
func exchangeCodeForToken(code string, c *gin.Context) (*LinuxDoTokenResponse, error) {
	tokenURL := config.AppConfig.LinuxDoBaseURL + "/oauth2/token"

	// 获取动态回调地址
	redirectURI := getLinuxDoRedirectURI(c)

	// 构建请求参数
	data := url.Values{}
	data.Set("client_id", config.AppConfig.LinuxDoClientID)
	data.Set("client_secret", config.AppConfig.LinuxDoClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	// 发送POST请求
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errorResp LinuxDoErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("获取令牌失败: %s - %s", errorResp.Error, errorResp.ErrorDescription)
		}
		return nil, fmt.Errorf("获取令牌失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析令牌响应
	var tokenResp LinuxDoTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析令牌响应失败: %w", err)
	}

	return &tokenResp, nil
}

// 获取Linux Do用户信息
func fetchLinuxDoUserInfo(accessToken string) (*LinuxDoUserInfo, error) {
	userInfoURL := config.AppConfig.LinuxDoBaseURL + "/api/user"

	// 创建请求
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置授权头
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取用户信息失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析用户信息
	var userInfo LinuxDoUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}

	return &userInfo, nil
}

// 处理Linux Do登录逻辑
func processLinuxDoLogin(userInfo *LinuxDoUserInfo, tokenResp *LinuxDoTokenResponse) (*models.User, string, bool, error) {
	// 首先尝试通过provider_uid查找已绑定的OAuth账号
	var oauthAccount models.OAuthAccount
	err := database.DB.Where("provider = ? AND provider_uid = ?", "linux_do", userInfo.ID).First(&oauthAccount).Error

	if err == nil {
		// 找到已绑定的OAuth账号，更新信息并登录
		err = updateOAuthAccount(&oauthAccount, userInfo, tokenResp)
		if err != nil {
			return nil, "", false, fmt.Errorf("更新OAuth账号失败: %w", err)
		}

		// 获取关联的用户
		var user models.User
		err = database.DB.Where("id = ?", oauthAccount.UserID).First(&user).Error
		if err != nil {
			return nil, "", false, fmt.Errorf("查找关联用户失败: %w", err)
		}

		// 生成访问令牌
		token, err := utils.GenerateAccessToken(user.ID, user.Email)
		if err != nil {
			return nil, "", false, fmt.Errorf("生成访问令牌失败: %w", err)
		}

		return &user, token, false, nil
	}

	// 如果没有找到已绑定的OAuth账号，引导用户完成注册
	// 注意：为了安全起见，不自动匹配现有用户，避免账号劫持风险
	tempToken, err := storeTemporaryLinuxDoUser(userInfo, tokenResp)
	if err != nil {
		return nil, "", false, fmt.Errorf("存储临时用户信息失败: %w", err)
	}

	return nil, tempToken, true, nil
}

// 临时存储Linux Do用户信息结构
type TemporaryLinuxDoUser struct {
	UserInfo    *LinuxDoUserInfo     `json:"user_info"`
	TokenResp   *LinuxDoTokenResponse `json:"token_resp"`
	CreatedAt   time.Time            `json:"created_at"`
}

// 存储临时Linux Do用户信息到Redis
func storeTemporaryLinuxDoUser(userInfo *LinuxDoUserInfo, tokenResp *LinuxDoTokenResponse) (string, error) {
	// 生成临时token
	tempToken := generateRandomState()
	
	// 创建临时用户信息
	tempUser := TemporaryLinuxDoUser{
		UserInfo:  userInfo,
		TokenResp: tokenResp,
		CreatedAt: time.Now(),
	}
	
	// 序列化为JSON
	jsonData, err := json.Marshal(tempUser)
	if err != nil {
		return "", fmt.Errorf("序列化临时用户信息失败: %w", err)
	}
	
	// 存储到Redis，设置30分钟过期时间
	ctx := context.Background()
	key := fmt.Sprintf("temp_linuxdo_user:%s", tempToken)
	err = database.TokenRedisClient.Set(ctx, key, string(jsonData), 30*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("存储到Redis失败: %w", err)
	}
	
	return tempToken, nil
}

// 从Linux Do用户信息创建新用户
func createUserFromLinuxDo(userInfo *LinuxDoUserInfo) (*models.User, error) {
	// 使用Linux Do用户名作为邮箱：username@linux.do
	email := fmt.Sprintf("%s@linux.do", userInfo.Username)
	
	// 直接使用Linux Do用户名，如果冲突则添加#1, #2等后缀
	username := userInfo.Username
	if username == "" {
		username = fmt.Sprintf("linuxdo_user_%d", userInfo.ID)
	}

	// 检查用户名是否已存在，如果存在则添加#后缀
	var existingUser models.User
	originalUsername := username
	counter := 1
	for {
		err := database.DB.Where("username = ?", username).First(&existingUser).Error
		if err != nil {
			// 用户名不存在，可以使用
			break
		}
		// 用户名已存在，添加#数字后缀
		username = fmt.Sprintf("%s#%d", originalUsername, counter)
		counter++
	}

	// 创建新用户
	user := models.User{
		Username:              username,
		Email:                 email,
		Password:              nil, // Linux Do登录的用户没有密码
		DegradationGuaranteed: config.AppConfig.DefaultDegradationGuaranteed,
		DegradationSource:     "system",
		DegradationLocked:     false,
		DegradationCounter:    0,
	}

	err := database.DB.Create(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// 创建OAuth账号绑定
func createOAuthAccount(user *models.User, userInfo *LinuxDoUserInfo, tokenResp *LinuxDoTokenResponse) error {
	// 将ExternalIDs转换为JSON字符串
	externalIDsJSON := ""
	if userInfo.ExternalIDs != nil {
		if jsonBytes, err := json.Marshal(userInfo.ExternalIDs); err == nil {
			externalIDsJSON = string(jsonBytes)
		}
	}

	// 计算token过期时间
	var tokenExpiry *time.Time
	if tokenResp.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		tokenExpiry = &expiry
	}

	oauthAccount := models.OAuthAccount{
		UserID:         user.ID,
		Provider:       "linux_do",
		ProviderUID:    fmt.Sprintf("%d", userInfo.ID),
		Email:          user.Email, // 使用本地用户的邮箱
		Username:       userInfo.Username,
		Name:           userInfo.Name,
		AvatarTemplate: userInfo.AvatarTemplate,
		Active:         userInfo.Active,
		TrustLevel:     userInfo.TrustLevel,
		Silenced:       userInfo.Silenced,
		ExternalIDs:    externalIDsJSON,
		APIKey:         userInfo.APIKey,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		TokenExpiry:    tokenExpiry,
		SyncEnabled:    true,
	}

	now := time.Now()
	oauthAccount.LastSyncAt = &now

	return database.DB.Create(&oauthAccount).Error
}

// 更新OAuth账号信息
func updateOAuthAccount(oauthAccount *models.OAuthAccount, userInfo *LinuxDoUserInfo, tokenResp *LinuxDoTokenResponse) error {
	// 将ExternalIDs转换为JSON字符串
	externalIDsJSON := ""
	if userInfo.ExternalIDs != nil {
		if jsonBytes, err := json.Marshal(userInfo.ExternalIDs); err == nil {
			externalIDsJSON = string(jsonBytes)
		}
	}

	// 计算token过期时间
	var tokenExpiry *time.Time
	if tokenResp.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		tokenExpiry = &expiry
	}

	// 更新OAuth账号信息
	now := time.Now()
	updates := map[string]interface{}{
		"username":        userInfo.Username,
		"name":            userInfo.Name,
		"avatar_template": userInfo.AvatarTemplate,
		"active":          userInfo.Active,
		"trust_level":     userInfo.TrustLevel,
		"silenced":        userInfo.Silenced,
		"external_ids":    externalIDsJSON,
		"api_key":         userInfo.APIKey,
		"access_token":    tokenResp.AccessToken,
		"refresh_token":   tokenResp.RefreshToken,
		"token_expiry":    tokenExpiry,
		"last_sync_at":    &now,
		"updated_at":      now,
	}

	return database.DB.Model(oauthAccount).Updates(updates).Error
}

// 获取临时Linux Do用户信息
func HandleGetTemporaryLinuxDoUser(c *gin.Context) {
	tempToken := c.Query("temp_token")
	if tempToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少临时令牌",
		})
		return
	}

	// 从Redis获取临时用户信息
	ctx := context.Background()
	key := fmt.Sprintf("temp_linuxdo_user:%s", tempToken)
	jsonData, err := database.TokenRedisClient.Get(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "临时用户信息不存在或已过期",
		})
		return
	}

	// 解析JSON数据
	var tempUser TemporaryLinuxDoUser
	err = json.Unmarshal([]byte(jsonData), &tempUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "解析用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tempUser,
	})
}

// 完成OAuth注册请求结构
type CompleteOAuthRegistrationRequest struct {
	TempToken string `json:"temp_token" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

// 完成Linux Do OAuth注册
func HandleCompleteLinuxDoRegistration(c *gin.Context) {
	var req CompleteOAuthRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 从Redis获取临时用户信息
	ctx := context.Background()
	key := fmt.Sprintf("temp_linuxdo_user:%s", req.TempToken)
	jsonData, err := database.TokenRedisClient.Get(ctx, key).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "临时用户信息不存在或已过期",
		})
		return
	}

	// 解析临时用户信息
	var tempUser TemporaryLinuxDoUser
	err = json.Unmarshal([]byte(jsonData), &tempUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "解析用户信息失败",
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	err = database.DB.Where("username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "用户名已被使用，请选择其他用户名",
		})
		return
	}

	// 检查邮箱是否已存在
	err = database.DB.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "邮箱已被使用，请选择其他邮箱",
		})
		return
	}

	// 创建新用户
	user := models.User{
		Username:              req.Username,
		Email:                 req.Email,
		Password:              nil, // Linux Do登录的用户没有密码
		DegradationGuaranteed: config.AppConfig.DefaultDegradationGuaranteed,
		DegradationSource:     "system",
		DegradationLocked:     false,
		DegradationCounter:    0,
	}

	err = database.DB.Create(&user).Error
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建用户失败",
		})
		return
	}

	// 创建OAuth账号绑定
	err = createOAuthAccount(&user, tempUser.UserInfo, tempUser.TokenResp)
	if err != nil {
		log.Printf("创建OAuth绑定失败: %v", err)
		// 如果OAuth绑定失败，删除已创建的用户
		database.DB.Delete(&user)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建OAuth绑定失败",
		})
		return
	}

	// 生成访问令牌
	token, err := utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		log.Printf("生成访问令牌失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "生成访问令牌失败",
		})
		return
	}

	// 注册设备到Redis
	deviceManager := utils.NewDeviceManager(database.TokenRedisClient, database.UserRedisClient, database.DeviceRedisClient)
	device, err := deviceManager.RegisterDevice(
		user.ID,
		token,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		"web",
	)
	if err != nil {
		log.Printf("设备注册失败: user_id=%d, error=%v", user.ID, err)
		// 设备注册失败不影响注册，继续处理
	} else {
		log.Printf("Linux Do用户注册成功: user_id=%d, device_id=%s, ip=%s",
			user.ID, device.ID, device.IP)
	}

	// 处理新用户注册套餐赠送
	if err := utils.ProcessRegistrationPlanGift(user.ID, "linux_do"); err != nil {
		log.Printf("Linux Do新用户套餐赠送失败: user_id=%d, error=%v", user.ID, err)
		// 套餐赠送失败不影响注册，继续处理
	}

	// 删除临时用户信息
	ctx = context.Background()
	database.TokenRedisClient.Del(ctx, key)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "注册成功",
		"token":   token,
		"user":    user,
	})
}
