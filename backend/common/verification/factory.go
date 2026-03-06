package verification

import (
	"fmt"
	"strings"
	"time"

	"football-backend/common/config"
)

// NewCodeProviderFromConfig 按配置创建验证码发送器。
func NewCodeProviderFromConfig(cfg config.VerificationConfig) (CodeProvider, error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if mode == "" || mode == "mock" {
		return NewMockCodeProvider(), nil
	}
	if mode != "real" {
		return nil, fmt.Errorf("unsupported verification provider mode: %s", cfg.Provider)
	}

	return NewHTTPCodeProvider(HTTPProviderConfig{
		Timeout:     time.Duration(cfg.HTTPTimeoutSec) * time.Second,
		EmailAPIURL: cfg.Email.APIURL,
		EmailAPIKey: cfg.Email.APIKey,
		EmailFrom:   cfg.Email.From,
		SMSAPIURL:   cfg.SMS.APIURL,
		SMSAPIKey:   cfg.SMS.APIKey,
		SMSSender:   cfg.SMS.Sender,
	})
}
