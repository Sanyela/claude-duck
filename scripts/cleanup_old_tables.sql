-- 1. 删除老架构的核心表
-- 注意：只删除已完全迁移且不再使用的表

-- 删除外键约束
-- 注意：如果外键不存在会报错，但不影响继续执行
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS daily_points_usage;
DROP TABLE IF EXISTS subscriptions;
SET FOREIGN_KEY_CHECKS = 1;

-- 表已在上面删除，这里不需要重复

-- 4. 重建性能优化索引
CREATE INDEX IF NOT EXISTS idx_user_wallets_status ON user_wallets(status);
CREATE INDEX IF NOT EXISTS idx_user_wallets_expires_at ON user_wallets(wallet_expires_at);
CREATE INDEX IF NOT EXISTS idx_redemption_records_user_source ON redemption_records(user_id, source_type);
CREATE INDEX IF NOT EXISTS idx_redemption_records_activated_at ON redemption_records(activated_at);
CREATE INDEX IF NOT EXISTS idx_user_daily_usage_user_date ON user_daily_usage(user_id, usage_date);

-- 5. 优化表结构
-- 添加复合索引提升查询性能
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_daily_usage_unique ON user_daily_usage(user_id, usage_date);

-- 6. 清理系统配置中不再使用的配置项
DELETE FROM system_configs WHERE config_key IN (
    'daily_checkin_multi_subscription_strategy'  -- 新架构中不再需要多订阅策略
);

-- 7. 更新系统配置说明
UPDATE system_configs 
SET description = '每日签到奖励积分数量（钱包架构）'
WHERE config_key = 'daily_checkin_points';

-- =====================================================
-- 验证清理结果
-- =====================================================

-- 检查表是否成功删除
SELECT table_name FROM information_schema.tables 
WHERE table_schema = DATABASE() 
AND table_name IN ('subscriptions', 'daily_points_usage', 'gift_records');

-- 检查新表数据完整性
SELECT 
    (SELECT COUNT(*) FROM users) as user_count,
    (SELECT COUNT(*) FROM user_wallets) as wallet_count,
    (SELECT COUNT(*) FROM redemption_records) as redemption_count;

-- =====================================================
-- 完成清理
-- =====================================================
-- 如果所有验证通过，老架构清理完成
-- 新的钱包架构已完全生效 