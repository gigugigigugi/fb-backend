package model

import "time"

// Booking 报名模型
type Booking struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	MatchID uint   `gorm:"index:idx_bookings_match_id;not null" json:"match_id"`
	Match   *Match `gorm:"foreignKey:MatchID" json:"match,omitempty"` // 关联比赛
	UserID  uint   `gorm:"index:idx_bookings_user_id;not null" json:"user_id"`
	User    *User  `gorm:"foreignKey:UserID" json:"user,omitempty"` // 关联用户

	GuestName     string `gorm:"size:50;default:''" json:"guest_name"`      // 代报名人名
	Status        string `gorm:"size:20;default:'CONFIRMED'" json:"status"` // CONFIRMED, WAITING, CANCELED
	PaymentStatus string `gorm:"size:20;default:'UNPAID'" json:"payment_status"`
	SubTeam       string `gorm:"size:10;default:''" json:"sub_team"` // 分队: A, B

	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Booking) TableName() string {
	return "bookings"
}
