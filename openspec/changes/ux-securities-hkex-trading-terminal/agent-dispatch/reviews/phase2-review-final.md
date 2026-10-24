# Phase 2 Independent Review (Final Post-Contract-Rework)

Date: 2026-10-24 (Asia/Hong_Kong)
Recommendation: `FAIL`

This review was performed against the current post-contract-rework tree. The
review explicitly rechecked the prior OpenAPI error-response and candle-query
findings. Those two findings are resolved. Phase 2 remains blocked by the
required IAB evidence and has two additional implementation gaps below.

## Review Scope

Reviewed the current Phase 2 source, shared OpenAPI contract, market/watchlist
specs, RED/GREEN tests, rework handoff, runtime/DB reports, orchestration state,
frontend workbench, and screenshot directory. No task checkbox or production
implementation was edited by this review. The user-selected in-app browser is
the only accepted browser evidence surface; no Chrome or Computer Use evidence
is claimed.

## Findings

1. **[P1] Required IAB workbench evidence is missing.**
   `results/screenshots/phase2-market-workbench/` contains only `.gitkeep`, and
   the runtime/rework handoffs record IAB as pending/unavailable pending user
   confirmation for synthetic local credentials. The required 1440x900,
   1280x800, and 390x844 workbench checks therefore do not exist for search,
   quote, daily candles, favorites add/delete/reorder, empty/error/stale/
   unavailable states, mobile overflow, or scroll-anchor restoration. Phase 2
   cannot be accepted until these artifacts and interaction observations are
   captured and labeled `IAB`.

2. **[P2] A provider outage is reported as a missing security on quote/candle
   routes.** `marketdata.Provider.Quote` returns `ErrUnavailable` with an
   `unavailable` quote after all candidates fail (`apps/backend/internal/
   integration/marketdata/provider.go:38-55`), but `getQuote` maps every error
   to `404 not_found` before checking `DataStatus` (`apps/backend/internal/
   httpapi/market_handler.go:28-36`). `getDailyCandles` likewise maps every
   provider error to `404` (`market_handler.go:50-55`). This contradicts the
   declared 503 unavailable response and the market-data requirement that
   provider failures be explainable. Add typed unavailable/error mapping and
   tests so dependency failure is not presented as a nonexistent symbol.

3. **[P2] A search that returns zero matches has no frontend empty-result
   state.** `WorkbenchPage.search` stores an empty array on a successful
   no-match response (`apps/frontend/src/pages/WorkbenchPage.tsx:37-42`), while
   the render only shows the result container when `results.length > 0`
   (`WorkbenchPage.tsx:57`). Unless the favorites list is also empty, the user
   receives no “no matches” explanation. This misses the market-data scenario
   requiring an understandable empty result state; track a searched/no-match
   state and render a message distinct from transport failure.

## Prior Findings Rechecked

- **OpenAPI error-response parity:** resolved. `tests/phase2_contract_rework_red.py`
  passes and the current contract declares the handler branches: search 400/
  503, candles 400, watchlist add 400, and list/reorder/delete 500 as well as
  the authenticated statuses.
- **Strict candle query validation:** resolved. `getDailyCandles` rejects
  malformed or repeated date values, `from > to`, non-numeric/zero/out-of-range
  limits before provider invocation; `TestGetDailyCandlesRejectsInvalidQuery-
  BeforeProviderCall` and the contract guard pass.
- **Earlier rework findings:** resolved in source and tests. Search input is
  validated before the provider call; idempotent watchlist adds return 201 only
  for a new relation and 200 for an existing one; reorder controls and
  `ux_watchlist_scroll_anchor` persistence are present.

## Verified Checks

```text
python3 tests/phase2_contract_rework_red.py       PASS (2)
python3 tests/phase2_rework_contract_test.py      PASS (3)
python3 tests/phase2_market_red_test.py            PASS (5)
cd apps/backend && go test ./... -count=1         PASS
cd apps/backend && go vet ./...                   PASS
cd apps/frontend && npm run build                 PASS
openspec validate ux-securities-hkex-trading-terminal --strict  PASS
docker compose -f deploy/docker-compose.yml config -q          PASS
OpenAPI YAML parse                                  PASS
Backend /health                                     200 OK
Frontend /login                                     200 OK
```

The isolated migration/API smoke and actor-isolation evidence remains recorded
in `results/phase2-market-runtime.md` and
`results/db-change-reports/2026-10-24-phase2-market.md`. This review did not
perform additional database writes. The workspace has no Git metadata, so
`git diff --check` cannot be claimed.

## Residual Evidence and Risk

- Browser evidence is **pending** and must remain separate from terminal/API
  checks; no alternate browser surface may substitute for IAB.
- Runtime evidence exercises the mock `simulated` provider and an empty daily
  history. Populated candle rendering and stale/unavailable visual states are
  not runtime-proven.
- `0.4`, Phase 2 task checkboxes, and Phase 3 remain gated by the parent agent;
  this review did not change them.

