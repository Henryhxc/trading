# Phase 1 Runtime Verification

Date: 2026-10-24 (Asia/Hong_Kong)
Status: `NEEDS_REWORK`

## Scope and environment

- Existing Compose project: `ux-securities-hkex-trading-terminal`; no rebuild or stop was issued.
- Frontend `http://localhost:3100`, backend `http://localhost:8182`, MySQL host port `3308`, phpMyAdmin `http://localhost:8183`.
- All auth test identities/passwords were synthetic. No real trading or execution API exists in this phase.
- The integration test created four known synthetic users across the initial failed harness attempt and the successful run. Their app DB rows were retained rather than deleting from the shared application database. The temporary migration schemas were removed.
- Test source: [`tests/test_phase1_runtime.py`](../../../../tests/test_phase1_runtime.py).

## Results

| Check | Result | Evidence |
|---|---|---|
| Register and account creation | PASS | Runtime integration test returned 201 and verified response user and cookie attributes. |
| Login by email and username | PASS | Login returned 200 and the same user; new session cookie differed from registration cookie. |
| Wrong and unknown credentials | PASS | Both returned 401 `invalid_credentials`; wrong-password message matched the contract. |
| Duplicate username and email | PASS | Both returned 409 `identity_unavailable`. |
| Protected `/me` | PASS | Anonymous returned 401; authenticated session returned 200 with current user. |
| Logout/revocation | PASS | Logout returned 204; the same cookie then returned 401 from `/me`. |
| Expired session | PASS | SQL expired only the test fixture's session digest; `/me` then returned 401. |
| Per-user account separation | PASS | SQL asserted two distinct users, two account rows, and configured initial/balance/available cash for each. |
| Compose services / MySQL health | PASS | `docker compose ... ps --all` showed backend/frontend/MySQL healthy; phpMyAdmin was Up (no healthcheck configured). |
| Backend health | PASS | `GET /health` returned 200 `{"status":"ok"}`. |
| Frontend and API proxy | PASS | Frontend returned 200 HTML; proxied unauthenticated `/api/v1/me` returned the backend's 401 JSON contract. |
| phpMyAdmin HTTP page | PASS | `GET http://localhost:8183/` returned 200 and rendered the phpMyAdmin login page. |
| phpMyAdmin database/table browser | UNVERIFIED | No phpMyAdmin UI login/table navigation was performed after the main agent requested lane closure. Schema/table assertions were made directly via SQL. |
| Migration direct up/down/reapply | PASS | Isolated `ux_phase1_qa_20261024`: up yielded `accounts`, `sessions`, `users`; down yielded 0 base tables; raw up reapplied successfully. |
| Migration runner rollback/reapply | FAIL | Isolated `ux_phase1_reapply_qa_20261024`: runner recorded version, down removed auth tables but retained ledger row; runner reapply skipped migration, leaving only `schema_migrations`. See DB report. |
| Logo asset / desktop UI | FAIL | Chrome displayed broken-image icon and alt text. `GET /ux-securities-logo.png` returned `403 Forbidden`; independent review attributed this to mode `0600` preserved into nginx. |
| Requested exact-size screenshots | PENDING | Chrome's available viewport was 1359x768. Requested 1440x900, 1280x800 and 390x844 emulation was not completed. The two screenshot paths referenced below were checked after handoff and are absent; no screenshot file is accepted as evidence. |
| Keyboard focus / UI error-state / mobile overflow | PENDING | Not covered in this lane before closure. API invalid-credential branches passed. |

## Commands and outputs

Successful runtime auth test:

```text
$ python3 tests/test_phase1_runtime.py
test_auth_lifecycle_and_session_guards (__main__.Phase1RuntimeAuthTest) ... ok

----------------------------------------------------------------------
Ran 1 test in 1.892s

OK
exit_code=0
```

The first harness attempt reached account/session setup but failed before the expiry assertion because the MySQL container has `MYSQL_DATABASE`, not `DB_NAME`, in its environment (`ERROR 1046 (3D000): No database selected`). The helper was corrected; the subsequent fresh-fixture run above passed. No test was run against real orders.

Runtime service checks:

```text
$ docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml ps --all
backend   Up About a minute (healthy)  0.0.0.0:8182->8080/tcp
frontend  Up 53 seconds (healthy)      0.0.0.0:3100->80/tcp
mysql     Up About a minute (healthy)  0.0.0.0:3308->3306/tcp
phpmyadmin Up About a minute

$ curl -sS -i http://localhost:8182/health
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
{"status":"ok"}

$ curl -sS -i http://localhost:3100/api/v1/me
HTTP/1.1 401 Unauthorized
Content-Type: application/json; charset=utf-8
{"code":"unauthenticated","message":"Please sign in","request_id":"<generated request id>"}

$ curl -sS -I http://localhost:3100/ux-securities-logo.png
HTTP/1.1 403 Forbidden
```

The phpMyAdmin HTTP page returned `HTTP/1.1 200 OK` and the `Welcome to phpMyAdmin` login form. Cookie/session identifiers from its HTML response are intentionally not copied into this report.

Migration runner verification (the temporary isolated schema was dropped after evidence capture):

```text
1. DB_NAME=ux_phase1_reapply_qa_20261024 PORT=0 timeout 2 /app/ux-api
   output: 2026/10/23 18:21:39 api listening port=0
   exit_code=0
   tables: accounts, schema_migrations, sessions, users
   ledger: 001_users_sessions_accounts.sql
2. Apply db/migrations/001_users_sessions_accounts.down.sql to that isolated schema
   exit_code=0
   tables: schema_migrations
   ledger: 001_users_sessions_accounts.sql
3. Run the same backend migration runner again
   output: 2026/10/23 18:22:26 api listening port=0
   exit_code=0
   tables: schema_migrations
   ledger row count: 1
```

The first raw migration-only schema reported `accounts`, `sessions`, `users` and the expected unique username/email/account-user constraints plus user foreign keys. The down migration left 0 base tables; a raw SQL reapply recreated all three. This does not clear the runner defect above because the runner's ledger row is what makes startup skip a rolled-back migration.

## Docker log notes

- Backend log: `api listening port=8080`; repeated health requests logged status 200 with request IDs and no credentials/tokens.
- Frontend log: nginx started and health-check GET `/` returned 200.
- MySQL log: MySQL 8.4.9 reached `ready for connections`; Compose health check was healthy. Startup also emitted the standard initialization warning `root@localhost is created with an empty password` before the entrypoint configured the Compose root password; this is isolated local development configuration, not evidence of a production posture.
- phpMyAdmin log: Apache started; HTTP probe returned 200. Apache emitted a non-fatal ServerName warning.

## Browser evidence

- The QA lane reported native-size login and registration captures, but the referenced files `login-desktop-1359x768.jpg` and `register-desktop-1359x768.jpg` are absent from `screenshots/phase1-foundation/`; that directory contains only `.gitkeep`.
- Chrome observation: expected UX Securities logo is unavailable (`403`), so browser renders the broken-image icon plus alt text. Exact requested target viewports and mobile overflow remain pending.

## Required rework before Phase 1 acceptance

1. Fix migration rollback/reapply semantics so the runner can reapply a deliberately rolled-back version, then prove runner up/down/reapply on a fresh isolated schema.
2. Fix logo file readability in the nginx image and recapture browser evidence.
3. Capture/inspect login and registration at 1440x900, 1280x800 and 390x844; verify keyboard focus, visible UI credential error mapping, and horizontal overflow.
4. Log into local phpMyAdmin and verify the application database/table list in the UI, then retain a screenshot.

No production source or app DB schema was modified by this lane. Runtime test-created account rows remain in the app DB; isolated QA schemas and their temporary grant were removed.
