# Phase 1 Independent Review (Post-Rework)

Review date: 2026-10-24 (Asia/Hong_Kong)
Recommendation: `FAIL`

The earlier six implementation findings are resolved in the current source and supported by rework evidence. The user-selected IAB visual gate and phpMyAdmin table-list gate are now satisfied. Phase 1 is not ready for acceptance because the auth API contract remains incomplete and database lookup failures are misreported as invalid credentials.

## Review Scope

Reviewed the proposal, design, all six capability specs, Phase 1 packet, tasks 1.1-1.12, phase controller, rework handoff, current QA and DB reports, current implementation, OpenAPI contract, migration, handoffs, current orchestration state, and saved IAB screenshots. Phase 1 is the implementation scope; market, watchlist, and paper-trading behavior is not a Phase 1 omission.

The latest phase-controller decision explicitly selects IAB instead of Chrome/Computer Use. This review treats IAB as the required surface for this run and does not claim CUA evidence. The workspace has no `.git` metadata, so a baseline diff, `git status`, and `git diff --check` cannot be produced; source review is against the current tree and the recorded handoffs.

## Findings

1. **[P2] Login converts repository outages into invalid credentials.** `apps/backend/internal/service/auth/service.go:69-72` returns `ErrCredentials` for every `FindUserByIdentity` error, not just a missing identity. Consequently a MySQL timeout/outage is mapped by `apps/backend/internal/httpapi/server.go:168-172` to HTTP 401 `invalid_credentials`. The backend design requires unexpected repository failures to map to a generic 500 (`docs/design/develop-backend-design.md`, “Runtime configuration and failure branches”), and the health endpoint's 503 does not make a concurrent login failure truthful to the caller. Distinguish “not found” from storage errors, preserve the generic safe 500 response for storage errors, add a service/handler test, and document the 500 response in OpenAPI.

2. **[P2] Registration password bound is missing from the API schema.** The service accepts 12-72 UTF-8 bytes (`apps/backend/internal/service/auth/service.go:46`), but `docs/api/openapi.yaml:49` only specifies `minLength: 12`; an otherwise schema-valid password longer than bcrypt's 72-byte limit is rejected at runtime. Add a compatible maximum to the contract and make its byte/character semantics explicit, or change the implementation/policy consistently. The login operation also has no password bound; at minimum ensure its behavior is documented consistently with bcrypt comparison.

3. **[P2] Register/login OpenAPI responses omit an implemented server failure.** `server.go:174-175` returns the documented safe `500 internal_error` shape for unexpected errors from registration/session persistence, but the register and login response lists in `docs/api/openapi.yaml:50-59` and `:74-82` omit 500. This leaves generated/API consumers without a response contract for a real implementation branch. Document the shared 500 response (and include it for login, especially after finding 1 is corrected).

## Spec And Phase 1 Gate Assessment

- **1.1-1.3 Frontend:** current source has the route/session/API/error-boundary structure, actual Logo asset, and UX token; frontend design archive is present. Auth shell is intentionally not a workbench in this phase.
- **1.4-1.6 Backend/auth:** MVC responsibilities are separated; registration writes user/account/session in one transaction, passwords use bcrypt, session tokens are opaque with SHA-256 digests persisted, and session checks cover expiry/revocation/active-user status. Runtime evidence covers register, username/email login, logout, duplicate identities, expiry and per-user accounts. Finding 1 remains.
- **1.5 Migration:** schema has the required tables, unique constraints, foreign keys and indexes. The DB report records isolated apply/down/reapply evidence and runner rollback/reapply integration coverage; the post-rework runner test passed. Direct raw down SQL without an existing `schema_migrations` ledger was not rerun after the fix and remains unverified. Expected database-expert source files were unavailable; the DB report records this and the substitute review basis.
- **1.7 Runtime/Compose:** read-only current observation showed frontend, backend and MySQL healthy, phpMyAdmin running, and ports 3000/8080/3306/8081. Current smoke returned backend `/health` 200, frontend `/login` 200, proxied anonymous `/api/v1/me` 401, and phpMyAdmin HTTP 200. Read-only MySQL inspection showed `users`, `sessions`, `accounts`, `schema_migrations`, the applied `001_users_sessions_accounts.sql` version and expected unique indexes. The main-agent QA report records the Compose cold-start and retained-volume check; this review did not restart services.
- **1.8 TDD/API/DB:** RED/GREEN records, Go tests and runtime lifecycle evidence exist. This review independently reran `go test ./... -count=1`, `go vet ./...`, both Phase 1 Python guards, frontend production build, and strict OpenSpec validation; all passed. The migration integration test skips without its disposable-schema DSN; the isolated runner integration result is cited from the DB report, not claimed as rerun here. No shared app DB write or real broker connection was performed.
- An exploratory `npx @redocly/cli lint docs/api/openapi.yaml` using Redocly's built-in recommended rules failed on missing operation summaries across the contract and emitted generic metadata/4xx warnings. No Redocly configuration or lint gate is defined by this project, so this is not counted as a Phase 1 test failure; the specific undocumented 500 responses are reported in finding 3.
- **1.9 UI:** I inspected the saved IAB login screenshot at 1440x900 and phpMyAdmin table screenshot at 1280x720. The login view shows the real, correctly proportioned Logo and complete sign-in form. The phpMyAdmin image visibly lists `accounts`, `schema_migrations`, `sessions`, and `users` in `ux_securities`. The main-agent QA record includes the remaining 1280x800/390x844 login/register images, focus, loading/error, redirect, and mobile-overflow observations. Evidence remains labeled IAB, not CUA.
- **1.10 Review:** this artifact is the requested post-rework independent review; recommendation is FAIL because of findings 1-3.
- **1.11 DB report:** `results/db-change-reports/2026-10-24-phase1-auth.md` exists and records target isolation, SQL/assertions, cleanup, rollback/reapply behavior, failures, and residual risks.
- **1.12 Main acceptance:** not performed by this review. `tasks.md` checkboxes remain open and orchestration must not be advanced based on this report.

## Unresolved Or Unverified

- Task 0.4 product decisions remain open: registration identity/verification, initial paper cash and fees, lot size and matching delay, instrument scope, minute bars, and authorized data provider. Configuration/risk notes exist, but this task remains explicitly unchecked and is not resolved by Phase 1 review.
- The repository has no Git metadata. Git-only diff and `git diff --check` evidence is unavailable; this is recorded, not treated as a pass.
- Direct raw migration rollback with no runner-created migration ledger remains unverified, as documented in the DB report.
- Database expert files were not available in the implementation environment; no expert review is claimed.
- The three findings above need correction and targeted verification before review can recommend Phase 1 acceptance. Task 1.12/main acceptance and agent termination confirmation remain the main agent's responsibility.

## Evidence Paths

- Source/contracts: `apps/backend/internal/service/auth/service.go`, `apps/backend/internal/httpapi/server.go`, `docs/api/openapi.yaml`, `db/migrations/001_users_sessions_accounts.sql`, `db/migrations/001_users_sessions_accounts.down.sql`.
- Tests: `apps/backend/internal/service/auth/service_test.go`, `apps/backend/internal/httpapi/server_test.go`, `apps/backend/internal/migrations/runner_integration_test.go`, `tests/test_phase1_runtime.py`, `tests/phase1_rework_contract_test.py`, `tests/phase1_source_guard_test.py`.
- Handoffs/reports: `agent-dispatch/handoffs/phase1-code-handoff.md`, `agent-dispatch/handoffs/phase1-tdd-handoff.md`, `agent-dispatch/handoffs/phase1-rework-handoff.md`, `agent-dispatch/handoffs/phase1-runtime-qa-handoff.md`, `results/phase1-main-agent-qa-2026-10-24.md`, `results/db-change-reports/2026-10-24-phase1-auth.md`.
- IAB screenshots: `results/screenshots/phase1-foundation/login-1440x900-iab.jpg`, `login-1280x800-iab.jpg`, `login-390x844-iab.jpg`, `login-focus-390x844-iab.jpg`, `login-empty-submit-loading-390x844-iab.jpg`, `login-empty-submit-error-390x844-iab.jpg`, `register-1440x900-iab.jpg`, `register-1280x800-iab.jpg`, `register-390x844-iab.jpg`, `phpmyadmin-tables-1280x720-iab.jpg`.
