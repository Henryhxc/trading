# Phase 2/3 Independent Review Refresh

- Review date: 2026-10-24 (Asia/Singapore)
- Review scope: Phase 2 AkShare scheduler/market-data evidence and Phase 3 Go paper-trading backend, React frontend, Compose, migrations, tests, API smoke evidence, browser gate, and remaining task items.
- Review mode: read-only review; production code, `tasks.md`, and orchestration state were not modified.
- Decision: **NEEDS_REWORK**

## Main-agent follow-up — 2026-10-24 17:47 Asia/Hong_Kong

This is a factual follow-up, not a replacement independent review or a change to the decision above. The browser-unavailable statements below describe the earlier review session. The main agent has since used the authenticated in-app browser to capture the deployed workbench, favorites, empty portfolio/history, no-quote and minute-empty states at the actual **1311×1076** viewport. See `results/workbench-layout/verification-2026-10-24.md` and its linked screenshots. The coordinated navy background, left directory / center chart / right paper-order layout and default 2,791-security directory were visually checked; no order was submitted.

Fresh full Go tests/vet, frontend tests/build and strict OpenSpec validation passed. Runtime checks used the correct Compose project, `ux-securities-hkex-trading-terminal`, and observed healthy frontend/backend/MySQL/AkShare plus running phpMyAdmin. This is not a fresh cold-start claim. The additional independent source review in `results/workbench-layout/independent-review.md` is PASS_WITH_NOTES; its independent visual review remains BLOCKED.

Required three-viewport Phase 2/3 trading/stream scenarios, successful real AkShare upstream ingestion and aggregate independent visual acceptance remain open. Existing actual-viewport screenshots do not replace those gates. Overall status remains **rework / NEEDS_REWORK**.

## Executive decision

The prior Phase 3 idempotency finding is fixed in the current frontend source: `tradingApi.createOrder` generates one UUID, sends it as JSON `client_order_id`, and sends the same value as `Idempotency-Key` (`apps/frontend/src/shared/api/trading.ts:21-24`). The frontend source guard now has a dedicated assertion for this contract and passes.

The phase gate still does not pass. Required in-app-browser evidence is absent and real AkShare upstream evidence is a recorded failure rather than a successful sample. The permitted current full Go suite, frontend checks, AkShare tests, strict OpenSpec validation, and Docker runtime are green; refreshed API evidence also proves two-user order isolation. Missing screenshots and real upstream success remain blockers.

## Findings requiring rework

### 1. Browser gate remains blocked

`results/browser-tool-blocker.txt` records that the configured in-app browser failed before navigation with `sandboxCwd must be an absolute file URI: relative URL without a base`. `results/screenshots/phase2-market-workbench/` and `results/screenshots/phase3-paper-trading/` contain only `.gitkeep`. Therefore the required Phase 3 register/login → search → quote → watchlist → order → simulated fill → positions/history flow, three viewports (`1440x900`, `1280x800`, `390x844`), responsive overflow check, and visible error-state check remain unverified.

### 2. Real AkShare upstream evidence is failure evidence, not a pass

`results/akshare-live-sample.txt` records a real upstream pull failure (`RemoteDisconnected` / connection aborted). The Python scheduler contract tests pass, and the AkShare container is reported healthy in the saved Docker evidence, but there is no successful real upstream sample proving scheduler → signed webhook → MySQL row → REST → authenticated WebSocket. The requested failure evidence is preserved accurately; it cannot be rewritten as real-source success. Tasks `2.11` and `2.15` remain open.

### 3. Full Go test gate is green after permitted rerun

The initial sandbox rerun hit an IPv6 listener restriction, but the permitted rerun passed `GOCACHE=/tmp/ux-go-build go test ./... -count=1` and `go vet ./...`. Evidence is in `results/go-test-final-after-volume-fix-escalated.txt` and `results/go-vet-final-after-volume-fix-escalated.txt`.

### 4. Two-user isolation evidence is refreshed

The refreshed `results/api-two-user-isolation.txt` proves user A can create an order while user B receives `404 not_found` for that order and has empty positions. Concurrent/rollback evidence remains covered by Go integration tests rather than browser evidence.

### 5. AkShare cache fallback defect fixed

The repository quote path now converts DECIMAL volume values safely and the provider keeps stale unofficial cache visible instead of silently selecting mock. `results/phase2-stale-provider-red.txt`, `results/phase2-stale-provider-green.txt`, `results/phase2-runtime-provider-green.txt`, and `results/api-akshare-after-volume-fix.json` show RED → GREEN and runtime `source=akshare`, `data_status=stale`.

### 6. Remaining task/orchestration state is not complete

`tasks.md` still has unchecked items including `0.5`, `2.9-2.12`, `2.15`, `3.9`, `3.10`, `3.12`, `4.1`, and `4.2`. `agent-dispatch/orchestration-state.md` remains `status: rework`, with Phase 2 browser blocked and review needing rework. This review does not change either file.

## Resolved finding

- **Frontend idempotency contract:** resolved in source. `apps/frontend/src/shared/api/trading.ts:22-23` uses one `clientOrderID` for both body `client_order_id` and header `Idempotency-Key`; `apps/frontend/tests/source-guard.mjs` asserts both fields. `npm run test:frontend` passed all 6 assertions. The old finding that the production UI omitted `client_order_id` should not be carried forward as an active blocker.

## Verified or substantially aligned

- **Frontend checks:** `npm run test:frontend` passed; `npm run build` passed (`tsc -b && vite build`).
- **Python checks:** `python3 -m unittest discover -s services/akshare/tests -p 'test_*.py'` passed 8 tests. These are contract tests, not a real upstream success proof.
- **Focused Go checks:** `GOCACHE=/tmp/ux-go-build go test ./internal/service/trading ./internal/repository/mysqlrepo -count=1` passed.
- **Recorded backend evidence:** prior full Go artifact is present; current full-suite rerun is separately recorded above as sandbox-blocked. `go vet` output artifact is empty, so no vet failure is evidenced, but this review does not convert an empty artifact into a fresh PASS.
- **Compose/runtime observation:** `docker ps` shows the target frontend, backend, MySQL, and AkShare containers running; saved `docker-compose-ps-final.txt` and `docker-health-final.txt` report the target core services healthy. A fresh `docker compose -f deploy/docker-compose.yml ps` from the default project only showed `deploy-mysql`, so a fresh complete Compose cold-start/health proof is not claimed here.
- **Schema and boundary:** reviewed migrations keep Phase 3 to `orders` and `positions`; no real broker/execution venue is connected; frontend market/trading calls remain internal API calls.
- **Security design:** webhook HMAC, timestamp/replay checks, authenticated stream wiring, user-scoped service/repository calls, and paper-only execution remain aligned by source inspection and recorded focused tests.

## Required next evidence before PASS

1. Repair/select the configured in-app browser and capture the required Phase 2 and Phase 3 screenshots at all three viewports, including error/empty/loading and overflow checks.
2. Rerun full Go tests and vet in an environment where the stream test can bind its local listener; retain the exact command/output.
3. Capture a successful real AkShare pull and the complete signed webhook → DB → REST → authenticated WebSocket chain, while retaining the existing upstream failure as failure-path evidence.
4. Add two authenticated synthetic users and record cross-user read/cancel/detail/account/position denial plus transaction rollback/concurrency evidence.
5. Only after those gates pass should the main agent update task checkboxes and orchestration state.

## Commands checked in this refresh

```text
npm run test:frontend
npm run build
python3 -m unittest discover -s services/akshare/tests -p 'test_*.py'
GOCACHE=/tmp/ux-go-build go test ./... -count=1
GOCACHE=/tmp/ux-go-build go test ./internal/service/trading ./internal/repository/mysqlrepo -count=1
docker compose -f deploy/docker-compose.yml ps
docker ps --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}'
```
