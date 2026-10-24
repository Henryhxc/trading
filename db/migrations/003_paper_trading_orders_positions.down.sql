DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS orders;
DELETE FROM schema_migrations WHERE version = '003_paper_trading_orders_positions.sql';
