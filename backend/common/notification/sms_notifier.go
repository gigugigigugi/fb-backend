package notification

import (
	"context"
	"errors"
	"log/slog"
)

// SMSNotifier 是短信通道发送器。
// 当前为占位实现：仅做参数校验并记录发送日志。
// 后续可替换为真实短信网关（如阿里云短信、Twilio 等）。
type SMSNotifier struct{}

// NewSMSNotifier 创建短信发送器实例。
func NewSMSNotifier() *SMSNotifier {
	return &SMSNotifier{}
}

// Channel 返回该发送器对应的通道类型。
func (n *SMSNotifier) Channel() Channel {
	return ChannelSMS
}

// Send 执行短信发送。
// 当前实现不会真正下发短信，而是用日志模拟成功路径。
func (n *SMSNotifier) Send(_ context.Context, recipient Recipient, msg Message) error {
	// 当前项目尚未接入手机号字段，phone 为空时会返回错误并记录日志。
	if recipient.Phone == "" {
		return errors.New("phone is empty")
	}

	slog.Info("SMS notification sent",
		slog.Uint64("user_id", uint64(recipient.UserID)),
		slog.String("phone", recipient.Phone),
		slog.String("subject", msg.Subject),
		slog.String("body", msg.Body),
	)
	return nil
}
