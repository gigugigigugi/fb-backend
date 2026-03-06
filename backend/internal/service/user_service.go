package service

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"strings"
)

// UserService 处理当前登录用户资料相关业务。
type UserService struct {
	repo repository.UserRepository
}

// UpdateMeInput 定义用户资料更新入参（均为可选字段）。
type UpdateMeInput struct {
	Nickname *string // 用户昵称。
	Avatar   *string // 用户头像 URL。
}

// NewUserService 创建用户服务实例。
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetMe 查询当前登录用户信息。
func (s *UserService) GetMe(ctx context.Context, userID uint) (*model.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// UpdateMe 更新当前登录用户的昵称和头像。
func (s *UserService) UpdateMe(ctx context.Context, userID uint, in UpdateMeInput) (*model.User, error) {
	if in.Nickname == nil && in.Avatar == nil {
		return nil, errors.New("nothing to update")
	}

	if in.Nickname != nil {
		nickname := strings.TrimSpace(*in.Nickname)
		if len(nickname) < 2 || len(nickname) > 50 {
			return nil, errors.New("nickname length must be between 2 and 50")
		}
		in.Nickname = &nickname
	}

	if in.Avatar != nil {
		avatar := strings.TrimSpace(*in.Avatar)
		if len(avatar) > 255 {
			return nil, errors.New("avatar url is too long")
		}
		in.Avatar = &avatar
	}

	return s.repo.UpdateUserProfile(ctx, userID, in.Nickname, in.Avatar)
}
