package postgres

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type bookingRepository struct {
	db *gorm.DB
}

// NewBookingRepository 创建 BookingRepository 的 PostgreSQL 实现。
func NewBookingRepository(db *gorm.DB) repository.BookingRepository {
	return &bookingRepository{db: db}
}

// Transaction 在事务中执行报名仓储操作。
func (r *bookingRepository) Transaction(ctx context.Context, fn func(txRepo repository.BookingRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewBookingRepository(tx)
		return fn(txRepo)
	})
}

// CreateBooking 创建报名记录。
func (r *bookingRepository) CreateBooking(ctx context.Context, booking *model.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

// HasUserBooked 判断用户是否已报名（排除已取消记录）。
func (r *bookingRepository) HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error) {
	var existingCount int64
	err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("match_id = ? AND user_id = ? AND status != ?", matchID, userID, "CANCELED").
		Count(&existingCount).Error
	return existingCount > 0, err
}

// CountConfirmedPlayers 统计已确认人数。
func (r *bookingRepository) CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error) {
	var currentPlayers int64
	err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("match_id = ? AND status = ?", matchID, "CONFIRMED").
		Count(&currentPlayers).Error
	return currentPlayers, err
}

// CountWaitingPlayers 统计候补人数。
func (r *bookingRepository) CountWaitingPlayers(ctx context.Context, matchID uint) (int64, error) {
	var waitingPlayers int64
	err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("match_id = ? AND status = ?", matchID, "WAITING").
		Count(&waitingPlayers).Error
	return waitingPlayers, err
}

// GetBookingsByMatchID 查询比赛下的全部报名（含用户信息）。
func (r *bookingRepository) GetBookingsByMatchID(ctx context.Context, matchID uint) ([]*model.Booking, error) {
	var bookings []*model.Booking
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("match_id = ?", matchID).
		Order("created_at ASC").
		Find(&bookings).Error
	return bookings, err
}

// GetUserBookings 查询用户的报名历史（含比赛、球队、场地）。
func (r *bookingRepository) GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error) {
	var bookings []*model.Booking
	err := r.db.WithContext(ctx).
		Preload("Match").
		Preload("Match.Team").
		Preload("Match.Venue").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bookings).Error
	return bookings, err
}

// GetBookingWithLock 按 ID 查询报名并加行级锁。
func (r *bookingRepository) GetBookingWithLock(ctx context.Context, bookingID uint) (*model.Booking, error) {
	var booking model.Booking
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Match").First(&booking, bookingID).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

// CancelBookingTransaction 在事务内取消报名，并返回需要通知的候补用户列表。
func (r *bookingRepository) CancelBookingTransaction(ctx context.Context, bookingID uint, userID uint) (uint, []uint, error) {
	var matchID uint
	var waitingBookingUserIDs []uint

	err := r.Transaction(ctx, func(txRepo repository.BookingRepository) error {
		internalRepo := txRepo.(*bookingRepository)
		txDB := internalRepo.db.WithContext(ctx)

		booking, err := internalRepo.GetBookingWithLock(ctx, bookingID)
		if err != nil {
			return err
		}
		matchID = booking.MatchID

		if booking.UserID != userID && booking.Status != "CANCELED" {
			return errors.New("unauthorized or invalid booking record")
		}
		if booking.Status == "CANCELED" {
			return errors.New("booking already canceled")
		}

		// 赛前 24 小时内取消会扣信誉分。
		timeLeft := time.Until(booking.Match.StartTime)
		if timeLeft > 0 && timeLeft < 24*time.Hour {
			if err := txDB.Model(&model.User{}).Where("id = ?", userID).
				UpdateColumn("reputation", gorm.Expr("reputation - ?", 20)).Error; err != nil {
				return err
			}
		}

		oldStatus := booking.Status
		if err := txDB.Model(booking).Update("status", "CANCELED").Error; err != nil {
			return err
		}

		// 当前策略：取消确认席位后仅通知 WAITING 用户，不在事务中执行自动转正。
		if oldStatus == "CONFIRMED" && timeLeft > 0 {
			const waitlistNotifyLimit = 10
			if err := txDB.Model(&model.Booking{}).
				Where("match_id = ? AND status = ?", booking.MatchID, "WAITING").
				Order("created_at ASC").
				Limit(waitlistNotifyLimit).
				Pluck("user_id", &waitingBookingUserIDs).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return matchID, waitingBookingUserIDs, nil
}
