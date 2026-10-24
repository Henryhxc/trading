package market

import (
	"context"
	"errors"
	"time"
)

// ErrUnavailable identifies a provider or dependency outage.
var ErrUnavailable = errors.New("market data unavailable")

// ErrNotFound identifies a symbol that is not present in the market universe.
var ErrNotFound = errors.New("market symbol not found")

type DataStatus string

const (
	DataStatusSimulated   DataStatus = "simulated"
	DataStatusUnofficial  DataStatus = "unofficial"
	DataStatusDelayed     DataStatus = "delayed"
	DataStatusStale       DataStatus = "stale"
	DataStatusUnavailable DataStatus = "unavailable"
	DataStatusLive        DataStatus = "live"
)

type Symbol struct {
	Symbol      string `json:"symbol"`
	DisplayName string `json:"display_name"`
	Quote       *Quote `json:"quote,omitempty"`
}

type Quote struct {
	Symbol        string     `json:"symbol"`
	DisplayName   string     `json:"display_name"`
	LastPrice     *float64   `json:"last_price"`
	Bid           *float64   `json:"bid"`
	Ask           *float64   `json:"ask"`
	Change        *float64   `json:"change"`
	ChangePercent *float64   `json:"change_percent"`
	Volume        int64      `json:"volume"`
	Turnover      float64    `json:"turnover"`
	AsOf          time.Time  `json:"as_of"`
	Source        string     `json:"source"`
	DataStatus    DataStatus `json:"data_status"`
}

type Candle struct {
	TradeDate   time.Time  `json:"trade_date"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	Open        float64    `json:"open"`
	High        float64    `json:"high"`
	Low         float64    `json:"low"`
	Close       float64    `json:"close"`
	Volume      int64      `json:"volume"`
	Turnover    *float64   `json:"turnover"`
	Source      string     `json:"source"`
	DataStatus  DataStatus `json:"data_status"`
	CollectedAt time.Time  `json:"collected_at"`
}

type CandleSeries struct {
	Symbol     string     `json:"symbol"`
	Interval   string     `json:"interval"`
	Items      []Candle   `json:"items"`
	Source     string     `json:"source"`
	DataStatus DataStatus `json:"data_status"`
	AsOf       *time.Time `json:"as_of"`
}

// NativeCandleProvider returns actual bars for the requested intraday period.
// Missing minute data must never be replaced by daily or synthetic bars.
type NativeCandleProvider interface {
	NativeCandles(context.Context, string, string, time.Time, time.Time, int) ([]Candle, error)
}

func IsMinuteInterval(value string) bool {
	switch value {
	case "1m", "5m", "15m", "30m", "60m":
		return true
	}
	return false
}

func IsCandleInterval(value string) bool {
	return IsMinuteInterval(value) || value == "1d" || value == "1w" || value == "1M" || value == "1y"
}

type MarketDataProvider interface {
	Search(context.Context, string) ([]Symbol, error)
	Quote(context.Context, string) (Quote, error)
	Candles(context.Context, string, time.Time, time.Time, int) ([]Candle, error)
}
