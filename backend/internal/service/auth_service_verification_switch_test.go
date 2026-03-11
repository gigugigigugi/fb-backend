package service

import (
	"context"
	"errors"
	"football-backend/internal/model"
	"football-backend/internal/repository"
	"testing"
)

// panicVerificationRepo 用于验证“关闭开关时不会访问验证码仓储”。
// 任一方法被调用都会触发测试失败。
type panicVerificationRepo struct{}

func (p *panicVerificationRepo) GetByKey(ctx context.Context, bizKey string) (*repository.VerificationState, error) {
	panic("unexpected call: GetByKey")
}

func (p *panicVerificationRepo) Upsert(ctx context.Context, state *repository.VerificationState) error {
	panic("unexpected call: Upsert")
}

func (p *panicVerificationRepo) DeleteByKey(ctx context.Context, bizKey string) error {
	panic("unexpected call: DeleteByKey")
}

// countingCodeProvider 统计验证码发送次数，帮助断言“关闭开关时不会发码”。
type countingCodeProvider struct {
	emailSendCount int
	smsSendCount   int
}

func (c *countingCodeProvider) SendEmailCode(ctx context.Context, toEmail string, code string) error {
	c.emailSendCount++
	return nil
}

func (c *countingCodeProvider) SendSMSCode(ctx context.Context, toPhone string, code string) error {
	c.smsSendCount++
	return nil
}

func (c *countingCodeProvider) Mode() string {
	return "mock"
}

// TestAuthServiceVerificationSwitchDisabled 验证在开关关闭时，验证码接口会短路返回，
// 并且不触发任何仓储写入或发送动作。
func TestAuthServiceVerificationSwitchDisabled(t *testing.T) {
	t.Run("send email disabled", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByIDFn: func(ctx context.Context, id uint) (*model.User, error) {
				return &model.User{
					ID:    id,
					Email: "switch@example.com",
				}, nil
			},
		}
		codeProvider := &countingCodeProvider{}

		svc := NewAuthService(
			userRepo,
			&panicVerificationRepo{},
			codeProvider,
			AuthServiceOptions{
				EmailVerificationEnabled: false,
				PhoneVerificationEnabled: false,
			},
		)

		err := svc.SendEmailVerificationCode(context.Background(), 1)
		if !errors.Is(err, ErrEmailVerificationDisabled) {
			t.Fatalf("expected ErrEmailVerificationDisabled, got %v", err)
		}
		if codeProvider.emailSendCount != 0 {
			t.Fatalf("expected no email send, got %d", codeProvider.emailSendCount)
		}
	})

	t.Run("verify email disabled", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updateEmailVerifyFn: func(ctx context.Context, userID uint, verified bool) error {
				t.Fatalf("unexpected call: UpdateEmailVerified")
				return nil
			},
		}
		codeProvider := &countingCodeProvider{}

		svc := NewAuthService(
			userRepo,
			&panicVerificationRepo{},
			codeProvider,
			AuthServiceOptions{
				EmailVerificationEnabled: false,
				PhoneVerificationEnabled: false,
			},
		)

		err := svc.VerifyEmailCode(context.Background(), 1, "123456")
		if !errors.Is(err, ErrEmailVerificationDisabled) {
			t.Fatalf("expected ErrEmailVerificationDisabled, got %v", err)
		}
	})

	t.Run("send phone disabled", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getUserByPhoneFn: func(ctx context.Context, phone string) (*model.User, error) {
				t.Fatalf("unexpected call: GetUserByPhone")
				return nil, nil
			},
		}
		codeProvider := &countingCodeProvider{}

		svc := NewAuthService(
			userRepo,
			&panicVerificationRepo{},
			codeProvider,
			AuthServiceOptions{
				EmailVerificationEnabled: false,
				PhoneVerificationEnabled: false,
			},
		)

		err := svc.SendPhoneVerificationCode(context.Background(), 1, "+819012345678")
		if !errors.Is(err, ErrPhoneVerificationDisabled) {
			t.Fatalf("expected ErrPhoneVerificationDisabled, got %v", err)
		}
		if codeProvider.smsSendCount != 0 {
			t.Fatalf("expected no sms send, got %d", codeProvider.smsSendCount)
		}
	})

	t.Run("verify phone disabled", func(t *testing.T) {
		userRepo := &mockUserRepo{
			updatePhoneVerifyFn: func(ctx context.Context, userID uint, phone string, verified bool) error {
				t.Fatalf("unexpected call: UpdatePhoneVerified")
				return nil
			},
		}
		codeProvider := &countingCodeProvider{}

		svc := NewAuthService(
			userRepo,
			&panicVerificationRepo{},
			codeProvider,
			AuthServiceOptions{
				EmailVerificationEnabled: false,
				PhoneVerificationEnabled: false,
			},
		)

		err := svc.VerifyPhoneCode(context.Background(), 1, "+819012345678", "123456")
		if !errors.Is(err, ErrPhoneVerificationDisabled) {
			t.Fatalf("expected ErrPhoneVerificationDisabled, got %v", err)
		}
	})
}
