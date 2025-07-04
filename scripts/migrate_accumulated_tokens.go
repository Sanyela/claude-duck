package main

import (
	"log"
	"os"

	"claude/config"
	"claude/database"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 检查命令行参数
	if len(os.Args) > 1 && os.Args[1] == "migrate-accumulated-tokens" {
		// 执行累计token计费迁移
		if err := database.ExecuteAccumulatedTokensMigration(); err != nil {
			log.Fatal("Migration failed:", err)
		}
		log.Println("✅ 迁移执行成功！")
		return
	}

	log.Println("使用方法:")
	log.Println("  go run migrate_accumulated_tokens.go migrate-accumulated-tokens")
	log.Println("")
	log.Println("这将执行以下迁移:")
	log.Println("  1. 为user_wallets表添加accumulated_tokens字段")
	log.Println("  2. 删除旧的token_pricing_table配置")
	log.Println("  3. 添加新的token_threshold和points_per_threshold配置")
	log.Println("  4. 重置所有用户的累计token计数为0")
}