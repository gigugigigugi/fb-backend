package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"football-backend/common/utils"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"log/slog"
	"math/big"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type verificationRecord struct {
	Code      string
	ExpiresAt time.Time
}

type AuthService struct {
	repo repository.UserRepository

	mu         sync.Mutex
	emailCodes map[uint]verificationRecord
	phoneCodes map[string]verificationRecord
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{
		repo:       repo,
		emailCodes: make(map[uint]verificationRecord),
		phoneCodes: make(map[string]verificationRecord),
	}
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

func (s *AuthService) SendEmailVerificationCode(ctx context.Context, userID uint) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Email == "" {
		return errors.New("email is empty")
	}

	code, err := generate6DigitCode()
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.emailCodes[userID] = verificationRecord{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	s.mu.Unlock()

	// 这里是发送占位，后续可替换为真实邮件服务商。
	slog.Info("Email verification code generated",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("email", user.Email),
		slog.String("code", code),
	)
	return nil
}

func (s *AuthService) VerifyEmailCode(ctx context.Context, userID uint, code string) error {
	s.mu.Lock()
	record, ok := s.emailCodes[userID]
	if !ok {
		s.mu.Unlock()
		return errors.New("verification code not found")
	}
	if time.Now().After(record.ExpiresAt) {
		delete(s.emailCodes, userID)
		s.mu.Unlock()
		return errors.New("verification code expired")
	}
	if record.Code != code {
		s.mu.Unlock()
		return errors.New("invalid verification code")
	}
	delete(s.emailCodes, userID)
	s.mu.Unlock()

	return s.repo.UpdateEmailVerified(ctx, userID, true)
}

func (s *AuthService) SendPhoneVerificationCode(ctx context.Context, userID uint, phone string) error {
	normalized := utils.NormalizePhone(phone)
	if !utils.IsValidE164(normalized) {
		return errors.New("phone must be valid E.164 format")
	}

	existing, err := s.repo.GetUserByPhone(ctx, normalized)
	if err == nil && existing != nil && existing.ID != userID {
		return errors.New("phone already bound to another account")
	}
	if err != nil && err.Error() != "user not found" {
		return err
	}

	code, err := generate6DigitCode()
	if err != nil {
		return err
	}

	key := phoneCodeKey(userID, normalized)
	s.mu.Lock()
	s.phoneCodes[key] = verificationRecord{
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	s.mu.Unlock()

	// 这里是发送占位，后续可替换为真实短信服务商。
	slog.Info("Phone verification code generated",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("phone", normalized),
		slog.String("code", code),
	)
	return nil
}

func (s *AuthService) VerifyPhoneCode(ctx context.Context, userID uint, phone string, code string) error {
	normalized := utils.NormalizePhone(phone)
	if !utils.IsValidE164(normalized) {
		return errors.New("phone must be valid E.164 format")
	}

	key := phoneCodeKey(userID, normalized)
	s.mu.Lock()
	record, ok := s.phoneCodes[key]
	if !ok {
		s.mu.Unlock()
		return errors.New("verification code not found")
	}
	if time.Now().After(record.ExpiresAt) {
		delete(s.phoneCodes, key)
		s.mu.Unlock()
		return errors.New("verification code expired")
	}
	if record.Code != code {
		s.mu.Unlock()
		return errors.New("invalid verification code")
	}
	delete(s.phoneCodes, key)
	s.mu.Unlock()

	return s.repo.UpdatePhoneVerified(ctx, userID, normalized, true)
}

func phoneCodeKey(userID uint, phone string) string {
	return fmt.Sprintf("%d:%s", userID, phone)
}

func generate6DigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
