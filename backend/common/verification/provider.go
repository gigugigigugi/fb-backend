package verification

import "context"

// CodeProvider 定义验证码发送能力。
type CodeProvider interface {
	SendEmailCode(ctx context.Context, toEmail string, code string) error
	SendSMSCode(ctx context.Context, toPhone string, code string) error
	Mode() string
}
