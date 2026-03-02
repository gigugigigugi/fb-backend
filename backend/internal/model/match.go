package model

import (
	"time"

	"gorm.io/gorm"
)

// Match 比赛模型
type Match struct {
	ID         uint           `gorm:"primaryKey" json:"id"`                                    // 比赛场的全局唯一主键 ID
	TeamID     uint           `gorm:"index:idx_matches_team_id" json:"team_id"`                // 发布比赛的宿主球队的 ID (可用于查某队历史战局)
	Team       *Team          `gorm:"foreignKey:TeamID" json:"team,omitempty"`                 // 指向球队的实体对象关联
	VenueID    uint           `gorm:"index:idx_matches_venue_id" json:"venue_id"`              // 比赛所在地的物理球场场地 ID
	Venue      *Venue         `gorm:"foreignKey:VenueID" json:"venue,omitempty"`               // 指向场地的实体对象关联
	StartTime  time.Time      `gorm:"index:idx_matches_start_time;not null" json:"start_time"` // 比赛哨响开始的具体时间
	EndTime    time.Time      `gorm:"not null" json:"end_time"`                                // 比赛预定结束的时间
	Price      float64        `gorm:"type:decimal(10,2);default:0" json:"price"`               // 单人的报名费用或者总场地摊薄均摊费用
	MaxPlayers int            `gorm:"default:14" json:"max_players"`                           // 容量上限锁定阈值，判定是否允许普通加入并转为等待名单 (WAITING) 的关键参数
	Format     int            `gorm:"default:7" json:"format"`                                 // 赛制规格：常见值为 5 (五人制), 7 (七人制), 11 (十一人制正规足球)
	Note       string         `gorm:"type:text" json:"note"`                                   // 队长发出的参赛公告、要求带什么颜色衣服或是分队背心等富文本备注
	Status     string         `gorm:"size:20;default:'RECRUITING'" json:"status"`              // 状态机流转控制节点：RECRUITING(招募中), FULL(满员锁定), FINISHED(已完赛进入结算), CANCELED(因故取消)
	CreatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`             // 局子被创建的时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                                          // Gorm提供的软删哨兵时间戳，避免直接丢失比赛资料
}

// TableName 指定数据库表名
func (Match) TableName() string {
	return "matches"
}
