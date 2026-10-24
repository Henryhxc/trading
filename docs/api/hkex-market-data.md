# HKEX Market Data Adapter Boundary

Phase 2 uses `MockMarketDataProvider` by default. Responses are marked `source: mock` and `data_status: simulated`; no exchange or broker connection is made.

Runtime selection uses the backend-only `MARKET_DATA_MODE` environment variable. Its default is `mock` in both the Go configuration and Docker Compose. `repository` reads the local MySQL market repository as a cache/prototype source; `akshare` reads cache rows populated by the separate Python scheduler through the signed internal webhook. Both are explicitly non-authorized sources, not HKEX feeds. Cached `live`, `simulated`, or unknown statuses are downgraded to `stale`; valid non-authorized statuses are limited to `unofficial`, `delayed`, `stale`, and `unavailable`.

There is currently no authorized HKEX or broker adapter. The optional AkShare scheduler makes outbound requests to its upstream and sends normalized batches only to the backend's authenticated internal webhook; it is not a Go provider, official HKEX feed, or live quote source. Values such as `authorized`, `public`, or any unknown `MARKET_DATA_MODE` cause backend configuration to fail at startup instead of silently selecting a provider. An authorized slot may be enabled only after a real licensed adapter exists and its contract and credentials have been reviewed. Provider credentials, when such an adapter is added, remain backend-only environment or secret-manager values.

The Go `MarketDataProvider` contract normalizes symbol, display name, prices, change, volume, turnover, `as_of`, source, and data status. Canonical internal symbols are five-character zero-padded Hong Kong codes such as `00700`; legacy five/six-digit input remains accepted and normalized where applicable. The optional isolated Python scheduler is for development/simulated use: it polls no more often than every 15 minutes and backfills daily/native-minute candles. It normalizes symbols and OHLCV, retains source/collection-time provenance, validates rows, and sends idempotent deliveries to Go; it is never an official HKEX feed or live quote source. Review upstream availability, terms, and schema changes before use. The scheduler has no published port and no user, account, or order permissions.

The provider composition supports the contract order authorized provider, explicitly non-official provider, then mock, but the present runtime deliberately leaves the authorized slot empty. Free sources may use only `unofficial`, `delayed`, `stale`, or `unavailable`; they must never return `live`. React calls only `/api/v1/markets/*` and never receives provider credentials.

## Phase 2 AkShare scheduler packet

The optional `akshare` Compose profile runs `services/akshare` as a Python scheduler. It has no `ports` mapping and therefore no host-facing or public business port. It only makes outbound requests to the upstream data source and to the backend's internal webhook:

`POST /api/v1/internal/market-data/akshare`

The image pins `akshare==1.19.3`. The scheduler uses `stock_hk_spot_em()` for the full spot universe, normalizes codes to five-character zero-padded values such as `00700`, and labels spot rows `source: akshare` and `data_status: delayed`. History rows use `stock_hk_hist(..., period="daily", adjust="")`, carry `adjustment: none`, and use `data_status: unofficial`. Neither path may emit `live` or `simulated`.

`AKSHARE_POLL_INTERVAL` is seconds and defaults to `900`; values below 900 are rejected. `AKSHARE_SYMBOLS`, `AKSHARE_HISTORY_START`, and `AKSHARE_HISTORY_END` are deployment inputs for the configured universe and history backfill. The scheduler retries webhook delivery up to three times with exponential backoff and jitter. A failed delivery does not mark the scheduler healthy.

Each delivery has these headers:

| Header | Meaning |
| --- | --- |
| `X-UX-Timestamp` | Unix seconds used for the replay window |
| `X-UX-Delivery-ID` | UUID idempotency/replay key |
| `X-UX-Signature` | Lowercase hex HMAC-SHA256 of `timestamp.delivery_id.raw_body` |

The secret is `AKSHARE_WEBHOOK_SECRET`, shared only through service environment/secret configuration. The backend independently verifies timestamp age, signature, and delivery replay before validation and transaction commit. Scheduler logs redact secrets, token-like fields, and webhook URL credentials. The container health check verifies the scheduler module can import; it does not expose a health HTTP server.

The payload is `{version: 1, quotes: [...], candles: [...]}`. Quote fields are `symbol`, `display_name`, `last_price`, `bid`, `ask`, `change`, `change_percent`, `volume`, `turnover`, `collected_at`, `source`, and `data_status`. Candle fields are `symbol`, `trade_date`, `open`, `high`, `low`, `close`, `volume`, `turnover`, `source`, `data_status`, `collected_at`, and `adjustment`. The Go service remains responsible for validation, idempotent MySQL upsert, and post-commit WebSocket publication; the scheduler never writes MySQL or user/order tables.

## Multi-period history extension (2026-10-24)

The detailed contract is `docs/design/candlestick-history.md`. React uses pinned Apache-2.0 `lightweight-charts@5.2.1`, with TradingView attribution retained, and calls only the internal candle endpoint. Periods are `1m/5m/15m/30m/60m/1d/1w/1M/1y`; the initial request covers the preceding six calendar months in Hong Kong time. Week/month/year bars aggregate available unadjusted daily rows by Hong Kong calendar boundaries. Native minutes are never synthesized from daily or mock data.

Explicit interval responses contain `symbol`, `interval`, `items`, `source`, `data_status`, and nullable `as_of`. They allow up to 5000 bars and a maximum 366-day window. Requests without `interval` preserve the legacy daily response and its default 90/maximum 365 limits, without series-level provenance fields; individual candles retain their provenance. Unknown securities return 404; known securities without native history return an empty `unavailable` series. The UI displays actual coverage and refreshes its current series every 60 seconds; this is cache reading, not a claim of one-minute upstream updates.

`HISTORY_PROVIDER` selects the helper's history strategy independently of Go's `MARKET_DATA_MODE`. `akshare` attempts AkShare history then public fallback; `public` directly requests Tencent unadjusted daily data and Yahoo native minutes, with Eastmoney minutes as a fallback. Unknown values fail startup. Compose defaults to `public` because AkShare/Eastmoney requests currently disconnect; standalone Python configuration retains the AkShare-first default. Quotes still independently attempt AkShare, and history success must never freshen or relabel a stale quote.

`AKSHARE_MINUTE_PERIODS` configures the native minute periods. The default daily backfill is six calendar months; `AKSHARE_HISTORY_START/END` override dates. Yahoo one-minute requests cover approximately five trading days, 5/15/30-minute requests use at most the recent 59 days, and hourly requests attempt the six-month window. These are requests rather than coverage guarantees. Missing OHLC rows are skipped, and the UI reports the actual returned range. All public data is non-official and for simulation only; no provider is presented as an HKEX feed.

Candle provenance is preserved as `akshare`, `tencent`, `eastmoney`, or `yahoo`. Tencent daily and Yahoo native-minute sources do not supply turnover; those rows store and return `null`, never zero or an estimate. Aggregates with any unknown turnover also return null. Migration `004_intraday_candles.sql` adds `stock_intraday_quotes` with unique `(stock_id, interval_code, bar_time)`; `005_nullable_daily_turnover.sql` permits unknown turnover in daily/minute history. Rolling back 004 deletes only minute history. Rolling back 005 requires real turnover to be present and deliberately fails on remaining nulls instead of fabricating zeros.

The signed candle input adds `interval` (default `1d`) and a UTC RFC3339 `timestamp` for minutes, with the Hong Kong `trade_date`. Deliveries are bounded to 400 records. The backend rejects duplicate keys within a batch, commits atomically, prevents older `collected_at` overwrites, and broadcasts `candles.updated` only after commit. Repeated collection updates existing keys idempotently. The helper remains on the internal network without a published port.
