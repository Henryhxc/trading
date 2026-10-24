# Phase 1 Main-Agent Acceptance

Date: 2026-10-24 (Asia/Hong_Kong)
Phase: `phase1-foundation`
Decision: `PASS_WITH_NOTES`

## Evidence reviewed

- Phase 1 development, TDD/e2e, runtime, database and auth-contract rework handoffs.
- Fresh independent review: `agent-dispatch/reviews/phase1-review-post-rework.md` (`PASS`).
- IAB screenshots and interaction evidence under `results/screenshots/phase1-foundation/`, including login/register at `1440x900`, `1280x800`, `390x844`, focus/loading/error/route/overflow states, and read-only phpMyAdmin table listing at `1280x720`.
- Database report: `results/db-change-reports/2026-10-24-phase1-auth.md`.
- Compose/API evidence: backend and MySQL healthy, frontend and phpMyAdmin reachable, `/health` 200, frontend `/login` 200, anonymous `/api/v1/me` 401, and auth lifecycle runtime smoke PASS.

## Main-agent verification

```text
cd apps/backend && go test ./... -count=1 && go vet ./...       PASS
python3 tests/phase1_rework_contract_test.py                  PASS (4)
python3 tests/phase1_source_guard_test.py                     PASS (14)
cd apps/frontend && npm run build                             PASS
python3 PyYAML parse of docs/api/openapi.yaml                 PASS
npx openspec validate ux-securities-hkex-trading-terminal --strict PASS
UX_BASE_URL=http://localhost:8080 python3 tests/test_phase1_runtime.py PASS
```

The rebuilt Compose backend was healthy on ports `3000/8080/3306/8081` before the runtime smoke. The largest production source file is `apps/backend/internal/httpapi/server.go` at 206 lines.

## Acceptance notes

- The latest direct user instruction selected the in-app browser; all browser evidence is labeled IAB and is not represented as Chrome/Computer Use evidence.
- Task `0.4` remains unchecked because product choices about identity, fees/lot size/matching, instrument scope, minute bars, and authorized provider are not user-confirmed. The current foundation keeps initial cash/session settings configurable and records the remaining risk; Phase 2/3 must not hard-code or silently resolve those decisions.
- The workspace has no Git metadata, so `git status` and `git diff --check` remain unavailable.
- Direct raw SQL rollback without a runner-created migration ledger and the unavailable database-expert path remain documented residual risks; the supported isolated runner rollback/reapply path passed.

## Gate result

Phase 1 tasks `1.1` through `1.12` are complete and may be checked. Phase 2 remains pending until the orchestration state is updated and a new three-lane Phase 2 cycle is started. No real broker, external provider, or production database was contacted.
