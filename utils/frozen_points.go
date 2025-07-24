package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"claude/database"
	"claude/models"

	"gorm.io/gorm"
)

// CardUsageDetail 卡密使用详情
type CardUsageDetail struct {
	CardCode        string `json:"card_code"`        // 卡密代码
	OriginalPoints  int64  `json:"original_points"`  // 原始积分
	ConsumedPoints  int64  `json:"consumed_points"`  // 已消费积分
	RemainingPoints int64  `json:"remaining_points"` // 剩余积分
	Status          string `json:"status"`           // unused, partially_consumed, fully_consumed
	ExpiresAt       string `json:"expires_at"`       // 到期时间
}

// CardConsumptionResult 卡密消费计算结果
type CardConsumptionResult struct {
	TargetCard      models.RedemptionRecord `json:"target_card"`      // 目标卡密
	RemainingPoints int64                   `json:"remaining_points"` // 剩余积分
	ConsumedPoints  int64                   `json:"consumed_points"`  // 已消费积分
	AllCards        []CardUsageDetail       `json:"all_cards"`        // 所有卡密详情
}

// VirtualConsumptionCalculator 虚拟消费计算器
type VirtualConsumptionCalculator struct {
	UserID          uint                      `json:"user_id"`
	AllRedemptions  []models.RedemptionRecord `json:"all_redemptions"`
	TotalUsedPoints int64                     `json:"total_used_points"`
}

// CalculateCardConsumption 计算卡密的虚拟消费情况
func (calc *VirtualConsumptionCalculator) CalculateCardConsumption(targetCardCode string) (*CardConsumptionResult, error) {
	// 1. 获取所有激活码兑换记录，按优先级排序
	activeCards := calc.getSortedActiveCards()

	// 2. 模拟消费过程
	remainingUsage := calc.TotalUsedPoints
	consumptionDetails := make([]CardUsageDetail, 0)
	var targetResult *CardConsumptionResult

	for _, card := range activeCards {
		// 计算这张卡被消费了多少
		consumed := int64(0)
		remaining := card.PointsAmount

		if remainingUsage > 0 {
			consumed = min(remainingUsage, card.PointsAmount)
			remaining = card.PointsAmount - consumed
			remainingUsage -= consumed
		}

		status := "unused"
		if consumed > 0 && remaining > 0 {
			status = "partially_consumed"
		} else if consumed > 0 && remaining == 0 {
			status = "fully_consumed"
		}

		detail := CardUsageDetail{
			CardCode:        card.SourceID,
			OriginalPoints:  card.PointsAmount,
			ConsumedPoints:  consumed,
			RemainingPoints: remaining,
			Status:          status,
			ExpiresAt:       card.ExpiresAt.Format("2006-01-02 15:04:05"),
		}

		consumptionDetails = append(consumptionDetails, detail)

		// 如果是目标卡密，记录结果
		if card.SourceID == targetCardCode {
			targetResult = &CardConsumptionResult{
				TargetCard:      card,
				RemainingPoints: remaining,
				ConsumedPoints:  consumed,
				AllCards:        consumptionDetails,
			}
		}
	}

	if targetResult == nil {
		return nil, errors.New("未找到目标卡密")
	}

	// 设置完整的卡密详情列表
	targetResult.AllCards = consumptionDetails
	return targetResult, nil
}

// getSortedActiveCards 获取排序后的激活卡密
func (calc *VirtualConsumptionCalculator) getSortedActiveCards() []models.RedemptionRecord {
	var cards []models.RedemptionRecord
	var lastSameLevelTime time.Time

	// 找到最后一次同级兑换的时间
	for _, record := range calc.AllRedemptions {
		if record.SourceType == "activation_code" &&
			strings.Contains(record.Reason, "same_level服务") &&
			record.ActivatedAt.After(lastSameLevelTime) {
			lastSameLevelTime = record.ActivatedAt
		}
	}

	// 获取所有activation_code类型的兑换记录
	for _, record := range calc.AllRedemptions {
		if record.SourceType == "activation_code" {
			// 如果存在同级兑换，只考虑该时间之后的记录
			if !lastSameLevelTime.IsZero() && record.ActivatedAt.Before(lastSameLevelTime) {
				continue // 跳过同级兑换之前的记录
			}
			cards = append(cards, record)
		}
	}

	// 排序规则：先到期先使用，同时到期优先消耗积分少的
	sort.Slice(cards, func(i, j int) bool {
		expireI := cards[i].ExpiresAt
		expireJ := cards[j].ExpiresAt

		if expireI.Equal(expireJ) {
			// 同时到期，积分少的优先
			return cards[i].PointsAmount < cards[j].PointsAmount
		}
		// 先到期的优先
		return expireI.Before(expireJ)
	})

	return cards
}

// min 辅助函数
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// BanActivationCode 封禁激活码主函数
func BanActivationCode(userID uint, activationCode string, reason string, adminUserID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 检查是否已经被封禁
		var existingBan models.FrozenPointsRecord
		if err := tx.Where("user_id = ? AND banned_activation_code = ? AND status = 'frozen'",
			userID, activationCode).First(&existingBan).Error; err == nil {
			return errors.New("该激活码已经被封禁")
		}

		// 2. 获取用户当前钱包状态
		wallet, err := GetOrCreateUserWallet(userID)
		if err != nil {
			return fmt.Errorf("获取用户钱包失败: %w", err)
		}

		// 3. 获取所有兑换记录
		var allRedemptions []models.RedemptionRecord
		if err := tx.Where("user_id = ?", userID).Find(&allRedemptions).Error; err != nil {
			return fmt.Errorf("获取兑换记录失败: %w", err)
		}

		// 4. 计算虚拟消费情况
		calculator := &VirtualConsumptionCalculator{
			UserID:          userID,
			AllRedemptions:  allRedemptions,
			TotalUsedPoints: wallet.UsedPoints,
		}

		consumptionResult, err := calculator.CalculateCardConsumption(activationCode)
		if err != nil {
			return fmt.Errorf("计算卡密消费情况失败: %w", err)
		}

		// 5. 创建封禁前状态快照
		beforeBanSnapshot, err := createWalletSnapshot(wallet)
		if err != nil {
			return fmt.Errorf("创建钱包快照失败: %w", err)
		}

		beforeBenefitsSnapshot, err := createBenefitsSnapshot(wallet)
		if err != nil {
			return fmt.Errorf("创建权益快照失败: %w", err)
		}

		// 6. 从钱包扣除该卡密的剩余积分
		frozenPoints := consumptionResult.RemainingPoints
		if frozenPoints > 0 {
			wallet.AvailablePoints -= frozenPoints
			wallet.TotalPoints -= frozenPoints
		}

		// 7. 重新计算剩余卡密的权益
		newBenefits, err := calculateRemainingBenefits(userID, activationCode, consumptionResult.AllCards, tx)
		if err != nil {
			return fmt.Errorf("重新计算权益失败: %w", err)
		}

		// 8. 更新钱包权益
		updateWalletBenefits(wallet, newBenefits)

		// 9. 检查是否没有剩余有效卡密，如果是，清空钱包并设为过期
		hasRemainingCards := false
		for _, card := range consumptionResult.AllCards {
			if card.CardCode != activationCode && card.RemainingPoints > 0 {
				hasRemainingCards = true
				break
			}
		}

		if !hasRemainingCards {
			// 没有剩余有效卡密，清空钱包积分并设为过期状态
			wallet.AvailablePoints = 0
			wallet.TotalPoints = 0
			wallet.UsedPoints = 0
			wallet.Status = "expired"
			wallet.WalletExpiresAt = time.Now()
		}

		// 10. 生成计算日志
		calculationLog := generateCalculationLog(consumptionResult)

		// 11. 创建冻结记录
		frozenRecord := &models.FrozenPointsRecord{
			UserID:               userID,
			BannedActivationCode: activationCode,
			BannedCodeID:         getBannedCodeID(activationCode, tx),
			FrozenPoints:         frozenPoints,
			FrozenBenefits:       extractCardBenefitsJSON(consumptionResult.TargetCard),
			BeforeBanWalletState: beforeBanSnapshot,
			BeforeBanBenefits:    beforeBenefitsSnapshot,
			CalculationMethod:    calculationLog,
			EstimatedUsage:       consumptionResult.ConsumedPoints,
			BanReason:            reason,
			AdminUserID:          &adminUserID,
			Status:               "frozen",
		}

		// 12. 更新数据库
		if err := tx.Save(wallet).Error; err != nil {
			return fmt.Errorf("更新钱包失败: %w", err)
		}

		if err := tx.Create(frozenRecord).Error; err != nil {
			return fmt.Errorf("创建冻结记录失败: %w", err)
		}

		return nil
	})
}

// UnbanActivationCode 解禁激活码
func UnbanActivationCode(userID uint, activationCode string, adminUserID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 查找冻结记录
		var frozenRecord models.FrozenPointsRecord
		err := tx.Where("user_id = ? AND banned_activation_code = ? AND status = 'frozen'",
			userID, activationCode).First(&frozenRecord).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("未找到该卡密的封禁记录")
			}
			return fmt.Errorf("查询冻结记录失败: %w", err)
		}

		// 2. 获取当前钱包状态
		wallet, err := GetOrCreateUserWallet(userID)
		if err != nil {
			return fmt.Errorf("获取用户钱包失败: %w", err)
		}

		// 3. 恢复冻结的积分
		wallet.AvailablePoints += frozenRecord.FrozenPoints
		wallet.TotalPoints += frozenRecord.FrozenPoints

		// 4. 重新计算包含该卡密的权益
		newBenefits, err := recalculateBenefitsWithUnbannedCard(userID, activationCode, tx)
		if err != nil {
			return fmt.Errorf("重新计算权益失败: %w", err)
		}

		// 5. 更新钱包权益
		updateWalletBenefits(wallet, newBenefits)

		// 6. 更新冻结记录状态
		frozenRecord.Status = "restored"
		frozenRecord.UpdatedAt = time.Now()

		// 7. 更新数据库
		if err := tx.Save(wallet).Error; err != nil {
			return fmt.Errorf("更新钱包失败: %w", err)
		}

		if err := tx.Save(&frozenRecord).Error; err != nil {
			return fmt.Errorf("更新冻结记录失败: %w", err)
		}

		return nil
	})
}

// 辅助函数们...

// createWalletSnapshot 创建钱包状态快照
func createWalletSnapshot(wallet *models.UserWallet) (string, error) {
	snapshot := map[string]interface{}{
		"total_points":             wallet.TotalPoints,
		"available_points":         wallet.AvailablePoints,
		"used_points":              wallet.UsedPoints,
		"accumulated_tokens":       wallet.AccumulatedTokens,
		"daily_max_points":         wallet.DailyMaxPoints,
		"degradation_guaranteed":   wallet.DegradationGuaranteed,
		"daily_checkin_points":     wallet.DailyCheckinPoints,
		"daily_checkin_points_max": wallet.DailyCheckinPointsMax,
		"auto_refill_enabled":      wallet.AutoRefillEnabled,
		"auto_refill_threshold":    wallet.AutoRefillThreshold,
		"auto_refill_amount":       wallet.AutoRefillAmount,
		"wallet_expires_at":        wallet.WalletExpiresAt,
		"status":                   wallet.Status,
		"last_checkin_date":        wallet.LastCheckinDate,
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// createBenefitsSnapshot 创建权益状态快照
func createBenefitsSnapshot(wallet *models.UserWallet) (string, error) {
	benefits := map[string]interface{}{
		"daily_max_points":         wallet.DailyMaxPoints,
		"degradation_guaranteed":   wallet.DegradationGuaranteed,
		"daily_checkin_points":     wallet.DailyCheckinPoints,
		"daily_checkin_points_max": wallet.DailyCheckinPointsMax,
		"auto_refill_enabled":      wallet.AutoRefillEnabled,
		"auto_refill_threshold":    wallet.AutoRefillThreshold,
		"auto_refill_amount":       wallet.AutoRefillAmount,
	}

	data, err := json.Marshal(benefits)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// calculateRemainingBenefits 计算剩余卡密的综合权益
func calculateRemainingBenefits(userID uint, bannedCode string, allCards []CardUsageDetail, tx *gorm.DB) (map[string]interface{}, error) {
	remainingCards := make([]CardUsageDetail, 0)

	// 找出所有还有剩余积分的卡密（除了被封禁的）
	for _, card := range allCards {
		if card.CardCode != bannedCode && card.RemainingPoints > 0 {
			remainingCards = append(remainingCards, card)
		}
	}

	if len(remainingCards) == 0 {
		// 没有剩余卡密，恢复到初始状态
		return map[string]interface{}{
			"daily_max_points":         int64(0),
			"degradation_guaranteed":   0,
			"daily_checkin_points":     int64(0),
			"daily_checkin_points_max": int64(0),
			"auto_refill_enabled":      false,
			"auto_refill_threshold":    int64(0),
			"auto_refill_amount":       int64(0),
		}, nil
	}

	// 有剩余卡密，计算综合权益（取最优配置）
	return mergeBenefitsFromCards(remainingCards, tx)
}

// mergeBenefitsFromCards 合并多张卡密的权益（取最优）
func mergeBenefitsFromCards(cards []CardUsageDetail, tx *gorm.DB) (map[string]interface{}, error) {
	benefits := map[string]interface{}{
		"daily_max_points":         int64(0),
		"degradation_guaranteed":   0,
		"daily_checkin_points":     int64(0),
		"daily_checkin_points_max": int64(0),
		"auto_refill_enabled":      false,
		"auto_refill_threshold":    int64(0),
		"auto_refill_amount":       int64(0),
	}

	for _, card := range cards {
		cardBenefits, err := getCardBenefits(card.CardCode, tx)
		if err != nil {
			continue // 跳过获取失败的卡密
		}

		// 取最大值策略
		if dailyMax, ok := cardBenefits["daily_max_points"].(int64); ok && dailyMax > benefits["daily_max_points"].(int64) {
			benefits["daily_max_points"] = dailyMax
		}

		if degradation, ok := cardBenefits["degradation_guaranteed"].(int); ok && degradation > benefits["degradation_guaranteed"].(int) {
			benefits["degradation_guaranteed"] = degradation
		}

		if checkinMin, ok := cardBenefits["daily_checkin_points"].(int64); ok && checkinMin > benefits["daily_checkin_points"].(int64) {
			benefits["daily_checkin_points"] = checkinMin
		}

		if checkinMax, ok := cardBenefits["daily_checkin_points_max"].(int64); ok && checkinMax > benefits["daily_checkin_points_max"].(int64) {
			benefits["daily_checkin_points_max"] = checkinMax
		}

		// 布尔值取或操作
		if autoRefill, ok := cardBenefits["auto_refill_enabled"].(bool); ok && autoRefill {
			benefits["auto_refill_enabled"] = true
			if threshold, ok := cardBenefits["auto_refill_threshold"].(int64); ok {
				benefits["auto_refill_threshold"] = threshold
			}
			if amount, ok := cardBenefits["auto_refill_amount"].(int64); ok {
				benefits["auto_refill_amount"] = amount
			}
		}
	}

	return benefits, nil
}

// getCardBenefits 获取卡密的权益配置
func getCardBenefits(cardCode string, tx *gorm.DB) (map[string]interface{}, error) {
	var redemption models.RedemptionRecord
	err := tx.Where("source_id = ? AND source_type = 'activation_code'", cardCode).First(&redemption).Error
	if err != nil {
		return nil, err
	}

	benefits := map[string]interface{}{
		"daily_max_points":         redemption.DailyMaxPoints,
		"degradation_guaranteed":   redemption.DegradationGuaranteed,
		"daily_checkin_points":     redemption.DailyCheckinPoints,
		"daily_checkin_points_max": redemption.DailyCheckinPointsMax,
		"auto_refill_enabled":      redemption.AutoRefillEnabled,
		"auto_refill_threshold":    redemption.AutoRefillThreshold,
		"auto_refill_amount":       redemption.AutoRefillAmount,
	}

	return benefits, nil
}

// updateWalletBenefits 更新钱包权益
func updateWalletBenefits(wallet *models.UserWallet, benefits map[string]interface{}) {
	if val, ok := benefits["daily_max_points"].(int64); ok {
		wallet.DailyMaxPoints = val
	}
	if val, ok := benefits["degradation_guaranteed"].(int); ok {
		wallet.DegradationGuaranteed = val
	}
	if val, ok := benefits["daily_checkin_points"].(int64); ok {
		wallet.DailyCheckinPoints = val
	}
	if val, ok := benefits["daily_checkin_points_max"].(int64); ok {
		wallet.DailyCheckinPointsMax = val
	}
	if val, ok := benefits["auto_refill_enabled"].(bool); ok {
		wallet.AutoRefillEnabled = val
	}
	if val, ok := benefits["auto_refill_threshold"].(int64); ok {
		wallet.AutoRefillThreshold = val
	}
	if val, ok := benefits["auto_refill_amount"].(int64); ok {
		wallet.AutoRefillAmount = val
	}
}

// recalculateBenefitsWithUnbannedCard 重新计算包含解禁卡密的权益
func recalculateBenefitsWithUnbannedCard(userID uint, unbannedCode string, tx *gorm.DB) (map[string]interface{}, error) {
	// 1. 获取所有有效的兑换记录（包括刚解禁的）
	var allRedemptions []models.RedemptionRecord
	if err := tx.Where("user_id = ? AND source_type = 'activation_code'", userID).Find(&allRedemptions).Error; err != nil {
		return nil, err
	}

	// 2. 排除其他仍被封禁的卡密
	var otherBannedCodes []string
	if err := tx.Table("frozen_points_records").
		Where("user_id = ? AND status = 'frozen' AND banned_activation_code != ?",
			userID, unbannedCode).
		Pluck("banned_activation_code", &otherBannedCodes).Error; err != nil {
		return nil, err
	}

	// 3. 筛选出有效的兑换记录
	validRedemptions := make([]models.RedemptionRecord, 0)
	for _, record := range allRedemptions {
		// 检查是否被封禁
		isBanned := false
		for _, bannedCode := range otherBannedCodes {
			if record.SourceID == bannedCode {
				isBanned = true
				break
			}
		}
		if !isBanned {
			validRedemptions = append(validRedemptions, record)
		}
	}

	// 4. 基于有效记录计算综合权益
	return calculateCombinedBenefits(validRedemptions)
}

// calculateCombinedBenefits 计算综合权益
func calculateCombinedBenefits(redemptions []models.RedemptionRecord) (map[string]interface{}, error) {
	benefits := map[string]interface{}{
		"daily_max_points":         int64(0),
		"degradation_guaranteed":   0,
		"daily_checkin_points":     int64(0),
		"daily_checkin_points_max": int64(0),
		"auto_refill_enabled":      false,
		"auto_refill_threshold":    int64(0),
		"auto_refill_amount":       int64(0),
	}

	for _, record := range redemptions {
		// 取最大值策略
		if record.DailyMaxPoints > benefits["daily_max_points"].(int64) {
			benefits["daily_max_points"] = record.DailyMaxPoints
		}
		if record.DegradationGuaranteed > benefits["degradation_guaranteed"].(int) {
			benefits["degradation_guaranteed"] = record.DegradationGuaranteed
		}
		if record.DailyCheckinPoints > benefits["daily_checkin_points"].(int64) {
			benefits["daily_checkin_points"] = record.DailyCheckinPoints
		}
		if record.DailyCheckinPointsMax > benefits["daily_checkin_points_max"].(int64) {
			benefits["daily_checkin_points_max"] = record.DailyCheckinPointsMax
		}

		// 布尔值取或操作
		if record.AutoRefillEnabled {
			benefits["auto_refill_enabled"] = true
			benefits["auto_refill_threshold"] = record.AutoRefillThreshold
			benefits["auto_refill_amount"] = record.AutoRefillAmount
		}
	}

	return benefits, nil
}

// getBannedCodeID 获取被封禁卡密的ID
func getBannedCodeID(activationCode string, tx *gorm.DB) uint {
	var code models.ActivationCode
	if err := tx.Where("code = ?", activationCode).First(&code).Error; err != nil {
		return 0
	}
	return code.ID
}

// extractCardBenefitsJSON 提取卡密权益为JSON
func extractCardBenefitsJSON(card models.RedemptionRecord) string {
	benefits := map[string]interface{}{
		"daily_max_points":         card.DailyMaxPoints,
		"degradation_guaranteed":   card.DegradationGuaranteed,
		"daily_checkin_points":     card.DailyCheckinPoints,
		"daily_checkin_points_max": card.DailyCheckinPointsMax,
		"auto_refill_enabled":      card.AutoRefillEnabled,
		"auto_refill_threshold":    card.AutoRefillThreshold,
		"auto_refill_amount":       card.AutoRefillAmount,
	}

	data, err := json.Marshal(benefits)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// generateCalculationLog 生成计算过程日志
func generateCalculationLog(result *CardConsumptionResult) string {
	log := "虚拟消费计算结果:\n"
	log += fmt.Sprintf("目标卡密: %s\n", result.TargetCard.SourceID)
	log += fmt.Sprintf("剩余积分: %d\n", result.RemainingPoints)
	log += fmt.Sprintf("已消费积分: %d\n", result.ConsumedPoints)
	log += "消费顺序:\n"

	for i, card := range result.AllCards {
		log += fmt.Sprintf("  %d. %s: 原始%d, 消费%d, 剩余%d, 状态%s\n",
			i+1, card.CardCode, card.OriginalPoints,
			card.ConsumedPoints, card.RemainingPoints, card.Status)
	}

	return log
}
