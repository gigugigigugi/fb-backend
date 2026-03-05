package notification

import (
	"context"
	"errors"
	"log/slog"
)

// EmailNotifier 是邮件通道发送器。
// 当前为占位实现：仅做参数校验并记录发送日志。
// 后续可替换为真实的邮件服务商 SDK（如 SES、SendGrid、Mailgun）。
type EmailNotifier struct{}

// NewEmailNotifier 创建邮件发送器实例。
func NewEmailNotifier() *EmailNotifier {
	return &EmailNotifier{}
}

// Channel 返回该发送器对应的通道类型。
func (n *EmailNotifier) Channel() Channel {
	return ChannelEmail
}

// Send 执行邮件发送。
// 当前实现不会真正发送邮件，而是以日志模拟成功路径。
func (n *EmailNotifier) Send(_ context.Context, recipient Recipient, msg Message) error {
	// 输入校验由通道实现兜底，避免上层漏判导致“假成功”。
	if recipient.Email == "" {
		return errors.New("email is empty")
	}

	slog.Info("Email notification sent",
		slog.Uint64("user_id", uint64(recipient.UserID)),
		slog.String("email", recipient.Email),
		slog.String("subject", msg.Subject),
	)
	return nil
}
