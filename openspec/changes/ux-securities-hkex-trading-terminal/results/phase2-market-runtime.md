# Phase 2 Market Runtime Verification

Date: 2026-10-24 (Asia/Hong_Kong)
Status: `GREEN_API_DB; BROWSER_PENDING`

## Runtime environment

- Compose project: `ux-securities-hkex-trading-terminal`; rebuilt from the Phase 2 implementation with `docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml up -d --build`.
- Services remained on the configured default ports: frontend `3000`, backend `8080`, MySQL `3306`, phpMyAdmin `8081`.
- All user identities used in smoke were synthetic. No real market provider, broker or order API was called.

## Targeted and full tests

```text
$ cd apps/backend && go test ./internal/integration/marketdata ./internal/service/market ./internal/service/watchlist ./internal/repository/mysqlrepo ./...
ok   ux-securities/backend/internal/integration/marketdata
ok   ux-securities/backend/internal/service/market
ok   ux-securities/backend/internal/service/watchlist
ok   ux-securities/backend/internal/repository/mysqlrepo
?    ux-securities/backend/cmd/api [no test files]
?    ux-securities/backend/internal/config [no test files]
?    ux-securities/backend/internal/domain/auth [no test files]
?    ux-securities/backend/internal/domain/market [no test files]
?    ux-securities/backend/internal/domain/watchlist [no test files]
ok   ux-securities/backend/internal/httpapi
ok   ux-securities/backend/internal/migrations
ok   ux-securities/backend/internal/service/auth

exit_code=0
```

Focused tests added in this lane:

- `internal/service/market/service_test.go`: query trimming, provider error propagation, candle delegation.
- `internal/service/watchlist/service_test.go`: actor propagation, symbol normalization, quote enrichment and delete scoping.
- `internal/repository/mysqlrepo/watchlist_test.go`: actor predicates, duplicate upsert and delete SQL guard.
- `internal/integration/marketdata/provider_test.go`: mock fallback, simulated status and unavailable status.

## API smoke

The first smoke run against the stale pre-Phase-2 image returned 404 for all market/watchlist paths. After the rebuild, the same synthetic-user smoke produced:

```text
health [200,{"status":"ok"}]
register_a [201,{"user":{"id":25,"username":"p2qa...","email":"p2qa...@example.test"}}]
register_b [201,{"user":{"id":26,"username":"p2qb...","email":"p2qb...@example.test"}}]
search_a [200,{"items":[{"symbol":"00700","display_name":"Tencent Holdings"}]}]
search_empty [200,{"items":[]}]
quote_a [200,{"symbol":"00700","display_name":"Tencent Holdings","last_price":380.2,"source":"mock","data_status":"simulated"}]
candles_a [200,{"interval":"1d","items":[],"symbol":"00700"}]
watch_add_a [400,{"code":"invalid_input","message":"Check the submitted details"}]  # input was intentionally "700"
watch_add_duplicate_a [201,{"id":1,"symbol":"00700",...}]
watch_list_a [200,{"items":[{"symbol":"00700",...,"data_status":"simulated"}]}]
watch_list_b_isolated [200,{"items":[]}]
watch_delete_a [204,null]
watch_list_a_empty [200,{"items":[]}]
```

A second proper duplicate test with `00700` returned 201 for the first add and 201 with the same `id` for the second add, then one list row. This verifies idempotency. The first test's `700` rejection is an API validation behavior: the route requires a five- or six-digit symbol; frontend submits six-digit symbols from search results.

Search, quote, empty search, watchlist isolation/list/delete, and mock source/status passed. Daily candles returned the contract-valid empty array because the current `stock_daily_quotes` seed contains no rows; this is an explicit empty-state condition, not a fabricated series.

## Database and migration evidence

Current app schema SQL inspection after rebuild:

```text
SHOW TABLES: accounts, schema_migrations, sessions, stock_daily_quotes, stocks, users, watchlists
schema_migrations: 001_users_sessions_accounts.sql, 002_stocks_daily_quotes.sql, 003_watchlists.sql
stocks: 3
stock_daily_quotes: 0
watchlists: 1 (from synthetic smoke fixture before deletion)
```

An isolated schema `ux_phase2_qa_20261024` was created, granted only to the app DB user, and removed after verification. The backend migration runner produced all seven expected tables and three ledger versions. Applying down migrations in reverse order left only `schema_migrations` with zero rows; rerunning the backend migration runner recreated all tables and all three ledger versions. Full SQL and outputs are in [`2026-10-24-phase2-market.md`](db-change-reports/2026-10-24-phase2-market.md).

## Browser evidence

The required in-app browser (IAB) surface was unavailable in this environment (`Browser is not available: iab`; browser list was empty). Per the browser skill, no alternate browser surface was substituted. A separate IAB session had a local login page open and requested confirmation before entering synthetic credentials; no credential entry or screenshot action was performed from this lane without that confirmation.

Therefore the following remain `PENDING`: IAB screenshots at 1440x900, 1280x800 and 390x844; search/quote/candles/watchlist empty/error/stale/unavailable visual states; scroll-anchor and mobile overflow inspection. No screenshot files are claimed by this report.

## Residual risks

- No daily candle seed rows means the populated chart state is not visually/runtime-proven yet.
- Free/unofficial provider adapters are not configured; mock responses correctly report `simulated` and no `live` status was observed.
- The `700` watchlist request is rejected rather than normalized; this matches the route validator but should remain aligned with the shared symbol contract.
