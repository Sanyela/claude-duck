package handlers

import (
	"claude/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// BingImageResponse 定义了Bing API返回的图片数据结构
type BingImageResponse struct {
	Images []struct {
		StartDate     string `json:"startdate"`
		FullStartDate string `json:"fullstartdate"`
		EndDate       string `json:"enddate"`
		URL           string `json:"url"`
		URLBase       string `json:"urlbase"`
		Copyright     string `json:"copyright"`
		CopyrightLink string `json:"copyrightlink"`
		Title         string `json:"title"`
		Quiz          string `json:"quiz"`
		Wp            bool   `json:"wp"`
		Hsh           string `json:"hsh"`
		Drk           int    `json:"drk"`
		Top           int    `json:"top"`
		Bot           int    `json:"bot"`
		Hs            []any  `json:"hs"`
	} `json:"images"`
	Tooltips struct {
		Loading  string `json:"loading"`
		Previous string `json:"previous"`
		Next     string `json:"next"`
		Walle    string `json:"walle"`
		Walls    string `json:"walls"`
	} `json:"tooltips"`
}

// Redis客户端
var redisClient *redis.Client

// 初始化Redis客户端
func InitRedisClient() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
		Password: config.AppConfig.RedisPassword,
		DB:       2, // 使用DB 2避免与设备管理(DB 0)和认证(DB 1)冲突
	})
}

// GetBingDailyImage 处理获取Bing每日图片的请求
func GetBingDailyImage(c *gin.Context) {
	// 获取今天的日期作为缓存键
	today := time.Now().Format("20060102")
	cacheKey := fmt.Sprintf("bing:daily:image:%s", today)

	// 检查Redis缓存
	ctx := context.Background()
	cachedURL, err := redisClient.Get(ctx, cacheKey).Result()

	// 如果缓存存在且没有错误，直接返回缓存的URL
	if err == nil && cachedURL != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"url":     cachedURL,
			"cached":  true,
		})
		return
	}

	// 如果缓存不存在或出错，请求Bing API
	bingAPIURL := "https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1"
	resp, err := http.Get(bingAPIURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "无法连接到Bing API",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "无法读取Bing API响应",
		})
		return
	}

	// 解析JSON响应
	var bingResp BingImageResponse
	if err := json.Unmarshal(body, &bingResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "无法解析Bing API响应",
		})
		return
	}

	// 检查是否有图片数据
	if len(bingResp.Images) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Bing API未返回图片数据",
		})
		return
	}

	// 构建完整的图片URL
	imageURL := fmt.Sprintf("https://cn.bing.com%s", bingResp.Images[0].URL)

	// 将URL缓存到Redis，设置过期时间为24小时
	err = redisClient.Set(ctx, cacheKey, imageURL, 24*time.Hour).Err()
	if err != nil {
		// 即使缓存失败，也继续返回图片URL
		fmt.Printf("Redis缓存失败: %v\n", err)
	}

	// 返回图片URL
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"url":     imageURL,
		"cached":  false,
	})
}
