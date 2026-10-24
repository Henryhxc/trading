# Phase 1 Main-Agent QA (2026-10-24)

Status: `BLOCKED`; Phase 1 remains unaccepted and all `tasks.md` checkboxes remain open.

## Computer Use

- Opened the running frontend in Chrome at `http://localhost:3000/register`.
- Computer Use displayed the registration form and UX Securities Logo with device emulation set to 390x844. A capture was saved as `results/screenshots/phase1-foundation/register-390x844-cua.jpg`; it is cropped to exclude unrelated Chrome tabs.
- The registration page showed username, email, password, password guidance, and the paper-trading-only notice. This partial capture does not prove keyboard focus, login errors, loading/retry behavior, protected-route redirect, or absence of horizontal overflow.
- Subsequent Computer Use calls returned: “The Mac is locked and automatic unlock could not unlock it.” Required login/error/focus and protected-route observations, 1440x900 and 1280x800 captures, and phpMyAdmin UI table browsing remain pending.

## Independent Verification

Commands run against the current Phase 1 source state:

| Command | Result |
|---|---|
| `cd apps/backend && go test ./...` | PASS; package tests passed, cached results |
| `cd apps/backend && go vet ./...` | PASS |
| `cd apps/frontend && npm run build` | PASS; TypeScript and Vite production build |
| `python3 tests/phase1_rework_contract_test.py` | PASS; 3 tests |
| `python3 tests/phase1_source_guard_test.py` | PASS; 14 tests |
| `UX_BASE_URL=http://localhost:8080 python3 tests/test_phase1_runtime.py` | PASS; 1 auth lifecycle test |
| `PHASE1_MIGRATION_TEST_DSN=<redacted isolated schema> go test ./internal/migrations -run TestRunnerReappliesMigrationAfterRollback -count=1 -v` | PASS; clean main-agent rerun on `ux_phase1_test_codex_final_20261024`; schema and schema-scoped grant cleanup independently verified |
| `npx openspec validate ux-securities-hkex-trading-terminal --strict` | PASS |

Main-agent recovery rerun (2026-10-24T03:41–03:46+08:00) independently repeated `go test ./... -count=1`, `go vet ./...`, `npm run build`, both Phase 1 Python guards, the isolated MySQL runner rollback/reapply test, and strict OpenSpec validation; all passed. No auth lifecycle runtime smoke was rerun in this recovery to avoid adding more synthetic rows to the shared app DB. The final isolated migration test schema was verified absent before creation, then dropped by the successful cleanup hook; metadata confirmed both the schema count and its schema-privilege count returned to `0`.

## Docker Cold Start (2026-10-24T03:52+08:00)

The main agent ran `docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml down` without `-v`, then `docker compose -p ux-securities-hkex-trading-terminal --profile admin -f deploy/docker-compose.yml up -d --build --wait --wait-timeout 120`; the second command exited 0. The first command omitted the optional admin profile, so Compose removed core services while the existing phpMyAdmin container briefly kept the network in use. Starting again with `--profile admin` recreated phpMyAdmin as well. No named volume was removed; `ux-securities-hkex-trading-terminal_ux_mysql` remained present.

Final `docker compose ps` showed frontend, backend, and MySQL `healthy`; phpMyAdmin was running. Current resolved mappings are `3000/8080/3306/8081`, matching documented defaults. Earlier observations at `8183` were from a previous environment override that was not set for this cold-start command.

After cold start, host HTTP checks returned frontend login `200`, Logo `200`, backend `/health` `200`, phpMyAdmin `200` at `http://localhost:8081/`, and frontend-proxied anonymous `/api/v1/me` `401`. Read-only MySQL inspection found `users`, `sessions`, `accounts`, `schema_migrations`, and the applied migration version `001_users_sessions_accounts.sql`. This proves the retained database volume and migration state survived; the HTTP/SQL checks still do not substitute for browsing the table list in phpMyAdmin.

Read-only MySQL inspection found the Phase 1 tables `users`, `sessions`, `accounts`, and `schema_migrations`, plus unique constraints for username, email, session token hash, and one account per user. The auth runtime test writes synthetic QA accounts and sessions; they are retained and were not deleted. The Phase 1 DB report separately records isolated migration rollback/reapply evidence.

The production-source line-count scan found no file above 500 lines (largest scanned production source: `apps/backend/internal/httpapi/server.go`, 206 lines). `git status`, `git diff`, and `git diff --check` remain unavailable because this workspace has no `.git` metadata.

## In-app Browser Checks (2026-10-24T03:35+08:00)

The user explicitly authorized the in-app browser for this resumed check. It opened the running app at `http://localhost:3000/login`; the page exposed the UX Securities logo and labeled sign-in controls, and no browser console warnings/errors were observed. Screenshots are saved under `results/screenshots/phase1-foundation/`:

- `login-1440x900-iab.jpg`, `login-1280x800-iab.jpg`, `login-390x844-iab.jpg`
- `login-focus-390x844-iab.jpg`, `login-empty-submit-loading-390x844-iab.jpg`, `login-empty-submit-error-390x844-iab.jpg`
- `register-1440x900-iab.jpg`, `register-390x844-iab.jpg`, `register-1280x800-iab.jpg`
- `phpmyadmin-login-1440x810-iab.jpg` (actual captured viewport height: 810px)
- `phpmyadmin-login-current-port-iab.jpg` (1440x900 at the post-cold-start URL `http://localhost:8081/`)

Observed: focus reached the sign-in identity input; empty submit showed a transient `Please wait` state and then a visible `alert` reading “Username or password is incorrect”; `/workspace` while anonymous and an unknown path both redirected to `/login`; the 390px page had no horizontal overflow. The backend was also called directly with a synthetic nonexistent identity and synthetic password; it returned HTTP 401 with code `invalid_credentials` and the same user-safe message. No user-provided credential was used. After cold start, phpMyAdmin is reachable at `http://localhost:8081/` and displays its login form; no credentials were entered and database-table browsing remains unverified. The viewport override was reset after capture.

This in-app-browser evidence is supplemental and is not represented as Chrome/Computer Use evidence. The phase packet explicitly makes Chrome/Computer Use the primary UI gate, so browser-check remains blocked pending that gate or an explicit project-level change. Action-time confirmation is needed before entering the local MySQL credential into phpMyAdmin at `http://localhost:8081/`. No tasks were checked, no application data was changed, and no new independent review was started.

## Gate Decision

The latest direct user request selects IAB for browser work, superseding the older Chrome/Computer Use surface in the attached execution prompt. The login/register and phpMyAdmin IAB checks now satisfy the active browser gate; evidence remains labeled IAB, not CUA. Phase 1 remains in validation because the independent post-rework review has not run and product decisions in task 0.4 remain open. Do not start Phase 2 until the review and main acceptance pass.

## Resumed Verification (2026-10-24T04:14+08:00)

The main agent rechecked the current workspace and running project without changing production code or application data:

| Command / observation | Result |
|---|---|
| `go test ./... -count=1 && go vet ./...` in `apps/backend` | PASS; package tests and vet completed, command exit 0 |
| `npm run build` in `apps/frontend` | PASS; TypeScript and Vite production build, exit 0 |
| `python3 tests/phase1_rework_contract_test.py` | PASS; 3 tests |
| `python3 tests/phase1_source_guard_test.py` | PASS; 14 tests |
| `openspec validate ux-securities-hkex-trading-terminal --strict` | PASS |
| `docker compose -p ux-securities-hkex-trading-terminal --profile admin -f deploy/docker-compose.yml ps` | PASS; frontend/backend/MySQL healthy, phpMyAdmin running; ports 3000/8080/3306/8081 |
| `GET /health`, `GET /login`, read-only MySQL `SHOW TABLES` | PASS; backend `{"status":"ok"}`, frontend 200, tables `accounts`, `schema_migrations`, `sessions`, `users` |
| In-app browser live page observation | PASS for the observed 1280x720 viewport only; login route loaded, logo `naturalWidth=428`, document width equaled viewport width |
| Visual review of saved IAB screenshots | PASS for the observed login/register states and available 1440x900, 1280x800, and 390x844 images; no apparent clipping or horizontal overflow in the inspected images |

At 04:14, the project packet still named Chrome/Computer Use as its primary UI gate. The IAB observations and screenshots remain explicitly supplemental. The saved `register-390x844-cua.jpg` is a partial 500x500 capture, not proof of the complete viewport gate. At that time phpMyAdmin had only been opened to its unauthenticated login page; direct SQL table inspection did not prove UI browsing. The post-rework independent review was not started because the browser gate remained unresolved; the attempted review lane was interrupted before writing or changing any review artifact to preserve gate order. At 04:14 all 43 task checkboxes remained open and Phase 2 was pending. The 04:24 phpMyAdmin follow-up below supersedes only the UI table-list observation.

## phpMyAdmin UI Table Browsing (2026-10-24T04:24+08:00)

Using the local Compose development application account, the main agent signed in to `http://localhost:8081/`, opened the `ux_securities` database structure page, verified the visible table list (`accounts`, `schema_migrations`, `sessions`, `users`) and logged out. No query was submitted and no database write was performed. The page showed row counts only; no user row contents or credentials were captured. Screenshot: `results/screenshots/phase1-foundation/phpmyadmin-tables-1280x720-iab.jpg` (1280x720). This closes the phpMyAdmin UI table-list observation, but does not satisfy the separate Chrome/Computer Use browser gate.
