package model

import (
	"time"

	"gorm.io/gorm"
)

// Booking 报名模型
type Booking struct {
	ID      uint   `gorm:"primaryKey" json:"id"`                                 // 报名记录的唯一主键
	MatchID uint   `gorm:"index:idx_bookings_match_id;not null" json:"match_id"` // 关联的比赛 ID (外键)
	Match   *Match `gorm:"foreignKey:MatchID" json:"match,omitempty"`            // 预加载的比赛结构体引用
	UserID  uint   `gorm:"index:idx_bookings_user_id;not null" json:"user_id"`   // 报名的用户 ID (外键)
	User    *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`              // 预加载的用户结构体引用

	GuestName     string `gorm:"size:50;default:''" json:"guest_name"`           // 帮朋友代报名时填写的代称 (如果为空则代表是本人报名)
	Status        string `gorm:"size:20;default:'CONFIRMED'" json:"status"`      // 报名状态值枚举：CONFIRMED(已确认报名), WAITING(排队候补中), CANCELED(已取消)
	PaymentStatus string `gorm:"size:20;default:'UNPAID'" json:"payment_status"` // 赛后记账及支付状态：UNPAID(未付款), PAID(已付款), REFUNDED(已退款)
	SubTeam       string `gorm:"size:10;default:''" json:"sub_team"`             // 在比赛时分配的分队代号，如 "A队", "B队"

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"` // 报名创建的 UTC+9 时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                              // 取消/物理删除时的软删时间戳哨兵
}

func (Booking) TableName() string {
	return "bookings"
}
