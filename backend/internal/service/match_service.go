package service

import (
	"context"
	"errors"
	"football-backend/common/database"
	"football-backend/internal/model"
	"time"
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

// MatchCommonInfo 用于批量创建比赛的基础信息
type MatchCommonInfo struct {
	Price      float64
	MaxPlayers int
	Format     int
	Note       string
}

// MatchSchedule 比赛的时间安排表
type MatchSchedule struct {
	StartTime time.Time
	EndTime   time.Time
}

// CreateMatchBatch 开启事务批量创建比赛
func (s *MatchService) CreateMatchBatch(ctx context.Context, teamID uint, venueID uint, info MatchCommonInfo, schedules []MatchSchedule) ([]model.Match, error) {
	if len(schedules) == 0 {
		return nil, errors.New("schedules cannot be empty")
	}

	var createdMatches []model.Match

	err := s.repo.Transaction(ctx, func(txRepo database.Repository) error {
		for _, sch := range schedules {
			if sch.EndTime.Before(sch.StartTime) {
				return errors.New("end time cannot be before start time")
			}

			newMatch := model.Match{
				TeamID:     teamID,
				VenueID:    venueID,
				Price:      info.Price,
				MaxPlayers: info.MaxPlayers,
				Format:     info.Format,
				Note:       info.Note,
				StartTime:  sch.StartTime,
				EndTime:    sch.EndTime,
				Status:     "RECRUITING",
			}

			if err := txRepo.Create(ctx, &newMatch); err != nil {
				return err // Any failure will automatically rollback the entire transaction
			}

			createdMatches = append(createdMatches, newMatch)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdMatches, nil
}
