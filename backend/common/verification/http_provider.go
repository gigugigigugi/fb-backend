package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPProviderConfig 是通用 HTTP 验证码网关配置。
type HTTPProviderConfig struct {
	Timeout time.Duration

	EmailAPIURL string
	EmailAPIKey string
	EmailFrom   string

	SMSAPIURL string
	SMSAPIKey string
	SMSSender string
}

// HTTPCodeProvider 通过 HTTP 调用外部短信/邮件网关。
type HTTPCodeProvider struct {
	client *http.Client
	cfg    HTTPProviderConfig
}

func NewHTTPCodeProvider(cfg HTTPProviderConfig) (*HTTPCodeProvider, error) {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 8 * time.Second
	}
	if cfg.EmailAPIURL == "" || cfg.SMSAPIURL == "" {
		return nil, errors.New("email/sms api url is required in real mode")
	}
	return &HTTPCodeProvider{
		client: &http.Client{Timeout: cfg.Timeout},
		cfg:    cfg,
	}, nil
}

func (h *HTTPCodeProvider) Mode() string {
	return "real"
}

func (h *HTTPCodeProvider) SendEmailCode(ctx context.Context, toEmail string, code string) error {
	payload := map[string]string{
		"to":      toEmail,
		"from":    h.cfg.EmailFrom,
		"subject": "Verification Code",
		"body":    emailBody(code),
	}
	return h.postJSON(ctx, h.cfg.EmailAPIURL, h.cfg.EmailAPIKey, payload)
}

func (h *HTTPCodeProvider) SendSMSCode(ctx context.Context, toPhone string, code string) error {
	payload := map[string]string{
		"to":      toPhone,
		"sender":  h.cfg.SMSSender,
		"message": smsBody(code),
	}
	return h.postJSON(ctx, h.cfg.SMSAPIURL, h.cfg.SMSAPIKey, payload)
}

func (h *HTTPCodeProvider) postJSON(ctx context.Context, url, apiKey string, payload map[string]string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("provider http status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}
