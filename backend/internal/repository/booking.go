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
	CountWaitingPlayers(ctx context.Context, matchID uint) (int64, error)
	GetUserBookings(ctx context.Context, userID uint) ([]*model.Booking, error)
	// CancelBookingTransaction returns:
	// 1) matchID: 用于上层构造通知消息上下文
	// 2) []uint: 当前比赛下所有 WAITING 用户ID（用于通知，不做自动转正）
	CancelBookingTransaction(ctx context.Context, bookingID uint, userID uint) (uint, []uint, error)
	Transaction(ctx context.Context, fn func(txRepo BookingRepository) error) error
}
