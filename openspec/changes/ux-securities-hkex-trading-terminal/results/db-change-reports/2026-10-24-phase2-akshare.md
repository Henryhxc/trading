# Phase 2 AkShare Market Data Report

Date: 2026-10-24 (Asia/Singapore)

## Database contract

- Target DB: `ux_securities`
- Migrations: `002_stocks_daily_quotes_watchlists.sql`, `002_webhook_market_provenance.sql`
- Tables remain the required wide tables: `stocks`, `stock_daily_quotes`, `watchlists`.
- Provenance is stored on `stock_daily_quotes`; no ninth business table was introduced.

## Verification

The runtime ledger and table output are in `results/db-runtime-verification.txt`.
Focused Go tests cover HMAC verification, replay rejection, validation, transactional
upsert, stale/older update protection, rollback without broadcast, and idempotency.
The synthetic AkShare-compatible fixture uses `00700`, `source=akshare`, and delayed or
unofficial status; no payload is marked `live`.

## Live upstream result

The real container pull was attempted with `ak.stock_hk_spot_em()` and failed at the
upstream HTTP request with `requests.exceptions.ConnectionError` / `RemoteDisconnected`.
The complete failure is preserved in `results/akshare-live-sample.txt`; therefore no
real upstream sample or real AkShare-to-MySQL delivery is claimed as passed.

## Rollback and risks

The canonical Phase 2 down migration and provenance down migration are present and
covered by isolated migration tests. The WebSocket broadcast is post-commit and
best-effort without an outbox. AkShare is unofficial/delayed and remains subject to
upstream availability, field changes, and rate limits.
