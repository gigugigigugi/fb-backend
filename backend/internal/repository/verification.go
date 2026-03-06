package repository

import (
	"context"
	"errors"
	"time"
)

var ErrVerificationStateNotFound = errors.New("verification state not found")

// VerificationState 是验证码业务在存储层的统一状态结构。
// 通过该抽象，可以将底层实现从 PostgreSQL 切换为 Redis 而不影响 service 层。
type VerificationState struct {
	BizKey      string     // 业务唯一键（email:uid / phone:uid:phone）。
	Code        string     // 当前验证码。
	ExpiresAt   time.Time  // 验证码过期时间。
	SendCount   int        // 当前发送窗口内的发送次数。
	WindowStart time.Time  // 当前发送窗口开始时间。
	LastSentAt  *time.Time // 最近一次发送时间。
	FailCount   int        // 连续失败次数。
	LockUntil   *time.Time // 锁定截止时间。
}

// VerificationRepository 定义验证码状态持久化接口。
// 当前由 PostgreSQL 实现，后续可切换 Redis 实现。
type VerificationRepository interface {
	// GetByKey 按业务键读取状态。
	GetByKey(ctx context.Context, bizKey string) (*VerificationState, error)
	// Upsert 创建或更新状态。
	Upsert(ctx context.Context, state *VerificationState) error
	// DeleteByKey 按业务键删除状态。
	DeleteByKey(ctx context.Context, bizKey string) error
}
