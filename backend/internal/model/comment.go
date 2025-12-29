package model

import "time"

// Comment 留言模型
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MatchID   uint      `gorm:"index:idx_comments_match_id;not null" json:"match_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"` // 预加载用户信息
	Content   string    `gorm:"size:500" json:"content"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Comment) TableName() string {
	return "comments"
}
