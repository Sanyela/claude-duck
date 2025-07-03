-- =====================================================
-- 清理老架构的表和字段
-- 警告：执行此脚本前请确保新架构运行正常且数据一致性验证通过
-- =====================================================

-- 1. 删除老架构相关的表
-- 注意：这些表在新架构中已被 user_wallets 和 redemption_records 替代

-- 备份表（可选，如果需要保留备份）
-- CREATE TABLE subscriptions_backup AS SELECT * FROM subscriptions;
-- CREATE TABLE daily_points_usage_backup AS SELECT * FROM daily_points_usage;
-- CREATE TABLE gift_records_backup AS SELECT * FROM gift_records;

-- 删除外键约束
ALTER TABLE daily_points_usage DROP FOREIGN KEY IF EXISTS daily_points_usage_subscription_id_foreign;
ALTER TABLE gift_records DROP FOREIGN KEY IF EXISTS gift_records_subscription_id_foreign;

-- 删除老表
DROP TABLE IF EXISTS daily_points_usage;
DROP TABLE IF EXISTS gift_records;
DROP TABLE IF EXISTS subscriptions;

-- 2. 从 users 表中删除冗余字段
-- 这些字段的功能已移到 user_wallets 表中
ALTER TABLE users 
DROP COLUMN IF EXISTS degradation_guaranteed,
DROP COLUMN IF EXISTS degradation_source,
DROP COLUMN IF EXISTS degradation_locked,
DROP COLUMN IF EXISTS degradation_counter;

-- 3. 删除不再使用的索引
-- DROP INDEX IF EXISTS idx_subscriptions_user_id ON subscriptions;
-- DROP INDEX IF EXISTS idx_subscriptions_status ON subscriptions;
-- DROP INDEX IF EXISTS idx_daily_points_usage_user_date ON daily_points_usage;

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
-- SELECT table_name FROM information_schema.tables 
-- WHERE table_schema = DATABASE() 
-- AND table_name IN ('subscriptions', 'daily_points_usage', 'gift_records');

-- 检查新表数据完整性
-- SELECT 
--     (SELECT COUNT(*) FROM users) as user_count,
--     (SELECT COUNT(*) FROM user_wallets) as wallet_count,
--     (SELECT COUNT(*) FROM redemption_records) as redemption_count;

-- =====================================================
-- 完成清理
-- =====================================================
-- 如果所有验证通过，老架构清理完成
-- 新的钱包架构已完全生效 