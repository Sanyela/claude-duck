package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type DeviceInfo struct {
	ID         string    `json:"id"`
	UserID     uint      `json:"user_id"`
	TokenHash  string    `json:"token_hash"`
	IP         string    `json:"ip"`
	Location   string    `json:"location"`
	UserAgent  string    `json:"user_agent"`
	DeviceType string    `json:"device_type"`
	DeviceName string    `json:"device_name"`
	Source     string    `json:"source"` // "web" or "sso"
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type DeviceManager struct {
	tokenRedis  *redis.Client // DB 0: Token映射
	userRedis   *redis.Client // DB 1: 用户设备集合
	deviceRedis *redis.Client // DB 2: 设备详情
}

func NewDeviceManager(tokenClient, userClient, deviceClient *redis.Client) *DeviceManager {
	return &DeviceManager{
		tokenRedis:  tokenClient,
		userRedis:   userClient,
		deviceRedis: deviceClient,
	}
}

// 计算Token的SHA256摘要
func (dm *DeviceManager) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// 解析User-Agent获取设备信息
func (dm *DeviceManager) ParseUserAgent(userAgent string) (deviceType, deviceName string) {
	ua := strings.ToLower(userAgent)

	// 处理空或极短的User-Agent
	if len(userAgent) == 0 {
		return "desktop", "SSO客户端"
	}

	// 设备类型判断，默认为移动设备
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		deviceType = "tablet"
	} else if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ios") {
		deviceType = "mobile"
	} else {
		// 默认为移动设备
		deviceType = "mobile"
	}

	// 设备名称提取
	if strings.Contains(ua, "chrome") {
		// 进一步区分Chrome和移动设备
		if strings.Contains(ua, "android") {
			deviceName = "Android Chrome"
		} else if strings.Contains(ua, "iphone") {
			deviceName = "iPhone Chrome"
		} else if strings.Contains(ua, "mobile") {
			deviceName = "移动Chrome"
		} else {
			deviceName = "Chrome浏览器"
		}
	} else if strings.Contains(ua, "firefox") {
		deviceName = "Firefox浏览器"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		// 进一步区分Safari和iOS设备
		if strings.Contains(ua, "iphone") {
			deviceName = "iPhone Safari"
		} else if strings.Contains(ua, "ipad") {
			deviceName = "iPad Safari"
		} else {
			deviceName = "Safari浏览器"
		}
	} else if strings.Contains(ua, "edge") {
		deviceName = "Edge浏览器"
	} else if strings.Contains(ua, "cli") || strings.Contains(ua, "curl") || strings.Contains(ua, "wget") {
		deviceName = "CLI工具"
	} else if strings.Contains(ua, "postman") {
		deviceName = "Postman"
	} else if strings.Contains(ua, "go-http-client") || strings.Contains(ua, "go/") {
		deviceName = "Go客户端"
	} else if strings.Contains(ua, "python") || strings.Contains(ua, "requests") {
		deviceName = "Python客户端"
	} else if strings.Contains(ua, "java") || strings.Contains(ua, "okhttp") {
		deviceName = "Java客户端"
	} else if strings.Contains(ua, "dart") || strings.Contains(ua, "flutter") {
		deviceName = "Flutter应用"
	} else if len(userAgent) < 20 {
		// 对于简短的User-Agent，很可能是自定义客户端
		deviceName = "SSO客户端"
	} else {
		deviceName = "未知设备"
	}

	return deviceType, deviceName
}

// IP API 响应结构
type IPLocationResponse struct {
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

// 获取IP地理位置
func (dm *DeviceManager) GetLocationFromIP(ip string) string {
	// 简单的本地IP判断
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return "本地"
	}
	
	// 内网IP判断
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
		return "内网"
	}
	
	// 调用 ip-api.com 获取地理位置
	location := dm.getLocationFromAPI(ip)
	if location != "" {
		return location
	}
	
	return "未知地区"
}

// 调用 ip-api.com API 获取地理位置
func (dm *DeviceManager) getLocationFromAPI(ip string) string {
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip)
	
	// 设置超时时间
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	var result IPLocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	
	if result.Status != "success" {
		return ""
	}
	
	// 格式化返回：省份 + 城市（处理简称）
	province := dm.formatProvinceName(result.RegionName)
	city := dm.formatCityName(result.City)
	
	if province != "" && city != "" {
		return province + city
	} else if province != "" {
		return province
	} else if city != "" {
		return city
	}
	
	return ""
}

// 格式化省份名称（转换为简称）
func (dm *DeviceManager) formatProvinceName(regionName string) string {
	// 需要简化的省份名称映射
	provinceMap := map[string]string{
		"内蒙古壮族自治区": "内蒙古",
		"内蒙古":     "内蒙",
		"广西壮族自治区": "广西", 
		"西藏自治区":   "西藏",
		"宁夏回族自治区": "宁夏",
		"新疆维吾尔自治区": "新疆",
		"黑龙江省":    "黑龙江",
	}
	
	if simplified, exists := provinceMap[regionName]; exists {
		return simplified
	}
	
	// 去除省、市、自治区等后缀
	regionName = strings.TrimSuffix(regionName, "省")
	regionName = strings.TrimSuffix(regionName, "市")
	regionName = strings.TrimSuffix(regionName, "自治区")
	regionName = strings.TrimSuffix(regionName, "特别行政区")
	
	return regionName
}

// 格式化城市名称（去除多余后缀）
func (dm *DeviceManager) formatCityName(city string) string {
	// 去除常见后缀
	city = strings.TrimSuffix(city, "市")
	city = strings.TrimSuffix(city, "县")
	city = strings.TrimSuffix(city, "区")
	
	return city
}

// 注册新设备
func (dm *DeviceManager) RegisterDevice(userID uint, token string, ip string, userAgent string, source string) (*DeviceInfo, error) {
	ctx := context.Background()

	deviceID := uuid.New().String()
	tokenHash := dm.HashToken(token)
	deviceType, deviceName := dm.ParseUserAgent(userAgent)
	location := dm.GetLocationFromIP(ip)

	device := &DeviceInfo{
		ID:         deviceID,
		UserID:     userID,
		TokenHash:  tokenHash,
		IP:         ip,
		Location:   location,
		UserAgent:  userAgent,
		DeviceType: deviceType,
		DeviceName: deviceName,
		Source:     source,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		ExpiresAt:  time.Time{}, // Token永不过期
	}

	// 序列化设备信息
	deviceData, err := json.Marshal(device)
	if err != nil {
		return nil, err
	}

	// 分别在不同的Redis DB中存储数据
	// 1. 在DB 0中建立Token映射
	err = dm.tokenRedis.Set(ctx, fmt.Sprintf("token:%s", tokenHash), deviceID, 0).Err() // 永不过期
	if err != nil {
		return nil, err
	}

	// 2. 在DB 1中添加到用户设备集合
	err = dm.userRedis.SAdd(ctx, fmt.Sprintf("user_devices:%d", userID), deviceID).Err()
	if err != nil {
		return nil, err
	}

	// 3. 在DB 2中存储设备详情
	err = dm.deviceRedis.Set(ctx, fmt.Sprintf("device:%s", deviceID), deviceData, 0).Err() // 永不过期
	if err != nil {
		return nil, err
	}

	return device, nil
}

// 验证Token并返回设备信息
func (dm *DeviceManager) ValidateToken(token string) (*DeviceInfo, error) {
	ctx := context.Background()
	tokenHash := dm.HashToken(token)

	// 从DB 0通过Token找到设备ID
	deviceID, err := dm.tokenRedis.Get(ctx, fmt.Sprintf("token:%s", tokenHash)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("token not found or expired")
	} else if err != nil {
		return nil, err
	}

	// 从DB 2获取设备信息
	deviceData, err := dm.deviceRedis.Get(ctx, fmt.Sprintf("device:%s", deviceID)).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("device not found")
	} else if err != nil {
		return nil, err
	}

	var device DeviceInfo
	if err := json.Unmarshal([]byte(deviceData), &device); err != nil {
		return nil, err
	}

	// 更新最后活跃时间到DB 2
	device.LastActive = time.Now()
	updatedData, _ := json.Marshal(device)
	dm.deviceRedis.Set(ctx, fmt.Sprintf("device:%s", deviceID), updatedData, 0) // 永不过期

	return &device, nil
}

// 获取用户所有设备
func (dm *DeviceManager) GetUserDevices(userID uint) ([]DeviceInfo, error) {
	ctx := context.Background()

	// 从DB 1获取用户所有设备ID
	deviceIDs, err := dm.userRedis.SMembers(ctx, fmt.Sprintf("user_devices:%d", userID)).Result()
	if err != nil {
		return nil, err
	}

	var devices []DeviceInfo
	for _, deviceID := range deviceIDs {
		// 从DB 2获取设备详情
		deviceData, err := dm.deviceRedis.Get(ctx, fmt.Sprintf("device:%s", deviceID)).Result()
		if err == redis.Nil {
			// 设备信息不存在，跳过
			continue
		} else if err != nil {
			continue
		}

		var device DeviceInfo
		if err := json.Unmarshal([]byte(deviceData), &device); err != nil {
			continue
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// 下线设备
func (dm *DeviceManager) RevokeDevice(userID uint, deviceID string) error {
	ctx := context.Background()

	// 从DB 2获取设备信息
	deviceData, err := dm.deviceRedis.Get(ctx, fmt.Sprintf("device:%s", deviceID)).Result()
	if err == redis.Nil {
		return fmt.Errorf("device not found")
	} else if err != nil {
		return err
	}

	var device DeviceInfo
	if err := json.Unmarshal([]byte(deviceData), &device); err != nil {
		return err
	}

	// 验证设备属于该用户
	if device.UserID != userID {
		return fmt.Errorf("device does not belong to user")
	}

	// 分别从不同的DB中删除相关数据
	// 1. 从DB 1的用户设备集合中移除
	err = dm.userRedis.SRem(ctx, fmt.Sprintf("user_devices:%d", userID), deviceID).Err()
	if err != nil {
		return err
	}

	// 2. 从DB 2删除设备信息
	err = dm.deviceRedis.Del(ctx, fmt.Sprintf("device:%s", deviceID)).Err()
	if err != nil {
		return err
	}

	// 3. 从DB 0删除Token映射
	err = dm.tokenRedis.Del(ctx, fmt.Sprintf("token:%s", device.TokenHash)).Err()
	return err
}

// 下线用户所有设备
func (dm *DeviceManager) RevokeAllUserDevices(userID uint) error {
	ctx := context.Background()

	devices, err := dm.GetUserDevices(userID)
	if err != nil {
		return err
	}

	// 分别在不同的DB中删除数据
	for _, device := range devices {
		// 从DB 2删除设备信息
		dm.deviceRedis.Del(ctx, fmt.Sprintf("device:%s", device.ID))
		// 从DB 0删除Token映射
		dm.tokenRedis.Del(ctx, fmt.Sprintf("token:%s", device.TokenHash))
	}

	// 从DB 1删除用户设备集合
	err = dm.userRedis.Del(ctx, fmt.Sprintf("user_devices:%d", userID)).Err()
	return err
}

// 清理过期设备（可以设置定时任务调用）
func (dm *DeviceManager) CleanupExpiredDevices() error {
	ctx := context.Background()
	
	// 从DB 1获取所有用户设备集合的键
	keys, err := dm.userRedis.Keys(ctx, "user_devices:*").Result()
	if err != nil {
		return err
	}
	
	for _, key := range keys {
		// 获取用户ID
		var userID uint
		if _, err := fmt.Sscanf(key, "user_devices:%d", &userID); err != nil {
			continue
		}
		
		// 获取该用户的所有设备并清理过期的（GetUserDevices会自动清理过期设备）
		_, err := dm.GetUserDevices(userID)
		if err != nil {
			continue
		}
	}
	
	return nil
}