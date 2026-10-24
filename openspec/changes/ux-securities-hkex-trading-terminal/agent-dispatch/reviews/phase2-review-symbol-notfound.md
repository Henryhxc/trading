# Phase 2 Independent Review (Symbol-Not-Found Rework)

Date: 2026-10-24 (Asia/Hong_Kong)
Recommendation: `PASS_WITH_BROWSER_GATE`

This review independently checked the current Phase 2 tree after the
symbol-not-found rework. The implementation and contract checks pass. The
required user-selected in-app browser (IAB) workbench evidence remains a
separate pending gate; no Chrome or Computer Use evidence is claimed.

## Scope

Reviewed the Phase 2 market specification, current production composition,
domain/provider/repository/service/handler code, OpenAPI contract, focused RED
and GREEN tests, migration/runtime/DB reports, orchestration state, frontend
workbench source, and the Phase 2 screenshot directory. No production source,
task checkbox, or orchestration state was changed by this review.

## Required Symbol and Error Semantics

- **Unknown quote:** verified through the production composition in
  `apps/backend/cmd/api/main.go` (`MarketRepository -> marketdata.Provider ->
  mock provider`) after rebuilding the backend image from the current tree.
  A synthetic local user received `GET /api/v1/markets/quotes/12345` ->
  `404 not_found`.
- **Unknown candles:** the same production smoke received
  `GET /api/v1/markets/candles/12345` -> `404 not_found`. The repository checks
  stock existence before querying daily rows; the provider preserves a typed
  not-found result when all candidates report the unknown symbol; the handler
  maps it to 404.
- **Known symbol with no candle rows:** `GET
  /api/v1/markets/candles/00700?limit=30` -> `200` with `items: []`. The
  repository returns a non-nil empty slice for an existing stock with no
  `stock_daily_quotes` rows, so this is a valid empty response rather than a
  missing-symbol or outage response.
- **Provider outage:** `TestProviderReturnsUnavailableStatusWhenAllSourcesFail`
  verifies the fallback provider returns `ErrUnavailable` and unavailable
  status when all candidates fail. `TestGetQuoteMapsUnavailableProviderErrorToServiceUnavailable`
  and `TestGetDailyCandlesMapsUnavailableProviderErrorToServiceUnavailable`
  verify the HTTP layer maps that typed error to 503.
- **OpenAPI:** `docs/api/openapi.yaml` declares `404` and `503` for quote and
  candle routes, including `503: '#/components/responses/Unavailable'` for
  `/api/v1/markets/candles/{symbol}`.

## Findings

1. **[P1] IAB browser gate remains pending.**
   `orchestration-state.md` still records `browser-check: pending`, and
   `results/screenshots/phase2-market-workbench/` contains only `.gitkeep`.
   The required IAB screenshots and interaction evidence at 1440x900,
   1280x800, and 390x844 are therefore not present for search, quote, candles,
   favorites add/delete/reorder, empty/error/stale/unavailable states,
   mobile overflow, and scroll-anchor restoration. This is the only remaining
   review finding; Phase 2 must not be accepted until this evidence is
   captured and labeled `IAB`.

No additional implementation or API-contract finding was identified in this
review. The prior symbol-not-found and candle-503 findings are resolved in the
current source and focused tests.

## Independently Re-run Checks

```text
python3 tests/phase2_symbol_notfound_red.py             PASS (3)
python3 tests/phase2_contract_rework_red.py             PASS (2)
python3 tests/phase2_rework_contract_test.py            PASS (3)
cd apps/backend && go test ./internal/integration/marketdata ./internal/httpapi ./internal/repository/mysqlrepo -count=1  PASS
cd apps/backend && go test ./... -count=1               PASS
cd apps/backend && go vet ./...                        PASS
cd apps/frontend && npm run build                       PASS
OpenAPI YAML parse and candle 503 guard                PASS
npx openspec validate ux-securities-hkex-trading-terminal --strict  PASS
docker compose -f deploy/docker-compose.yml config --quiet                 PASS
```

The backend image was rebuilt from the current source before the live smoke;
Compose services were healthy on ports 3000/8080/3306/8081. Synthetic local
credentials were used only for the terminal smoke; no external provider,
broker, or real order endpoint was contacted. The workspace has no Git
metadata, so Git diff/status/diff-check cannot be claimed.

## Evidence and Residual Risk

- Runtime and isolated migration evidence remains in
  `results/phase2-market-runtime.md` and
  `results/db-change-reports/2026-10-24-phase2-market.md`.
- The current database has no seeded daily quote rows, so the populated chart
  state is not runtime-proven; the valid empty state is proven.
- The Compose runtime uses the mock provider (`source: mock`,
  `data_status: simulated`); no authorized or unofficial provider is configured.
- Phase 2 task checkboxes and Phase 3 remain gated by the parent agent. This
  review did not edit them.
