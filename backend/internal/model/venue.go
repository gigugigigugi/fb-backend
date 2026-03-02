package model

import (
	"gorm.io/gorm"
)

// Venue 场地模型
type Venue struct {
	ID          uint           `gorm:"primaryKey" json:"id"`                                  // 场地的实体唯一编号
	Name        string         `gorm:"size:100;not null" json:"name"`                         // 场地的官方宣发或标准名字 (例如 辰巳の森公园)
	Prefecture  string         `gorm:"size:50;index:idx_venues_prefecture" json:"prefecture"` // 高频用于 App 发现页筛选的一级大圈行政区化划分 ("東京都", "埼玉県")
	City        string         `gorm:"size:50;index:idx_venues_city" json:"city"`             // 可支持下钻至二级行政体系定位 ("江東区", "品川区")
	Address     string         `gorm:"type:text" json:"address"`                              // 最终导航用的邮政与地理全名称的文本拼接字符串
	Latitude    float64        `json:"latitude"`                                              // 连接 Google Maps 用于距离演算的维度坐标
	Longitude   float64        `json:"longitude"`                                             // 连接 Google Maps 用于定位映射的经度坐标
	Website     string         `gorm:"size:255" json:"website"`                               // [UGC扩充] 球场官方的预约网站或设施说明页面链接
	Description string         `gorm:"type:text" json:"description"`                          // [UGC扩充] 玩家自行填写的场地图文经验描述（如：入口在C栋后面天桥下）
	CreatedBy   int            `gorm:"default:0" json:"created_by"`                           // 特供判断口：0 代表开发组手工导入的官方验证红V级场地，其它用户 ID 是普通人自定义加盖的野生球场
	IsVerified  bool           `gorm:"default:false" json:"is_verified"`                      // 配合上面的控制位。开启可直接挂靠为白皮书信誉推荐的置顶高级场地
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                                        // 防止物理抹除造成的关联比赛找不着该场地的软删除隔离带
}

// TableName 指定数据库表名
func (Venue) TableName() string {
	return "venues"
}
