package notification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

// Dispatcher 负责异步通知调度。
// 当前策略是“单次投递”：任务入队后只发送一次，失败仅记录日志，不自动重试。
type Dispatcher struct {
	queue     chan Task            // 内存任务队列
	notifiers map[Channel]Notifier // 通道 -> 发送器实例
}

// NewDispatcher 创建并启动调度器。
// 参数说明：
// - bufferSize: 队列容量，<=0 时使用默认值 128。
func NewDispatcher(bufferSize int) *Dispatcher {
	if bufferSize <= 0 {
		bufferSize = 128
	}

	d := &Dispatcher{
		queue:     make(chan Task, bufferSize),
		notifiers: make(map[Channel]Notifier),
	}

	// 启动后台消费协程，持续处理队列中的任务。
	go d.worker()
	return d
}

// RegisterNotifier 注册某个通道的发送器。
// 如果同一通道重复注册，后注册的实例会覆盖前一个实例。
func (d *Dispatcher) RegisterNotifier(notifier Notifier) {
	d.notifiers[notifier.Channel()] = notifier
}

// Enqueue 将任务放入异步队列。
// 采用“快速失败”策略：队列满时立即返回错误，避免阻塞调用方线程。
func (d *Dispatcher) Enqueue(task Task) error {
	if len(task.Channels) == 0 {
		return errors.New("no notification channels configured")
	}

	select {
	case d.queue <- task:
		return nil
	default:
		return errors.New("notification queue is full")
	}
}

// worker 常驻消费队列。
// 只要队列未关闭，就会不断取任务并交给 process 处理。
func (d *Dispatcher) worker() {
	for task := range d.queue {
		d.process(task)
	}
}

// process 执行一次任务发送。
// 成功记录 info，失败记录 error，不进行重试。
func (d *Dispatcher) process(task Task) {
	err := d.dispatch(context.Background(), task)
	if err != nil {
		slog.Error("Notification task failed",
			slog.Uint64("match_id", uint64(task.MatchID)),
			slog.Uint64("user_id", uint64(task.Recipient.UserID)),
			slog.String("error", err.Error()),
		)
		return
	}

	slog.Info("Notification task delivered",
		slog.Uint64("match_id", uint64(task.MatchID)),
		slog.Uint64("user_id", uint64(task.Recipient.UserID)),
	)
}

// dispatch 执行“单次发送尝试”。
// 它会遍历任务中的所有通道并逐个发送，收集所有失败信息后统一返回。
func (d *Dispatcher) dispatch(ctx context.Context, task Task) error {
	var sendErrs []string

	for _, channel := range task.Channels {
		notifier, ok := d.notifiers[channel]
		if !ok {
			sendErrs = append(sendErrs, fmt.Sprintf("notifier not registered for channel=%s", channel))
			continue
		}

		if err := notifier.Send(ctx, task.Recipient, task.Message); err != nil {
			sendErrs = append(sendErrs, fmt.Sprintf("channel=%s err=%s", channel, err.Error()))
		}
	}

	if len(sendErrs) > 0 {
		return errors.New(strings.Join(sendErrs, "; "))
	}

	return nil
}
