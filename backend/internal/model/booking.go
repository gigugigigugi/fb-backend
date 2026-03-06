package model

import (
	"time"

	"gorm.io/gorm"
)

// Booking 表示用户对某场比赛的一条报名记录。
type Booking struct {
	ID      uint   `gorm:"primaryKey" json:"id"`                                 // 报名记录主键 ID。
	MatchID uint   `gorm:"index:idx_bookings_match_id;not null" json:"match_id"` // 关联的比赛 ID（外键）。
	Match   *Match `gorm:"foreignKey:MatchID" json:"match,omitempty"`            // 预加载时挂载的比赛对象。
	UserID  uint   `gorm:"index:idx_bookings_user_id;not null" json:"user_id"`   // 报名用户 ID（外键）。
	User    *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`              // 预加载时挂载的用户对象。

	GuestName     string `gorm:"size:50;default:''" json:"guest_name"`           // 代报名时填写的来宾名称；为空表示本人报名。
	Status        string `gorm:"size:20;default:'CONFIRMED'" json:"status"`      // 报名状态：CONFIRMED / WAITING / CANCELED。
	PaymentStatus string `gorm:"size:20;default:'UNPAID'" json:"payment_status"` // 费用状态：UNPAID / PAID / REFUNDED。
	SubTeam       string `gorm:"size:10;default:''" json:"sub_team"`             // 比赛分组标识（如 A/B 队）。

	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"` // 记录创建时间。
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                              // 软删除时间（为空表示未删除）。
}

// TableName 指定 Booking 对应的数据表名。
func (Booking) TableName() string {
	return "bookings"
}
