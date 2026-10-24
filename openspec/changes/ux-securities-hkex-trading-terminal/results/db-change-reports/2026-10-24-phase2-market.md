# Phase 2 Market/Watchlist Migration Runtime Report

Date: 2026-10-24 (Asia/Hong_Kong)
Current migration: `002_stocks_daily_quotes_watchlists.sql`
Current down migration: `002_stocks_daily_quotes_watchlists.down.sql`
Current status: `PASS` for canonical fresh apply/down/reapply and legacy-ledger upgrade/down/reapply in an isolated schema.

The original split-migration runtime record is retained in the historical sections below. Its commands and outputs were true for the pre-rework files, but its filenames and ledger are superseded by this addendum; they are not evidence of the current contract.

## Migration-contract rework evidence

### Accepted contract and compatibility behavior

The accepted sequence is `001_users_sessions_accounts.sql` →
`002_stocks_daily_quotes_watchlists.sql` → reserved
`003_orders_positions.sql`. The earlier implementation incorrectly split Phase 2 into
`002_stocks_daily_quotes.sql` and `003_watchlists.sql`.

The canonical `002` is safe for an already-used local development database because its
table statements are idempotent, its seed is an upsert, and its final statement removes
both obsolete ledger rows. The runner then records only the canonical filename. A future
`003_orders_positions.sql` therefore has its own unoccupied canonical ledger name and is
executed normally. The integration test proves this by supplying a temporary
`003_orders_positions.sql` probe after simulating both legacy ledger rows; the probe is
test-only and does not add Phase 3 production schema.

The canonical down migration drops `watchlists`, `stock_daily_quotes`, then `stocks`, and
deletes the canonical and both obsolete ledger names. This permits reapply from either a
fresh canonical history or an upgraded legacy history.

### RED

The pre-fix guard failed with exit `1` and four expected assertion failures: canonical up
and down files were missing, `003_watchlists.sql` occupied the reserved slot, and the
single-file Phase 2 contract was absent. Full output is preserved in
`../phase2-migration-contract-red.md`.

### GREEN and regression commands

```text
$ python3 tests/phase2_migration_contract_guard_test.py
Ran 4 tests in 0.005s
OK
exit=0

$ python3 -m unittest discover -s tests -p 'phase2*.py' -v
Ran 19 tests in 0.133s
OK
exit=0

$ cd apps/backend && go test ./... && go vet ./...
all Go packages passed; go vet emitted no findings
exit=0
```

### Isolated MySQL lifecycle

Disposable schema: `ux_phase2_migration_test_20261024_rework`.

Before creation, the schema existence query returned `0`; after creation it returned `1`.
The schema was created only after that absence check and was granted only to the existing
local application user. The shared `ux_securities` schema was not dropped, rolled back,
or migrated by this rework.

The real migration runner test used the local credential through an environment variable;
the credential value was neither printed nor recorded:

```text
$ cd apps/backend
$ PHASE2_MIGRATION_TEST_DSN="${APP_DB_USER_VALUE}:${APP_DB_PASSWORD_VALUE}@tcp(127.0.0.1:3306)/ux_phase2_migration_test_20261024_rework?parseTime=true" \
    go test ./internal/migrations -run '^TestPhase2MigrationContractLifecycle$' -count=1 -v
```

Fresh apply result:

```text
tables: [accounts schema_migrations sessions stock_daily_quotes stocks users watchlists]
schema_migrations: [001_users_sessions_accounts.sql 002_stocks_daily_quotes_watchlists.sql]
```

Fresh canonical down result:

```text
tables: [accounts schema_migrations sessions users]
schema_migrations: [001_users_sessions_accounts.sql]
```

Fresh reapply restored the same seven tables and canonical two-row ledger.

Legacy-ledger fixture started with the Phase 2 tables present and rows named
`002_stocks_daily_quotes.sql` plus `003_watchlists.sql`. Canonical upgrade plus the
temporary future migration produced:

```text
schema_migrations: [001_users_sessions_accounts.sql 002_stocks_daily_quotes_watchlists.sql 003_orders_positions.sql]
```

Neither obsolete ledger row remained, and the temporary Phase 3 probe table existed,
proving that the future filename executed. Canonical down after that upgraded history
left the probe and future ledger row intact while removing Phase 2:

```text
tables: [accounts migration_phase3_probe schema_migrations sessions users]
schema_migrations: [001_users_sessions_accounts.sql 003_orders_positions.sql]
```

Reapply restored all Phase 2 tables and the canonical row without re-running or skipping
the future probe:

```text
tables: [accounts migration_phase3_probe schema_migrations sessions stock_daily_quotes stocks users watchlists]
schema_migrations: [001_users_sessions_accounts.sql 002_stocks_daily_quotes_watchlists.sql 003_orders_positions.sql]
PASS; exit=0
```

The integration test cleanup then reported `0` tables in the disposable schema. The
schema existed (`1`) immediately before the explicit drop and did not exist (`0`) after
the drop. No disposable schema or test probe remains.

### Current residual risk

- The already-running shared application container was deliberately not restarted or
  mutated during this isolated test. Its old ledger will converge on the next deployment
  or migration-runner start; that shared upgrade remains a separate runtime observation.
- MySQL DDL auto-commits. The migration is rerunnable and has a tested rollback, but an
  operator must still use the down migration if a non-idempotent future statement is
  added after a partial DDL failure.
- No daily quote rows are seeded, so candle persistence remains structurally verified but
  not populated-data verified.
- The recorded database-expert paths remain unavailable; this rework relies on the
  accepted project contract and actual MySQL behavior.

## Historical pre-rework scope (superseded filenames)

The shared application database was inspected but never dropped or rolled back. An isolated schema named `ux_phase2_qa_20261024` was created inside the running MySQL service, granted to `ux_app`, exercised with the actual backend migration runner, then dropped. No production or user-service data was removed.

## Historical pre-rework apply verification (retained)

Backend runner command:

```text
docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml exec -T backend \
  sh -c 'DB_NAME=ux_phase2_qa_20261024 PORT=0 timeout 2 /app/ux-api'
2026/10/23 21:18:05 api listening port=0
runner_exit=0
```

Verification SQL:

```sql
SHOW TABLES;
SELECT version FROM schema_migrations ORDER BY version;
SELECT CONSTRAINT_NAME
FROM information_schema.TABLE_CONSTRAINTS
WHERE TABLE_SCHEMA = 'ux_phase2_qa_20261024'
  AND TABLE_NAME = 'watchlists'
ORDER BY CONSTRAINT_NAME;
```

Output:

```text
accounts
schema_migrations
sessions
stock_daily_quotes
stocks
users
watchlists
001_users_sessions_accounts.sql
002_stocks_daily_quotes.sql
003_watchlists.sql
fk_watchlist_stock
fk_watchlist_user
PRIMARY
uq_watchlist_user_stock
```

The watchlist migration has the expected user/stock foreign keys, primary key and `(user_id, stock_id)` uniqueness constraint. The stock migration has three seeded stocks and zero daily candle rows in the shared app DB; empty candles are therefore expected until daily fixtures/provider sync are added.

## Historical pre-rework rollback/reapply verification (retained)

The isolated schema ran down migrations in reverse dependency order:

```text
003_watchlists.down.sql
002_stocks_daily_quotes.down.sql
001_users_sessions_accounts.down.sql
```

Verification output:

```text
tables: schema_migrations
schema_migrations row count: 0
```

The backend runner was invoked again:

```text
2026/10/23 21:18:45 api listening port=0
runner_reapply_exit=0
tables: accounts, schema_migrations, sessions, stock_daily_quotes, stocks, users, watchlists
versions: 001_users_sessions_accounts.sql, 002_stocks_daily_quotes.sql, 003_watchlists.sql
```

This confirms all current down migrations remove their ledger rows and permit reapply. The temporary schema was dropped afterward.

## Historical pre-rework user-isolation/runtime impact (retained)

The API smoke created two synthetic users. User A's watchlist contained one `00700` row; user B's list remained empty. A duplicate add returned the same row ID and did not create another relation. User A's delete returned 204 and its list became empty. These checks exercised `watchlists(user_id, stock_id)` uniqueness and actor-scoped SQL predicates; no cross-user row was returned.

## Historical pre-rework residual risk (retained)

- No daily quote rows are seeded, so candle persistence is structurally verified but not populated-data verified.
- The project database-expert paths recorded by the prior handoff remain unavailable; this report relies on the project migration contract and actual MySQL execution.
