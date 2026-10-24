package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"ux-securities/backend/internal/config"
	market "ux-securities/backend/internal/domain/market"
	"ux-securities/backend/internal/httpapi"
	"ux-securities/backend/internal/integration/marketdata"
	"ux-securities/backend/internal/migrations"
	"ux-securities/backend/internal/repository/mysqlrepo"
	"ux-securities/backend/internal/service/auth"
	marketservice "ux-securities/backend/internal/service/market"
	trading "ux-securities/backend/internal/service/trading"
	watchlistservice "ux-securities/backend/internal/service/watchlist"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("mysql", cfg.DatabaseDSN())
	if err != nil {
		log.Fatal("database configuration failed")
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for {
		if err = db.PingContext(ctx); err == nil {
			break
		}
		if ctx.Err() != nil {
			log.Fatal("database is unavailable")
		}
		time.Sleep(time.Second)
	}
	if err = migrations.Apply(ctx, db, cfg.MigrationsDir); err != nil {
		log.Fatal("database migrations failed")
	}

	repo := mysqlrepo.New(db)
	service := auth.NewService(repo, cfg.SessionTTL, cfg.InitialCash)
	marketRepo := mysqlrepo.NewMarketRepository(db)
	provider, err := buildMarketDataProvider(cfg.MarketDataMode, marketRepo)
	if err != nil {
		log.Fatal(err)
	}
	markets := buildMarketService(cfg, provider, time.Now)
	watchlists := watchlistservice.NewService(marketRepo, markets)
	tradingRepo := mysqlrepo.NewTradingRepository(db)
	trades := trading.NewService(tradingRepo, markets)
	stream := httpapi.NewMarketStreamHub(cfg.FrontendOrigin)
	var ingestion httpapi.IngestionAPI
	var verifier httpapi.WebhookVerification
	if cfg.MarketDataMode == config.MarketDataModeAkShare {
		ingestion = marketservice.NewIngestionService(marketRepo, stream)
		verifier = marketdata.NewWebhookVerifier(cfg.AkShareWebhookSecret, cfg.AkShareWebhookMaxAge, time.Now)
	}
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewWithTradingRealtime(service, cfg.CookieSecure, markets, watchlists, trades, ingestion, verifier, stream),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("api listening port=%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http server stopped")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	_ = server.Shutdown(shutdown)
}

func buildMarketService(cfg config.Config, provider market.MarketDataProvider, now func() time.Time) *marketservice.Service {
	staleAfter := cfg.MarketDataStaleAfter
	if staleAfter <= 0 {
		staleAfter = 30 * time.Second
	}
	return marketservice.NewServiceWithClock(provider, staleAfter, now)
}

func buildMarketDataProvider(mode config.MarketDataMode, repository market.MarketDataProvider) (market.MarketDataProvider, error) {
	mock := marketdata.NewMockProvider()
	switch mode {
	case config.MarketDataModeMock:
		return marketdata.New(nil, nil, mock), nil
	case config.MarketDataModeRepository:
		if repository == nil {
			return nil, fmt.Errorf("MARKET_DATA_MODE %q requires a repository provider", mode)
		}
		return marketdata.New(nil, repository, mock), nil
	case config.MarketDataModeAkShare:
		if repository == nil {
			return nil, fmt.Errorf("MARKET_DATA_MODE %q requires a repository provider", mode)
		}
		return marketdata.New(nil, repository, mock), nil
	default:
		return nil, fmt.Errorf("unsupported MARKET_DATA_MODE %q", mode)
	}
}
