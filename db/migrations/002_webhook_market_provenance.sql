ALTER TABLE stock_daily_quotes
  ADD COLUMN source VARCHAR(32) NOT NULL DEFAULT 'mock' AFTER turnover,
  ADD COLUMN data_status ENUM('simulated','unofficial','delayed','stale','unavailable','live') NOT NULL DEFAULT 'simulated' AFTER source,
  ADD COLUMN collected_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) AFTER data_status,
  ADD COLUMN adjustment ENUM('none') NOT NULL DEFAULT 'none' AFTER collected_at;

CREATE INDEX ix_daily_source_collected ON stock_daily_quotes (source, collected_at);
