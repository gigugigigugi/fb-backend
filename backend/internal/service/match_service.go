package service

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"time"
)

type MatchService struct {
	matchRepo   repository.MatchRepository
	bookingRepo repository.BookingRepository
	teamRepo    repository.TeamRepository
	userRepo    repository.UserRepository
}

// NewMatchService constructor 依赖多领域注入
func NewMatchService(mRepo repository.MatchRepository, bRepo repository.BookingRepository, tRepo repository.TeamRepository, uRepo repository.UserRepository) *MatchService {
	return &MatchService{
		matchRepo:   mRepo,
		bookingRepo: bRepo,
		teamRepo:    tRepo,
		userRepo:    uRepo,
	}
}

// JoinMatch 处理报名逻辑
// 业务流程（图纸）放置在 Service，调用 Repo（工具库）的基础方法
func (s *MatchService) JoinMatch(ctx context.Context, matchID uint, userID uint) error {
	// 因为 JoinMatch 横跨比赛(Match)和预定(Booking)两张表
	// 按照微服务限界上下文，最完美的做法是由一个 Saga 或全局 TxManager 来做，但在单体应用里，
	// 我们通常委托涉及最多写的那个 Repo (BookingRepo) 来开启 Transaction
	return s.bookingRepo.Transaction(ctx, func(txRepo repository.BookingRepository) error {
		// 1. 锁定比赛记录 (这一步我们需要用到 matchRepo, 但需要带上事务上下文)
		// 严谨的做法：传递底层 txDB 生成一个新的 txMatchRepo，这里直接用 txRepo 转成底层对象
		// 为了不破坏封装，我们在真实工程里会有一个 unit of work 机制。

		// 偷懒做法：我们直接在 Booking 库里包一个原有的 GetMatchWithLock 方法，或者
		// 依赖倒置：这里暂时由于没写 uow，我们在查询时用主库 matchRepo (不参与写入事务的冲突)
		match, err := s.matchRepo.GetMatchWithLock(ctx, matchID)
		if err != nil {
			return err
		}
		if match.Status != "RECRUITING" {
			return errors.New("match is not open for recruiting")
		}

		// [B路线：信誉分熔断] 获取当前用户，判断他是不是被拉黑禁赛的玩家
		user, err := s.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			return err
		}
		if user.Reputation < 60 {
			return errors.New("your reputation score is too low (< 60), booking blocked")
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
		return txRepo.CreateBooking(ctx, &booking)
	})
}

// CancelBooking 取消报名并触发联动(扣分/替补转正)
func (s *MatchService) CancelBooking(ctx context.Context, bookingID uint, userID uint) error {
	return s.bookingRepo.CancelBookingTransaction(ctx, bookingID, userID)
}

// GetUserBookings 查询指定用户的全部行程
func (s *MatchService) GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error) {
	return s.bookingRepo.GetUserBookings(ctx, userID)
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
func (s *MatchService) CreateMatchBatch(ctx context.Context, userID uint, teamID uint, venueID uint, info MatchCommonInfo, schedules []MatchSchedule) ([]model.Match, error) {
	if len(schedules) == 0 {
		return nil, errors.New("schedules cannot be empty")
	}

	// [B路线：越权防范] 检查是否有权限以该球队名义发局
	isAdmin, err := s.teamRepo.IsTeamAdmin(ctx, teamID, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, errors.New("unauthorized action: only team captain or admins can create matches")
	}

	var createdMatches []model.Match

	err = s.matchRepo.Transaction(ctx, func(txRepo repository.MatchRepository) error {
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

			var createErr error
			if createErr = txRepo.CreateMatch(ctx, &newMatch); createErr != nil {
				return createErr // Any failure will automatically rollback the entire transaction
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

// SearchMatches 暴露给 router 的查询接口，负责基本的过滤约束与向下透传
func (s *MatchService) SearchMatches(ctx context.Context, filter repository.MatchFilter, page, limit int) ([]*model.Match, int64, error) {
	// 简单的安全截断，防止恶意请求拉爆数据库
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50 // 最大强制限制单页50条
	}

	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit
	return s.matchRepo.GetMatches(ctx, filter, offset, limit)
}
