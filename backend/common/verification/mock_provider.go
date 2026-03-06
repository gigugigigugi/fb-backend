package verification

import (
	"context"
	"fmt"
	"log/slog"
)

// MockCodeProvider 是开发环境默认实现，只打印日志不真实发送。
type MockCodeProvider struct{}

func NewMockCodeProvider() *MockCodeProvider {
	return &MockCodeProvider{}
}

func (m *MockCodeProvider) Mode() string {
	return "mock"
}

func (m *MockCodeProvider) SendEmailCode(_ context.Context, toEmail string, code string) error {
	slog.Info("[MOCK] send email verification code",
		slog.String("to", toEmail),
		slog.String("code", code),
	)
	return nil
}

func (m *MockCodeProvider) SendSMSCode(_ context.Context, toPhone string, code string) error {
	slog.Info("[MOCK] send sms verification code",
		slog.String("to", toPhone),
		slog.String("code", code),
	)
	return nil
}

func emailBody(code string) string {
	return fmt.Sprintf("Your verification code is %s. It expires in 10 minutes.", code)
}

func smsBody(code string) string {
	return fmt.Sprintf("Verification code: %s (valid for 10 minutes)", code)
}
