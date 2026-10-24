package marketdata

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidWebhook          = errors.New("invalid webhook metadata")
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
	ErrExpiredWebhook          = errors.New("expired webhook")
	ErrWebhookReplay           = errors.New("webhook replay")
)

type WebhookVerifier struct {
	secret []byte
	maxAge time.Duration
	now    func() time.Time

	mu   sync.Mutex
	seen map[string]time.Time
}

func NewWebhookVerifier(secret string, maxAge time.Duration, now func() time.Time) *WebhookVerifier {
	return &WebhookVerifier{
		secret: []byte(secret),
		maxAge: maxAge,
		now:    now,
		seen:   make(map[string]time.Time),
	}
}

func (v *WebhookVerifier) Verify(timestamp, deliveryID, signature string, body []byte) error {
	if timestamp == "" || deliveryID == "" || len(deliveryID) > 128 || signature == "" {
		return ErrInvalidWebhook
	}
	unixSeconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrInvalidWebhook
	}
	now := v.now().UTC()
	age := now.Sub(time.Unix(unixSeconds, 0).UTC())
	if age < -v.maxAge || age > v.maxAge {
		return ErrExpiredWebhook
	}

	provided, err := hex.DecodeString(strings.TrimSpace(signature))
	if err != nil || len(provided) != sha256.Size {
		return ErrInvalidWebhookSignature
	}
	mac := hmac.New(sha256.New, v.secret)
	_, _ = mac.Write([]byte(timestamp))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write([]byte(deliveryID))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(body)
	if !hmac.Equal(provided, mac.Sum(nil)) {
		return ErrInvalidWebhookSignature
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	for id, expiresAt := range v.seen {
		if !expiresAt.After(now) {
			delete(v.seen, id)
		}
	}
	if _, exists := v.seen[deliveryID]; exists {
		return ErrWebhookReplay
	}
	v.seen[deliveryID] = now.Add(v.maxAge)
	return nil
}
