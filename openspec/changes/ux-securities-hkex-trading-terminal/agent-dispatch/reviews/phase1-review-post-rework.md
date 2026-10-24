# Phase 1 Independent Review (Post-Auth-Contract Rework)

Date: 2026-10-24 (Asia/Hong_Kong)
Recommendation: `PASS`

The three findings in the previous independent review are resolved in the current source and contract. No blocking implementation finding remains for the Phase 1 review gate. Main-agent acceptance is still a separate gate: task 0.4 remains open, task 1.x checkboxes remain unchecked, and Phase 2/3 remain pending.

## Review Scope

Reviewed the current auth service and HTTP handlers, auth unit/handler tests, OpenAPI contract, Phase 1 design/spec/task/phase-controller files, auth-contract rework handoff, prior review, runtime and database reports, current orchestration state, Compose configuration, and saved user-selected in-app-browser (IAB) screenshots. The latest user instruction selected IAB; this report labels that evidence IAB and makes no Chrome/Computer Use claim.

## Prior Findings Rechecked

1. **Repository failures are no longer treated as invalid credentials.** `apps/backend/internal/service/auth/service.go:75-82` maps only `sql.ErrNoRows` to `ErrCredentials`; any other `FindUserByIdentity` error is returned unchanged. `apps/backend/internal/httpapi/server.go:166-176` maps that unexpected error to the generic 500 `internal_error` response without exposing storage details. `TestLoginPreservesUnexpectedRepositoryError` and `TestAuthUnexpectedErrorsReturnGenericInternalServerError` pass.

2. **Password byte semantics are explicit in OpenAPI.** `docs/api/openapi.yaml:392-402` defines the shared registration/login `Password` schema with `minLength: 12`, `maxLength: 72`, and a description stating that the server enforces 12-72 UTF-8 bytes for bcrypt's 72-byte input limit. It explicitly explains that JSON Schema length keywords count Unicode characters and that the server performs the additional non-ASCII byte validation. The Go service enforces the same byte bounds for registration via `len(input.Password)`; login mismatch and over-limit inputs remain invalid credentials through bcrypt comparison.

3. **Auth server errors are listed in the API contract.** Both register and login list `500: InternalServerError` at `docs/api/openapi.yaml:50-60` and `:75-85`, and the shared response is defined at `:373-377`. The response shape matches the generic handler output (`code`, `message`, `request_id`).

## Findings

No blocking findings.

### Process notes for main-agent acceptance

- `orchestration-state.md` correctly remains `status: validating`, Phase 1 `validating`, Phase 2/3 `pending`, with `active_agents` containing this review lane. Tasks 0.4, 0.5, and all Phase 1 tasks remain unchecked; no phase or task was incorrectly advanced by this review.
- The `agent_history` section still contains an older duplicate `/root/phase1_independent_review_2` entry marked `running`, although the current live-agent list has no such active agent. Reconcile that stale lifecycle record before `termination-check`; it does not change this review's implementation PASS.
- The workspace has no Git metadata, so `git status`, baseline diff, and `git diff --check` cannot be claimed.
- Direct raw SQL rollback without a runner-created `schema_migrations` ledger remains documented as unverified; the supported runner apply -> down -> reapply path has independent isolated-schema evidence.
- The database expert source path was unavailable; the DB report records the substitute DDL/MySQL execution review. Synthetic QA rows remain in the local development database by design of the recorded runtime evidence.

## Independently Re-run Checks

All commands below were run against the current tree on 2026-10-24:

```text
cd apps/backend && go test ./internal/service/auth ./internal/httpapi -count=1
PASS

cd apps/backend && go test ./... -count=1
PASS (all backend packages)

cd apps/backend && go vet ./...
PASS

python3 tests/phase1_rework_contract_test.py
Ran 4 tests ... OK

python3 tests/phase1_source_guard_test.py
Ran 14 tests ... OK

cd apps/frontend && npm run build
PASS (TypeScript and Vite production build)

python3 YAML parse/contract inspection for docs/api/openapi.yaml
PASS; OpenAPI 3.1 parsed; register responses include 201/400/409/500;
login responses include 200/400/401/500

openspec validate ux-securities-hkex-trading-terminal --strict
PASS; change is valid
```

The focused Go tests include missing-identity-to-401, unexpected repository error propagation, register 500, login 500, and invalid-identity 401 assertions. No production code, task checkbox, or orchestration state was changed during this review.

## Runtime, Database, and UI Evidence

- `docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml ps` showed frontend, backend, and MySQL healthy; phpMyAdmin running on ports `3000/8080/3306/8081`.
- Read-only HTTP checks returned backend `/health` 200, frontend `/login` 200, anonymous `/api/v1/me` 401, and phpMyAdmin HTTP 200.
- Read-only MySQL inspection returned `accounts`, `schema_migrations`, `sessions`, and `users`, with migration `001_users_sessions_accounts.sql` applied.
- Production source line-count review found no file over 500 lines; the largest application source is `apps/backend/internal/httpapi/server.go` at 206 lines (dependency `node_modules` excluded).
- IAB screenshots reviewed: `results/screenshots/phase1-foundation/login-1440x900-iab.jpg`, `login-390x844-iab.jpg`, and `phpmyadmin-tables-1280x720-iab.jpg`; the login form/logo are visible at desktop and mobile sizes without apparent clipping or horizontal overflow, and the phpMyAdmin UI lists the four expected tables. The broader IAB set and interaction observations are recorded in `results/phase1-main-agent-qa-2026-10-24.md`.
- Runtime and isolated migration apply/down/reapply evidence is recorded in `results/phase1-main-agent-qa-2026-10-24.md` and `results/db-change-reports/2026-10-24-phase1-auth.md`; this review did not rerun the mutating shared-database runtime lifecycle test.

## Phase Decision

The auth-contract rework is verified and the Phase 1 independent review gate passes. The main agent must still reconcile lifecycle metadata, review task 0.4 and all remaining evidence, and explicitly perform main acceptance before changing Phase 1 to `passed` or starting Phase 2. No real broker, production database, or external market-data provider was contacted.

## Evidence Paths

- Source/contracts: `apps/backend/internal/service/auth/service.go`, `apps/backend/internal/httpapi/server.go`, `apps/backend/internal/service/auth/service_test.go`, `apps/backend/internal/httpapi/server_test.go`, `docs/api/openapi.yaml`.
- Handoffs/reviews: `agent-dispatch/handoffs/phase1-auth-contract-rework-handoff.md`, `agent-dispatch/reviews/phase1-review.md`, `agent-dispatch/orchestration-state.md`.
- Runtime/DB: `results/phase1-main-agent-qa-2026-10-24.md`, `results/db-change-reports/2026-10-24-phase1-auth.md`.
- IAB screenshots: `results/screenshots/phase1-foundation/login-1440x900-iab.jpg`, `login-1280x800-iab.jpg`, `login-390x844-iab.jpg`, `login-focus-390x844-iab.jpg`, `login-empty-submit-loading-390x844-iab.jpg`, `login-empty-submit-error-390x844-iab.jpg`, `register-1440x900-iab.jpg`, `register-1280x800-iab.jpg`, `register-390x844-iab.jpg`, `phpmyadmin-tables-1280x720-iab.jpg`.
