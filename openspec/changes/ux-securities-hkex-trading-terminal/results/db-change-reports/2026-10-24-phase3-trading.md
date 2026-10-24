# Phase 3 Paper Trading DB Change Report

Date: 2026-10-24 (Asia/Singapore)

## Migration

- Migration ID: `003_paper_trading_orders_positions.sql`
- Target DB: `ux_securities`
- Tables: `orders`, `positions`
- Contract: orders store fill fields directly; no `executions` or `position_lots` tables.

## Runtime evidence

The running migration ledger contains `003_paper_trading_orders_positions.sql`, and
`SHOW TABLES` confirms both tables. The verification output is in
`results/db-runtime-verification.txt`.

The localhost smoke flow registered a synthetic user, created a buy order for `00700`,
observed `filled`, reduced account cash from HKD 1,000,000 to HKD 961,980, and exposed
100 shares in `/api/v1/positions` and `/api/v1/orders`.

A two-user smoke is preserved in `results/api-two-user-isolation.txt`: user A's filled
order is returned to user A, user B receives `404 not_found` when requesting that order,
and user B's positions remain empty.

## Verification SQL

```sql
SHOW TABLES;
SELECT version, applied_at FROM schema_migrations ORDER BY version;
SELECT COUNT(*) FROM orders;
SELECT COUNT(*) FROM positions;
```

## Rollback

`003_paper_trading_orders_positions.down.sql` drops `positions` before `orders` and
removes its migration ledger row. Foreign keys and unique `(user_id, client_order_id)`
and `(user_id, symbol)` constraints are included in the up migration.

## Residual risk

The product remains paper-only. There is no real broker, exchange, or execution venue
connection. Concurrent transaction behavior is covered by backend service/repository
tests; a production outbox is intentionally not introduced, so WebSocket delivery is
best-effort after the database commit.
