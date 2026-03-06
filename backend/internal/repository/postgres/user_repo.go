package postgres

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 的 PostgreSQL 实现。
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// CreateUser 创建用户。
func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetUserByID 按用户 ID 查询。
func (r *userRepository) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 按邮箱查询。
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone 按手机号查询。
func (r *userRepository) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByGoogleID 按 GoogleID 查询。
func (r *userRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserProfile 更新用户昵称和头像，并返回更新后的用户信息。
func (r *userRepository) UpdateUserProfile(ctx context.Context, userID uint, nickname *string, avatar *string) (*model.User, error) {
	updates := make(map[string]interface{})
	if nickname != nil {
		updates["nickname"] = *nickname
	}
	if avatar != nil {
		updates["avatar"] = *avatar
	}

	if len(updates) > 0 {
		if err := r.db.WithContext(ctx).Model(&model.User{}).
			Where("id = ?", userID).
			Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return r.GetUserByID(ctx, userID)
}

// UpdateEmailVerified 更新邮箱验证状态。
func (r *userRepository) UpdateEmailVerified(ctx context.Context, userID uint, verified bool) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("email_verified", verified).Error
}

// UpdatePhoneVerified 更新手机号与手机验证状态。
func (r *userRepository) UpdatePhoneVerified(ctx context.Context, userID uint, phone string, verified bool) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"phone":          phone,
			"phone_verified": verified,
		}).Error
}
