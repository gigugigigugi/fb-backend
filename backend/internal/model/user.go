package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`                            // 核心用户 ID，系统所有外键参照根基
	Email        string         `gorm:"size:100;unique;not null" json:"email"`           // 用户的关键登录凭证，具有独立约束
	PasswordHash string         `gorm:"size:255" json:"-"`                               // 存储 bcrypt 处理后的加密指纹密码，允许为空(第三方登录时无密码)
	GoogleID     *string        `gorm:"size:100;uniqueIndex" json:"google_id,omitempty"` // Google 单点登录绑定的唯一凭证
	WechatID     *string        `gorm:"size:100;uniqueIndex" json:"wechat_id,omitempty"` // 微信小程序/APP绑定的唯一凭证
	Nickname     string         `gorm:"size:50" json:"nickname"`                         // App 界面和比赛阵容里对所有人公开的称谓名字
	Avatar       string         `gorm:"size:255" json:"avatar"`                          // OSS 或图像储存桶回传的真实头像外部 URL
	Reputation   int            `gorm:"default:100" json:"-"`                            // 用户的风控信誉积分(起点为100)，报名后鸽子不来可直接扣分，后端隐式校验
	Stats        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"stats"`            // 基于 PGSQL 原生的 JSON(B) 扩展列，支持动态挂载大量游戏化徽章奖励如 {"mvp": 5, "late_count": 0}
	CreatedAt    time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`     // 用户完成首次注册并存入库的归档时间点
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`                                  // 注销账户的逻辑软性删除指令时间戳

	Teams []Team `gorm:"many2many:team_members;" json:"teams,omitempty"` // 反向映射：该用户所加入的所有球队列表
}

// TableName 指定数据库表名
func (User) TableName() string {
	return "users"
}
