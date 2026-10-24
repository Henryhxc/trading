DROP INDEX ix_daily_source_collected ON stock_daily_quotes;

ALTER TABLE stock_daily_quotes
  DROP COLUMN adjustment,
  DROP COLUMN collected_at,
  DROP COLUMN data_status,
  DROP COLUMN source;

DELETE FROM schema_migrations WHERE version = '002_webhook_market_provenance.sql';
