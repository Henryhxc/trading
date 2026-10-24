CREATE TABLE IF NOT EXISTS cash_movements (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  movement_type ENUM('deposit','withdraw') NOT NULL,
  amount DECIMAL(20,4) NOT NULL,
  currency CHAR(3) NOT NULL DEFAULT 'HKD',
  method VARCHAR(64) NOT NULL,
  status ENUM('pending','completed') NOT NULL DEFAULT 'completed',
  prototype TINYINT(1) NOT NULL DEFAULT 1,
  created_at DATETIME(3) NOT NULL,
  PRIMARY KEY (id),
  KEY ix_cash_movements_user_created (user_id, created_at),
  CONSTRAINT fk_cash_movements_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
