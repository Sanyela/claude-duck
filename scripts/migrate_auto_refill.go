package main

import (
	"fmt"
	"log"

	"claude/config"
	"claude/database"
)

func main() {
	log.Println("🚀 开始执行自动补给字段迁移脚本...")

	// 加载配置
	config.LoadConfig()

	// 连接数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	log.Println("✅ 数据库连接成功")

	// 执行自动补给字段迁移
	if err := database.ExecuteAutoRefillMigration(); err != nil {
		log.Fatalf("❌ 自动补给字段迁移失败: %v", err)
	}

	log.Println("🎉 自动补给字段迁移完成！")
	fmt.Println("")
	fmt.Println("现在可以在订阅计划中配置自动补给功能：")
	fmt.Println("- auto_refill_enabled: 是否启用自动补给")
	fmt.Println("- auto_refill_threshold: 自动补给阈值（积分低于此值时触发）")
	fmt.Println("- auto_refill_amount: 每次补给的积分数量")
	fmt.Println("")
	fmt.Println("系统将在每天的 0点、4点、8点、12点、16点、20点 自动检查并补给用户积分")
}
