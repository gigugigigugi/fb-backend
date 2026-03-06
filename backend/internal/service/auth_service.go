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
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	verificationCodeTTL = 10 * time.Minute // 验证码有效期。
	sendCooldown        = 60 * time.Second // 同一业务键两次发送的最小间隔。
	maxSendPerHour      = 5                // 每小时发送上限。
	maxVerifyFailures   = 5                // 连续失败达到该值后进入锁定。
	verifyLockDuration  = 10 * time.Minute // 锁定时长。
)

// AuthService 负责注册登录与验证码验证相关业务。
type AuthService struct {
	userRepo   repository.UserRepository
	verifyRepo repository.VerificationRepository
}

// NewAuthService 创建认证服务实例。
func NewAuthService(userRepo repository.UserRepository, verifyRepo repository.VerificationRepository) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		verifyRepo: verifyRepo,
	}
}

// RegisterEmail 使用邮箱密码注册并返回 JWT。
func (s *AuthService) RegisterEmail(ctx context.Context, email, password, nickname string) (string, error) {
	if existingUser, _ := s.userRepo.GetUserByEmail(ctx, email); existingUser != nil {
		return "", errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	newUser := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Nickname:     nickname,
	}

	if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
		return "", err
	}

	return utils.GenerateToken(newUser.ID)
}

// LoginEmail 使用邮箱密码登录并返回 JWT。
func (s *AuthService) LoginEmail(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	return utils.GenerateToken(user.ID)
}

// LoginGoogle 使用 Google 身份登录；若用户不存在则自动创建。
func (s *AuthService) LoginGoogle(ctx context.Context, googleID, email, name, avatar string) (string, error) {
	user, _ := s.userRepo.GetUserByGoogleID(ctx, googleID)
	if user == nil {
		existingEmailUser, _ := s.userRepo.GetUserByEmail(ctx, email)
		if existingEmailUser != nil {
			return "", errors.New("email already bounded to an existed native account")
		}

		newUser := &model.User{
			Email:    email,
			GoogleID: &googleID,
			Nickname: name,
			Avatar:   avatar,
		}

		if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
			return "", err
		}
		user = newUser
	}

	return utils.GenerateToken(user.ID)
}

// SendEmailVerificationCode 发送邮箱验证码（仅生成并记录，实际发送由外部渠道处理）。
func (s *AuthService) SendEmailVerificationCode(ctx context.Context, userID uint) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Email == "" {
		return errors.New("email is empty")
	}

	key := fmt.Sprintf("email:%d", userID)
	now := time.Now()
	state, err := s.getOrInitVerificationState(ctx, key, now)
	if err != nil {
		return err
	}

	// 统一执行发送频控与窗口计数推进。
	if err := s.checkSendAllowedAndAdvance(state, now); err != nil {
		return err
	}

	code, err := generate6DigitCode()
	if err != nil {
		return err
	}
	state.Code = code
	state.ExpiresAt = now.Add(verificationCodeTTL)

	if err := s.verifyRepo.Upsert(ctx, state); err != nil {
		return err
	}

	slog.Info("Email verification code generated",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("email", user.Email),
		slog.String("code", code),
	)
	return nil
}

// VerifyEmailCode 校验邮箱验证码并更新邮箱已验证状态。
func (s *AuthService) VerifyEmailCode(ctx context.Context, userID uint, code string) error {
	key := fmt.Sprintf("email:%d", userID)
	now := time.Now()

	state, err := s.verifyRepo.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, repository.ErrVerificationStateNotFound) {
			return errors.New("verification code not found")
		}
		return err
	}

	if err := checkVerifyAllowed(state, now); err != nil {
		return err
	}

	if state.Code == "" {
		return errors.New("verification code not found")
	}

	if now.After(state.ExpiresAt) {
		if markErr := s.markVerifyFailed(ctx, state, now); markErr != nil {
			return markErr
		}
		return errors.New("verification code expired")
	}

	if state.Code != code {
		if markErr := s.markVerifyFailed(ctx, state, now); markErr != nil {
			return markErr
		}
		return errors.New("invalid verification code")
	}

	if err := s.userRepo.UpdateEmailVerified(ctx, userID, true); err != nil {
		return err
	}

	// 校验成功后清空验证码和失败状态，避免重复使用。
	s.clearVerifyState(state, now)
	return s.verifyRepo.Upsert(ctx, state)
}

// SendPhoneVerificationCode 发送手机验证码（当前仅写入状态并打印日志）。
func (s *AuthService) SendPhoneVerificationCode(ctx context.Context, userID uint, phone string) error {
	normalized := utils.NormalizePhone(phone)
	if !utils.IsValidE164(normalized) {
		return errors.New("phone must be valid E.164 format")
	}

	existing, err := s.userRepo.GetUserByPhone(ctx, normalized)
	if err == nil && existing != nil && existing.ID != userID {
		return errors.New("phone already bound to another account")
	}
	if err != nil && err.Error() != "user not found" {
		return err
	}

	key := phoneCodeKey(userID, normalized)
	now := time.Now()
	state, err := s.getOrInitVerificationState(ctx, key, now)
	if err != nil {
		return err
	}

	if err := s.checkSendAllowedAndAdvance(state, now); err != nil {
		return err
	}

	code, err := generate6DigitCode()
	if err != nil {
		return err
	}
	state.Code = code
	state.ExpiresAt = now.Add(verificationCodeTTL)

	if err := s.verifyRepo.Upsert(ctx, state); err != nil {
		return err
	}

	slog.Info("Phone verification code generated",
		slog.Uint64("user_id", uint64(userID)),
		slog.String("phone", normalized),
		slog.String("code", code),
	)
	return nil
}

// VerifyPhoneCode 校验手机验证码并写入手机号与已验证状态。
func (s *AuthService) VerifyPhoneCode(ctx context.Context, userID uint, phone string, code string) error {
	normalized := utils.NormalizePhone(phone)
	if !utils.IsValidE164(normalized) {
		return errors.New("phone must be valid E.164 format")
	}

	key := phoneCodeKey(userID, normalized)
	now := time.Now()

	state, err := s.verifyRepo.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, repository.ErrVerificationStateNotFound) {
			return errors.New("verification code not found")
		}
		return err
	}

	if err := checkVerifyAllowed(state, now); err != nil {
		return err
	}

	if state.Code == "" {
		return errors.New("verification code not found")
	}

	if now.After(state.ExpiresAt) {
		if markErr := s.markVerifyFailed(ctx, state, now); markErr != nil {
			return markErr
		}
		return errors.New("verification code expired")
	}

	if state.Code != code {
		if markErr := s.markVerifyFailed(ctx, state, now); markErr != nil {
			return markErr
		}
		return errors.New("invalid verification code")
	}

	if err := s.userRepo.UpdatePhoneVerified(ctx, userID, normalized, true); err != nil {
		return err
	}

	s.clearVerifyState(state, now)
	return s.verifyRepo.Upsert(ctx, state)
}

// getOrInitVerificationState 读取验证码状态，不存在则初始化默认状态。
func (s *AuthService) getOrInitVerificationState(ctx context.Context, key string, now time.Time) (*repository.VerificationState, error) {
	state, err := s.verifyRepo.GetByKey(ctx, key)
	if err == nil {
		return state, nil
	}
	if !errors.Is(err, repository.ErrVerificationStateNotFound) {
		return nil, err
	}

	return &repository.VerificationState{
		BizKey:      key,
		WindowStart: now,
		SendCount:   0,
		FailCount:   0,
	}, nil
}

// checkSendAllowedAndAdvance 校验发送限制并推进发送计数。
func (s *AuthService) checkSendAllowedAndAdvance(state *repository.VerificationState, now time.Time) error {
	// 超过 1 小时则开启新统计窗口。
	if now.Sub(state.WindowStart) >= time.Hour {
		state.WindowStart = now
		state.SendCount = 0
	}

	if state.LastSentAt != nil && now.Sub(*state.LastSentAt) < sendCooldown {
		return errors.New("send too frequent, please try again later")
	}
	if state.SendCount >= maxSendPerHour {
		return errors.New("send limit reached, please try again in 1 hour")
	}

	state.SendCount++
	state.LastSentAt = &now
	return nil
}

// checkVerifyAllowed 校验当前是否处于失败锁定期。
func checkVerifyAllowed(state *repository.VerificationState, now time.Time) error {
	if state.LockUntil != nil && now.Before(*state.LockUntil) {
		return errors.New("too many failed attempts, please try again later")
	}
	return nil
}

// markVerifyFailed 记录一次校验失败，并在达到阈值后写入锁定时间。
func (s *AuthService) markVerifyFailed(ctx context.Context, state *repository.VerificationState, now time.Time) error {
	state.FailCount++
	if state.FailCount >= maxVerifyFailures {
		lockUntil := now.Add(verifyLockDuration)
		state.LockUntil = &lockUntil
		state.FailCount = 0
	}
	return s.verifyRepo.Upsert(ctx, state)
}

// clearVerifyState 清空验证码与失败状态（用于校验成功后）。
func (s *AuthService) clearVerifyState(state *repository.VerificationState, now time.Time) {
	state.Code = ""
	state.ExpiresAt = now
	state.FailCount = 0
	state.LockUntil = nil
}

// phoneCodeKey 生成手机号验证码状态的业务键。
func phoneCodeKey(userID uint, phone string) string {
	return fmt.Sprintf("phone:%d:%s", userID, phone)
}

// generate6DigitCode 生成 6 位数字验证码。
func generate6DigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
