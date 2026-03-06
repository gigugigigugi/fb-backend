package model

import (
	"time"

	"gorm.io/gorm"
)

// Team 表示一个球队实体。
type Team struct {
	ID           uint           `gorm:"primaryKey" json:"id"`                          // 球队主键 ID。
	Name         string         `gorm:"size:100;not null" json:"name"`                 // 球队名称。
	Logo         string         `gorm:"size:255" json:"logo"`                          // 球队 Logo 地址。
	Slogan       string         `gorm:"size:255" json:"slogan"`                        // 球队口号。
	Description  string         `gorm:"type:text" json:"description"`                  // 球队简介。
	CaptainID    uint           `json:"captain_id"`                                    // 队长用户 ID。
	Captain      *User          `gorm:"foreignKey:CaptainID" json:"captain,omitempty"` // 预加载时挂载的队长用户对象。
	InviteCode   string         `gorm:"size:20;unique" json:"invite_code"`             // 球队邀请编码（唯一）。
	TotalMatches int            `gorm:"default:0" json:"total_matches"`                // 球队累计比赛场次。
	WinRate      float64        `gorm:"default:0.0" json:"win_rate"`                   // 球队胜率。
	CreatedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`   // 球队创建时间。
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                // 软删除时间（为空表示未删除）。

	Members []User `gorm:"many2many:team_members;" json:"members,omitempty"` // 通过中间表关联的球队成员列表。
}

// TableName 指定 Team 对应的数据表名。
func (Team) TableName() string {
	return "teams"
}

// TeamMember 表示用户与球队的成员关系。
type TeamMember struct {
	ID           uint           `gorm:"primaryKey" json:"id"`                          // 成员关系主键 ID。
	TeamID       uint           `gorm:"index:idx_team_members_team_id" json:"team_id"` // 球队 ID。
	UserID       uint           `gorm:"index:idx_team_members_user_id" json:"user_id"` // 用户 ID。
	Role         string         `gorm:"size:20;default:'MEMBER'" json:"role"`          // 成员角色：OWNER / ADMIN / MEMBER。
	JoinTime     time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"join_time"`    // 入队时间。
	JerseyNumber int            `json:"jersey_number"`                                 // 球衣号码。
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                // 软删除时间（为空表示未删除）。
}

// TableName 指定 TeamMember 对应的数据表名。
func (TeamMember) TableName() string {
	return "team_members"
}
