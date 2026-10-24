DROP TABLE IF EXISTS watchlists;
DROP TABLE IF EXISTS stock_daily_quotes;
DROP TABLE IF EXISTS stocks;
DELETE FROM schema_migrations
WHERE version IN (
  '002_stocks_daily_quotes_watchlists.sql',
  '002_stocks_daily_quotes.sql',
  '003_watchlists.sql'
);
