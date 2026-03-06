package model

import (
	"gorm.io/gorm"
)

// Venue 表示比赛场地。
type Venue struct {
	ID          uint           `gorm:"primaryKey" json:"id"`                                  // 场地主键 ID。
	Name        string         `gorm:"size:100;not null" json:"name"`                         // 场地名称。
	Prefecture  string         `gorm:"size:50;index:idx_venues_prefecture" json:"prefecture"` // 一级行政区（都道府县/省州）。
	City        string         `gorm:"size:50;index:idx_venues_city" json:"city"`             // 城市或区。
	Address     string         `gorm:"type:text" json:"address"`                              // 场地详细地址。
	Latitude    float64        `json:"latitude"`                                              // 纬度坐标。
	Longitude   float64        `json:"longitude"`                                             // 经度坐标。
	Website     string         `gorm:"size:255" json:"website"`                               // 场地官网或预约链接。
	Description string         `gorm:"type:text" json:"description"`                          // 场地说明。
	CreatedBy   int            `gorm:"default:0" json:"created_by"`                           // 创建者用户 ID（0 表示系统预置）。
	IsVerified  bool           `gorm:"default:false" json:"is_verified"`                      // 是否已审核认证。
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间（为空表示未删除）。
}

// TableName 指定 Venue 对应的数据表名。
func (Venue) TableName() string {
	return "venues"
}
