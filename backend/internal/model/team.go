package model

import (
	"time"

	"gorm.io/gorm"
)

// Team 球队模型
type Team struct {
	ID           uint           `gorm:"primaryKey" json:"id"`                          // 球队体系内的唯一主键ID标识
	Name         string         `gorm:"size:100;not null" json:"name"`                 // 球队的公开展示名称 (例如: 皇家马德里)
	Logo         string         `gorm:"size:255" json:"logo"`                          // 放置球队队徽图标的 CDN/对象存储网络访问地址
	Slogan       string         `gorm:"size:255" json:"slogan"`                        // 球队对外展示的口号短语 (例如: Hala Madrid)
	Description  string         `gorm:"type:text" json:"description"`                  // 球队长篇文字详情面板说明
	CaptainID    uint           `json:"captain_id"`                                    // 创建并拥有此球队的第一任队长的用户编号，拥有最高删库权限
	Captain      *User          `gorm:"foreignKey:CaptainID" json:"captain,omitempty"` // 关联指向当前队伍的拥有者(创建者)用户模型引用
	InviteCode   string         `gorm:"size:20;unique" json:"invite_code"`             // 一个短巧的外部进队邀请码(六位英文字段)，保证全局唯一性
	TotalMatches int            `gorm:"default:0" json:"total_matches"`                // 从球队创立至今打过的所有历史局数的总计数器
	WinRate      float64        `gorm:"default:0.0" json:"win_rate"`                   // 该球队累计赛事表现推算出的胜率统计 (0.0~1.0)
	CreatedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`   // 建队成立的大喜日子
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                // 球队解散的逻辑软删标记位

	Members []User `gorm:"many2many:team_members;" json:"members,omitempty"` // Gorm 的高级魔法特性：通过交叉映射表 team_members 获取所有拥有关联身份的队员体数组
}

// TableName 指定数据库表名
func (Team) TableName() string {
	return "teams"
}

// TeamMember 球队成员关联模型
type TeamMember struct {
	ID           uint           `gorm:"primaryKey" json:"id"`                          // 权限授予映射表唯一主键
	TeamID       uint           `gorm:"index:idx_team_members_team_id" json:"team_id"` // 所属挂靠的具体球队 ID
	UserID       uint           `gorm:"index:idx_team_members_user_id" json:"user_id"` // 被签下的用户身份 ID
	Role         string         `gorm:"size:20;default:'MEMBER'" json:"role"`          // 核心的RBAC权限体系枚举：OWNER(所有者与创建者), ADMIN(队副/财务管理员), MEMBER(普通球员)
	JoinTime     time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"join_time"`    // 用户点击同意进队当天的加盟记录时间
	JerseyNumber int            `json:"jersey_number"`                                 // （尚未完全应用）该用户在队伍中注册挂靠和印号的球衣专属背面数字
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                // 脱离球队（退会）时保留入会黑历史的软删标志位
}

// TableName 指定数据库表名
func (TeamMember) TableName() string {
	return "team_members"
}
