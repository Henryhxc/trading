# Phase 2 TDD Handoff

Status: `GREEN_API_DB; BROWSER_PENDING`

## RED evidence

- Test: [`tests/phase2_market_red_test.py`](../../../../tests/phase2_market_red_test.py)
- Result: [`results/phase2-market-red.md`](../results/phase2-market-red.md)
- Command: `python3 tests/phase2_market_red_test.py`
- Result: exit code 1; five tests ran and failed as expected.

The failures identify absent `MarketDataProvider`/quote status domain types, fallback adapter, market handler, watchlist repository/domain files, and targeted backend test files. The suite also asserts the existing OpenAPI market/watchlist route contract and will continue to guard against free-source `live` status.

## GREEN and runtime evidence

- Added and passed the expected focused Go tests: `service/market/service_test.go`, `service/watchlist/service_test.go`, and `repository/mysqlrepo/watchlist_test.go`; also added provider fallback/status tests under `integration/marketdata/provider_test.go`.
- Re-ran the original `python3 tests/phase2_market_red_test.py` contract after implementation: 5 tests passed, exit code 0. The historical RED output remains unchanged in `results/phase2-market-red.md`.
- Targeted plus full command passed: `cd apps/backend && go test ./internal/integration/marketdata ./internal/service/market ./internal/service/watchlist ./internal/repository/mysqlrepo ./...` (exit 0).
- Rebuilt Compose and reran smoke. Search, empty search, mock quote, watchlist idempotency, actor isolation, delete and empty list passed. Quote returned `source: mock`, `data_status: simulated`. Candles returned contract-valid empty `items` because no daily rows are seeded.
- An isolated `ux_phase2_qa_20261024` migration run passed apply, reverse down (ledger rows removed), and runner reapply; the temporary schema/grant were removed. See [`results/phase2-market-runtime.md`](../results/phase2-market-runtime.md) and [`results/db-change-reports/2026-10-24-phase2-market.md`](../results/db-change-reports/2026-10-24-phase2-market.md).

## Remaining Phase 2 work

- Code lane: implement the smallest provider/status/fallback, search/quote/candles, watchlist migration/repository/service/handler and frontend workbench slice; then run targeted GREEN tests and API smoke.
- Runtime/browser lane: exercise mock/unofficial/delayed/stale/unavailable fixtures, symbol search, quote, candles, watchlist empty/error/idempotency/user isolation, and capture user-selected in-app browser screenshots at 1440x900, 1280x800 and 390x844. Label evidence `IAB`. IAB is currently unavailable in this environment (`Browser is not available: iab`); no alternate browser was substituted. A separate IAB session also requires confirmation before synthetic credentials are entered.
- Review lane: verify no direct third-party requests from React, no free-source `live` status, actor predicates, migration rollback and visual/error/overflow evidence.

No production code, database, OpenAPI, global tasks, orchestration state, or running service was modified by this TDD lane. Phase 2 must remain pending until GREEN, API smoke, browser evidence and independent review are complete.
