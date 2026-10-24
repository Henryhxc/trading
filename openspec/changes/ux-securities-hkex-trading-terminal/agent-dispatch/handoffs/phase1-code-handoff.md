# Phase 1 Code Handoff

Status: `GREEN_IMPLEMENTATION; RUNTIME_VALIDATION_PENDING`

## Scope and facts

Implemented only Phase 1 frontend auth shell, Go auth foundation, users/sessions/accounts schema and rollback, Compose topology, and Phase 1 design records. No Phase 2 market, watchlist, or Phase 3 trading behavior was added. No test, result, task, orchestration, packet, or review file was changed by this code lane; this handoff is the assigned delivery artifact.

Facts-source precedence followed: current user request and UX Securities Logo → this change's `proposal.md`, `design.md`, `tasks.md`, and relevant identity/deployment specs → `AGENTS.md` and `agent.md` → `docs/PRD.md` → `docs/research/hkex-free-data.md` → official HKEX sources. Phase 1 implementation does not make a claim about live HKEX data or real trading.

Unresolved product decisions were not treated as confirmed: registration uses username plus email per the current spec, but email verification is absent; default simulated cash is `HKD 1,000,000` and configurable with `INITIAL_CASH_HKD`; session lifetime defaults to `24h` and is configurable with `SESSION_TTL`; cookie secure mode defaults off for local HTTP and is configurable with `COOKIE_SECURE`; fees, lot sizes, matching delay, supported instrument scope, minute bars, and licensed data vendor remain unimplemented/unconfirmed for their later design/phase. Compose's `3000/8080/3306/8081` defaults are environment-overridable. Development DB credentials are insecure examples and must be overridden outside isolated local use.

## Files changed

- `apps/backend/go.mod`, `apps/backend/go.sum`: Go module and MySQL/bcrypt dependency locks.
- `apps/backend/cmd/api/main.go`: database readiness, migrations, server lifecycle, dependency wiring.
- `apps/backend/internal/config/config.go`: environment-driven ports, database, session, account, and migration settings.
- `apps/backend/internal/domain/auth/user.go`: auth domain values and identity-conflict contract.
- `apps/backend/internal/service/auth/service.go`: registration validation, bcrypt, opaque session creation, SHA-256 token hashing, authentication and logout.
- `apps/backend/internal/repository/mysqlrepo/auth.go`: SQL persistence, transactional account/user/session creation, session lookup and revocation.
- `apps/backend/internal/httpapi/server.go`: health/register/login/logout/me routes, request IDs, safe error mapping, cookie and actor middleware.
- `apps/backend/internal/migrations/runner.go`: sorted up-migration runner and version ledger.
- `apps/backend/Dockerfile`: non-root Go runtime image with migration SQL.
- `apps/frontend/package.json`, `package-lock.json`, `tsconfig.json`, `vite.config.ts`, `index.html`: pinned React/Vite/TypeScript build setup.
- `apps/frontend/src/main.tsx`, `src/pages/AuthPage.tsx`, `src/shared/api/auth.ts`, `src/styles.css`: branded auth shell, same-origin API client, state and focus/error layout.
- `apps/frontend/Dockerfile`, `apps/frontend/nginx.conf`: logo inclusion, static server and internal API proxy.
- `db/migrations/001_users_sessions_accounts.sql`: user, session and account tables, unique constraints, foreign keys and indexes.
- `db/migrations/001_users_sessions_accounts.down.sql`: reverse-order table rollback.
- `deploy/docker-compose.yml`: frontend/backend/mysql services, optional `admin` phpMyAdmin profile, defaults/overrides and health checks.
- `docs/design/develop-backend-design.md`, `develop-frontend-design.md`, `develop-design-task-linkage.md`, `project-archive.md`: Phase 1 file responsibilities, API/failure contracts, traceability and archive baseline.

## Behavior and failure branches

- Wire contract: `POST /api/v1/auth/register` accepts `{ "username": string, "email": string, "password": string }` and returns 201 `{ "user": { "id": number, "username": string, "email": string } }`. `POST /api/v1/auth/login` accepts `{ "identity": string, "password": string }` and returns 200 with the same `user` object. `POST /api/v1/auth/logout` accepts the session cookie when present and returns 204 with no body. `GET /api/v1/me` requires the session cookie and returns 200 with the same `user` object. `GET /health` returns 200 `{ "status": "ok" }`.
- Session cookie wire attributes: name `ux_session`; `HttpOnly`; `SameSite=Lax`; `Path=/`; `Expires` set to session expiry; `Secure` follows `COOKIE_SECURE` (false for local HTTP, must be true behind HTTPS). The token is not in JSON or local storage. `X-Request-ID` is returned on responses; a safe incoming ID is reused, otherwise a random ID is generated. Error bodies are `{ "code": string, "message": string, "request_id": string }`.
- Error mappings: malformed JSON/unknown field/trailing JSON -> 400 `invalid_request`, `Request body is invalid`; registration validation -> 400 `invalid_input`, `Check the submitted details`; duplicate username or email -> 409 `identity_unavailable`, `Username or email is unavailable`; unknown identity or wrong password -> 401 `invalid_credentials`, `Username or password is incorrect`; missing/expired/revoked session -> 401 `unauthenticated`, `Please sign in`; unexpected storage failure -> 500 `internal_error`, `Request could not be completed`. There is no `details` field in current errors.
- Registration validates username/email/password lengths, stores bcrypt only, and atomically creates user, default account, and initial session. Duplicate username/email share one stable 409 response.
- Login returns the same 401 response for unknown identity and wrong password. The session cookie is HttpOnly, SameSite=Lax, path `/`, and optionally Secure. Only SHA-256 token digest is stored.
- `/me` resolves session actor on every request; revoked, expired, missing, or disabled-user sessions return 401. Logout revokes the digest server-side and clears the cookie; logout without a cookie remains idempotent.
- Errors expose stable `{code,message,request_id}` and do not include SQL/stack/credentials. Logs include request ID, method, path, and status, not credentials or tokens.
- Registration DB writes have one transaction; pre-commit failure rolls back rows. MySQL DDL is auto-committing, so migration interruption can leave prior tables; DDL is idempotent and the explicit down migration is the recovery route. The API runner never applies `.down.sql` automatically.
- User-owned account tables exist, but no account endpoint is exposed in Phase 1. Future protected reads must use the session actor predicate; Phase 2/3 data isolation remains unimplemented here.

## Verification

| Command | Result |
|---|---|
| Initial RED `python3 tests/phase1_source_guard_test.py` (recorded by TDD lane) | Exit 1; 14 tests / 14 assertion failures for absent Phase 1 routes, auth/session behavior and tables. No syntax or environment error. |
| Final `python3 tests/phase1_source_guard_test.py` | Exit 0; 14 tests ran, `OK`. Static/source contract only. |
| `cd apps/backend && go test ./...` | Exit 0; all packages compile; no Go test files exist yet. |
| `cd apps/backend && go vet ./...` | Exit 0, no findings. |
| `cd apps/frontend && npm install && npm run build` | Exit 0; TypeScript and Vite production build passed. One initial build attempt identified missing React typings; those were added before the successful build. |
| `cd apps/frontend && npm audit` | Exit 0; 0 vulnerabilities after pinning patched Vite 7.3.7. npm emitted install-script approval warnings for `esbuild`/`fsevents`; bundle build nevertheless completed. |
| `docker compose -f deploy/docker-compose.yml config -q` | Exit 0; Compose model parses. |
| Production source line count | All listed production files are under 500 lines; largest is `server.go` at 201 lines. |
| DB expert file recheck | All five planned files under `/Users/henry/.codex/plugins/database-design/` and `database-migrations/` are absent. No expert recommendation was available; schema was checked against project `db-schema.md`, `AGENTS.md`, and MySQL constraints. |

Generated `apps/frontend/node_modules` and `apps/frontend/dist` were moved out of the project tree after successful validation; package lock remains. Parent agent added root `.dockerignore` to keep generated/dependency directories out of Docker contexts.

## Pending validation and residual risks

- Do not mark Phase 1 passed from this handoff. Project Compose has not been started by this lane. Cold start, MySQL health, migration apply/reapply/rollback, backend `/health`, frontend proxy, phpMyAdmin, and actual host URLs remain pending.
- Runtime API tests for registration, cookie attributes, duplicate identity, session expiry/revocation, and actor isolation remain with the TDD/e2e lane. The source guard does not prove those behaviors.
- Chrome/Computer Use screenshots at 1440x900, 1280x800, and 390x844, and independent review remain pending.
- The required `results/db-change-reports/<timestamp>-phase1-auth.md` is not written by this code lane; the validation/DB owner must record isolated DB target, SQL, dry-run/apply/rollback evidence, and failure results.
- Session lookup is DB-backed per request. There is no auth rate limiter yet; add one before production exposure. Local example DB credentials and non-Secure local cookie mode are not production settings.
- Migration runner splits the simple current SQL script on semicolons. If future migrations contain procedure bodies or semicolons in literals, replace this with a migration format/runner that safely parses them.
- The workspace has no Git metadata; `git diff --check` and repository diff review are unavailable here.

Next owner: Phase 1 TDD/e2e and runtime validation, then independent Phase 1 review and main-agent acceptance. Phase 2 must not start until the controller's Phase 1 gates pass and all Phase 1 agents are terminated.
