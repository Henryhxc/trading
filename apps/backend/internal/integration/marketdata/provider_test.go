package marketdata

import (
	"context"
	"errors"
	"testing"
	"time"

	market "ux-securities/backend/internal/domain/market"
)

type providerFixture struct {
	quote   market.Quote
	candles []market.Candle
	err     error
}

func TestMockDirectoryListsAllSymbolsInStableOrder(t *testing.T) {
	items, err := NewMockProvider().Search(context.Background(), "")
	if err != nil || len(items) != 3 {
		t.Fatalf("directory size=%d err=%v, want 3", len(items), err)
	}
	for i, want := range []string{"00005", "00700", "09988"} {
		if items[i].Symbol != want {
			t.Fatalf("directory[%d]=%s, want %s", i, items[i].Symbol, want)
		}
	}
}

func (p providerFixture) Search(context.Context, string) ([]market.Symbol, error) {
	return nil, p.err
}
func (p providerFixture) Quote(context.Context, string) (market.Quote, error) {
	return p.quote, p.err
}
func (p providerFixture) Candles(context.Context, string, time.Time, time.Time, int) ([]market.Candle, error) {
	return p.candles, p.err
}

func TestProviderFallsBackToMockAndPreservesSimulatedStatus(t *testing.T) {
	price := 1.0
	want := market.Quote{Symbol: "00700", LastPrice: &price, Source: "mock", DataStatus: market.DataStatusSimulated}
	provider := New(providerFixture{err: errors.New("authorized timeout")}, nil, providerFixture{quote: want})
	got, err := provider.Quote(context.Background(), "00700")
	if err != nil || got.Source != "mock" || got.DataStatus != market.DataStatusSimulated {
		t.Fatalf("Quote() got=%+v err=%v, want mock simulated fallback", got, err)
	}
}

func TestProviderDowngradesStatusesNotAllowedForUnofficialSources(t *testing.T) {
	price := 1.0
	for _, status := range []market.DataStatus{market.DataStatusLive, market.DataStatusSimulated, "unexpected"} {
		t.Run(string(status), func(t *testing.T) {
			provider := New(nil, providerFixture{quote: market.Quote{
				Symbol:     "00700",
				LastPrice:  &price,
				Source:     "local-cache",
				DataStatus: status,
			}}, nil)

			got, err := provider.Quote(context.Background(), "00700")
			if err != nil {
				t.Fatalf("Quote() error = %v, want nil", err)
			}
			if got.DataStatus != market.DataStatusStale {
				t.Fatalf("Quote() status = %q, want stale safety downgrade", got.DataStatus)
			}
		})
	}
}

func TestProviderKeepsStaleUnofficialQuoteVisible(t *testing.T) {
	price := 380.2
	provider := New(
		nil,
		providerFixture{quote: market.Quote{
			Symbol:     "00700",
			LastPrice:  &price,
			Source:     "akshare",
			DataStatus: market.DataStatusStale,
		}},
		providerFixture{quote: market.Quote{
			Symbol:     "00700",
			LastPrice:  &price,
			Source:     "mock",
			DataStatus: market.DataStatusSimulated,
		}},
	)

	got, err := provider.Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("Quote() error = %v, want stale cache", err)
	}
	if got.Source != "akshare" || got.DataStatus != market.DataStatusStale {
		t.Fatalf("Quote() = %+v, want akshare stale quote", got)
	}
}

func TestProviderReturnsUnavailableStatusWhenAllSourcesFail(t *testing.T) {
	provider := New(providerFixture{err: errors.New("authorized down")}, providerFixture{err: errors.New("unofficial down")}, providerFixture{err: errors.New("mock down")})
	got, err := provider.Quote(context.Background(), "00700")
	if !errors.Is(err, ErrUnavailable) || got.DataStatus != market.DataStatusUnavailable {
		t.Fatalf("Quote() got=%+v err=%v, want unavailable status and error", got, err)
	}
}

func TestProviderPreservesUnknownQuoteThroughFallbackComposition(t *testing.T) {
	provider := New(
		providerFixture{err: market.ErrNotFound},
		providerFixture{err: market.ErrNotFound},
		providerFixture{err: market.ErrNotFound},
	)
	_, err := provider.Quote(context.Background(), "12345")
	if !errors.Is(err, market.ErrNotFound) {
		t.Fatalf("unknown quote error = %v, want market.ErrNotFound", err)
	}
}

func TestProviderFallsThroughRepositoryNotFoundToKnownMockQuote(t *testing.T) {
	price := 380.2
	provider := New(
		providerFixture{err: market.ErrNotFound}, // repository-shaped miss
		nil,
		providerFixture{
			quote: market.Quote{Symbol: "00700", LastPrice: &price, Source: "mock", DataStatus: market.DataStatusSimulated},
		},
	)
	got, err := provider.Quote(context.Background(), "00700")
	if err != nil || got.Source != "mock" || got.DataStatus != market.DataStatusSimulated {
		t.Fatalf("repository miss fallback quote = %+v, err=%v, want mock quote", got, err)
	}
}

func TestProviderKeepsKnownSymbolWithoutCachedQuoteUnavailable(t *testing.T) {
	provider := New(
		providerFixture{quote: market.Quote{
			Symbol:     "00700",
			Source:     "none",
			DataStatus: market.DataStatusUnavailable,
		}},
		nil,
		NewMockProvider(),
	)

	got, err := provider.Quote(context.Background(), "00700")
	if err != nil {
		t.Fatalf("Quote() error = %v, want explicit unavailable quote", err)
	}
	if got.LastPrice != nil || got.DataStatus != market.DataStatusUnavailable || got.Source != "none" {
		t.Fatalf("Quote() = %+v, want unavailable quote without a fabricated mock price", got)
	}
}

func TestProviderPreservesUnknownCandlesButEmptyKnownCandlesSucceed(t *testing.T) {
	provider := New(
		providerFixture{err: market.ErrNotFound},
		providerFixture{err: market.ErrNotFound},
		providerFixture{err: market.ErrNotFound},
	)
	_, err := provider.Candles(context.Background(), "12345", time.Time{}, time.Time{}, 30)
	if !errors.Is(err, market.ErrNotFound) {
		t.Fatalf("unknown candles error = %v, want market.ErrNotFound", err)
	}

	known := New(nil, nil, providerFixture{candles: []market.Candle{}})
	items, err := known.Candles(context.Background(), "00700", time.Time{}, time.Time{}, 30)
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("known empty candles = %#v, err=%v, want non-nil empty slice and nil error", items, err)
	}
}

func TestMockProviderUsesNotFoundForUnknownQuoteAndCandles(t *testing.T) {
	mock := NewMockProvider()
	if _, err := mock.Quote(context.Background(), "12345"); !errors.Is(err, market.ErrNotFound) {
		t.Fatalf("mock unknown quote error = %v, want market.ErrNotFound", err)
	}
	if _, err := mock.Candles(context.Background(), "12345", time.Time{}, time.Time{}, 30); !errors.Is(err, market.ErrNotFound) {
		t.Fatalf("mock unknown candles error = %v, want market.ErrNotFound", err)
	}
}

func TestComposedMockNativeCandlesDistinguishesUnknownSymbolFromMissingMinutes(t *testing.T) {
	provider := New(nil, nil, NewMockProvider())
	for _, interval := range []string{"1m", "5m", "15m", "30m", "60m"} {
		t.Run(interval, func(t *testing.T) {
			_, err := provider.NativeCandles(context.Background(), "12345", interval, time.Time{}, time.Time{}, 5000)
			if !errors.Is(err, market.ErrNotFound) {
				t.Fatalf("unknown native candles error = %v, want ErrNotFound", err)
			}
			items, err := provider.NativeCandles(context.Background(), "00700", interval, time.Time{}, time.Time{}, 5000)
			if err != nil || items == nil || len(items) != 0 {
				t.Fatalf("known mock native candles = %#v, err=%v, want empty without fabricated minutes", items, err)
			}
		})
	}
}
