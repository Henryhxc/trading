package config

import (
	"testing"
	"time"
)

func TestLoadDefaultsMarketDataModeToMock(t *testing.T) {
	t.Setenv("MARKET_DATA_MODE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.MarketDataMode != MarketDataModeMock {
		t.Fatalf("MarketDataMode = %q, want %q", cfg.MarketDataMode, MarketDataModeMock)
	}
	if cfg.MarketDataStaleAfter != 20*time.Minute {
		t.Fatalf("MarketDataStaleAfter = %s, want 20m", cfg.MarketDataStaleAfter)
	}
}

func TestLoadAcceptsRepositoryAsExplicitUnofficialMode(t *testing.T) {
	t.Setenv("MARKET_DATA_MODE", "repository")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.MarketDataMode != MarketDataModeRepository {
		t.Fatalf("MarketDataMode = %q, want %q", cfg.MarketDataMode, MarketDataModeRepository)
	}
}

func TestLoadRejectsModesWithoutAnImplementedAdapter(t *testing.T) {
	for _, mode := range []string{"authorized", "public", "something-else"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("MARKET_DATA_MODE", mode)

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() error = nil for unsupported mode %q, want fail-closed error", mode)
			}
		})
	}
}

func TestLoadAcceptsAkShareModeWithRealtimeSecuritySettings(t *testing.T) {
	t.Setenv("MARKET_DATA_MODE", "akshare")
	t.Setenv("AKSHARE_WEBHOOK_SECRET", "synthetic-test-secret")
	t.Setenv("MARKET_DATA_STALE_AFTER", "45s")
	t.Setenv("AKSHARE_WEBHOOK_MAX_AGE", "20s")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.MarketDataMode != MarketDataModeAkShare {
		t.Fatalf("MarketDataMode = %q, want akshare", cfg.MarketDataMode)
	}
	if cfg.AkShareWebhookSecret != "synthetic-test-secret" {
		t.Fatal("AkShareWebhookSecret was not loaded from the environment")
	}
	if cfg.MarketDataStaleAfter != 45*time.Second || cfg.AkShareWebhookMaxAge != 20*time.Second {
		t.Fatalf("durations = %s/%s, want 45s/20s", cfg.MarketDataStaleAfter, cfg.AkShareWebhookMaxAge)
	}
	if cfg.FrontendOrigin != "http://localhost:3000" {
		t.Fatalf("FrontendOrigin = %q, want configured origin", cfg.FrontendOrigin)
	}
}

func TestLoadRejectsAkShareModeWithoutWebhookSecret(t *testing.T) {
	t.Setenv("MARKET_DATA_MODE", "akshare")
	t.Setenv("AKSHARE_WEBHOOK_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want fail-closed missing secret error")
	}
}
