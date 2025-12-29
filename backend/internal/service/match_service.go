package service

import (
	"errors"
	"football-backend/common/database"
	"football-backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause" // 必须导入这个包以支持锁
)

type MatchService struct{}

// JoinMatch 处理报名逻辑
// 包含：悲观锁防超卖、状态检查、幂等性检查
func (s *MatchService) JoinMatch(matchID uint, userID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var match model.Match

		// 1. 锁定比赛记录 (Pessimistic Locking)
		// 使用 FOR UPDATE 锁住该行，防止其他事务同时修改或读取旧数据导致超卖
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&match, matchID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("match not found")
			}
			return err
		}

		// 2. 检查比赛状态
		if match.Status != "RECRUITING" {
			return errors.New("match is not open for recruiting")
		}

		// 3. 幂等性检查 (Idempotency Check)
		// 防止同一个用户重复报名 (手抖点两次或网络重试)
		var existingCount int64
		tx.Model(&model.Booking{}).
			Where("match_id = ? AND user_id = ? AND status != ?", matchID, userID, "CANCELED").
			Count(&existingCount)

		if existingCount > 0 {
			return errors.New("user already joined this match")
		}

		// 4. 检查人数是否已满 (Capacity Check)
		var currentPlayers int64
		// 只统计状态为 CONFIRMED 的报名人数
		tx.Model(&model.Booking{}).
			Where("match_id = ? AND status = ?", matchID, "CONFIRMED").
			Count(&currentPlayers)

		if int(currentPlayers) >= match.MaxPlayers {
			return errors.New("match is full")
		}

		// 5. 创建报名记录
		booking := model.Booking{
			MatchID:       matchID,
			UserID:        userID,
			Status:        "CONFIRMED",
			PaymentStatus: "UNPAID", // 默认为未支付
		}

		if err := tx.Create(&booking).Error; err != nil {
			return err
		}

		return nil
	})
}
