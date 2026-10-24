-- Refuses rollback while missing turnover remains; do not fabricate zero amounts.
ALTER TABLE stock_daily_quotes MODIFY COLUMN turnover DECIMAL(24,4) NOT NULL DEFAULT 0;
ALTER TABLE stock_intraday_quotes MODIFY COLUMN turnover DECIMAL(24,4) NOT NULL DEFAULT 0;
