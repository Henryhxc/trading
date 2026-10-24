package market

import "time"

type IngestionBatch struct {
	Version int            `json:"version"`
	Quotes  []QuoteUpdate  `json:"quotes"`
	Candles []CandleUpdate `json:"candles"`
}

type QuoteUpdate struct {
	Symbol        string     `json:"symbol"`
	DisplayName   string     `json:"display_name"`
	LastPrice     string     `json:"last_price"`
	Bid           *string    `json:"bid"`
	Ask           *string    `json:"ask"`
	Change        string     `json:"change"`
	ChangePercent string     `json:"change_percent"`
	Volume        string     `json:"volume"`
	Turnover      string     `json:"turnover"`
	CollectedAt   time.Time  `json:"collected_at"`
	Source        string     `json:"source"`
	DataStatus    DataStatus `json:"data_status"`
}

type CandleUpdate struct {
	Symbol      string     `json:"symbol"`
	TradeDate   string     `json:"trade_date"`
	Interval    string     `json:"interval,omitempty"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	Open        string     `json:"open"`
	High        string     `json:"high"`
	Low         string     `json:"low"`
	Close       string     `json:"close"`
	Volume      string     `json:"volume"`
	Turnover    string     `json:"turnover"`
	Source      string     `json:"source"`
	DataStatus  DataStatus `json:"data_status"`
	CollectedAt time.Time  `json:"collected_at"`
	Adjustment  string     `json:"adjustment"`
}

type MarketEvent struct {
	Type       string     `json:"type"`
	Symbol     string     `json:"symbol"`
	Interval   string     `json:"interval,omitempty"`
	Source     string     `json:"source"`
	DataStatus DataStatus `json:"data_status"`
	AsOf       time.Time  `json:"as_of"`
	Quote      *Quote     `json:"quote,omitempty"`
}

type IngestionResult struct {
	QuoteCount  int           `json:"quote_count"`
	CandleCount int           `json:"candle_count"`
	Events      []MarketEvent `json:"-"`
}
