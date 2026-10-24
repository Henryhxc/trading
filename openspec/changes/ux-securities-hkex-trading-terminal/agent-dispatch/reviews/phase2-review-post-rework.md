# Phase 2 Independent Review (Post-Rework)

Date: 2026-10-24 (Asia/Hong_Kong)
Recommendation: `FAIL`

This is a fresh review of the post-rework tree. The three findings from `phase2-review.md` are resolved in the current implementation and tests, but Phase 2 still cannot pass because the required IAB browser evidence is absent. One remaining API-contract gap is also documented below.

## Review Scope

Reviewed the post-rework handoff and RED evidence, current market/watchlist domain/provider/service/repository/handler code, frontend workbench/API client/styles, OpenAPI contract, migration/down scripts, runtime/DB reports, orchestration state, and the Phase 2 screenshot directory. The user-selected in-app browser remains the required surface; no Chrome/Computer Use evidence is claimed.

## Findings

1. **[P1] Required IAB browser evidence is still missing.** `orchestration-state.md` records `browser-check: pending`; `agent-dispatch/handoffs/phase2-rework-handoff.md` is `GREEN_CONTRACTS; BROWSER_PENDING`; and `results/phase2-market-runtime.md` reports that IAB is unavailable. `results/screenshots/phase2-market-workbench/` contains only `.gitkeep`. The required 1440x900, 1280x800, and 390x844 workbench screenshots and interaction/state evidence therefore do not exist for search, quote, candles, favorites, empty/error/stale/unavailable states, mobile overflow, or scroll-anchor behavior. Phase 2 must remain unaccepted until this evidence is supplied or an explicitly authorized equivalent gate is recorded.

2. **[P2] Market/watchlist error responses introduced or exercised by the current handlers are not fully represented in OpenAPI.** `apps/backend/internal/httpapi/market_handler.go` now returns 400 for invalid/empty search input (`searchSymbols`), 503 for provider failure, 400 for malformed reorder requests, and 500 for watchlist list/reorder/delete storage failures. It also returns 400 for invalid watchlist symbols. `docs/api/openapi.yaml` lists only 200/401 for search, omits search 400/503, omits add-watchlist 400, and omits reorder/list/delete 400/500 responses. This leaves generated clients without the documented failure branches required by the Phase 2 error-state contract. Add the shared response references or change the handlers consistently.

3. **[P2] Candle query validation remains weaker than the declared schema.** `docs/api/openapi.yaml:145-153` declares `from`/`to` date formats and `limit` minimum 1/maximum 365, but `getDailyCandles` silently ignores malformed dates, maps zero to 90, and passes negative or over-365 limits to the service. Add explicit 400 validation (including `from <= to`) and tests, or revise the contract to document the permissive behavior.

## Prior Rework Findings Rechecked

- **Empty search validation:** resolved. `searchSymbols` trims and rejects empty or over-80-character input before invoking the provider; `TestSearchRejectsEmptyQueryBeforeProviderCall` passes.
- **Idempotent watchlist status:** resolved. `MarketRepository.AddWatchlist` returns a creation flag from `RowsAffected`; the service/handler return 201 for new relationships and 200 for existing relationships; `TestAddWatchlistUsesCreatedStatusOnlyForNewRelation` passes.
- **Frontend reorder and scroll anchor:** resolved in source. `marketApi.reorderWatchlist` calls `PUT /watchlist/order`; `WorkbenchPage` exposes accessible up/down controls, optimistically reorders and reloads on failure, and stores/restores `ux_watchlist_scroll_anchor` on the scrollable favorites list.

## Verified Implementation Areas

- `MarketDataProvider` and normalized quote/candle DTOs include source and data status fields. Provider fallback order is authorized -> unofficial -> mock, and non-authorized `live` quote status is normalized through `safeFreeStatus`; mock responses are `simulated`.
- React calls only same-origin `/api/v1/*`; no third-party provider URL or credential exists in `apps/frontend/src` or the production bundle source.
- Market and watchlist routes are session-protected. Watchlist SQL binds `user_id`, has a unique `(user_id, stock_id)` constraint, supports deletion/reorder, and preserves an unavailable quote state rather than fabricating a price.
- Migrations `002_stocks_daily_quotes.sql` and `003_watchlists.sql` have matching down scripts. The isolated runner apply/down/reapply and actor-isolation evidence remains valid for this rework.
- The production source line-count scan remains below 500 lines. This workspace has no Git metadata, so Git diff/status/diff-check cannot be claimed.

## Independently Re-run Checks

All commands below were run against the post-rework tree:

```text
python3 tests/phase2_rework_contract_test.py
Ran 3 tests ... OK

python3 tests/phase2_market_red_test.py
Ran 5 tests ... OK

cd apps/backend && go test ./... -count=1
PASS (all backend packages)

cd apps/backend && go vet ./...
PASS

cd apps/frontend && npm run build
PASS (TypeScript and Vite production build)

openspec validate ux-securities-hkex-trading-terminal --strict
PASS; change is valid

docker compose -f deploy/docker-compose.yml config -q
PASS
```

Live read-only smoke during this review showed the named Compose project healthy on ports `3000/8080/3306/8081`, backend `/health` 200, and frontend `/workspace` reachable. The runtime lane's API/DB smoke and migration evidence is recorded in `results/phase2-market-runtime.md` and `results/db-change-reports/2026-10-24-phase2-market.md`; no additional shared-database mutation was performed in this review.

## Evidence and Gate State

- RED: `results/phase2-market-red.md` and the rework-specific `results/phase2-rework-red.md` record expected pre-fix failures.
- GREEN/rework: `agent-dispatch/handoffs/phase2-code-handoff.md` and `agent-dispatch/handoffs/phase2-rework-handoff.md` record implementation and targeted checks.
- Runtime/DB: `agent-dispatch/handoffs/phase2-tdd-handoff.md`, `results/phase2-market-runtime.md`, and `results/db-change-reports/2026-10-24-phase2-market.md` record API smoke, mock status, actor isolation, and isolated rollback/reapply.
- Browser: **pending**. No Phase 2 IAB screenshots or visual state evidence exist; the directory has only `.gitkeep`.
- Orchestration remains constrained: Phase 2 is `validating`, browser-check pending, review-check pending; Phase 3 remains `pending`; task 2.x checkboxes remain unchecked. This review did not edit tasks or global state.

## Residual Risks

- The current database has three synthetic stock rows but zero daily quote rows, so a populated candle chart is not runtime-proven; only the empty state is observed in API smoke.
- Free/unofficial providers are not configured in Compose; runtime observed only the mock `simulated` path.
- Database expert paths remain unavailable, as recorded in the DB report.
- Synthetic QA identities and local database fixtures are development-only; no broker or real market endpoint was contacted.

## Evidence Paths

- Source/contracts: `apps/backend/internal/httpapi/market_handler.go`, `apps/backend/internal/httpapi/market_handler_test.go`, `apps/backend/internal/repository/mysqlrepo/watchlist.go`, `apps/frontend/src/pages/WorkbenchPage.tsx`, `apps/frontend/src/shared/api/market.ts`, `docs/api/openapi.yaml`.
- Handoffs/reports: `agent-dispatch/handoffs/phase2-rework-handoff.md`, `agent-dispatch/handoffs/phase2-tdd-handoff.md`, `results/phase2-market-runtime.md`, `results/db-change-reports/2026-10-24-phase2-market.md`.
- Tests: `tests/phase2_rework_contract_test.py`, `tests/phase2_market_red_test.py`, `apps/backend/internal/httpapi/market_handler_test.go`, `apps/backend/internal/integration/marketdata/provider_test.go`, `apps/backend/internal/service/market/service_test.go`, `apps/backend/internal/service/watchlist/service_test.go`.
- Browser evidence: none; `results/screenshots/phase2-market-workbench/.gitkeep` only.
