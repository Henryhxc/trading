# Phase 2 Independent Review (Error/Empty-State Rework)

Date: 2026-10-24 (Asia/Hong_Kong)
Recommendation: `FAIL`

This review covers the current Phase 2 implementation after the error/empty-state
rework. It rechecked the shared OpenAPI contract, provider/handler behavior,
frontend outcome states, tests, handoffs, runtime/DB evidence, orchestration
state, and the required screenshot directory. No production implementation or
task checkbox was changed by this review. The selected browser surface remains
the in-app browser (IAB); no Chrome or Computer Use evidence is claimed.

## Findings

1. **[P1] Required IAB workbench evidence is still missing.**
   `results/screenshots/phase2-market-workbench/` contains only `.gitkeep`, and
   `orchestration-state.md` and `phase2-error-empty-rework-handoff.md` keep the
   browser gate pending. The required IAB screenshots and interaction evidence
   at 1440x900, 1280x800, and 390x844 therefore remain absent for search,
   quote, candles, favorites add/delete/reorder, zero-match/error/unavailable
   states, mobile overflow, and scroll-anchor restoration. Phase 2 cannot be
   accepted on terminal/API checks alone.

2. **[P1] Unknown symbols do not remain 404 on the production provider path.**
   `apps/backend/internal/integration/marketdata/provider.go:39-57` retries
   every provider error and returns `market.ErrUnavailable` after the mock
   provider also rejects an unknown symbol. In the production wiring,
   `apps/backend/cmd/api/main.go:51-56` supplies the MySQL repository and mock,
   so an unknown quote such as `12345` reaches `getQuote`, which maps
   `ErrUnavailable` to HTTP 503 (`apps/backend/internal/httpapi/market_handler.go:29-37`)
   rather than the OpenAPI 404 branch. The same distinction is not preserved
   by the provider contract: an outage and an unknown security are collapsed
   into one sentinel. Add a typed not-found result (or preserve `sql.ErrNoRows`
   as a domain error) and test the real provider composition so outages map to
   503 while unknown symbols map to 404.

3. **[P1] Unknown candle symbols can return a successful empty series instead
   of 404.** `MarketRepository.Candles` (`apps/backend/internal/repository/
   mysqlrepo/market.go:45-74`) returns an empty slice with nil error when the
   symbol has no matching stock row. `Provider.Candles` then treats that as a
   successful provider response, and `getDailyCandles` returns 200 with empty
   `items` (`market_handler.go:55-64`). This violates the declared 404
   not-found behavior for `/markets/candles/{symbol}` and makes an unknown
   security indistinguishable from a known security with no history. Validate
   symbol existence before reading candles or return a typed not-found error,
   while keeping provider outages as `ErrUnavailable`/503.

4. **[P2] OpenAPI omits the candle provider-outage response.** The handler
   returns 503 `market_unavailable` for `ErrUnavailable`
   (`apps/backend/internal/httpapi/market_handler.go:55-60`), but
   `docs/api/openapi.yaml:142-172` documents only 200, 400, 401, and 404 for
   `/api/v1/markets/candles/{symbol}`. Add the shared `Unavailable` response
   so the contract describes the exercised outage branch and generated clients
   can model it.

## Rechecked and passing

- `python3 tests/phase2_error_empty_rework_red.py` PASS (2 tests): shared
  `ErrUnavailable` mapping exists and successful zero-match search renders a
  distinct frontend empty state.
- `python3 tests/phase2_rework_contract_test.py` PASS (3 tests).
- `python3 tests/phase2_market_red_test.py` PASS (5 source-contract tests).
- `cd apps/backend && go test ./... -count=1` PASS.
- `cd apps/backend && go vet ./...` PASS.
- `cd apps/frontend && npm run build` PASS.
- OpenAPI YAML parses successfully. Search and watchlist failure responses are
  present; candle 503 remains the exception noted above.
- Prior migration/API smoke and isolated rollback/reapply evidence remains in
  `results/phase2-market-runtime.md` and
  `results/db-change-reports/2026-10-24-phase2-market.md`.

## Evidence and residual risk

- IAB browser evidence is **pending** and must remain a separate gate. No
  alternate browser surface may substitute for it, and no screenshot is
  labeled as Chrome or Computer Use.
- Runtime evidence exercises the mock `simulated` provider and an empty daily
  history. Populated candles, stale/unavailable visual states, and the unknown
  symbol status split are not runtime-proven.
- `0.4`, Phase 2 task checkboxes, and Phase 3 remain gated; this review did not
  change them. The workspace has no Git metadata, so `git diff --check` cannot
  be claimed.

