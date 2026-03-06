package verification

import (
	"testing"

	"football-backend/common/config"
)

func TestNewCodeProviderFromConfigMock(t *testing.T) {
	p, err := NewCodeProviderFromConfig(config.VerificationConfig{Provider: "mock"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.Mode() != "mock" {
		t.Fatalf("expected mock mode, got %s", p.Mode())
	}
}

func TestNewCodeProviderFromConfigRealMissingURL(t *testing.T) {
	_, err := NewCodeProviderFromConfig(config.VerificationConfig{Provider: "real"})
	if err == nil {
		t.Fatal("expected error when real mode urls are missing")
	}
}

func TestNewCodeProviderFromConfigRealSuccess(t *testing.T) {
	p, err := NewCodeProviderFromConfig(config.VerificationConfig{
		Provider:       "real",
		HTTPTimeoutSec: 3,
		Email: config.VerificationEmailConfig{
			APIURL: "http://localhost/email",
		},
		SMS: config.VerificationSMSConfig{
			APIURL: "http://localhost/sms",
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.Mode() != "real" {
		t.Fatalf("expected real mode, got %s", p.Mode())
	}
}
