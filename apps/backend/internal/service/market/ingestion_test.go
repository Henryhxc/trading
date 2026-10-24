package market

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "ux-securities/backend/internal/domain/market"
)

type ingestionRepositoryStub struct {
	result domain.IngestionResult
	err    error
	calls  int
}

func (r *ingestionRepositoryStub) UpsertMarketBatch(context.Context, domain.IngestionBatch) (domain.IngestionResult, error) {
	r.calls++
	return r.result, r.err
}

type ingestionPublisherStub struct{ events []domain.MarketEvent }

func (p *ingestionPublisherStub) Publish(event domain.MarketEvent) {
	p.events = append(p.events, event)
}

func validIngestionBatch() domain.IngestionBatch {
	return domain.IngestionBatch{
		Version: 1,
		Quotes: []domain.QuoteUpdate{{
			Symbol: "00700", DisplayName: "Tencent Holdings",
			LastPrice: "380.2000", Change: "2.1000", ChangePercent: "0.5543",
			Volume: "1200000", Turnover: "456240000.0000",
			CollectedAt: time.Date(2026, 10, 24, 2, 0, 0, 0, time.UTC),
			Source:      "akshare", DataStatus: domain.DataStatusDelayed,
		}},
		Candles: []domain.CandleUpdate{{
			Symbol: "00700", TradeDate: "2026-10-23",
			Open: "376.0000", High: "382.0000", Low: "375.0000", Close: "380.2000",
			Volume: "1000000", Turnover: "379000000.0000",
			Source: "akshare", DataStatus: domain.DataStatusUnofficial,
			CollectedAt: time.Date(2026, 10, 24, 2, 0, 0, 0, time.UTC), Adjustment: "none",
		}},
	}
}

func TestIngestionRejectsLiveAkShareDataBeforeRepository(t *testing.T) {
	batch := validIngestionBatch()
	batch.Quotes[0].DataStatus = domain.DataStatusLive
	repository := &ingestionRepositoryStub{}
	service := NewIngestionService(repository, &ingestionPublisherStub{})

	if _, err := service.Ingest(context.Background(), batch); !errors.Is(err, ErrInvalidIngestion) {
		t.Fatalf("Ingest() error = %v, want invalid ingestion", err)
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0", repository.calls)
	}
}

func TestIngestionAcceptsExplicitPublicCandleSources(t *testing.T) {
	for _, source := range []string{"tencent", "eastmoney", "yahoo"} {
		batch := validIngestionBatch()
		batch.Quotes = nil
		batch.Candles[0].Source = source
		if source == "tencent" || source == "yahoo" {
			batch.Candles[0].Turnover = "" // JSON null, never a made-up zero.
		}
		if err := validateIngestionBatch(batch); err != nil {
			t.Fatalf("source %s rejected: %v", source, err)
		}
	}
	for _, source := range []string{"unknown", "mock"} {
		batch := validIngestionBatch()
		batch.Candles[0].Source = source
		if !errors.Is(validateIngestionBatch(batch), ErrInvalidIngestion) {
			t.Fatalf("unsupported source %s accepted", source)
		}
	}
}

func TestIngestionRejectsInvalidOHLCBeforeRepository(t *testing.T) {
	batch := validIngestionBatch()
	batch.Candles[0].Low = "381.0000"
	repository := &ingestionRepositoryStub{}
	service := NewIngestionService(repository, &ingestionPublisherStub{})

	if _, err := service.Ingest(context.Background(), batch); !errors.Is(err, ErrInvalidIngestion) {
		t.Fatalf("Ingest() error = %v, want invalid ingestion", err)
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0", repository.calls)
	}
}

func TestIngestionDoesNotPublishWhenRepositoryTransactionFails(t *testing.T) {
	repository := &ingestionRepositoryStub{err: errors.New("commit failed")}
	publisher := &ingestionPublisherStub{}
	service := NewIngestionService(repository, publisher)

	if _, err := service.Ingest(context.Background(), validIngestionBatch()); err == nil {
		t.Fatal("Ingest() error = nil, want transaction failure")
	}
	if len(publisher.events) != 0 {
		t.Fatalf("published events = %d, want 0 after failed transaction", len(publisher.events))
	}
}

func TestIngestionPublishesRepositoryResultAfterSuccessfulTransaction(t *testing.T) {
	want := domain.MarketEvent{Type: "quote.updated", Symbol: "00700", Source: "akshare", DataStatus: domain.DataStatusDelayed}
	repository := &ingestionRepositoryStub{result: domain.IngestionResult{Events: []domain.MarketEvent{want}}}
	publisher := &ingestionPublisherStub{}
	service := NewIngestionService(repository, publisher)

	result, err := service.Ingest(context.Background(), validIngestionBatch())
	if err != nil {
		t.Fatalf("Ingest() error = %v, want nil", err)
	}
	if result.QuoteCount != 1 || result.CandleCount != 1 {
		t.Fatalf("result counts = %d/%d, want 1/1", result.QuoteCount, result.CandleCount)
	}
	if len(publisher.events) != 1 || publisher.events[0] != want {
		t.Fatalf("published events = %#v, want %#v", publisher.events, []domain.MarketEvent{want})
	}
}
