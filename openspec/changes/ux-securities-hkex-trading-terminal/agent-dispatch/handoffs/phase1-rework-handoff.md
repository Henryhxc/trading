# Phase 1 Rework Handoff

Date: 2026-10-24 (Asia/Hong_Kong)
Status: `READY_FOR_MAIN_AGENT_QA`; Phase 1 is not accepted by this handoff.

## Scope

Addressed the six implementation findings from `agent-dispatch/reviews/phase1-review.md`: unreadable logo asset; runner rollback/reapply ledger defect; missing React routes/session/API/error-boundary architecture; email format/length mismatch; idempotent logout/OpenAPI mismatch; and constant-only database health. No Phase 2 functionality, task checkbox, phase state, or global orchestration state was changed.

## Files Changed

- `apps/frontend/Dockerfile`: make the copied Logo readable by nginx workers during the build.
- `docs/assets/ux-securities-logo.png`: change source mode from `0600` to `0644`.
- `db/migrations/001_users_sessions_accounts.down.sql`: remove the matching migration ledger row after dropping auth tables.
- `apps/backend/internal/migrations/runner_integration_test.go`: add a real MySQL runner apply/down/reapply test guarded to `ux_phase1_test_*` schemas.
- `apps/backend/internal/service/auth/service.go`: parse email syntax with `net/mail`, enforce the existing 128-byte column limit, and expose repository readiness through the service.
- `apps/backend/internal/repository/mysqlrepo/auth.go`: implement DB readiness with `PingContext`.
- `apps/backend/internal/httpapi/server.go`: return 503 when DB readiness fails or is unavailable.
- `docs/api/openapi.yaml`: set email maximum to 128, document health 503, and describe logout as public/idempotent with only a 204 response.
- `apps/frontend/src/main.tsx`, `src/app/*`, and `src/shared/api/*`: add routing, protected `/workspace`, shared session bootstrap, API transport/status errors, and a React error boundary.
- `apps/frontend/src/pages/AuthPage.tsx`, `src/styles.css`: use shared session state and route navigation; distinguish bootstrap loading/error from signed-out state without adding a workbench.
- `tests/phase1_rework_contract_test.py`, Go unit tests, and `docs/design/develop-frontend-design.md`: durable rework checks and updated frontend archive.
- `results/db-change-reports/2026-10-24-phase1-auth.md`: append rework evidence and migration-test limitations.

## RED Evidence

The focused tests were added before implementation.

```text
cd apps/backend && go test ./internal/service/auth ./internal/httpapi ./internal/migrations
FAIL: 5 malformed email cases were accepted; GET /health returned 200 instead of 503.
The logout-without-cookie behavior passed its target assertion.

python3 tests/phase1_rework_contract_test.py
FAIL: AppRouter missing; Logo source unreadable to unprivileged worker; logout OpenAPI missing security: [].
```

Runner integration RED was run against isolated schema `ux_phase1_test_rework`:

```text
PHASE1_MIGRATION_TEST_DSN=... go test ./internal/migrations -run TestRunnerReappliesMigrationAfterRollback -count=1 -v
FAIL: accounts exists after runner reapply = 0; sessions = 0; users = 0.
```

These failures were behavior/contract assertions; no compile or test-harness failures were used as RED evidence.

## GREEN Evidence

```text
cd apps/backend && go test ./... && go vet ./...
PASS

cd apps/frontend && npm run build
PASS (TypeScript build and Vite production bundle)

python3 tests/phase1_rework_contract_test.py
PASS (3 tests)

python3 tests/phase1_source_guard_test.py
PASS (14 tests)

openspec validate ux-securities-hkex-trading-terminal --strict
PASS
```

The durable MySQL runner integration passed on the isolated schema after the fix. The schema and test-only `phase1_runner` user were then dropped; information-schema checks returned `0` for both. The shared `ux_securities` database and its existing synthetic QA rows were not modified by this integration test. Running the integration test without `PHASE1_MIGRATION_TEST_DSN` reports `SKIP` as intended; ordinary `go test ./...` therefore does not touch MySQL.

The existing auth runtime lifecycle was rerun against the rebuilt backend:

```text
UX_BASE_URL=http://localhost:8080 python3 tests/test_phase1_runtime.py
Ran 1 test ... OK
```

No-cookie `POST /api/v1/auth/logout` returned `204`. A unit test verifies a database readiness error returns `/health` `503`; success returns `200`.

## Compose And Asset Smoke

Rebuilt only this project with:

```text
docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml up -d --build frontend backend
```

Final `ps --all`: backend healthy on host port `8080`; frontend healthy on `3000`; MySQL healthy on `3306`; phpMyAdmin remains up on `8183`. The Compose rollout recreated only services in this named project and used its existing named MySQL volume. No unrelated containers were targeted.

`GET /ux-securities-logo.png` returned HTTP `200`, `image/png`, 17,214 bytes; `file` identified a valid `428x96` PNG. The built asset mode is `644 root:root`, readable by nginx. Frontend `/` returned `200`, proxied anonymous `/api/v1/me` returned expected `401`, and backend `/health` returned `200 {"status":"ok"}`.

No browser was used in this rework lane. Consequently nonzero browser `naturalWidth`, required screenshots, keyboard focus, responsive overflow, visible UI error/success states, and authenticated phpMyAdmin table browsing remain pending for the main agent's Chrome/Computer Use QA.

## Limits And Recommendation

- Direct raw SQL-only down migration without a runner-created `schema_migrations` table was not rerun after adding ledger cleanup. The tested supported path is runner apply → down SQL → runner reapply; see the appended DB report note.
- Frontend architecture contract tests are source-level checks; the main agent must behaviorally inspect direct routes, bootstrap loading/failure/retry, and route protection in the browser.
- Existing screenshot and phpMyAdmin UI evidence gaps remain. This handoff does not authorize checking tasks or changing phase/orchestration state.
- Git metadata is absent in this workspace; no Git diff or `git diff --check` result is claimed.

Recommendation: proceed to main-agent browser and phpMyAdmin QA, then independent review. Do not mark Phase 1 passed until those pending evidence gates are reviewed.
