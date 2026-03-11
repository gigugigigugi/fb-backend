package main

import (
	"context"
	"errors"
	"fmt"
	"football-backend/common/config"
	verificationcode "football-backend/common/verification"
	"os"
	"strings"
	"time"
)

// smokeChannels 表示本次联调要验证的通道集合。
type smokeChannels struct {
	email bool
	sms   bool
}

func main() {
	config.Load()

	provider, err := verificationcode.NewCodeProviderFromConfig(config.App.Verification)
	if err != nil {
		fail("初始化验证码 provider 失败: %v", err)
	}

	if provider.Mode() != "real" {
		fail("当前 provider_mode=%s，联调验收要求 VERIFICATION_PROVIDER=real", provider.Mode())
	}

	channels, err := parseChannels(os.Getenv("PROVIDER_SMOKE_CHANNELS"))
	if err != nil {
		fail("解析 PROVIDER_SMOKE_CHANNELS 失败: %v", err)
	}

	emailTo := strings.TrimSpace(os.Getenv("PROVIDER_SMOKE_EMAIL_TO"))
	phoneTo := strings.TrimSpace(os.Getenv("PROVIDER_SMOKE_PHONE_TO"))
	code := strings.TrimSpace(os.Getenv("PROVIDER_SMOKE_CODE"))
	if code == "" {
		code = "654321"
	}

	if channels.email && emailTo == "" {
		fail("已选择 email 通道，但 PROVIDER_SMOKE_EMAIL_TO 为空")
	}
	if channels.sms && phoneTo == "" {
		fail("已选择 sms 通道，但 PROVIDER_SMOKE_PHONE_TO 为空")
	}

	timeout := time.Duration(config.App.Verification.HTTPTimeoutSec+3) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fmt.Printf("provider smoke start: mode=%s channels=%s timeout=%s\n",
		provider.Mode(), channels.String(), timeout.String())

	if channels.email {
		if err := provider.SendEmailCode(ctx, emailTo, code); err != nil {
			fail("Email 通道联调失败: %v", err)
		}
		fmt.Printf("Email 通道联调成功: to=%s\n", emailTo)
	}

	if channels.sms {
		if err := provider.SendSMSCode(ctx, phoneTo, code); err != nil {
			fail("SMS 通道联调失败: %v", err)
		}
		fmt.Printf("SMS 通道联调成功: to=%s\n", phoneTo)
	}

	fmt.Println("provider smoke success: 所选通道均已通过")
}

// parseChannels 解析联调通道配置，支持：email / sms / both（默认 both）。
func parseChannels(raw string) (smokeChannels, error) {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "" || mode == "both" {
		return smokeChannels{email: true, sms: true}, nil
	}
	switch mode {
	case "email":
		return smokeChannels{email: true, sms: false}, nil
	case "sms":
		return smokeChannels{email: false, sms: true}, nil
	default:
		return smokeChannels{}, errors.New("仅支持 email / sms / both")
	}
}

func (s smokeChannels) String() string {
	switch {
	case s.email && s.sms:
		return "email+sms"
	case s.email:
		return "email"
	case s.sms:
		return "sms"
	default:
		return "none"
	}
}

func fail(format string, args ...any) {
	fmt.Printf("provider smoke failed: "+format+"\n", args...)
	os.Exit(1)
}
