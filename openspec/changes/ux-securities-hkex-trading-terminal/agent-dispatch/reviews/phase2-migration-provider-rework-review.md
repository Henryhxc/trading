# Phase 2 Migration / Provider Rework Independent Review

Date: 2026-10-24 (Asia/Hong_Kong)

## Conclusion

- **Migration-contract rework packet: `PASS`.** The repository now has the accepted `001` → combined `002` sequence, leaves `003` reserved for `003_orders_positions.sql`, reconciles both legacy ledger names without consuming or deleting the future `003` row, and passes fresh/legacy apply/down/reapply against a real isolated MySQL schema.
- **Provider-config rework packet: `PASS_WITH_NOTES`.** Source, behavioral tests, Compose resolution, documentation, and the React network boundary prove mock-by-default, explicitly unofficial repository mode, status sanitization, and fail-closed unsupported modes. The named running backend container predates this rework and has not been rebuilt, so it is not runtime evidence for the new provider wiring; aggregate runtime acceptance must rebuild it.
- **Combined verdict for these two bounded rework packets: `PASS_WITH_NOTES`.** No must-fix defect was found in either packet's production changes.
- **Phase 2 overall: not passed.** The newly requested AkShare increment and the required IAB three-viewport browser gate are outside these two old packets and remain pending. This review does not authorize Phase 3.

## Scope and sources reviewed

I reviewed `AGENTS.md`, `agent.md`, the change design/database schema/tasks, the market-data/watchlist specs, the Phase 2 packet and phase controller, both RED records and handoffs, the current migration files and migration runner/integration/guard tests, provider configuration/runtime composition/tests, MySQL repository, Docker Compose, market-data boundary documentation, frontend market client, and the Phase 2 DB report. The conclusions below are based on current source plus independent execution, not only the handoffs.

The user's later decision to use AkShare is intentionally treated as a new, unreviewed increment. Its chosen operating mode and implementation evidence are not part of the migration-contract or provider-config rework packets reviewed here.

## Migration-contract review

### Specification compliance

- `db/migrations/` contains only `001_users_sessions_accounts.sql` and the combined `002_stocks_daily_quotes_watchlists.sql` up migrations; there is no Phase 2 `003`. The file guard permits the future slot to be absent or exactly `003_orders_positions.sql`.
- The canonical up migration creates `stocks`, `stock_daily_quotes`, and `watchlists`, then removes only obsolete ledger rows `002_stocks_daily_quotes.sql` and `003_watchlists.sql` (`db/migrations/002_stocks_daily_quotes_watchlists.sql:1-55`). It does not remove or suppress `003_orders_positions.sql`.
- The canonical down migration drops tables in reverse dependency order and deletes the canonical plus obsolete Phase 2 ledger rows, but not the future orders/positions row (`db/migrations/002_stocks_daily_quotes_watchlists.down.sql:1-9`).
- The real-MySQL lifecycle test uses the production runner and canonical SQL. Its legacy fixture adds a test-only `003_orders_positions.sql` probe and proves upgrade leaves a canonical `002` plus future `003`; canonical down leaves future `003` intact; reapply restores only canonical `002` without skipping or rerunning the future migration (`apps/backend/internal/migrations/phase2_contract_integration_test.go:62-122`).
- I independently created the absent disposable schema `ux_phase2_migration_test_review_20261024`, ran the lifecycle, observed both fresh and legacy subtests pass, verified zero tables after test cleanup, dropped that exact schema, and verified it no longer existed. The shared `ux_securities` schema was not migrated, rolled back, or dropped.

### Quality and compatibility

- The legacy upgrade is rerunnable: Phase 2 tables use `CREATE TABLE IF NOT EXISTS`, seed data uses an upsert, and obsolete ledger rows are removed before the runner records canonical `002`.
- The production runner still keys on the complete filename and applies sorted migrations, so a future canonical `003_orders_positions.sql` remains independently discoverable (`apps/backend/internal/migrations/runner.go:20-57`).
- The static guard protects naming/ownership and rollback ordering; the integration test supplies the stronger behavioral evidence against MySQL. This is not a grep-only acceptance.
- The report accurately labels the old split-migration sections historical and superseded. No credential value is present in the checked evidence.
- Residual operational note: MySQL DDL auto-commits, so future non-idempotent additions still require explicit partial-failure handling. This is not a defect in the reviewed canonical migration.

## Provider-config review

### Specification compliance

- Unset `MARKET_DATA_MODE` resolves to `mock` in Go (`apps/backend/internal/config/config.go:32-49`) and Docker Compose (`deploy/docker-compose.yml:25-35`). Independent Compose resolution returned `mock`.
- `mock` composition has no authorized or repository provider and preserves `source=mock`, `data_status=simulated`; explicit `repository` occupies only the unofficial slot and retains mock fallback (`apps/backend/cmd/api/main.go:85-97`).
- The unofficial-slot allow-list is exactly `unofficial`, `delayed`, `stale`, and `unavailable`; `live`, `simulated`, and unknown values are downgraded to `stale`, while the separate mock slot preserves `simulated` (`apps/backend/internal/integration/marketdata/provider.go:43-84`).
- Configuration accepts only `mock` and `repository`. `authorized`, `public`, and unknown strings fail closed during `Load`; repository mode without a dependency also fails closed. No unimplemented adapter is represented as connected.
- The behavioral tests instantiate the real composition, assert the status downgrade, exercise missing-dependency rejection, and send an authenticated quote request through the production HTTP router/service/provider path. They therefore prove behavior rather than source spelling alone (`apps/backend/cmd/api/main_test.go:50-115`).
- `apps/frontend/src/shared/api/market.ts` calls only relative internal API paths, and the frontend source contains no third-party quote URL or provider credential. `docs/api/hkex-market-data.md` and Compose match the implemented modes and safety boundary.

### Quality, errors, and sensitive data

- Unsupported-mode and missing-repository errors contain only the requested mode; database connection/migration startup errors remain generic. No password, token, DSN, provider credential, or raw upstream payload is logged by the changed paths.
- The provider sanitizer is centralized and preserves the distinct mock rule. Focused tests cover `live`, `simulated`, arbitrary status, repository missing, unsupported modes, and the mock API smoke.
- Reviewed production files remain below the 500-line limit: `config.go` 95 lines, `cmd/api/main.go` 98, and `provider.go` 173.
- Runtime note: the currently named backend container was created before the rework source timestamps, exposes an empty `MARKET_DATA_MODE`, and therefore cannot prove the rebuilt implementation. Its internal `/health` succeeds, but aggregate runtime acceptance must rebuild/recreate the service and recheck the quote response as `mock/simulated`. This note makes the packet `PASS_WITH_NOTES`, not `FAIL`, because current source behavior and resolved Compose configuration are independently proven.

## Independent verification

All successful checks below were run against the current tree:

```text
python3 -m unittest discover -s tests -p 'phase2*.py' -v
PASS: 19 tests

cd apps/backend && go test ./internal/config ./internal/integration/marketdata ./cmd/api ./internal/migrations -count=1
PASS: all four packages

cd apps/backend && go vet ./...
PASS

docker compose -f deploy/docker-compose.yml config -q
PASS

env -u MARKET_DATA_MODE docker compose -f deploy/docker-compose.yml config --format json | jq -r '.services.backend.environment.MARKET_DATA_MODE'
mock

npx openspec validate ux-securities-hkex-trading-terminal --strict
PASS

PHASE2_MIGRATION_TEST_DSN=<redacted isolated schema DSN> \
  go test ./internal/migrations -run '^TestPhase2MigrationContractLifecycle$' -count=1 -v
PASS: fresh apply/down/reapply and legacy upgrade/down/reapply
schema before create=0; after create=1; tables after test cleanup=0; schema after exact drop=0
```

## Remaining Phase 2 gates (not findings against these packets)

1. **AkShare increment:** implement and independently verify the newly scoped local historical daily-candle import path, provenance/status contract, idempotency, validation/failure behavior, MySQL/API read path, and absence of a Python runtime business service or extra port. Until that new scope has its own evidence and review, it is unaccepted.
2. **IAB browser gate:** `results/screenshots/phase2-market-workbench/` still contains only `.gitkeep`. The required `1440x900`, `1280x800`, and `390x844` IAB evidence for search, quote, candles, watchlist, empty/error/stale/unavailable states, responsive overflow, and scroll anchoring is still missing.
3. Rebuild/recreate the backend container from the reviewed source before aggregate runtime acceptance; do not use the older healthy container as proof of the provider-config change.

Phase 2 must remain in rework/validation, its completion checkboxes must remain under the main agent's gate, and Phase 3 must not start from this review.
