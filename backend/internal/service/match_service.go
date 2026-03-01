package service

import (
	"context"
	"errors"
	"football-backend/common/database"
	"football-backend/internal/model"
)

type MatchService struct {
	repo database.Repository
}

// NewMatchService constructor 依赖注入 Repository
func NewMatchService(repo database.Repository) *MatchService {
	return &MatchService{repo: repo}
}

// JoinMatch 处理报名逻辑
// 业务流程（图纸）放置在 Service，调用 Repo（工具库）的基础方法
func (s *MatchService) JoinMatch(ctx context.Context, matchID uint, userID uint) error {
	return s.repo.Transaction(ctx, func(txRepo database.Repository) error {
		// 1. 锁定比赛记录
		match, err := txRepo.GetMatchWithLock(ctx, matchID)
		if err != nil {
			return err
		}
		if match.Status != "RECRUITING" {
			return errors.New("match is not open for recruiting")
		}

		// 2. 幂等性检查
		hasBooked, err := txRepo.HasUserBooked(ctx, matchID, userID)
		if err != nil {
			return err
		}
		if hasBooked {
			return errors.New("user already joined this match")
		}

		// 3. 检查容量
		playersCount, err := txRepo.CountConfirmedPlayers(ctx, matchID)
		if err != nil {
			return err
		}
		if int(playersCount) >= match.MaxPlayers {
			return errors.New("match is full")
		}

		// 4. 创建记录
		booking := model.Booking{
			MatchID:       matchID,
			UserID:        userID,
			Status:        "CONFIRMED",
			PaymentStatus: "UNPAID",
		}
		return txRepo.Create(ctx, &booking)
	})
}
