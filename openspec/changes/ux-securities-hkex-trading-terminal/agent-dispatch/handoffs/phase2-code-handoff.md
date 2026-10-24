# Phase 2 Code Handoff

Date: 2026-10-24 (Asia/Hong_Kong)
Status: `GREEN_READY_FOR_RUNTIME_QA`

## Scope delivered

Implemented only the Phase 2 market/workbench slice after the TDD lane recorded five expected RED failures in `results/phase2-market-red.md`.

- Go market domain contract and status allowlist in `apps/backend/internal/domain/market/quote.go`.
- Explicit authorized -> non-official -> mock fallback and deterministic mock symbols/quotes/candles in `apps/backend/internal/integration/marketdata/provider.go`.
- Market and watchlist use cases in `apps/backend/internal/service/market/service.go` and `internal/service/watchlist/service.go`.
- MySQL market and watchlist adapters in `apps/backend/internal/repository/mysqlrepo/market.go` and `watchlist.go`, with actor predicates and duplicate-safe insertion.
- Authenticated market/search/quote/candle and watchlist routes in `apps/backend/internal/httpapi/market_handler.go`, wired through `server.go` and `cmd/api/main.go`.
- Synthetic market seed, daily quote schema, watchlist schema, indexes, foreign keys, and matching down scripts in `db/migrations/002_stocks_daily_quotes.sql`, `002_stocks_daily_quotes.down.sql`, `003_watchlists.sql`, and `003_watchlists.down.sql`.
- Protected React workbench, same-origin market/watchlist API client, active-symbol persistence, responsive layout, status badges, quote/candle view, loading/empty/error states in `apps/frontend/src/pages/WorkbenchPage.tsx`, `src/shared/api/market.ts`, `src/app/AppRouter.tsx`, `src/shared/api/client.ts`, and `src/styles.css`.
- Adapter boundary and Phase 2 design records in `docs/api/hkex-market-data.md`, `docs/design/develop-backend-design.md`, and `docs/design/develop-frontend-design.md`.

## Verification output

```text
python3 tests/phase2_market_red_test.py
Ran 5 tests ... OK

cd apps/backend && go test ./... && go vet ./...
PASS

cd apps/frontend && npm run build
PASS (TypeScript and Vite production build)

npx openspec validate ux-securities-hkex-trading-terminal --strict
Change 'ux-securities-hkex-trading-terminal' is valid

docker compose -f deploy/docker-compose.yml config -q
PASS (no output)
```

All production source files remain below 500 lines. No React code calls a third-party provider; browser requests are same-origin `/api/v1/*` only. Free-provider `live` status is normalized away; mock responses are `simulated`.

The frontend build was rerun after removing an unused search-state variable found during main-agent verification; the final TypeScript/Vite build is green.

## Pending owner evidence

Runtime QA still owns MySQL apply/rollback against an isolated database, API smoke with two authenticated actors, provider failure fixtures, Compose cold start, in-app browser checks and screenshots at 1440x900/1280x800/390x844, and unavailable/stale visual states. Independent review and main-agent acceptance are also pending. This handoff does not mark Phase 2 passed or alter tasks/orchestration state.

## Concerns

- The default Compose database contains only synthetic seed symbols; a licensed provider remains a future adapter and is not connected.
- Watchlist quote enrichment preserves the relationship and now emits an explicit `unavailable` status when quote lookup fails; runtime QA should verify the rendered failure state.
- The migration runner's existing semicolon splitter remains appropriate for these plain DDL/seed statements; future procedural SQL would need a safer migration format.
- The workspace has no Git metadata, so diff checks remain unavailable.
