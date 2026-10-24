# Phase 2 Migration Contract Rework — RED Evidence

Date: 2026-10-24 (Asia/Hong_Kong)

## Contract break guarded

Production change required for GREEN: replace the split Phase 2 migrations
`002_stocks_daily_quotes.sql` plus `003_watchlists.sql` with the accepted
`002_stocks_daily_quotes_watchlists.sql` up/down pair, and make canonical apply/rollback
reconcile the two legacy ledger names. The `003` slot must remain absent until it is
owned by `003_orders_positions.sql`.

The behavior integration test was added at
`apps/backend/internal/migrations/phase2_contract_integration_test.go`; it exercises a
real disposable MySQL schema when `PHASE2_MIGRATION_TEST_DSN` is supplied. The file
contract guard was run first so that the pre-fix state could be captured without
changing any production migration.

## RED command and result

```text
$ python3 tests/phase2_migration_contract_guard_test.py
FFFF

FAIL: test_canonical_down_reverses_dependencies_and_clears_old_and_new_ledger_rows
AssertionError: required canonical migration is missing: 002_stocks_daily_quotes_watchlists.down.sql

FAIL: test_canonical_up_owns_all_phase2_tables_and_reconciles_legacy_ledger
AssertionError: required canonical migration is missing: 002_stocks_daily_quotes_watchlists.sql

FAIL: test_only_orders_positions_may_own_the_003_slot
AssertionError: ['003_watchlists.sql'] not found in ([], ['003_orders_positions.sql'])

FAIL: test_phase2_is_one_canonical_002_migration
AssertionError: False is not true : Phase 2 must use 002_stocks_daily_quotes_watchlists.sql

Ran 4 tests in 0.001s
FAILED (failures=4)
exit=1
```

This is the expected RED: it fails on the current split files and the Phase 2 use of
the reserved `003` namespace, rather than on a test syntax/import error.
