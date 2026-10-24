# Phase 1 RED Evidence

Status: `NEEDS_CONTEXT`

The application implementation is not present. The authorized source guard scans project code under `apps/` and `deploy/`, plus SQL migrations under `db/migrations/`. It verifies source-level presence of Phase 1 auth behavior; it does not claim runtime API behavior.

Test file: [`tests/phase1_source_guard_test.py`](../../../../tests/phase1_source_guard_test.py)

Command: `python3 tests/phase1_source_guard_test.py`

Exit code: `1`

Exact output:

```text
FAIL test_accounts_table_migration_exists: Regex didn't match: '(?i)create\\s+table\\s+.*accounts' not found in '' : Missing Phase 1 behavior: accounts table migration
FAIL test_auth_migration_rollback_scripts_exist: [] is not true : Missing Phase 1 behavior: reversible auth migration rollback scripts
FAIL test_auth_routes_are_wired: Regex didn't match: '/auth/register' not found in '' : Missing Phase 1 behavior: HTTP route for POST /api/v1/auth/register
FAIL test_current_user_route_is_wired: Regex didn't match: '/me' not found in '' : Missing Phase 1 behavior: authenticated current-user route GET /api/v1/me
FAIL test_health_endpoint_is_wired: Regex didn't match: '(?i)(/health|/healthz|/readyz)' not found in '' : Missing Phase 1 behavior: backend health endpoint
FAIL test_login_route_is_wired: Regex didn't match: '/auth/login' not found in '' : Missing Phase 1 behavior: HTTP route for POST /api/v1/auth/login
FAIL test_logout_revokes_sessions: Regex didn't match: '(?i)(revoked_at|revokedat|revoke|revoked)' not found in '\\n' : Missing Phase 1 behavior: server-side session revocation on logout
FAIL test_logout_route_is_wired: Regex didn't match: '/auth/logout' not found in '' : Missing Phase 1 behavior: HTTP route for POST /api/v1/auth/logout
FAIL test_passwords_use_a_strong_hash: Regex didn't match: '(?i)(bcrypt|argon2id|argon2)' not found in '' : Missing Phase 1 behavior: bcrypt or Argon2 password hashing
FAIL test_protected_data_queries_are_actor_scoped: Regex didn't match: '(?i)(user_id\\s*=\\s*\\?|user_id\\s*==|user_id\\s*=\\s*actor|actor.*user_id|user_id.*actor)' not found in '' : Missing Phase 1 behavior: repository/data access constrained by the authenticated user
FAIL test_protected_handlers_receive_the_session_actor: Regex didn't match: '(?i)(user_id|userid|actor_id|actorid)' not found in '' : Missing Phase 1 behavior: authenticated actor identity available to protected handlers
FAIL test_session_expiry_is_enforced: Regex didn't match: '(?i)(expires_at|expiresat|expiry|expiration)' not found in '\\n' : Missing Phase 1 behavior: session expiry field or check
FAIL test_sessions_table_migration_exists: Regex didn't match: '(?i)create\\s+table\\s+.*sessions' not found in '' : Missing Phase 1 behavior: sessions table migration
FAIL test_users_table_migration_exists: Regex didn't match: '(?i)create\\s+table\\s+.*users' not found in '' : Missing Phase 1 behavior: users table migration
----------------------------------------------------------------------
Ran 14 tests in 0.003s

FAILED (failures=14)
```

All 14 failures are assertion failures naming absent capabilities; the harness completed without syntax, dependency, or service errors. Password hashing and actor-scoped query checks are source guards, not cryptographic or data-isolation runtime proofs.

Not run: API integration/GREEN, MySQL migration dry-run/apply/rollback, project Compose cold start and service checks, and browser screenshots. These remain pending until implementation exists.
