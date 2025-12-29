package model

// Venue 场地模型
type Venue struct {
	ID         uint    `gorm:"primaryKey" json:"id"`
	Name       string  `gorm:"size:100;not null" json:"name"`
	Prefecture string  `gorm:"size:50;index:idx_venues_prefecture" json:"prefecture"` // 一级行政区
	City       string  `gorm:"size:50;index:idx_venues_city" json:"city"`             // 二级行政区
	Address    string  `gorm:"type:text" json:"address"`                              // 完整地址
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	CreatedBy  int     `gorm:"default:0" json:"created_by"` // 0=官方, 其他=用户ID
	IsVerified bool    `gorm:"default:false" json:"is_verified"`
}

// TableName 指定数据库表名
func (Venue) TableName() string {
	return "venues"
}
