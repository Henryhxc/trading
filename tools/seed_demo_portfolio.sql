-- Development-only demo portfolio seed.
-- It targets the existing demo email and is safe to run repeatedly.
-- No credentials are stored here.

SET @demo_user_id := (SELECT id FROM users WHERE email = '021021hxc@gmail.com' LIMIT 1);

START TRANSACTION;

INSERT INTO orders (
  user_id, symbol, side, quantity, limit_price, estimated_amount, status,
  client_order_id, filled_quantity, filled_price, filled_at, quote_as_of,
  created_at, updated_at
)
SELECT @demo_user_id, '00700', 'buy', 300, 350.0000, 105000.0000, 'filled',
  'demo-portfolio-00700-buy-20260603', 300, 350.0000, '2026-06-03 08:00:00.000', '2026-06-03 08:00:00.000',
  '2026-06-03 08:00:00.000', '2026-06-03 08:00:00.000'
WHERE @demo_user_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-00700-buy-20260603');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '00700', 'sell', 50, 355.0000, 17750.0000, 'filled', 'demo-portfolio-00700-sell-20260708', 50, 355.0000, '2026-07-08 08:00:00.000', '2026-07-08 08:00:00.000', '2026-07-08 08:00:00.000', '2026-07-08 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-00700-sell-20260708');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '00700', 'sell', 50, 365.0000, 18250.0000, 'filled', 'demo-portfolio-00700-sell-20260912', 50, 365.0000, '2026-09-12 08:00:00.000', '2026-09-12 08:00:00.000', '2026-09-12 08:00:00.000', '2026-09-12 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-00700-sell-20260912');

INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '09988', 'buy', 400, 90.0000, 36000.0000, 'filled', 'demo-portfolio-09988-buy-20260610', 400, 90.0000, '2026-06-10 08:00:00.000', '2026-06-10 08:00:00.000', '2026-06-10 08:00:00.000', '2026-06-10 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-09988-buy-20260610');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '09988', 'sell', 50, 92.0000, 4600.0000, 'filled', 'demo-portfolio-09988-sell-20260722', 50, 92.0000, '2026-07-22 08:00:00.000', '2026-07-22 08:00:00.000', '2026-07-22 08:00:00.000', '2026-07-22 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-09988-sell-20260722');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '09988', 'sell', 50, 97.0000, 4850.0000, 'filled', 'demo-portfolio-09988-sell-20260918', 50, 97.0000, '2026-09-18 08:00:00.000', '2026-09-18 08:00:00.000', '2026-09-18 08:00:00.000', '2026-09-18 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-09988-sell-20260918');

INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '01810', 'buy', 600, 23.0000, 13800.0000, 'filled', 'demo-portfolio-01810-buy-20260617', 600, 23.0000, '2026-06-17 08:00:00.000', '2026-06-17 08:00:00.000', '2026-06-17 08:00:00.000', '2026-06-17 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-01810-buy-20260617');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '01810', 'sell', 50, 24.0000, 1200.0000, 'filled', 'demo-portfolio-01810-sell-20260804', 50, 24.0000, '2026-08-04 08:00:00.000', '2026-08-04 08:00:00.000', '2026-08-04 08:00:00.000', '2026-08-04 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-01810-sell-20260804');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '01810', 'sell', 50, 26.0000, 1300.0000, 'filled', 'demo-portfolio-01810-sell-20261003', 50, 26.0000, '2026-10-03 08:00:00.000', '2026-10-03 08:00:00.000', '2026-10-03 08:00:00.000', '2026-10-03 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-01810-sell-20261003');

INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '00005', 'buy', 1200, 60.0000, 72000.0000, 'filled', 'demo-portfolio-00005-buy-20260624', 1200, 60.0000, '2026-06-24 08:00:00.000', '2026-06-24 08:00:00.000', '2026-06-24 08:00:00.000', '2026-06-24 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-00005-buy-20260624');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '00005', 'sell', 100, 61.0000, 6100.0000, 'filled', 'demo-portfolio-00005-sell-20260819', 100, 61.0000, '2026-08-19 08:00:00.000', '2026-08-19 08:00:00.000', '2026-08-19 08:00:00.000', '2026-08-19 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-00005-sell-20260819');
INSERT INTO orders (user_id, symbol, side, quantity, limit_price, estimated_amount, status, client_order_id, filled_quantity, filled_price, filled_at, quote_as_of, created_at, updated_at)
SELECT @demo_user_id, '00005', 'sell', 100, 63.0000, 6300.0000, 'filled', 'demo-portfolio-00005-sell-20261010', 100, 63.0000, '2026-10-10 08:00:00.000', '2026-10-10 08:00:00.000', '2026-10-10 08:00:00.000', '2026-10-10 08:00:00.000'
WHERE @demo_user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM orders WHERE user_id = @demo_user_id AND client_order_id = 'demo-portfolio-00005-sell-20261010');

DELETE FROM positions WHERE user_id = @demo_user_id AND symbol IN ('00700', '09988', '01810', '00005');
INSERT INTO positions (user_id, symbol, quantity, available_quantity, average_cost, updated_at)
SELECT @demo_user_id, '00700', 200, 200, 350.0000, UTC_TIMESTAMP(3) WHERE @demo_user_id IS NOT NULL;
INSERT INTO positions (user_id, symbol, quantity, available_quantity, average_cost, updated_at)
SELECT @demo_user_id, '09988', 300, 300, 90.0000, UTC_TIMESTAMP(3) WHERE @demo_user_id IS NOT NULL;
INSERT INTO positions (user_id, symbol, quantity, available_quantity, average_cost, updated_at)
SELECT @demo_user_id, '01810', 500, 500, 23.0000, UTC_TIMESTAMP(3) WHERE @demo_user_id IS NOT NULL;
INSERT INTO positions (user_id, symbol, quantity, available_quantity, average_cost, updated_at)
SELECT @demo_user_id, '00005', 1000, 1000, 60.0000, UTC_TIMESTAMP(3) WHERE @demo_user_id IS NOT NULL;

UPDATE accounts
SET cash_balance = initial_cash - 226800.0000 + 42100.0000,
    available_cash = initial_cash - 226800.0000 + 42100.0000,
    updated_at = UTC_TIMESTAMP(3)
WHERE user_id = @demo_user_id;

COMMIT;
