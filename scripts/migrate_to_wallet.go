package main

import (
	"log"
	"os"

	"claude/config"
	"claude/database"
)

func main() {
	log.Println("🔄 开始执行钱包架构迁移...")

	// 初始化配置
	config.LoadConfig()

	// 初始化数据库连接
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 执行数据库迁移 (创建新表)
	if err := database.Migrate(); err != nil {
		log.Fatalf("❌ 数据库表迁移失败: %v", err)
	}

	// 检查是否有 --verify 参数
	if len(os.Args) > 1 && os.Args[1] == "--verify" {
		log.Println("🔍 仅验证迁移结果...")
		if err := database.VerifyMigration(); err != nil {
			log.Fatalf("❌ 迁移验证失败: %v", err)
		}
		log.Println("✅ 验证完成！")
		return
	}

	// 执行数据迁移
	if err := database.MigrateToWalletArchitecture(); err != nil {
		log.Fatalf("❌ 数据迁移失败: %v", err)
	}

	// 验证迁移结果
	if err := database.VerifyMigration(); err != nil {
		log.Fatalf("❌ 迁移验证失败: %v", err)
	}

	log.Println("🎉 钱包架构迁移成功完成！")
	log.Println("📋 下一步:")
	log.Println("   1. 测试新架构功能是否正常")
	log.Println("   2. 确认数据一致性")
	log.Println("   3. 运行 cleanup_old_tables.sql 清理老表")
}
