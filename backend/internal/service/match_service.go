package service

import (
	"errors"
	"football-backend/common/database"
	"football-backend/internal/model"

	"gorm.io/gorm"
)

type MatchService struct{}

// JoinMatch 处理报名逻辑
func (s *MatchService) JoinMatch(matchID uint, userID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 检查比赛是否存在
		var match model.Match
		if err := tx.First(&match, matchID).Error; err != nil {
			return err
		}

		// 2. 检查人数是否已满 (核心逻辑)
		var count int64
		tx.Model(&model.Booking{}).Where("match_id = ? AND status = ?", matchID, "CONFIRMED").Count(&count)
		if int(count) >= match.MaxPlayers {
			return errors.New("match is full")
		}

		// 3. 创建报名记录
		booking := model.Booking{
			MatchID: matchID,
			UserID:  userID,
			Status:  "CONFIRMED",
		}
		return tx.Create(&booking).Error
	})
}
