# MVP 数据库设计：传统宽表

目标数据库：MySQL 8。设计优先考虑开发速度、phpMyAdmin 可读性和单体应用排错，不引入事件溯源、复杂审计、分库或微服务数据模型。

## 表清单

### `users`

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `username` | VARCHAR(64) | NOT NULL, UNIQUE |
| `email` | VARCHAR(128) | NOT NULL, UNIQUE |
| `password_hash` | VARCHAR(255) | NOT NULL |
| `status` | ENUM('active','disabled') | NOT NULL DEFAULT 'active' |
| `created_at` / `updated_at` | DATETIME(3) | NOT NULL |

### `sessions`

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `user_id` | BIGINT UNSIGNED | FK `users.id`, INDEX |
| `token_hash` | CHAR(64) | NOT NULL, UNIQUE；仅此处保存 SHA-256 token hash |
| `expires_at` | DATETIME(3) | NOT NULL |
| `revoked_at` | DATETIME(3) | NULL |
| `created_at` | DATETIME(3) | NOT NULL |

### `stocks`

一只股票一行，最新行情直接放在同一行，方便行情页查询。

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `symbol` | CHAR(6) | NOT NULL, UNIQUE，例如 `00700` |
| `name` | VARCHAR(128) | NOT NULL |
| `market` | VARCHAR(16) | NOT NULL DEFAULT 'HKEX' |
| `lot_size` | INT UNSIGNED | NOT NULL DEFAULT 100 |
| `status` | ENUM('active','suspended','delisted') | NOT NULL |
| `last_price` / `bid` / `ask` | DECIMAL(18,4) | NULL |
| `change` / `change_percent` | DECIMAL(18,4) | NULL |
| `volume` / `turnover` | DECIMAL(24,4) | NULL |
| `quote_as_of` | DATETIME(3) | NULL |
| `quote_source` | VARCHAR(32) | NULL |
| `data_status` | ENUM('simulated','unofficial','delayed','stale','unavailable','live') | NOT NULL |
| `updated_at` | DATETIME(3) | NOT NULL |

### `stock_daily_quotes`

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `stock_id` | BIGINT UNSIGNED | FK `stocks.id`, INDEX |
| `trade_date` | DATE | NOT NULL |
| `open` / `high` / `low` / `close` | DECIMAL(18,4) | NOT NULL |
| `volume` / `turnover` | DECIMAL(24,4) | NULL |
| 组合唯一键 | — | UNIQUE(`stock_id`, `trade_date`) |

### `watchlists`

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `user_id` / `stock_id` | BIGINT UNSIGNED | FK，组合 UNIQUE |
| `sort_order` | INT | NOT NULL DEFAULT 0 |
| `created_at` | DATETIME(3) | NOT NULL |

### `accounts`

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `user_id` | BIGINT UNSIGNED | FK `users.id`, UNIQUE |
| `initial_cash` / `cash_balance` / `available_cash` | DECIMAL(20,4) | NOT NULL |
| `currency` | CHAR(3) | NOT NULL DEFAULT 'HKD' |
| `status` | ENUM('active','locked') | NOT NULL DEFAULT 'active' |
| `updated_at` | DATETIME(3) | NOT NULL |

### `orders`

订单和模拟成交结果放在一行，避免 MVP 为一笔模拟成交再拆表。

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `user_id` / `account_id` / `stock_id` | BIGINT UNSIGNED | FK，按 user_id 建查询索引 |
| `side` | ENUM('buy','sell') | NOT NULL |
| `order_type` | ENUM('limit') | NOT NULL |
| `quantity` / `filled_quantity` | DECIMAL(20,4) | NOT NULL |
| `limit_price` / `filled_price` / `estimated_amount` | DECIMAL(20,4) | NULL |
| `status` | ENUM('pending','filled','cancelled','rejected') | NOT NULL |
| `idempotency_key` | VARCHAR(128) | NOT NULL；UNIQUE(`user_id`, `idempotency_key`) |
| `reject_reason` | VARCHAR(255) | NULL |
| `created_at` / `filled_at` / `cancelled_at` | DATETIME(3) | 允许后两项 NULL |

### `positions`

每个用户和股票只保留一行当前汇总持仓。

| 字段 | 类型 | 约束 |
|---|---|---|
| `id` | BIGINT UNSIGNED | PK, auto increment |
| `user_id` / `stock_id` | BIGINT UNSIGNED | FK，组合 UNIQUE |
| `quantity` / `available_quantity` | DECIMAL(20,4) | NOT NULL DEFAULT 0 |
| `average_cost` / `market_price` / `unrealized_pnl` / `market_value` | DECIMAL(20,4) | NOT NULL DEFAULT 0 |
| `updated_at` | DATETIME(3) | NOT NULL |

## 关系和事务

```text
users 1---N sessions
users 1---1 accounts
users 1---N watchlists N---1 stocks
users 1---N orders N---1 stocks
users 1---N positions N---1 stocks
stocks 1---N stock_daily_quotes
```

买入/卖出模拟成交使用一个 MySQL 事务：锁定 `accounts` 和用户对应 `positions`，检查余额或可用数量，更新 `orders` 的成交字段和状态，再更新账户与持仓。任一步失败全部回滚。行情同步只更新 `stocks`、`stock_daily_quotes`。

## Hash 边界

只对密码和 session token 做 hash。股票代码、订单号、幂等键、用户邮箱、行情来源和业务状态不 hash，保留可查询性。Token 原文只在客户端安全 cookie 中存在，数据库和日志只出现 `token_hash`。

## Migration 顺序

`001_users_sessions_accounts.sql`（`users`、`sessions`、`accounts`）→ `002_stocks_daily_quotes_watchlists.sql`（`stocks`、`stock_daily_quotes`、`watchlists`）→ `003_orders_positions.sql`（`orders`、`positions`）。`accounts` 在 Phase 1 创建，Phase 3 复用；不得重复创建或迁移其初始结构。每个迁移必须有对应 rollback、验证 SQL 和 `results/db-change-reports/` 报告。
