# Phase 1 Runtime QA Handoff

Status: `NEEDS_REWORK`

## Completed

- Read Phase 1 packet, project rules, auth OpenAPI contract, code/TDD handoffs, and Computer Use skill.
- Ran [`tests/test_phase1_runtime.py`](../../../../../tests/test_phase1_runtime.py) against the already-running backend. Register, login, logout/revocation, wrong/unknown credentials, duplicate username/email, protected `/me`, expired session and per-user account-row assertions passed.
- Verified Compose process health, backend health, frontend response and API proxy, phpMyAdmin HTTP login page, direct MySQL migration up/down/raw reapply, and actual backend-runner up/down/reapply on isolated schemas.
- Chrome inspected login and registration at native 1359x768; the logo renders broken because the asset endpoint returns 403. The lane reported screenshot captures, but the named files are absent from the evidence directory on main-agent inspection; they are not accepted as saved evidence.
- Wrote runtime and DB findings in [`phase1-foundation-runtime.md`](../../results/phase1-foundation-runtime.md) and [`2026-10-24-phase1-auth.md`](../../results/db-change-reports/2026-10-24-phase1-auth.md).

## Blocking findings

1. Runner rollback/reapply fails: down migration leaves `schema_migrations` version, so runner startup exits successfully but leaves auth tables absent. The second runner invocation was performed on an isolated schema; application DB was not rolled back.
2. UX Securities logo request returns 403 and Chrome displays broken image/alt text. Independent review attributed this to source asset mode `0600` retained in nginx image.

## Pending evidence

- Screenshot files are absent; requested exact-size screenshots at 1440x900, 1280x800 and 390x844, mobile overflow inspection, keyboard focus, and UI-visible invalid-credential error remain pending.
- phpMyAdmin UI login and application database table-list screenshot. HTTP page is reachable; direct SQL checks passed for constraints, but UI visibility remains unverified.
- Independent review and main-agent acceptance.

No production code, OpenAPI, tasks, orchestration state, running service, or application schema was changed. Four known synthetic QA user rows remain in the app database; temporary migration schemas and their grant were removed. No trading API or real order path was exercised.

Recommendation: do not pass Phase 1 until the migration runner semantics and Logo permissions are fixed, then capture the remaining visual/database UI evidence and rerun review.
