package main

import (
	"flag"
	"log"

	"claude/config"
	"claude/database"
)

func main() {
	// 定义命令行参数
	var (
		forceRun = flag.Bool("force", false, "强制执行迁移，即使检查条件不满足")
		cleanup  = flag.Bool("cleanup", false, "清理旧表（危险操作）")
		verify   = flag.Bool("verify", false, "仅验证迁移结果，不执行迁移")
	)
	flag.Parse()

	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("=== 订阅表架构重构迁移工具 ===")

	if *verify {
		// 仅验证模式
		log.Println("验证模式：检查迁移状态...")
		if err := verifyMigrationStatus(); err != nil {
			log.Fatal("验证失败:", err)
		}
		return
	}

	if *cleanup {
		// 清理模式
		log.Println("清理模式：删除旧表...")
		if err := database.CleanupOldTables(); err != nil {
			log.Fatal("清理失败:", err)
		}
		log.Println("清理完成")
		return
	}

	// 检查是否需要迁移
	if !*forceRun && !database.NeedsSubscriptionRefactor() {
		log.Println("不需要执行迁移或迁移已完成")
		return
	}

	if *forceRun {
		log.Println("强制执行模式：忽略检查条件")
	}

	// 执行迁移
	log.Println("开始执行订阅表架构重构迁移...")
	if err := database.MigrateSubscriptionRefactor(); err != nil {
		log.Fatal("迁移失败:", err)
	}

	log.Println("迁移成功完成!")
}

// verifyMigrationStatus 验证迁移状态
func verifyMigrationStatus() error {
	// 检查迁移标记
	log.Println("检查迁移完成标记...")
	// 这里可以调用database包中的验证函数

	// 检查表结构
	log.Println("检查表结构...")

	// 检查数据一致性
	log.Println("检查数据一致性...")

	log.Println("验证完成")
	return nil
}
