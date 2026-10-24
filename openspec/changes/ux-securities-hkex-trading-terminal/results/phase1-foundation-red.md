# Phase 1 RED Evidence

> Superseded by [`phase1-foundation-red-final.md`](phase1-foundation-red-final.md), which records the final split-check test run and exact output. This initial transcript predates that test refinement.

Status: `NEEDS_CONTEXT`

The application implementation is not present in this workspace yet. A source guard was added as an explicitly permitted behavioral RED substitute so missing capabilities fail as assertions rather than as a missing-process, compile, or setup error. It scans project implementation under `apps/` and `deploy/`, and migrations under `db/migrations/`.

Test file: [`tests/phase1_source_guard_test.py`](../../../../tests/phase1_source_guard_test.py)

Command:

```text
python3 tests/phase1_source_guard_test.py
```

Exit code: `1`

Exact output:

```text
test_auth_routes_are_wired (__main__.Phase1SourceGuardTests) ... FAIL
test_auth_schema_has_up_and_rollback_migrations (__main__.Phase1SourceGuardTests) ... FAIL
test_health_endpoint_is_wired (__main__.Phase1SourceGuardTests) ... FAIL
test_passwords_use_a_strong_hash (__main__.Phase1SourceGuardTests) ... FAIL
test_protected_data_is_scoped_to_the_session_actor (__main__.Phase1SourceGuardTests) ... FAIL
test_session_expiry_and_revocation_are_enforced (__main__.Phase1SourceGuardTests) ... FAIL

======================================================================
FAIL: test_auth_routes_are_wired (__main__.Phase1SourceGuardTests) (action='register')
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 34, in test_auth_routes_are_wired
    self.require(
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '/auth/register' not found in '' : Missing Phase 1 behavior: HTTP route for POST /api/v1/auth/register

======================================================================
FAIL: test_auth_routes_are_wired (__main__.Phase1SourceGuardTests) (action='login')
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 34, in test_auth_routes_are_wired
    self.require(
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '/auth/login' not found in '' : Missing Phase 1 behavior: HTTP route for POST /api/v1/auth/login

======================================================================
FAIL: test_auth_routes_are_wired (__main__.Phase1SourceGuardTests) (action='logout')
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 34, in test_auth_routes_are_wired
    self.require(
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '/auth/logout' not found in '' : Missing Phase 1 behavior: HTTP route for POST /api/v1/auth/logout

======================================================================
FAIL: test_auth_routes_are_wired (__main__.Phase1SourceGuardTests)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 39, in test_auth_routes_are_wired
    self.require(r"/me", self.text, "authenticated current-user route GET /api/v1/me")
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '/me' not found in '' : Missing Phase 1 behavior: authenticated current-user route GET /api/v1/me

======================================================================
FAIL: test_auth_schema_has_up_and_rollback_migrations (__main__.Phase1SourceGuardTests)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 73, in test_auth_schema_has_up_and_rollback_migrations
    self.require(r"(?i)create\s+table\s+.*users", self.migration_text, "users table migration")
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '(?i)create\\s+table\\s+.*users' not found in '' : Missing Phase 1 behavior: users table migration

======================================================================
FAIL: test_health_endpoint_is_wired (__main__.Phase1SourceGuardTests)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 91, in test_health_endpoint_is_wired
    self.require(r"(?i)(/health|/healthz|/readyz)", self.text, "backend health endpoint")
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '(?i)(/health|/healthz|/readyz)' not found in '' : Missing Phase 1 behavior: backend health endpoint

======================================================================
FAIL: test_passwords_use_a_strong_hash (__main__.Phase1SourceGuardTests)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 54, in test_passwords_use_a_strong_hash
    self.require(
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '(?i)(bcrypt|argon2id|argon2)' not found in '' : Missing Phase 1 behavior: bcrypt or Argon2 password hashing

======================================================================
FAIL: test_protected_data_is_scoped_to_the_session_actor (__main__.Phase1SourceGuardTests)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 61, in test_protected_data_is_scoped_to_the_session_actor
    self.require(
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '(?i)(user_id|userid|actor_id|actorid)' not found in '' : Missing Phase 1 behavior: authenticated actor identity available to protected handlers

======================================================================
FAIL: test_session_expiry_and_revocation_are_enforced (__main__.Phase1SourceGuardTests)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 42, in test_session_expiry_and_revocation_are_enforced
    self.require(
  File "/Users/hongxincong/project/UX Security/tests/phase1_source_guard_test.py", line 29, in require
    self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")
AssertionError: Regex didn't match: '(?i)(expires_at|expiresat|expiry|expiration)' not found in '\\n' : Missing Phase 1 behavior: session expiry field or check

----------------------------------------------------------------------
Ran 6 tests in 0.007s

FAILED (failures=9)
```

Interpretation: all six test groups failed because their expected implementation source is absent. The failures are assertion failures naming missing auth routes, migration/rollback, health, password hashing, actor-scoped access, and session expiry/revocation. This establishes RED against the current source state, not runtime behavior.

Not run: behavioral API integration tests, GREEN tests, migration dry-run/apply/rollback against MySQL, Docker cold-start checks for this project, and browser screenshots. Those require the Phase 1 implementation.
