# Phase 1 Backend Design

## Scope

This phase provides only health, registration, login, logout, and current-user endpoints. It does not add market data, watchlists, orders, or broker execution. The backend starts in mock-safe mode and does not connect to external market providers.

## Directory and file contracts

```text
apps/backend/
  cmd/api/main.go                         process wiring, database readiness, graceful shutdown
  internal/config/config.go              environment configuration and DSN construction
  internal/domain/auth/user.go           user, registration, credentials, actor, identity conflict
  internal/service/auth/service.go       validation, bcrypt, opaque sessions, token hashing
  internal/repository/mysqlrepo/auth.go  MySQL auth persistence and registration transaction
  internal/httpapi/server.go             routes, request IDs, cookies, authentication, errors
  internal/migrations/runner.go          ordered up-migration application
  Dockerfile                              static Go binary runtime
db/migrations/
  001_users_sessions_accounts.sql        users, sessions, accounts and constraints
  001_users_sessions_accounts.down.sql   reverse dependency drop order
```

MVC direction is `httpapi handler -> auth service -> domain / repository interface -> mysqlrepo`. SQL is isolated in the repository. The `MarketDataProvider` and execution adapter remain absent until their assigned phases.

## HTTP contract

All routes use `/api/v1` except `GET /health`.

| Method and path | Access | Request | Success |
|---|---|---|---|
| `GET /health` | public | none | `200 {"status":"ok"}`; `503 {"status":"unavailable"}` when database readiness fails |
| `POST /api/v1/auth/register` | public | `{username,email,password}` | `201 {"user":{"id","username","email"}}` and session cookie |
| `POST /api/v1/auth/login` | public | `{identity,password}` | `200 {"user":{"id","username","email"}}` and session cookie |
| `POST /api/v1/auth/logout` | public/idempotent | cookie if present | `204`, revoke token digest, clear cookie |
| `GET /api/v1/me` | session | cookie | `200 {"user":{"id","username","email"}}` |

Errors have `{code,message,request_id}`. Registration requires username length 3-64, an address accepted by Go `net/mail` up to the database column limit of 128 characters, and password length 12-72 bytes (bcrypt limit). Duplicate username or email maps to the same `identity_unavailable` response. Login failures share one `invalid_credentials` response. Unknown fields and malformed/trailing JSON are rejected.

The cookie is named `ux_session`, is `HttpOnly`, `SameSite=Lax`, scoped to `/`, and uses `Secure` when `COOKIE_SECURE=true`. It carries a cryptographically random opaque token. Only SHA-256 hex is persisted in `sessions.token_hash`; authentication checks revocation, expiry, and active user status on each protected request. Default TTL is `24h`.

Registration writes `users`, the one-to-one HKD `accounts` row, and the first `sessions` row in one repository transaction. A unique constraint protects username, email, token hash, and account ownership. Failure before commit rolls back all three rows. Login creates a separate session row. Logout sets `revoked_at`; expired or revoked sessions receive 401. Repository access to user-owned records must always bind the session actor; Phase 1 currently exposes only `/me` and no account data endpoint.

## Runtime configuration and failure branches

Environment variables: `PORT=8080`, `DB_HOST=mysql`, `DB_PORT=3306`, `DB_NAME=ux_securities`, `DB_USER=ux_app`, `DB_PASSWORD`, `COOKIE_SECURE=false`, `SESSION_TTL=24h`, `INITIAL_CASH_HKD=1000000.0000`, and `MIGRATIONS_DIR=/app/db/migrations`. Local development Compose defaults are explicitly non-production and must be overridden outside an isolated local machine.

- Database unavailable or migration error: process does not serve HTTP and emits a generic startup failure without DSN/password.
- Invalid JSON/fields: 400 `invalid_request` with request ID.
- Invalid registration fields: 400 `invalid_input`.
- Duplicate identity: 409 `identity_unavailable`, without identifying which field/account exists.
- Bad credentials: 401 `invalid_credentials`.
- Missing, expired, revoked, or disabled-user session: 401 `unauthenticated`.
- Unexpected repository error: generic 500 `internal_error`; SQL details are not returned or logged.
- Logout with absent cookie: still clears cookie and returns 204.

Request logs include request ID, method, path, and status only. Passwords, cookie values, token digests, database credentials, and query values are not logged. The HTTP server has a read-header timeout and graceful shutdown.

## Database contract

MySQL 8.4 uses InnoDB and `utf8mb4_0900_ai_ci`. Migration `001_users_sessions_accounts.sql` creates users, sessions, accounts and indexes; the matching `.down.sql` drops dependent tables in order. The startup runner creates `schema_migrations`, applies sorted non-down SQL once, and records filenames after successful statements. MySQL DDL auto-commits, so a failed migration may leave earlier tables present; rerun is idempotent (`CREATE TABLE IF NOT EXISTS`) and the rollback script is the recovery path. The API migration runner never executes `.down.sql`.

## Verification

- `python3 tests/phase1_source_guard_test.py`
- `(cd apps/backend && go test ./... && go vet ./...)`
- `docker compose -f deploy/docker-compose.yml config -q`
- Runtime/API and database apply/rollback evidence is owned by the Phase 1 validation lane; do not treat source guard or Compose parsing as runtime proof.

Evidence destinations: `openspec/changes/ux-securities-hkex-trading-terminal/agent-dispatch/handoffs/phase1-code-handoff.md`, TDD lane handoff, and the phase's database change report. Runtime/browser gates remain pending until actual Compose and localhost checks complete.

## Phase 2 market and watchlist additions

`internal/domain/market/quote.go` owns normalized quote/candle DTOs, `DataStatus`, and the `MarketDataProvider` interface. `internal/integration/marketdata/provider.go` owns authorized -> non-official -> mock fallback and the deterministic `MockProvider`; it never labels a free provider `live`. `internal/service/market/service.go` is the use-case boundary for search, quote, and daily candles.

`internal/domain/watchlist/item.go`, `internal/service/watchlist/service.go`, and `internal/repository/mysqlrepo/watchlist.go` keep each relationship bound to the authenticated actor's `userID`. The unique `(user_id, stock_id)` key makes repeated additions idempotent. The single canonical migration `002_stocks_daily_quotes_watchlists.sql` creates `stocks`, `stock_daily_quotes`, and `watchlists`; its matching down migration drops them in reverse dependency order. Canonical apply removes the legacy ledger rows `002_stocks_daily_quotes.sql` and `003_watchlists.sql` before the runner records the combined filename. This lets an existing local database converge to the accepted ledger and leaves `003_orders_positions.sql` available for Phase 3.
