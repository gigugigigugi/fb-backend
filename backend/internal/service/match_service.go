package service

import (
	"context"
	"errors"
	"fmt"
	"football-backend/common/notification"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"log/slog"
	"strings"
	"time"
)

const maxWaitlistSize = 10 // 候补队列容量上限。

var (
	// ErrMatchNotFound 表示比赛不存在。
	ErrMatchNotFound = errors.New("match not found")
	// ErrMatchManageForbidden 表示当前用户没有比赛赛后管理权限。
	ErrMatchManageForbidden = errors.New("forbidden: only team captain or admins can manage post-match actions")
	// ErrInvalidPaymentStatus 表示 settlement 的 payment_status 参数非法。
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
	// ErrInvalidSubTeamAssignments 表示 subteams 参数非法。
	ErrInvalidSubTeamAssignments = errors.New("invalid subteam assignments")
)

// MatchService 负责比赛报名、取消、批量建赛与详情聚合。
type MatchService struct {
	matchRepo   repository.MatchRepository
	bookingRepo repository.BookingRepository
	teamRepo    repository.TeamRepository
	userRepo    repository.UserRepository
	notifier    *notification.Dispatcher
}

// MatchDetailRosterItem 表示详情页中的单条报名成员信息。
type MatchDetailRosterItem struct {
	BookingID uint   `json:"booking_id"` // 报名记录 ID。
	UserID    uint   `json:"user_id"`    // 用户 ID。
	Nickname  string `json:"nickname"`   // 用户昵称。
	Avatar    string `json:"avatar"`     // 用户头像。
	GuestName string `json:"guest_name"` // 来宾名称（代报名场景）。
	Status    string `json:"status"`     // 报名状态。
}

// MatchDetailCommentItem 表示详情页中的单条评论。
type MatchDetailCommentItem struct {
	ID        uint      `json:"id"`         // 评论 ID。
	UserID    uint      `json:"user_id"`    // 评论作者 ID。
	Nickname  string    `json:"nickname"`   // 评论作者昵称。
	Avatar    string    `json:"avatar"`     // 评论作者头像。
	Content   string    `json:"content"`    // 评论内容。
	CreatedAt time.Time `json:"created_at"` // 评论创建时间。
}

// MatchDetailRoster 表示详情页的报名阵容分组。
type MatchDetailRoster struct {
	Confirmed []MatchDetailRosterItem `json:"confirmed"` // 已确认名单。
	Waiting   []MatchDetailRosterItem `json:"waiting"`   // 候补名单。
}

// MatchDetailResponse 是比赛详情接口统一返回结构。
type MatchDetailResponse struct {
	MatchInfo  *model.Match             `json:"match_info"`  // 比赛基础信息。
	Roster     MatchDetailRoster        `json:"roster"`      // 报名阵容。
	Comments   []MatchDetailCommentItem `json:"comments"`    // 最新评论列表。
	UserStatus string                   `json:"user_status"` // 当前用户状态：NOT_JOINED/JOINED/WAITING/CANCELED。
}

// MatchCommonInfo 表示批量建赛时共用的固定字段。
type MatchCommonInfo struct {
	Price      float64 // 单人费用。
	MaxPlayers int     // 最大人数。
	Format     int     // 赛制。
	Note       string  // 备注。
}

// MatchSchedule 表示一场比赛的时间安排。
type MatchSchedule struct {
	StartTime time.Time // 开始时间。
	EndTime   time.Time // 结束时间。
}

// SubTeamAssignment 表示一条分队结果输入。
type SubTeamAssignment struct {
	BookingID uint   `json:"booking_id"` // 报名记录 ID。
	SubTeam   string `json:"sub_team"`   // 分队标识，例如 A/B。
}

// NewMatchService 创建比赛服务实例。
func NewMatchService(mRepo repository.MatchRepository, bRepo repository.BookingRepository, tRepo repository.TeamRepository, uRepo repository.UserRepository, notifier *notification.Dispatcher) *MatchService {
	return &MatchService{
		matchRepo:   mRepo,
		bookingRepo: bRepo,
		teamRepo:    tRepo,
		userRepo:    uRepo,
		notifier:    notifier,
	}
}

// JoinMatch 处理用户报名逻辑。
func (s *MatchService) JoinMatch(ctx context.Context, matchID uint, userID uint) error {
	// 报名流程需要同时读写比赛与报名记录，因此在事务里执行关键步骤。
	return s.bookingRepo.Transaction(ctx, func(txRepo repository.BookingRepository) error {
		match, err := s.matchRepo.GetMatchWithLock(ctx, matchID)
		if err != nil {
			return err
		}
		if match.Status != "RECRUITING" {
			return errors.New("match is not open for recruiting")
		}

		// 信用分低于阈值时禁止报名。
		user, err := s.userRepo.GetUserByID(ctx, userID)
		if err != nil {
			return err
		}
		if user.Reputation < 60 {
			return errors.New("your reputation score is too low (< 60), booking blocked")
		}

		// 幂等校验：同一用户不可重复报名同一比赛。
		hasBooked, err := txRepo.HasUserBooked(ctx, matchID, userID)
		if err != nil {
			return err
		}
		if hasBooked {
			return errors.New("user already joined this match")
		}

		playersCount, err := txRepo.CountConfirmedPlayers(ctx, matchID)
		if err != nil {
			return err
		}

		booking := model.Booking{
			MatchID:       matchID,
			UserID:        userID,
			Status:        "CONFIRMED",
			PaymentStatus: "UNPAID",
		}

		// 满员后进入 WAITING 队列，并限制候补最多 10 人。
		if int(playersCount) >= match.MaxPlayers {
			waitingCount, err := txRepo.CountWaitingPlayers(ctx, matchID)
			if err != nil {
				return err
			}
			if waitingCount >= maxWaitlistSize {
				return fmt.Errorf("waitlist is full (max %d)", maxWaitlistSize)
			}
			booking.Status = "WAITING"
		}

		return txRepo.CreateBooking(ctx, &booking)
	})
}

// CancelBooking 取消报名，并通知候补队列用户（不执行自动转正）。
func (s *MatchService) CancelBooking(ctx context.Context, bookingID uint, userID uint) error {
	// 事务内仅负责状态变更与候补名单提取；通知作为事务外副作用执行。
	matchID, waitingUserIDs, err := s.bookingRepo.CancelBookingTransaction(ctx, bookingID, userID)
	if err != nil {
		return err
	}

	// 防御性限制：即使仓储层返回超量数据，通知层最多处理前 10 位候补用户。
	if len(waitingUserIDs) > maxWaitlistSize {
		waitingUserIDs = waitingUserIDs[:maxWaitlistSize]
	}
	if s.notifier == nil || len(waitingUserIDs) == 0 {
		return nil
	}

	msg := notification.Message{
		Subject: "候补可报名提醒",
		Body:    fmt.Sprintf("比赛 %d 出现空位，请尽快重新报名。", matchID),
	}

	for _, waitingUserID := range waitingUserIDs {
		user, userErr := s.userRepo.GetUserByID(ctx, waitingUserID)
		if userErr != nil {
			slog.Warn("Skip waitlist notification because user lookup failed",
				slog.Uint64("match_id", uint64(matchID)),
				slog.Uint64("user_id", uint64(waitingUserID)),
				slog.String("error", userErr.Error()),
			)
			continue
		}

		channels := make([]notification.Channel, 0, 2)
		recipient := notification.Recipient{UserID: user.ID}

		// 仅使用已验证联系方式，防止误发。
		if user.EmailVerified && user.Email != "" {
			channels = append(channels, notification.ChannelEmail)
			recipient.Email = user.Email
		}
		if user.PhoneVerified && user.Phone != nil && *user.Phone != "" {
			channels = append(channels, notification.ChannelSMS)
			recipient.Phone = *user.Phone
		}
		if len(channels) == 0 {
			slog.Warn("Skip waitlist notification because user has no supported contact",
				slog.Uint64("match_id", uint64(matchID)),
				slog.Uint64("user_id", uint64(waitingUserID)),
			)
			continue
		}

		task := notification.Task{
			MatchID:   matchID,
			Recipient: recipient,
			Message:   msg,
			Channels:  channels,
		}

		// 通知失败不回滚主业务，只记录日志便于后续追踪。
		if enqueueErr := s.notifier.Enqueue(task); enqueueErr != nil {
			slog.Error("Failed to enqueue waitlist notification",
				slog.Uint64("match_id", uint64(matchID)),
				slog.Uint64("user_id", uint64(waitingUserID)),
				slog.String("error", enqueueErr.Error()),
			)
		}
	}

	return nil
}

// GetUserBookings 查询当前用户的报名记录。
func (s *MatchService) GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error) {
	return s.bookingRepo.GetUserBookings(ctx, userID)
}

// CreateMatchBatch 批量创建比赛。
func (s *MatchService) CreateMatchBatch(ctx context.Context, userID uint, teamID uint, venueID uint, info MatchCommonInfo, schedules []MatchSchedule) ([]model.Match, error) {
	if len(schedules) == 0 {
		return nil, errors.New("schedules cannot be empty")
	}

	// 仅队长/管理员允许以球队名义建赛。
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

			if createErr := txRepo.CreateMatch(ctx, &newMatch); createErr != nil {
				return createErr
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

// SettleMatch 处理赛后结算，批量更新 payment_status。
func (s *MatchService) SettleMatch(ctx context.Context, matchID uint, userID uint, paymentStatus string, bookingIDs []uint) (int64, error) {
	status := strings.ToUpper(strings.TrimSpace(paymentStatus))
	switch status {
	case "UNPAID", "PAID", "REFUNDED":
	default:
		return 0, ErrInvalidPaymentStatus
	}

	_, err := s.ensurePostMatchPermission(ctx, matchID, userID)
	if err != nil {
		return 0, err
	}

	uniqueBookingIDs := make([]uint, 0, len(bookingIDs))
	seen := make(map[uint]struct{}, len(bookingIDs))
	for _, id := range bookingIDs {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		uniqueBookingIDs = append(uniqueBookingIDs, id)
	}

	return s.bookingRepo.SettleMatchBookings(ctx, matchID, status, uniqueBookingIDs)
}

// AssignMatchSubTeams 保存赛后分队结果。
func (s *MatchService) AssignMatchSubTeams(ctx context.Context, matchID uint, userID uint, assignments []SubTeamAssignment) error {
	_, err := s.ensurePostMatchPermission(ctx, matchID, userID)
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		return ErrInvalidSubTeamAssignments
	}

	// 防止同一 booking 在一次请求中被重复分配，避免歧义。
	seen := make(map[uint]struct{}, len(assignments))
	repoAssignments := make([]repository.SubTeamAssignment, 0, len(assignments))
	for _, assignment := range assignments {
		subTeam := strings.TrimSpace(assignment.SubTeam)
		if assignment.BookingID == 0 || subTeam == "" || len(subTeam) > 10 {
			return ErrInvalidSubTeamAssignments
		}
		if _, exists := seen[assignment.BookingID]; exists {
			return ErrInvalidSubTeamAssignments
		}
		seen[assignment.BookingID] = struct{}{}

		repoAssignments = append(repoAssignments, repository.SubTeamAssignment{
			BookingID: assignment.BookingID,
			SubTeam:   subTeam,
		})
	}

	return s.bookingRepo.AssignSubTeams(ctx, matchID, repoAssignments)
}

// SearchMatches 按过滤条件分页查询比赛列表。
func (s *MatchService) SearchMatches(ctx context.Context, filter repository.MatchFilter, page, limit int) ([]*model.Match, int64, error) {
	// 分页参数做保护，避免异常请求导致大批量扫描。
	if limit <= 0 {
		limit = 10
	} else if limit > 50 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit
	return s.matchRepo.GetMatches(ctx, filter, offset, limit)
}

// GetMatchDetails 聚合比赛详情（基础信息、阵容、评论和当前用户状态）。
func (s *MatchService) GetMatchDetails(ctx context.Context, matchID uint, userID uint) (*MatchDetailResponse, error) {
	match, err := s.matchRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	bookings, err := s.bookingRepo.GetBookingsByMatchID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	comments, err := s.matchRepo.GetCommentsByMatchID(ctx, matchID, 50)
	if err != nil {
		return nil, err
	}

	resp := &MatchDetailResponse{
		MatchInfo: match,
		Roster: MatchDetailRoster{
			Confirmed: make([]MatchDetailRosterItem, 0),
			Waiting:   make([]MatchDetailRosterItem, 0),
		},
		Comments:   make([]MatchDetailCommentItem, 0),
		UserStatus: "NOT_JOINED",
	}

	for _, b := range bookings {
		item := MatchDetailRosterItem{
			BookingID: b.ID,
			UserID:    b.UserID,
			GuestName: b.GuestName,
			Status:    b.Status,
		}
		if b.User != nil {
			item.Nickname = b.User.Nickname
			item.Avatar = b.User.Avatar
		}

		switch b.Status {
		case "CONFIRMED":
			resp.Roster.Confirmed = append(resp.Roster.Confirmed, item)
		case "WAITING":
			resp.Roster.Waiting = append(resp.Roster.Waiting, item)
		}

		// 根据当前用户在报名列表中的状态推导 user_status。
		if b.UserID == userID {
			switch b.Status {
			case "CONFIRMED":
				resp.UserStatus = "JOINED"
			case "WAITING":
				if resp.UserStatus != "JOINED" {
					resp.UserStatus = "WAITING"
				}
			case "CANCELED":
				if resp.UserStatus == "NOT_JOINED" {
					resp.UserStatus = "CANCELED"
				}
			}
		}
	}

	for _, c := range comments {
		item := MatchDetailCommentItem{
			ID:        c.ID,
			UserID:    c.UserID,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		}
		if c.User != nil {
			item.Nickname = c.User.Nickname
			item.Avatar = c.User.Avatar
		}
		resp.Comments = append(resp.Comments, item)
	}

	return resp, nil
}

// ensurePostMatchPermission 校验赛后管理权限：仅队长/管理员可操作。
func (s *MatchService) ensurePostMatchPermission(ctx context.Context, matchID uint, userID uint) (*model.Match, error) {
	match, err := s.matchRepo.GetMatchByID(ctx, matchID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "match not found") {
			return nil, ErrMatchNotFound
		}
		return nil, err
	}

	isAdmin, err := s.teamRepo.IsTeamAdmin(ctx, match.TeamID, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrMatchManageForbidden
	}

	return match, nil
}
