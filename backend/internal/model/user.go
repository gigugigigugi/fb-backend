package model

import (
	"time"

	"gorm.io/datatypes"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Mobile       string         `gorm:"size:20;unique;not null" json:"mobile"`
	PasswordHash string         `gorm:"size:100" json:"-"` // JSON tag "-" 表示不序列化此字段
	Nickname     string         `gorm:"size:50" json:"nickname"`
	Avatar       string         `gorm:"size:255" json:"avatar"`
	Reputation   int            `gorm:"default:100" json:"reputation"`
	Stats        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"stats"` // 游戏化统计: {"mvp": 5, "badges": []}
	CreatedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定数据库表名
func (User) TableName() string {
	return "users"
}
