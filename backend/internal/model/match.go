package model

import "time"

// Match 比赛模型
type Match struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TeamID    uint      `gorm:"index:idx_matches_team_id" json:"team_id"`
	Team      *Team     `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	VenueID   uint      `gorm:"index:idx_matches_venue_id" json:"venue_id"`
	Venue     *Venue    `gorm:"foreignKey:VenueID" json:"venue,omitempty"`
	StartTime time.Time `gorm:"index:idx_matches_start_time;not null" json:"start_time"`
	EndTime   time.Time `gorm:"not null" json:"end_time"`
	Price     float64   `gorm:"type:decimal(10,2);default:0" json:"price"`
	MaxPlayers int      `gorm:"default:14" json:"max_players"`
	Format    int       `gorm:"default:7" json:"format"` // 5/7/11
	JerseyColor string  `gorm:"size:50" json:"jersey_color"`
	HasBibs   bool      `gorm:"default:false" json:"has_bibs"`
	Note      string    `gorm:"type:text" json:"note"`
	Status    string    `gorm:"size:20;default:'RECRUITING'" json:"status"` // RECRUITING, FULL, FINISHED, CANCELED
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定数据库表名
func (Match) TableName() string {
	return "matches"
}
