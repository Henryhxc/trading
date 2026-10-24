CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL, symbol CHAR(6) NOT NULL, side ENUM('buy','sell') NOT NULL,
  quantity BIGINT UNSIGNED NOT NULL, limit_price DECIMAL(18,4) NOT NULL, estimated_amount DECIMAL(20,4) NOT NULL,
  status ENUM('pending','filled','cancelled','rejected') NOT NULL, client_order_id VARCHAR(128) NOT NULL,
  filled_quantity BIGINT UNSIGNED NOT NULL DEFAULT 0, filled_price DECIMAL(18,4) NULL, filled_at DATETIME(3) NULL,
  reject_reason VARCHAR(255) NULL, quote_as_of DATETIME(3) NOT NULL, created_at DATETIME(3) NOT NULL, updated_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id), UNIQUE KEY uq_orders_user_client (user_id, client_order_id), KEY ix_orders_user_status_created (user_id, status, created_at), KEY ix_orders_user_symbol (user_id, symbol),
  CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE IF NOT EXISTS positions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, user_id BIGINT UNSIGNED NOT NULL, symbol CHAR(6) NOT NULL, quantity BIGINT UNSIGNED NOT NULL DEFAULT 0,
  available_quantity BIGINT UNSIGNED NOT NULL DEFAULT 0, average_cost DECIMAL(18,4) NOT NULL DEFAULT 0, updated_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id), UNIQUE KEY uq_positions_user_symbol (user_id, symbol), KEY ix_positions_user (user_id), CONSTRAINT fk_positions_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
