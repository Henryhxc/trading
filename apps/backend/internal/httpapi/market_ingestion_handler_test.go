package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/market"
	"ux-securities/backend/internal/integration/marketdata"
)

type ingestionAPIStub struct {
	batch domain.IngestionBatch
	calls int
}

func (s *ingestionAPIStub) Ingest(_ context.Context, batch domain.IngestionBatch) (domain.IngestionResult, error) {
	s.calls++
	s.batch = batch
	return domain.IngestionResult{QuoteCount: len(batch.Quotes), CandleCount: len(batch.Candles)}, nil
}

func TestAkShareWebhookVerifiesRawBodyBeforeIngestion(t *testing.T) {
	body := []byte(`{"version":1,"quotes":[{"symbol":"00700","display_name":"Tencent Holdings","last_price":"380.2000","change":"2.1000","change_percent":"0.5543","volume":"1200000","turnover":"456240000.0000","collected_at":"2023-11-14T22:13:20Z","source":"akshare","data_status":"delayed"}],"candles":[]}`)
	ingestion := &ingestionAPIStub{}
	verifier := marketdata.NewWebhookVerifier("synthetic-test-secret", 30*time.Second, fixedWebhookNow)
	handler := NewWithRealtime(&testAuthService{}, false, nil, nil, ingestion, verifier, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/internal/market-data/akshare", bytes.NewReader(body))
	request.Header.Set("X-UX-Timestamp", "1700000000")
	request.Header.Set("X-UX-Delivery-ID", "delivery-http-1")
	request.Header.Set("X-UX-Signature", webhookSignature("synthetic-test-secret", "1700000000", "delivery-http-1", body))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("webhook status = %d body=%s, want 200", response.Code, response.Body.String())
	}
	if ingestion.calls != 1 || len(ingestion.batch.Quotes) != 1 || ingestion.batch.Quotes[0].Symbol != "00700" {
		t.Fatalf("ingestion side effect = calls:%d batch:%#v", ingestion.calls, ingestion.batch)
	}
}

func TestAkShareWebhookRejectsInvalidSignatureBeforeIngestion(t *testing.T) {
	body := []byte(`{"version":1,"quotes":[],"candles":[]}`)
	ingestion := &ingestionAPIStub{}
	verifier := marketdata.NewWebhookVerifier("synthetic-test-secret", 30*time.Second, fixedWebhookNow)
	handler := NewWithRealtime(&testAuthService{}, false, nil, nil, ingestion, verifier, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/internal/market-data/akshare", bytes.NewReader(body))
	request.Header.Set("X-UX-Timestamp", "1700000000")
	request.Header.Set("X-UX-Delivery-ID", "delivery-http-2")
	request.Header.Set("X-UX-Signature", "bad")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("invalid signature status = %d, want 401", response.Code)
	}
	if ingestion.calls != 0 {
		t.Fatalf("ingestion calls = %d, want 0", ingestion.calls)
	}
}

func fixedWebhookNow() time.Time { return time.Unix(1700000000, 0).UTC() }

func webhookSignature(secret, timestamp, deliveryID string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "." + deliveryID + "."))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
