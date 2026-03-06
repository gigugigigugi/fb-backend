package postgres

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type verificationRepository struct {
	db *gorm.DB
}

// NewVerificationRepository 创建验证码状态仓储的 PostgreSQL 实现。
func NewVerificationRepository(db *gorm.DB) repository.VerificationRepository {
	return &verificationRepository{db: db}
}

// GetByKey 按业务键读取验证码状态。
// 若记录不存在，返回 repository.ErrVerificationStateNotFound。
func (r *verificationRepository) GetByKey(ctx context.Context, bizKey string) (*repository.VerificationState, error) {
	var row model.VerificationChallenge
	if err := r.db.WithContext(ctx).Where("biz_key = ?", bizKey).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrVerificationStateNotFound
		}
		return nil, err
	}

	return &repository.VerificationState{
		BizKey:      row.BizKey,
		Code:        row.Code,
		ExpiresAt:   row.ExpiresAt,
		SendCount:   row.SendCount,
		WindowStart: row.WindowStart,
		LastSentAt:  row.LastSentAt,
		FailCount:   row.FailCount,
		LockUntil:   row.LockUntil,
	}, nil
}

// Upsert 创建或更新验证码状态。
// 通过 biz_key 冲突更新，保证同一业务键只有一条记录。
func (r *verificationRepository) Upsert(ctx context.Context, state *repository.VerificationState) error {
	row := model.VerificationChallenge{
		BizKey:      state.BizKey,
		Code:        state.Code,
		ExpiresAt:   state.ExpiresAt,
		SendCount:   state.SendCount,
		WindowStart: state.WindowStart,
		LastSentAt:  state.LastSentAt,
		FailCount:   state.FailCount,
		LockUntil:   state.LockUntil,
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "biz_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"code":         row.Code,
			"expires_at":   row.ExpiresAt,
			"send_count":   row.SendCount,
			"window_start": row.WindowStart,
			"last_sent_at": row.LastSentAt,
			"fail_count":   row.FailCount,
			"lock_until":   row.LockUntil,
		}),
	}).Create(&row).Error
}

// DeleteByKey 按业务键删除验证码状态。
func (r *verificationRepository) DeleteByKey(ctx context.Context, bizKey string) error {
	return r.db.WithContext(ctx).Where("biz_key = ?", bizKey).Delete(&model.VerificationChallenge{}).Error
}
