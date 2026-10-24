package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	domain "ux-securities/backend/internal/domain/market"
	"ux-securities/backend/internal/integration/marketdata"
	marketservice "ux-securities/backend/internal/service/market"
)

const maxIngestionBody = 2 << 20

func (s *server) ingestAkShare(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxIngestionBody)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			s.fail(w, r, http.StatusRequestEntityTooLarge, "payload_too_large", "Market data batch is too large")
			return
		}
		s.fail(w, r, http.StatusBadRequest, "invalid_request", "Request body is invalid")
		return
	}
	err = s.verifier.Verify(
		r.Header.Get("X-UX-Timestamp"),
		r.Header.Get("X-UX-Delivery-ID"),
		r.Header.Get("X-UX-Signature"),
		body,
	)
	if err != nil {
		if errors.Is(err, marketdata.ErrWebhookReplay) {
			s.fail(w, r, http.StatusConflict, "webhook_replay", "Market data delivery was already accepted")
			return
		}
		s.fail(w, r, http.StatusUnauthorized, "invalid_webhook", "Market data delivery could not be authenticated")
		return
	}

	var batch domain.IngestionBatch
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&batch); err != nil || decoder.Decode(new(any)) != io.EOF {
		s.fail(w, r, http.StatusBadRequest, "invalid_request", "Request body is invalid")
		return
	}
	result, err := s.ingestion.Ingest(r.Context(), batch)
	if err != nil {
		if errors.Is(err, marketservice.ErrInvalidIngestion) {
			s.fail(w, r, http.StatusUnprocessableEntity, "invalid_market_data", "Market data batch failed validation")
			return
		}
		s.fail(w, r, http.StatusServiceUnavailable, "market_ingestion_failed", "Market data batch could not be committed")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
