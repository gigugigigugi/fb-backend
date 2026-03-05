package repository

import (
	"context"
	"football-backend/internal/model"
)

// BookingRepository 定义报名、占位与队列的数据访问接口
type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *model.Booking) error
	HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error)
	CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error)
	GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error)
	CancelBookingTransaction(ctx context.Context, bookingID uint, userID uint) error
	Transaction(ctx context.Context, fn func(txRepo BookingRepository) error) error
}
