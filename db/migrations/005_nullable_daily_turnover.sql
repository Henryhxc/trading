-- Tencent daily history has no turnover amount. Preserve missing values as NULL.
-- Repeating MODIFY is safe after DDL commit / migration-ledger interruption.
ALTER TABLE stock_daily_quotes MODIFY COLUMN turnover DECIMAL(24,4) NULL DEFAULT NULL;
ALTER TABLE stock_intraday_quotes MODIFY COLUMN turnover DECIMAL(24,4) NULL DEFAULT NULL;
