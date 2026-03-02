package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 留言模型
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`                                 // 留言的唯一标识主键
	MatchID   uint           `gorm:"index:idx_comments_match_id;not null" json:"match_id"` // 留言所属的比赛大厅 ID
	UserID    uint           `gorm:"not null" json:"user_id"`                              // 发表留言的用户 ID
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`              // 用于 ORM 预加载时挂载发送者的关联用户信息
	Content   string         `gorm:"size:500" json:"content"`                              // 留言的具体文字内容
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`          // 留言产生的时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                                       // 软删除追踪标识字段
}

func (Comment) TableName() string {
	return "comments"
}
