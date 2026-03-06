package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User 表示平台用户。
type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`                            // 用户主键 ID。
	Email         string         `gorm:"size:100;unique;not null" json:"email"`           // 登录邮箱（唯一）。
	Phone         *string        `gorm:"size:20;uniqueIndex" json:"phone,omitempty"`      // 手机号（E.164 格式，唯一，可空）。
	PasswordHash  string         `gorm:"size:255" json:"-"`                               // 密码哈希（bcrypt）。
	GoogleID      *string        `gorm:"size:100;uniqueIndex" json:"google_id,omitempty"` // Google 账号绑定 ID（唯一，可空）。
	WechatID      *string        `gorm:"size:100;uniqueIndex" json:"wechat_id,omitempty"` // 微信账号绑定 ID（唯一，可空）。
	EmailVerified bool           `gorm:"default:false" json:"email_verified"`             // 邮箱是否已完成验证。
	PhoneVerified bool           `gorm:"default:false" json:"phone_verified"`             // 手机号是否已完成验证。
	Nickname      string         `gorm:"size:50" json:"nickname"`                         // 显示昵称。
	Avatar        string         `gorm:"size:255" json:"avatar"`                          // 头像 URL。
	Reputation    int            `gorm:"default:100" json:"-"`                            // 信誉分（内部使用，不对外透出）。
	Stats         datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"stats"`            // 动态统计信息 JSON。
	CreatedAt     time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`     // 用户创建时间。
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`                                  // 软删除时间（为空表示未删除）。

	Teams []Team `gorm:"many2many:team_members;" json:"teams,omitempty"` // 用户加入的球队列表。
}

// TableName 指定 User 对应的数据表名。
func (User) TableName() string {
	return "users"
}
