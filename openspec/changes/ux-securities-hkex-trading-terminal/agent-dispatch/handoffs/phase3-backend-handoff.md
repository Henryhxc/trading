# Phase 3 Backend Handoff

## Scope

- Completed the backend paper-trading slice only; no frontend, OpenSpec `tasks.md`, orchestration state, or other phase files were changed.
- Reused the existing `accounts` table and added only `orders` and `positions`; no `executions` or `position_lots` tables were introduced.
- Followed the project fact-source order: repository `AGENTS.md` and existing code first, then the Phase 3 `orders`/`portfolio` specs and tasks, with the HKEX official links in `AGENTS.md` treated as market-data boundary references. This implementation remains simulated and does not connect to a real execution venue.

## Changes

- Added trading domain models for orders, accounts, and positions in `apps/backend/internal/domain/trading/trading.go`.
- Added service validation, quote freshness checks, idempotency lookup, user-scoped reads, cancellation, and service-level error contracts in `apps/backend/internal/service/trading/service.go`.
- Added `orders` and `positions` migrations plus rollback in `db/migrations/003_paper_trading_orders_positions.sql` and `db/migrations/003_paper_trading_orders_positions.down.sql`.
- Added MySQL repository methods for idempotent order creation, limit matching, account/position settlement in one transaction, cancellation, user-filtered history, account reads, and position valuation in `apps/backend/internal/repository/mysqlrepo/trading.go`.
- Added authenticated endpoints for order creation, order detail, cancellation, history filters, account, and positions in `apps/backend/internal/httpapi/trading_handler.go`; wired them through `NewWithTrading` in `apps/backend/internal/httpapi/server.go` and `cmd/api/main.go`.

## Evidence

- RED command: `cd UX\ Security/apps/backend && go test ./internal/service/trading -run 'TestCreateRejectsStaleQuoteBeforePersistence|TestCreateIsIdempotentAndSettlesAtomically' -count=1 2>&1 | tee ../../openspec/changes/ux-securities-hkex-trading-terminal/results/phase3-backend-red.txt`
- RED result: failed before implementation because trading domain/service types were absent; output preserved in `results/phase3-backend-red.txt`.
- Targeted GREEN command: `GOCACHE=/tmp/ux-go-cache go test ./internal/service/trading ./internal/repository/mysqlrepo ./internal/httpapi -count=1 2>&1 | tee ../../openspec/changes/ux-securities-hkex-trading-terminal/results/phase3-backend-targeted-green.txt`
- Full GREEN command: `GOCACHE=/tmp/ux-go-cache go test ./... -count=1 2>&1 | tee ../../openspec/changes/ux-securities-hkex-trading-terminal/results/phase3-backend-full-go.txt`
- Final full result: all Go packages passed, including `cmd/api`, `httpapi`, migrations, repository, and trading service.
- Production file line counts remain below 500 lines; the largest changed production file is `internal/repository/mysqlrepo/trading.go` at 214 lines.

## Risks and not done

- No live MySQL integration run was performed in this handoff; transaction behavior is implemented with InnoDB SQL and covered only by compile/unit-level verification in this environment.
- No real brokerage or HKEX execution adapter is connected; matching uses the stored simulated quote and remains paper trading only.
- Pending orders are persisted and cancellable, but there is no asynchronous re-matching worker in this phase.
- No frontend flow, browser evidence, OpenAPI artifact update, DB change report, independent review, or final cross-phase gate was performed; those remain outside this backend packet and user scope.
- Existing project migration naming/runner behavior was preserved; rollback SQL is provided but was not executed against a live database here.
