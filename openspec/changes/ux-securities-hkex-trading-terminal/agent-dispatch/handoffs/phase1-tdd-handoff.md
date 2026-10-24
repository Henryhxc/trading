# Phase 1 TDD Handoff

Status: `NEEDS_CONTEXT`

## RED evidence

- Source guard: [`tests/phase1_source_guard_test.py`](../../../../../tests/phase1_source_guard_test.py)
- Exact command output: [`results/phase1-foundation-red-final.md`](../../results/phase1-foundation-red-final.md)
- Command: `python3 tests/phase1_source_guard_test.py`
- Result: exit code 1; 14 tests ran and failed with 14 assertion failures. Failures identify missing register/login/logout/me route wiring, auth schema/rollback migrations, health endpoint, strong password hash, actor-scoped access, and session expiry/revocation.

The source guard was explicitly authorized for this empty-source state. It establishes source-level RED only; it does not establish runtime API behavior. It ran without syntax, dependency, or service-availability errors.

## Remaining verification

- Wait for Phase 1 implementation before targeted GREEN or API smoke.
- Run migration dry-run/apply/rollback against an isolated project database and save the DB report.
- Start the project's Compose stack and verify MySQL health, backend health, frontend/API connectivity, and phpMyAdmin access.
- Exercise registration/login/logout/session expiry/revocation and user isolation with runtime tests.
- Capture login/registration screenshots at 1440x900, 1280x800, and 390x844; inspect focus, errors, logo proportion, and overflow.
- Independent review and main-agent acceptance remain pending.

Docker is available in this environment, but no UX Securities services or implementation files existed when RED was recorded. The workspace is not a Git repository, so Git diff checks are unavailable.
