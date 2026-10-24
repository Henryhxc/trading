CREATE TABLE IF NOT EXISTS stocks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  symbol CHAR(6) NOT NULL,
  name VARCHAR(128) NOT NULL,
  market VARCHAR(16) NOT NULL DEFAULT 'HKEX',
  lot_size INT UNSIGNED NOT NULL DEFAULT 100,
  status ENUM('active','suspended','delisted') NOT NULL DEFAULT 'active',
  last_price DECIMAL(18,4) NULL,
  bid DECIMAL(18,4) NULL,
  ask DECIMAL(18,4) NULL,
  `change` DECIMAL(18,4) NULL,
  change_percent DECIMAL(18,4) NULL,
  volume DECIMAL(24,4) NULL,
  turnover DECIMAL(24,4) NULL,
  quote_as_of DATETIME(3) NULL,
  quote_source VARCHAR(32) NULL,
  data_status ENUM('simulated','unofficial','delayed','stale','unavailable','live') NOT NULL DEFAULT 'simulated',
  updated_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id), UNIQUE KEY uq_stocks_symbol (symbol), KEY ix_stocks_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS stock_daily_quotes (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  stock_id BIGINT UNSIGNED NOT NULL,
  trade_date DATE NOT NULL,
  `open` DECIMAL(18,4) NOT NULL,
  high DECIMAL(18,4) NOT NULL,
  low DECIMAL(18,4) NOT NULL,
  close DECIMAL(18,4) NOT NULL,
  volume DECIMAL(24,4) NULL,
  turnover DECIMAL(24,4) NULL,
  PRIMARY KEY (id), UNIQUE KEY uq_daily_stock_date (stock_id, trade_date), KEY ix_daily_stock (stock_id),
  CONSTRAINT fk_daily_stock FOREIGN KEY (stock_id) REFERENCES stocks (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

INSERT INTO stocks (symbol, name, status, last_price, bid, ask, `change`, change_percent, volume, turnover, quote_as_of, quote_source, data_status, updated_at)
VALUES
 ('00700', 'Tencent Holdings', 'active', 380.2000, 380.0000, 380.4000, 2.1000, 0.5543, 1200000, 456240000.0000, UTC_TIMESTAMP(3), 'mock', 'simulated', UTC_TIMESTAMP(3)),
 ('00005', 'HSBC Holdings', 'active', 63.4500, 63.2500, 63.6500, -0.3500, -0.5495, 1200000, 76140000.0000, UTC_TIMESTAMP(3), 'mock', 'simulated', UTC_TIMESTAMP(3)),
 ('09988', 'Alibaba Group', 'active', 88.1000, 87.9000, 88.3000, 1.2000, 1.3818, 1200000, 105720000.0000, UTC_TIMESTAMP(3), 'mock', 'simulated', UTC_TIMESTAMP(3))
ON DUPLICATE KEY UPDATE name = VALUES(name), updated_at = UTC_TIMESTAMP(3);

CREATE TABLE IF NOT EXISTS watchlists (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  stock_id BIGINT UNSIGNED NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id), UNIQUE KEY uq_watchlist_user_stock (user_id, stock_id), KEY ix_watchlist_user_order (user_id, sort_order),
  CONSTRAINT fk_watchlist_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_watchlist_stock FOREIGN KEY (stock_id) REFERENCES stocks (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

DELETE FROM schema_migrations
WHERE version IN ('002_stocks_daily_quotes.sql', '003_watchlists.sql');
