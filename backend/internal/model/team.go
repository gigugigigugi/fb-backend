package model

import "time"

// Team 球队模型
type Team struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Logo         string    `gorm:"size:255" json:"logo"`
	Slogan       string    `gorm:"size:255" json:"slogan"`
	Description  string    `gorm:"type:text" json:"description"`
	CaptainID    uint      `json:"captain_id"`
	Captain      *User     `gorm:"foreignKey:CaptainID" json:"captain,omitempty"`
	InviteCode   string    `gorm:"size:20;unique" json:"invite_code"`
	TotalMatches int       `gorm:"default:0" json:"total_matches"`
	WinRate      float64   `gorm:"default:0.0" json:"win_rate"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	Members []User `gorm:"many2many:team_members;" json:"members,omitempty"`
}

// TableName 指定数据库表名
func (Team) TableName() string {
	return "teams"
}

// TeamMember 球队成员关联模型
type TeamMember struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TeamID     uint      `gorm:"index:idx_team_members_team_id" json:"team_id"`
	UserID     uint      `gorm:"index:idx_team_members_user_id" json:"user_id"`
	Role       string    `gorm:"size:20;default:'MEMBER'" json:"role"` // OWNER, ADMIN, MEMBER
	JoinTime   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"join_time"`
	JerseyNumber int     `json:"jersey_number"`
}

// TableName 指定数据库表名
func (TeamMember) TableName() string {
	return "team_members"
}
