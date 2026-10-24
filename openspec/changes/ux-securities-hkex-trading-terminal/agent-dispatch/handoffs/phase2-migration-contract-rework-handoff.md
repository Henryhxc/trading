# Phase 2 Migration Contract Rework Handoff

Date: 2026-10-24 (Asia/Hong_Kong)
Status: `DONE`
Packet: Phase 2 migration-contract rework only

## Outcome

The accepted migration sequence is now represented on disk:

```text
001_users_sessions_accounts.sql
002_stocks_daily_quotes_watchlists.sql
003_orders_positions.sql (reserved; not implemented by this packet)
```

Phase 2 now has one combined up migration and one combined down migration. The former
split files `002_stocks_daily_quotes.sql`, `002_stocks_daily_quotes.down.sql`,
`003_watchlists.sql`, and `003_watchlists.down.sql` were removed from the release
directory.

No Phase 3 production code or schema was added. The backend migration runner did not need
to change: it already keys the ledger by complete filename. Compatibility is implemented
inside canonical `002`, where idempotent table creation/upsert is followed by deletion of
the two legacy ledger rows. The runner then inserts the canonical row. This converges an
old local database to one canonical Phase 2 row, and a future
`003_orders_positions.sql` remains a distinct unapplied filename.

The canonical down migration removes `watchlists`, `stock_daily_quotes`, and `stocks` in
reverse dependency order and clears both canonical and obsolete ledger names.

## RED → GREEN → REFACTOR

RED was written and observed before production migration changes:

```text
python3 tests/phase2_migration_contract_guard_test.py
exit=1
Ran 4 tests
FAILED (failures=4)
```

The failures were the intended contract failures: missing canonical up/down, the legacy
split `002`, and `003_watchlists.sql` occupying the reserved slot. Full failure output is
in `results/phase2-migration-contract-red.md`.

Minimal GREEN combined the unchanged Phase 2 table/seed semantics and added ledger
reconciliation. REFACTOR extended the real-MySQL lifecycle test so both fresh and legacy
histories exercise canonical down/reapply, log exact tables/ledger, and use a Phase 2-only
temporary migration directory so the test remains valid after the real Phase 3 migration
is later added.

## Automated verification

```text
python3 -m unittest discover -s tests -p 'phase2*.py' -v
exit=0; 19 tests passed

cd apps/backend && go test ./...
exit=0; all packages passed

cd apps/backend && go vet ./...
exit=0; no findings

docker compose -p ux-securities-hkex-trading-terminal \
  -f deploy/docker-compose.yml config -q
exit=0
```

The first accidental Python discovery from `apps/backend` found zero tests and is not
counted as evidence. The command above was rerun from the project root and executed all
19 Phase 2 tests.

## Isolated MySQL evidence

Disposable schema: `ux_phase2_migration_test_20261024_rework`.

- Pre-create existence count: `0`.
- Post-create existence count: `1`.
- Real command: `go test ./internal/migrations -run '^TestPhase2MigrationContractLifecycle$' -count=1 -v` with `PHASE2_MIGRATION_TEST_DSN` injected from the local container environment; no credential value was printed or persisted.
- Result: `PASS`, package exit `0`.

Fresh path:

```text
apply tables: accounts, schema_migrations, sessions, stock_daily_quotes, stocks, users, watchlists
apply ledger: 001_users_sessions_accounts.sql, 002_stocks_daily_quotes_watchlists.sql
down tables: accounts, schema_migrations, sessions, users
down ledger: 001_users_sessions_accounts.sql
reapply: restored the same Phase 2 tables and canonical two-row ledger
```

Legacy-ledger path:

```text
fixture ledger: 001 plus old 002_stocks_daily_quotes.sql and 003_watchlists.sql
upgrade ledger: 001, 002_stocks_daily_quotes_watchlists.sql, 003_orders_positions.sql
canonical down ledger: 001, 003_orders_positions.sql
canonical reapply ledger: 001, 002_stocks_daily_quotes_watchlists.sql, 003_orders_positions.sql
```

The `003_orders_positions.sql` in this test was a temporary probe migration that created
only `migration_phase3_probe` inside the disposable schema. It proves the future 003 file
executes after legacy reconciliation; it is not a repository migration or Phase 3
implementation.

Cleanup evidence:

```text
tables in disposable schema after test cleanup: 0
schema existence immediately before drop: 1
schema existence after drop: 0
```

The shared `ux_securities` schema was never dropped, rolled back, or migrated during this
packet.

## Files

Added:

- `tests/phase2_migration_contract_guard_test.py`
- `apps/backend/internal/migrations/phase2_contract_integration_test.go`
- `db/migrations/002_stocks_daily_quotes_watchlists.sql`
- `db/migrations/002_stocks_daily_quotes_watchlists.down.sql`
- `openspec/changes/ux-securities-hkex-trading-terminal/results/phase2-migration-contract-red.md`
- this handoff

Updated:

- `docs/design/develop-backend-design.md`
- `openspec/changes/ux-securities-hkex-trading-terminal/results/db-change-reports/2026-10-24-phase2-market.md`

Removed:

- `db/migrations/002_stocks_daily_quotes.sql`
- `db/migrations/002_stocks_daily_quotes.down.sql`
- `db/migrations/003_watchlists.sql`
- `db/migrations/003_watchlists.down.sql`

`apps/backend/internal/migrations/runner.go`, `tasks.md`, `phase-controller.md`, and
`orchestration-state.md` were not modified. Frontend, market/watchlist business logic,
and Phase 3 production code were not modified.

## Residual risk and next owner

- The running shared application container was intentionally not restarted. Its old
  ledger should be observed converging during the next authorized deployment/startup;
  isolated MySQL proves the compatibility path without mutating shared data.
- MySQL DDL auto-commits; canonical statements are rerunnable and rollback is tested, but
  future non-idempotent additions would need the same partial-failure review.
- Historical runtime/review/handoff files outside this packet still mention the old split
  names. They were left as historical evidence; the current DB report clearly labels
  those sections superseded.

Next action belongs to the parent/review lane: independently inspect this packet and rerun
the Phase 2 gate. This implementer does not advance Phase 3 or mark global tasks complete.
