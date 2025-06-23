package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"claude/config"
	"claude/database"
	"claude/models"

	"gorm.io/gorm"
)

const (
	baseURL = "http://127.0.0.1:9998/api"
)

type TestSuite struct {
	adminToken  string
	userToken   string
	userID      uint
	planID      uint
	codeID      string
}

func main() {
	// 加载配置和连接数据库
	config.LoadConfig()
	if err := database.InitDB(); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 执行数据库迁移
	if err := database.Migrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	ts := &TestSuite{}

	// 运行测试流程
	fmt.Println("=== 开始测试流程 ===")
	
	// 1. 创建管理员用户
	fmt.Println("\n1. 创建管理员用户...")
	if err := ts.createAdminUser(); err != nil {
		log.Fatalf("创建管理员失败: %v", err)
	}
	fmt.Println("✓ 管理员创建成功")

	// 2. 管理员登录
	fmt.Println("\n2. 管理员登录...")
	if err := ts.adminLogin(); err != nil {
		log.Fatalf("管理员登录失败: %v", err)
	}
	fmt.Println("✓ 管理员登录成功，Token:", ts.adminToken[:20]+"...")

	// 3. 创建订阅计划
	fmt.Println("\n3. 创建订阅计划...")
	if err := ts.createSubscriptionPlan(); err != nil {
		log.Fatalf("创建订阅计划失败: %v", err)
	}
	fmt.Println("✓ 订阅计划创建成功，ID:", ts.planID)

	// 4. 创建激活码
	fmt.Println("\n4. 创建激活码...")
	if err := ts.createActivationCode(); err != nil {
		log.Fatalf("创建激活码失败: %v", err)
	}
	fmt.Println("✓ 激活码创建成功:", ts.codeID)

	// 5. 用户注册
	fmt.Println("\n5. 用户注册...")
	if err := ts.userRegister(); err != nil {
		log.Fatalf("用户注册失败: %v", err)
	}
	fmt.Println("✓ 用户注册成功，ID:", ts.userID)

	// 6. 验证用户初始积分
	fmt.Println("\n6. 验证用户初始积分...")
	if err := ts.verifyUserCredits(0); err != nil {
		log.Fatalf("验证初始积分失败: %v", err)
	}
	fmt.Println("✓ 初始积分验证成功：0")

	// 7. 用户激活激活码
	fmt.Println("\n7. 用户激活激活码...")
	if err := ts.redeemActivationCode(); err != nil {
		log.Fatalf("激活码兑换失败: %v", err)
	}
	fmt.Println("✓ 激活码兑换成功")

	// 8. 验证用户积分增加
	fmt.Println("\n8. 验证用户积分增加...")
	if err := ts.verifyUserCredits(10000); err != nil {
		log.Fatalf("验证积分增加失败: %v", err)
	}
	fmt.Println("✓ 积分增加验证成功：10000")

	// 9. 模拟扣费
	fmt.Println("\n9. 模拟扣费...")
	if err := ts.simulatePointConsumption(500); err != nil {
		log.Fatalf("模拟扣费失败: %v", err)
	}
	fmt.Println("✓ 扣费成功：500")

	// 10. 验证用户积分减少
	fmt.Println("\n10. 验证用户积分减少...")
	if err := ts.verifyUserCredits(9500); err != nil {
		log.Fatalf("验证积分减少失败: %v", err)
	}
	fmt.Println("✓ 积分减少验证成功：9500")

	// 11. 验证数据库记录
	fmt.Println("\n11. 验证数据库记录...")
	if err := ts.verifyDatabaseRecords(); err != nil {
		log.Fatalf("数据库记录验证失败: %v", err)
	}
	fmt.Println("✓ 数据库记录验证成功")

	fmt.Println("\n=== 测试流程完成 ===")
}

// 创建管理员用户（直接操作数据库）
func (ts *TestSuite) createAdminUser() error {
	// 先检查是否已存在
	var existingUser models.User
	err := database.DB.Where("email = ?", "admin@test.com").First(&existingUser).Error
	if err == nil {
		// 用户已存在，更新为管理员
		existingUser.IsAdmin = true
		return database.DB.Save(&existingUser).Error
	}

	// 创建新管理员
	admin := models.User{
		Email:                 "admin@test.com",
		Username:              "admin",
		Password:              "$2a$10$X7.Kz6Xm9FZhYSO8qY7PFO.Qxn1M5aYRrJQ5YPgYRQqXGq5gXHX.a", // password: admin123
		IsAdmin:               true,
		DegradationGuaranteed: 0,
		DegradationSource:     "system",
		DegradationLocked:     false,
		DegradationCounter:    0,
	}

	return database.DB.Create(&admin).Error
}

// 管理员登录
func (ts *TestSuite) adminLogin() error {
	reqBody := map[string]string{
		"email":    "admin@test.com",
		"password": "admin123",
	}

	resp, err := makeRequest("POST", "/auth/login", nil, reqBody)
	if err != nil {
		return err
	}

	var result struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("登录失败: %s", result.Message)
	}

	ts.adminToken = result.Token
	return nil
}

// 创建订阅计划
func (ts *TestSuite) createSubscriptionPlan() error {
	reqBody := map[string]interface{}{
		"plan_id":                "TEST-PLAN-001",
		"title":                  "测试套餐",
		"description":            "用于测试的订阅套餐",
		"point_amount":           10000,
		"price":                  9.99,
		"currency":               "USD",
		"validity_days":          30,
		"degradation_guaranteed": 3,
		"features":               "[\"Feature 1\", \"Feature 2\"]",
		"active":                 true,
	}

	resp, err := makeRequest("POST", "/admin/subscription-plans", &ts.adminToken, reqBody)
	if err != nil {
		return err
	}

	var plan models.SubscriptionPlan
	if err := json.Unmarshal(resp, &plan); err != nil {
		return err
	}

	ts.planID = plan.ID
	return nil
}

// 创建激活码
func (ts *TestSuite) createActivationCode() error {
	reqBody := map[string]interface{}{
		"count":                1,
		"subscription_plan_id": ts.planID,
		"batch_number":         "TEST-BATCH-001",
	}

	resp, err := makeRequest("POST", "/admin/activation-codes", &ts.adminToken, reqBody)
	if err != nil {
		return err
	}

	var result struct {
		Message     string `json:"message"`
		Count       int    `json:"count"`
		BatchNumber string `json:"batch_number"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	// 从数据库获取创建的激活码
	var code models.ActivationCode
	err = database.DB.Where("batch_number = ?", "TEST-BATCH-001").First(&code).Error
	if err != nil {
		return err
	}

	ts.codeID = code.Code
	return nil
}

// 用户注册
func (ts *TestSuite) userRegister() error {
	timestamp := time.Now().Unix()
	reqBody := map[string]string{
		"username": fmt.Sprintf("testuser_%d", timestamp),
		"email":    fmt.Sprintf("test_%d@example.com", timestamp),
		"password": "test123456",
	}

	resp, err := makeRequest("POST", "/auth/register", nil, reqBody)
	if err != nil {
		return err
	}

	var result struct {
		Success bool   `json:"success"`
		Token   string `json:"token"`
		Message string `json:"message"`
		User    struct {
			ID uint `json:"id"`
		} `json:"user"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("注册失败: %s", result.Message)
	}

	ts.userToken = result.Token
	ts.userID = result.User.ID
	return nil
}

// 验证用户积分
func (ts *TestSuite) verifyUserCredits(expectedPoints int64) error {
	resp, err := makeRequest("GET", "/credits/balance", &ts.userToken, nil)
	if err != nil {
		return err
	}

	var result struct {
		Balance struct {
			TotalPoints     int64 `json:"total_points"`
			UsedPoints      int64 `json:"used_points"`
			AvailablePoints int64 `json:"available_points"`
		} `json:"balance"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if result.Balance.AvailablePoints != expectedPoints {
		return fmt.Errorf("积分不匹配: 期望 %d, 实际 %d", expectedPoints, result.Balance.AvailablePoints)
	}

	return nil
}

// 兑换激活码
func (ts *TestSuite) redeemActivationCode() error {
	reqBody := map[string]string{
		"couponCode": ts.codeID,
	}

	resp, err := makeRequest("POST", "/subscription/redeem", &ts.userToken, reqBody)
	if err != nil {
		return err
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("兑换失败: %s", result.Message)
	}

	return nil
}

// 模拟积分消费
func (ts *TestSuite) simulatePointConsumption(points int64) error {
	reqBody := map[string]interface{}{
		"user_id":           ts.userID,
		"points_to_consume": points,
		"model":             "claude-3-opus-20240229",
		"prompt_tokens":     1000,
		"completion_tokens": 500,
	}

	resp, err := makeRequest("POST", "/admin/test/consume-points", &ts.adminToken, reqBody)
	if err != nil {
		return err
	}

	var result struct {
		Message         string `json:"message"`
		PointsConsumed  int64  `json:"points_consumed"`
		RemainingPoints int64  `json:"remaining_points"`
		Error           string `json:"error"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if result.Error != "" {
		return fmt.Errorf("扣费失败: %s", result.Error)
	}

	return nil
}

// 验证数据库记录
func (ts *TestSuite) verifyDatabaseRecords() error {
	// 验证积分池
	var poolCount int64
	database.DB.Model(&models.PointPool{}).Where("user_id = ?", ts.userID).Count(&poolCount)
	if poolCount == 0 {
		return fmt.Errorf("积分池记录不存在")
	}

	// 验证使用历史
	var historyCount int64
	database.DB.Model(&models.PointUsageHistory{}).Where("user_id = ?", ts.userID).Count(&historyCount)
	if historyCount == 0 {
		return fmt.Errorf("使用历史记录不存在")
	}

	// 验证支付历史
	var paymentCount int64
	database.DB.Model(&models.PaymentHistory{}).Where("user_id = ?", ts.userID).Count(&paymentCount)
	if paymentCount == 0 {
		return fmt.Errorf("支付历史记录不存在")
	}

	// 验证用户服务降级配置
	var user models.User
	err := database.DB.Where("id = ?", ts.userID).First(&user).Error
	if err != nil {
		return err
	}

	if user.DegradationGuaranteed != 3 {
		return fmt.Errorf("用户降级配置未更新: 期望 3, 实际 %d", user.DegradationGuaranteed)
	}

	if user.DegradationSource != "subscription" {
		return fmt.Errorf("用户降级来源错误: 期望 subscription, 实际 %s", user.DegradationSource)
	}

	return nil
}

// 发送HTTP请求的辅助函数
func makeRequest(method, path string, token *string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != nil {
		req.Header.Set("Authorization", "Bearer "+*token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}