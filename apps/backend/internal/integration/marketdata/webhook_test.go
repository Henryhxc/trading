package marketdata

import (
	"errors"
	"testing"
	"time"
)

func TestWebhookVerifierAcceptsKnownHMACVector(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	verifier := NewWebhookVerifier("synthetic-test-secret", 30*time.Second, func() time.Time { return now })
	body := []byte(`{"version":1,"quotes":[]}`)

	err := verifier.Verify(
		"1700000000",
		"delivery-123",
		"17391157ee8f694e0df8bf7f973d5930cfa9602822b20dd5f3ac605d0f239406",
		body,
	)
	if err != nil {
		t.Fatalf("Verify() error = %v, want nil", err)
	}
}

func TestWebhookVerifierRejectsInvalidSignatureBeforeAcceptingDelivery(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	verifier := NewWebhookVerifier("synthetic-test-secret", 30*time.Second, func() time.Time { return now })
	body := []byte(`{"version":1,"quotes":[]}`)

	if err := verifier.Verify("1700000000", "delivery-123", "not-a-signature", body); !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Fatalf("Verify() error = %v, want invalid signature", err)
	}
	if err := verifier.Verify("1700000000", "delivery-123", "17391157ee8f694e0df8bf7f973d5930cfa9602822b20dd5f3ac605d0f239406", body); err != nil {
		t.Fatalf("valid retry after invalid signature error = %v, want nil", err)
	}
}

func TestWebhookVerifierRejectsExpiredTimestamp(t *testing.T) {
	now := time.Unix(1700000100, 0).UTC()
	verifier := NewWebhookVerifier("synthetic-test-secret", 30*time.Second, func() time.Time { return now })

	err := verifier.Verify("1700000000", "delivery-123", "17391157ee8f694e0df8bf7f973d5930cfa9602822b20dd5f3ac605d0f239406", []byte(`{"version":1,"quotes":[]}`))
	if !errors.Is(err, ErrExpiredWebhook) {
		t.Fatalf("Verify() error = %v, want expired webhook", err)
	}
}

func TestWebhookVerifierRejectsReplayWithinAcceptanceWindow(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	verifier := NewWebhookVerifier("synthetic-test-secret", 30*time.Second, func() time.Time { return now })
	body := []byte(`{"version":1,"quotes":[]}`)
	signature := "17391157ee8f694e0df8bf7f973d5930cfa9602822b20dd5f3ac605d0f239406"

	if err := verifier.Verify("1700000000", "delivery-123", signature, body); err != nil {
		t.Fatalf("first Verify() error = %v, want nil", err)
	}
	if err := verifier.Verify("1700000000", "delivery-123", signature, body); !errors.Is(err, ErrWebhookReplay) {
		t.Fatalf("second Verify() error = %v, want replay", err)
	}
}
