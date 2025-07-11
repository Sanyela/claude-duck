-- 开始事务
START TRANSACTION;

-- 设置安全模式
SET SQL_SAFE_UPDATES = 0;

-- =====================================================
-- 第一步：分析当前钱包状态
-- =====================================================

-- 查看当前钱包状态分布
SELECT 
    status,
    COUNT(*) as wallet_count,
    COUNT(*) * 100.0 / (SELECT COUNT(*) FROM user_wallets) as percentage
FROM user_wallets 
GROUP BY status
ORDER BY wallet_count DESC;

-- 查看过期但仍有积分的钱包
SELECT 
    COUNT(*) as expired_with_points_count,
    SUM(available_points) as total_available_points,
    AVG(available_points) as avg_available_points
FROM user_wallets 
WHERE status = 'expired' 
AND available_points > 0;

-- 查看即将过期的钱包（未来7天内）
SELECT 
    COUNT(*) as expiring_soon_count,
    SUM(available_points) as total_points_at_risk
FROM user_wallets 
WHERE status = 'active' 
AND wallet_expires_at <= DATE_ADD(NOW(), INTERVAL 7 DAY)
AND wallet_expires_at > NOW();

-- =====================================================
-- 第二步：更新钱包状态逻辑
-- =====================================================

-- 创建临时表记录更新前的状态
CREATE TEMPORARY TABLE wallet_status_backup AS
SELECT 
    user_id,
    status as old_status,
    wallet_expires_at as old_expires_at,
    available_points,
    total_points,
    NOW() as backup_time
FROM user_wallets;

-- =====================================================
-- 第三步：执行钱包状态更新
-- =====================================================

-- 1. 将有积分但状态为expired的钱包设为active
UPDATE user_wallets 
SET status = 'active',
    updated_at = NOW()
WHERE status = 'expired' 
AND available_points > 0;

-- 记录第一步更新结果
SELECT '步骤1：激活有积分的过期钱包' as step, ROW_COUNT() as affected_rows;

-- 2. 将有兑换记录且未过期的钱包设为active
UPDATE user_wallets uw
SET status = 'active',
    updated_at = NOW()
WHERE uw.status = 'expired'
AND EXISTS (
    SELECT 1 FROM redemption_records rr 
    WHERE rr.user_id = uw.user_id 
    AND rr.expires_at > NOW()
);

-- 记录第二步更新结果
SELECT '步骤2：基于兑换记录激活钱包' as step, ROW_COUNT() as affected_rows;

-- 3. 更新钱包过期时间为最新的兑换记录过期时间
UPDATE user_wallets uw
SET wallet_expires_at = (
    SELECT MAX(rr.expires_at)
    FROM redemption_records rr 
    WHERE rr.user_id = uw.user_id 
    AND rr.expires_at > NOW()
),
updated_at = NOW()
WHERE uw.status = 'active'
AND EXISTS (
    SELECT 1 FROM redemption_records rr 
    WHERE rr.user_id = uw.user_id 
    AND rr.expires_at > NOW()
    AND rr.expires_at > uw.wallet_expires_at
);

-- 记录第三步更新结果
SELECT '步骤3：更新钱包过期时间' as step, ROW_COUNT() as affected_rows;

-- 4. 对于没有有效兑换记录但有积分的钱包，延长过期时间至30天后
UPDATE user_wallets 
SET wallet_expires_at = DATE_ADD(NOW(), INTERVAL 30 DAY),
    updated_at = NOW()
WHERE status = 'active' 
AND available_points > 0
AND wallet_expires_at <= NOW();

-- 记录第四步更新结果
SELECT '步骤4：延长有积分钱包过期时间' as step, ROW_COUNT() as affected_rows;

-- 5. 同步用户钱包的套餐配置（从最新的兑换记录）
UPDATE user_wallets uw
INNER JOIN (
    SELECT 
        rr.user_id,
        rr.daily_max_points,
        rr.degradation_guaranteed,
        rr.daily_checkin_points,
        rr.daily_checkin_points_max,
        rr.auto_refill_enabled,
        rr.auto_refill_threshold,
        rr.auto_refill_amount,
        ROW_NUMBER() OVER (PARTITION BY rr.user_id ORDER BY rr.created_at DESC) as rn
    FROM redemption_records rr
    WHERE rr.expires_at > NOW()
    AND rr.subscription_plan_id IS NOT NULL
) latest_config ON uw.user_id = latest_config.user_id AND latest_config.rn = 1
SET 
    uw.daily_max_points = latest_config.daily_max_points,
    uw.degradation_guaranteed = latest_config.degradation_guaranteed,
    uw.daily_checkin_points = COALESCE(NULLIF(latest_config.daily_checkin_points, 0), uw.daily_checkin_points),
    uw.daily_checkin_points_max = COALESCE(NULLIF(latest_config.daily_checkin_points_max, 0), uw.daily_checkin_points_max),
    uw.auto_refill_enabled = latest_config.auto_refill_enabled,
    uw.auto_refill_threshold = latest_config.auto_refill_threshold,
    uw.auto_refill_amount = latest_config.auto_refill_amount,
    uw.updated_at = NOW()
WHERE uw.status = 'active';

-- 记录第五步更新结果
SELECT '步骤5：同步套餐配置' as step, ROW_COUNT() as affected_rows;

-- =====================================================
-- 第四步：数据验证和报告
-- =====================================================

-- 验证更新后的钱包状态
SELECT 
    '更新后钱包状态分布' as report_type,
    status,
    COUNT(*) as wallet_count,
    SUM(available_points) as total_points,
    AVG(available_points) as avg_points
FROM user_wallets 
GROUP BY status
ORDER BY wallet_count DESC;

-- 比较更新前后的变化
SELECT 
    '状态变更统计' as report_type,
    backup.old_status,
    uw.status as new_status,
    COUNT(*) as change_count,
    SUM(uw.available_points) as total_points_affected
FROM wallet_status_backup backup
JOIN user_wallets uw ON backup.user_id = uw.user_id
WHERE backup.old_status != uw.status
GROUP BY backup.old_status, uw.status
ORDER BY change_count DESC;

-- 检查异常情况
SELECT 
    '异常检查' as report_type,
    '有积分但仍过期的钱包' as issue_type,
    COUNT(*) as issue_count
FROM user_wallets 
WHERE status = 'expired' AND available_points > 0
UNION ALL
SELECT 
    '异常检查' as report_type,
    '活跃但已过期的钱包' as issue_type,
    COUNT(*) as issue_count
FROM user_wallets 
WHERE status = 'active' AND wallet_expires_at <= NOW()
UNION ALL
SELECT 
    '异常检查' as report_type,
    '签到配置为0的活跃钱包' as issue_type,
    COUNT(*) as issue_count
FROM user_wallets 
WHERE status = 'active' 
AND daily_checkin_points = 0 
AND daily_checkin_points_max = 0;

-- =====================================================
-- 第五步：清理和完成
-- =====================================================

-- 重置安全模式
SET SQL_SAFE_UPDATES = 1;

-- 提交事务
COMMIT;

-- =====================================================
-- 验证脚本（可选执行）
-- =====================================================

-- 显示详细的用户钱包信息样本
SELECT 
    uw.user_id,
    u.username,
    uw.status,
    uw.available_points,
    uw.wallet_expires_at,
    uw.daily_checkin_points,
    uw.daily_checkin_points_max,
    uw.auto_refill_enabled,
    (SELECT COUNT(*) FROM redemption_records rr 
     WHERE rr.user_id = uw.user_id AND rr.expires_at > NOW()) as active_redemptions
FROM user_wallets uw
JOIN users u ON uw.user_id = u.id
WHERE uw.status = 'active'
ORDER BY uw.available_points DESC
LIMIT 10;

-- =====================================================
-- 完成报告
-- =====================================================
SELECT 
    '=== 钱包状态迁移完成 ===' as message,
    NOW() as completion_time,
    (SELECT COUNT(*) FROM user_wallets WHERE status = 'active') as active_wallets,
    (SELECT COUNT(*) FROM user_wallets WHERE status = 'expired') as expired_wallets,
    (SELECT SUM(available_points) FROM user_wallets WHERE status = 'active') as total_active_points;
