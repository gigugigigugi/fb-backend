package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment 表示用户在比赛下发布的一条评论。
type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`                                 // 评论主键 ID。
	MatchID   uint           `gorm:"index:idx_comments_match_id;not null" json:"match_id"` // 所属比赛 ID。
	UserID    uint           `gorm:"not null" json:"user_id"`                              // 发布评论的用户 ID。
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`              // 预加载时挂载的评论作者信息。
	Content   string         `gorm:"size:500" json:"content"`                              // 评论正文内容。
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`          // 评论创建时间。
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`                                       // 软删除时间（为空表示未删除）。
}

// TableName 指定 Comment 对应的数据表名。
func (Comment) TableName() string {
	return "comments"
}
