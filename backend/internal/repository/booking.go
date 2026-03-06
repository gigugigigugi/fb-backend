package repository

import (
	"context"
	"football-backend/internal/model"
)

// BookingRepository 定义报名与候补相关数据访问接口。
type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *model.Booking) error
	HasUserBooked(ctx context.Context, matchID uint, userID uint) (bool, error)
	CountConfirmedPlayers(ctx context.Context, matchID uint) (int64, error)
	CountWaitingPlayers(ctx context.Context, matchID uint) (int64, error)
	GetBookingsByMatchID(ctx context.Context, matchID uint) ([]*model.Booking, error)
	GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error)
	// CancelBookingTransaction 返回：
	// 1) matchID：供上层构造通知内容。
	// 2) []uint：当前比赛 WAITING 用户 ID 列表（仅用于通知，不做自动转正）。
	CancelBookingTransaction(ctx context.Context, bookingID uint, userID uint) (uint, []uint, error)
	Transaction(ctx context.Context, fn func(txRepo BookingRepository) error) error
}
