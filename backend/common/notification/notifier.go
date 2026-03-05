package notification

import "context"

// Channel 表示通知发送通道类型。
// 通过通道枚举，业务层可以声明“发哪些渠道”，而不关心具体实现细节。
type Channel string

const (
	// ChannelEmail 邮件通知通道。
	ChannelEmail Channel = "email"
	// ChannelSMS 短信通知通道。
	ChannelSMS Channel = "sms"
)

// Recipient 描述通知接收人信息。
// 同时保留 Email 与 Phone 字段，便于单任务多通道发送。
type Recipient struct {
	UserID uint
	Email  string
	Phone  string
}

// Message 描述通知内容。
// Subject 可用于邮件标题或短信主题，Body 为正文。
type Message struct {
	Subject string
	Body    string
}

// Task 是进入异步队列的通知任务实体。
// MatchID 用于日志追踪和业务排障；Channels 指定本次任务尝试的通道集合。
type Task struct {
	MatchID   uint
	Recipient Recipient
	Message   Message
	Channels  []Channel // 同一任务支持多通道并发尝试，例如 Email + SMS
}

// Notifier 定义“单通道发送器”接口。
// 每种通道（Email/SMS/Push）都应实现该接口并向 Dispatcher 注册。
type Notifier interface {
	// Channel 返回当前发送器所支持的通道类型。
	Channel() Channel
	// Send 执行实际发送。
	// 当前 Dispatcher 采用单次投递策略，发送失败不会自动重试。
	Send(ctx context.Context, recipient Recipient, msg Message) error
}
