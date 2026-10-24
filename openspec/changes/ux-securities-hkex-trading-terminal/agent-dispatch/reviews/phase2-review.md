# Phase 2 Independent Review

Date: 2026-10-24 (Asia/Hong_Kong)
Recommendation: `FAIL`

The Phase 2 backend/provider/migration implementation is present and its GREEN/API/DB evidence is credible, but the phase cannot pass the review gate yet. Required IAB screenshots and state checks are unavailable, and the current frontend/API contract has two observable gaps that need correction or an explicit contract decision.

## Review Scope

Reviewed the Phase 2 market/watchlist specs, design and task packet, current domain/provider/service/repository/handler code, React workbench and API client, OpenAPI routes/schemas, migrations and down scripts, RED/GREEN handoffs, runtime and DB reports, current orchestration state, and the current workspace. The project uses the user-selected in-app browser surface; no Chrome/Computer Use claim is made.

## Findings

1. **[P1] Required IAB browser evidence is missing, so the Phase 2 visual gate is not verifiable.** `orchestration-state.md` records `browser-check: pending`; `results/phase2-market-runtime.md` and `agent-dispatch/handoffs/phase2-tdd-handoff.md` explicitly report that IAB is unavailable and no screenshots were captured. The required `results/screenshots/phase2-market-workbench/` directory contains only `.gitkeep`. Consequently search, quote, daily chart, favorites, empty/error/stale/unavailable states, responsive overflow, and the 1440x900/1280x800/390x844 layouts cannot be independently accepted. Keep Phase 2 pending until those IAB artifacts exist or the project records an authorized gate change and equivalent evidence.

2. **[P2] The Phase 2 workbench does not implement required favorite sorting or the specified scroll-anchor behavior.** `apps/frontend/src/pages/WorkbenchPage.tsx` provides search, quote loading and add/remove favorites, but no reorder control or call to `/api/v1/watchlist/order`; `apps/frontend/src/shared/api/market.ts` has no reorder method. The component persists only `ux_active_symbol` in `sessionStorage` and does not persist or restore a filtered-list scroll anchor. This leaves task 2.2's “筛选后的 scroll anchor” and task 2.3's add/delete/sort contract incomplete even though the backend `PUT /api/v1/watchlist/order` route exists.

3. **[P2] Duplicate watchlist POST status does not match the declared API contract.** `docs/api/openapi.yaml:195-199` declares `200` for an existing idempotent relationship and `201` for a newly created relationship. `apps/backend/internal/httpapi/market_handler.go:59-72` always writes `201 Created`, while `apps/backend/internal/repository/mysqlrepo/watchlist.go:48-57` deliberately uses `ON DUPLICATE KEY UPDATE` and returns no created/existing indication. Runtime evidence confirms both the first and duplicate `00700` requests returned `201`. Either return `200` on an existing row and preserve `201` for creation, or revise the shared OpenAPI contract and front-end expectations to state that both idempotent outcomes use `201`.

4. **[P2] Empty search input is not validated according to the OpenAPI request contract.** `docs/api/openapi.yaml:105-124` marks `query` required with `minLength: 1`, but `apps/backend/internal/httpapi/market_handler.go:13-20` passes an absent/empty query directly to the provider. The repository query in `apps/backend/internal/repository/mysqlrepo/market.go:15-29` treats an empty value as `LIKE '%%'` and can return every seeded stock instead of a 400 validation error or an intentional empty-result behavior. Add explicit query validation and a corresponding handler/API test, or change the OpenAPI contract to document empty-query semantics.

## Resolved/Verified Areas

- `MarketDataProvider` and quote/candle DTOs exist in `apps/backend/internal/domain/market/quote.go` with the required source/status fields.
- Provider fallback order is explicit (`authorized -> unofficial -> mock`) in `apps/backend/internal/integration/marketdata/provider.go`; non-authorized `live` status is normalized through `safeFreeStatus`, and mock data is marked `simulated`. React source calls only same-origin `/api/v1/*` routes; no provider URL or credential appears in `apps/frontend/src`.
- Search, quote, daily-candle, watchlist list/add/delete/reorder routes are authenticated and wired through `server.go`; repository queries bind `user_id` for watchlist reads/writes and the unique `(user_id, stock_id)` constraint prevents duplicate relationships.
- Migrations `002_stocks_daily_quotes.sql` and `003_watchlists.sql` have matching down scripts. The isolated apply/down/reapply and actor-isolation evidence is recorded in the Phase 2 runtime and DB reports.
- All inspected production source files are below 500 lines. The workspace has no Git metadata, so Git diff/status/diff-check evidence is unavailable.

## Independently Re-run Checks

All commands below were run against the current tree:

```text
python3 tests/phase2_market_red_test.py
Ran 5 tests ... OK

cd apps/backend && go test ./... -count=1
PASS (all backend packages, including provider, market, watchlist and repository tests)

cd apps/backend && go vet ./...
PASS

cd apps/frontend && npm run build
PASS (TypeScript and Vite production build)

openspec validate ux-securities-hkex-trading-terminal --strict
PASS; change is valid

docker compose -f deploy/docker-compose.yml config -q
PASS
```

The runtime lane's API smoke and isolated migration checks are recorded as passing in `results/phase2-market-runtime.md` and `results/db-change-reports/2026-10-24-phase2-market.md`. The smoke observed source `mock` and `data_status: simulated`, actor-isolated watchlists, duplicate-safe storage, deletion, and an empty daily-candle array. No daily quote rows are seeded, so a populated chart is not runtime-proven.

## Evidence and Gate State

- RED: `results/phase2-market-red.md` records five expected failures before implementation.
- GREEN: `agent-dispatch/handoffs/phase2-code-handoff.md` records the implementation and focused/full test/build/strict evidence.
- Runtime/DB: `agent-dispatch/handoffs/phase2-tdd-handoff.md`, `results/phase2-market-runtime.md`, and `results/db-change-reports/2026-10-24-phase2-market.md` record API smoke, actor isolation, provider status, and isolated migration rollback/reapply.
- Browser: **pending**; no Phase 2 IAB screenshots or visual state evidence exist. The phase packet requires IAB checks at `1440x900`, `1280x800`, and `390x844`, including search, quote, candles, favorites, stale/unavailable, overflow and scroll-anchor behavior.
- Orchestration remains correctly constrained: Phase 2 is `running` with browser-check pending, review-check running, main acceptance pending; Phase 3 remains `pending`; Phase 2 task checkboxes remain unchecked. This review did not edit tasks or global orchestration state.

## Residual Risks

- The default database has three synthetic stock rows but zero `stock_daily_quotes` rows; candle UI can only show its empty state until a fixture/provider history is loaded.
- Free/unofficial providers are not configured in the running Compose stack; only the mock path was observed at runtime.
- Database expert paths remain unavailable, as documented in the DB report.
- Synthetic runtime users/watchlist fixtures are local development data; no production database or real market/broker endpoint was contacted.

## Evidence Paths

- Source/contracts: `apps/backend/internal/domain/market/quote.go`, `apps/backend/internal/integration/marketdata/provider.go`, `apps/backend/internal/service/market/service.go`, `apps/backend/internal/service/watchlist/service.go`, `apps/backend/internal/repository/mysqlrepo/market.go`, `apps/backend/internal/repository/mysqlrepo/watchlist.go`, `apps/backend/internal/httpapi/market_handler.go`, `apps/frontend/src/pages/WorkbenchPage.tsx`, `apps/frontend/src/shared/api/market.ts`, `docs/api/openapi.yaml`.
- Handoffs/reports: `agent-dispatch/handoffs/phase2-code-handoff.md`, `agent-dispatch/handoffs/phase2-tdd-handoff.md`, `results/phase2-market-runtime.md`, `results/db-change-reports/2026-10-24-phase2-market.md`.
- RED/tests: `results/phase2-market-red.md`, `tests/phase2_market_red_test.py`, `apps/backend/internal/integration/marketdata/provider_test.go`, `apps/backend/internal/service/market/service_test.go`, `apps/backend/internal/service/watchlist/service_test.go`, `apps/backend/internal/repository/mysqlrepo/watchlist_test.go`.
- Browser evidence: none; `results/screenshots/phase2-market-workbench/.gitkeep` only.
