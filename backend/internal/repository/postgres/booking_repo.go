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

func NewBookingRepository(db *gorm.DB) repository.BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Transaction(ctx context.Context, fn func(txRepo repository.BookingRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewBookingRepository(tx)
		return fn(txRepo)
	})
}

func (r *bookingRepository) CreateBooking(ctx context.Context, booking *model.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

func (r *bookingRepository) HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error) {
	var existingCount int64
	err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("match_id = ? AND user_id = ? AND status != ?", matchID, userID, "CANCELED").
		Count(&existingCount).Error
	return existingCount > 0, err
}

func (r *bookingRepository) CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error) {
	var currentPlayers int64
	err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("match_id = ? AND status = ?", matchID, "CONFIRMED").
		Count(&currentPlayers).Error
	return currentPlayers, err
}

func (r *bookingRepository) GetBookingWithLock(ctx context.Context, bookingID uint) (*model.Booking, error) {
	var booking model.Booking
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Match").First(&booking, bookingID).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) CancelBookingTransaction(ctx context.Context, bookingID uint, userID uint) error {
	return r.Transaction(ctx, func(txRepo repository.BookingRepository) error {
		internalRepo := txRepo.(*bookingRepository)
		txDB := internalRepo.db.WithContext(ctx)

		booking, err := internalRepo.GetBookingWithLock(ctx, bookingID)
		if err != nil {
			return err
		}
		if booking.UserID != userID && booking.Status != "CANCELED" {
			return errors.New("unauthorized or invalid booking record")
		}

		if booking.Status == "CANCELED" {
			return errors.New("booking already canceled")
		}

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

		if oldStatus == "CONFIRMED" && timeLeft > 0 {
			var waitingBookingUserIDs []uint
			err := txDB.Model(&model.Booking{}).
				Where("match_id = ? AND status = ?", booking.MatchID, "WAITING").
				Pluck("user_id", &waitingBookingUserIDs).Error

			if err == nil && len(waitingBookingUserIDs) > 0 {
				// 发送 MQ
			}
		}

		return nil
	})
}
