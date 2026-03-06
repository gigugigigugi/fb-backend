package model

import (
	"time"

	"gorm.io/gorm"
)

// Match 表示一场可报名的足球比赛。
type Match struct {
	ID         uint           `gorm:"primaryKey" json:"id"`                                    // 比赛主键 ID。
	TeamID     uint           `gorm:"index:idx_matches_team_id" json:"team_id"`                // 发起比赛的球队 ID。
	Team       *Team          `gorm:"foreignKey:TeamID" json:"team,omitempty"`                 // 预加载时挂载的球队对象。
	VenueID    uint           `gorm:"index:idx_matches_venue_id" json:"venue_id"`              // 比赛场地 ID。
	Venue      *Venue         `gorm:"foreignKey:VenueID" json:"venue,omitempty"`               // 预加载时挂载的场地对象。
	StartTime  time.Time      `gorm:"index:idx_matches_start_time;not null" json:"start_time"` // 比赛开始时间。
	EndTime    time.Time      `gorm:"not null" json:"end_time"`                                // 比赛结束时间。
	Price      float64        `gorm:"type:decimal(10,2);default:0" json:"price"`               // 单人报名费用。
	MaxPlayers int            `gorm:"default:14" json:"max_players"`                           // 最多确认参赛人数。
	Format     int            `gorm:"default:7" json:"format"`                                 // 比赛制式（如 5/7/11 人制）。
	Note       string         `gorm:"type:text" json:"note"`                                   // 比赛备注（装备要求、集合说明等）。
	Status     string         `gorm:"size:20;default:'RECRUITING'" json:"status"`              // 比赛状态：RECRUITING / FULL / FINISHED / CANCELED。
	CreatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`             // 比赛创建时间。
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                                          // 软删除时间（为空表示未删除）。
}

// TableName 指定 Match 对应的数据表名。
func (Match) TableName() string {
	return "matches"
}
