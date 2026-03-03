package service

import (
	"context"
	"errors"
	"football-backend/common/utils"
	"football-backend/internal/model"

	"football-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// RegisterEmail 处理标准邮箱注册
func (s *AuthService) RegisterEmail(ctx context.Context, email, password, nickname string) (string, error) {
	// 检查 Email 是否已经注册
	if existingUser, _ := s.repo.GetUserByEmail(ctx, email); existingUser != nil {
		return "", errors.New("email already registered")
	}

	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	newUser := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Nickname:     nickname,
	}

	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		return "", err
	}

	// 直接签发并返回 JWT
	return utils.GenerateToken(newUser.ID)
}

// LoginEmail 处理标准邮箱登录
func (s *AuthService) LoginEmail(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid email or password") // 为了安全，模糊错误提示
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	return utils.GenerateToken(user.ID)
}

// LoginGoogle 处理 Google OAuth (静默注册 + 直接发 Token)
func (s *AuthService) LoginGoogle(ctx context.Context, googleID, email, name, avatar string) (string, error) {
	// 寻找是否已经注册
	user, _ := s.repo.GetUserByGoogleID(ctx, googleID)
	if user == nil {
		// 没找到 GoogleID 绑定的，看看 Email 有没有被老账户占用了
		existingEmailUser, _ := s.repo.GetUserByEmail(ctx, email)
		if existingEmailUser != nil {
			// [高级玩法] Email 被占用，自动进行账号关联！将 GoogleID 强行挂载到旧账户上
			// 或者返回错误让用户手动操作，为了演示流畅度，我们暂时返回要求用户先原路登录
			return "", errors.New("email already bounded to an existed native account")
		}

		// 全新静默注册
		newUser := &model.User{
			Email:    email,
			GoogleID: &googleID,
			Nickname: name,
			Avatar:   avatar,
			// 第三方进来的免密，PasswordHash为空
		}

		if err := s.repo.CreateUser(ctx, newUser); err != nil {
			return "", err
		}

		user = newUser
	}

	// 签发自己系统的 Token
	return utils.GenerateToken(user.ID)
}
