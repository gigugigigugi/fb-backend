package model

import "time"

// VerificationChallenge 存储验证码、发送频控和失败锁定状态。
// BizKey 示例：
// - email:123
// - phone:123:+819012345678
type VerificationChallenge struct {
	ID          uint       `gorm:"primaryKey" json:"id"`                         // 主键 ID。
	BizKey      string     `gorm:"size:200;uniqueIndex;not null" json:"biz_key"` // 业务唯一键（定位某个邮箱/手机号验证流程）。
	Code        string     `gorm:"size:6;not null" json:"code"`                  // 当前有效验证码。
	ExpiresAt   time.Time  `gorm:"not null" json:"expires_at"`                   // 验证码失效时间。
	SendCount   int        `gorm:"default:0" json:"send_count"`                  // 当前窗口内已发送次数。
	WindowStart time.Time  `gorm:"not null" json:"window_start"`                 // 发送计数窗口起始时间。
	LastSentAt  *time.Time `json:"last_sent_at,omitempty"`                       // 最近一次发送时间（用于冷却判断）。
	FailCount   int        `gorm:"default:0" json:"fail_count"`                  // 连续校验失败次数。
	LockUntil   *time.Time `json:"lock_until,omitempty"`                         // 达到失败阈值后的锁定截止时间。
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`  // 创建时间。
	UpdatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`  // 更新时间。
}

// TableName 指定 VerificationChallenge 对应的数据表名。
func (VerificationChallenge) TableName() string {
	return "verification_challenges"
}
