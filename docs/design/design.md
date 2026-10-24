# UX Securities 港股交易终端设计

## 业务设计基线

### 事实源与优先级

本 change 的事实源优先级为：当前用户请求 → 本 change 的 proposal/spec → 项目 [AGENTS.md](../../../../AGENTS.md) 与 [agent.md](../../../../agent.md) → [docs/PRD.md](../../../../docs/PRD.md) → [docs/research/hkex-free-data.md](../../../../docs/research/hkex-free-data.md) → 用户提供的 Logo 和 Futubull 截图 → HKEX 官方资料。外部项目只能用于校准，不能覆盖本项目业务合同。

用户提供的 Logo 作为唯一品牌素材，实施阶段需要固化到 `docs/assets/ux-securities-logo.png` 或前端资源目录，并记录实际路径。交易终端参考截图只用于信息密度和工作台分区，不复制第三方品牌、文案、标的或数值。当前仓库没有可直接启动的原型应用。

### 用户与核心流程

普通用户注册或登录后进入行情工作台，搜索港股、查看价格和日线、加入收藏、打开模拟订单、确认服务端校验结果，并在持仓和订单历史中追踪结果。系统服务账号负责行情同步，不具备用户登录能力。

第一版默认模拟交易。HKEX 官方实时数据通过 OMD-C/授权信息供应商接入属于后续生产能力；免费公开行情只用于开发或模拟场景，界面必须显示来源和数据状态。

### 品牌主题合同

Logo 样本的主视觉是蓝色图形、黑色文字和白色背景。由栅格样本提取的初始 token（实现前允许以源 Logo 色值复核）为：

| Token | 初始值 | 用途 |
|---|---|---|
| `--ux-primary` | `#1D41A9` | 主按钮、导航选中、品牌链接、焦点环 |
| `--ux-accent` | `#2857E2` | hover、强调操作、图表选中线 |
| `--ux-ink` | `#0B0B0D` | 标题、正文和 Logo 文字近似色 |
| `--ux-surface` | `#FFFFFF` | 页面和卡片背景 |
| `--ux-muted` | `#6B7280` | 次要文字、边框和禁用说明 |

涨跌颜色属于交易语义 token，不能替代品牌色：上涨使用可读绿色、下跌使用可读红色，并同时呈现数值和箭头/文字，避免只依靠颜色传达状态。Logo 不得被拉伸、裁切或改色；暗色工作台可使用深色表面，但 Logo 保持原始可读性。

### 业务边界

范围包括注册/登录、港股搜索和行情、收藏、模拟限价买卖、模拟撮合、余额、持仓和订单历史。真实开户、入金、出金、真实下单、杠杆、融资融券、期权行权、投顾和自动策略不在本 change 范围。

## 边界 / 用户流程 / 角色体验

### 页面合同

- 登录/注册：UX Securities Logo、账号字段、密码字段、错误提示和会话跳转。
- 行情工作台：左侧搜索与收藏，中间报价/日线，右侧模拟订单入口；桌面高密度，窄屏使用单列或抽屉。
- 收藏：代码、名称、最新价、涨跌幅、数据时间、数据状态、排序和删除。
- 模拟订单：买卖切换、数量、限价、估算金额、行情时间戳、余额和失败原因。
- 持仓：数量、可用数量、平均成本、现价、未实现盈亏和估值。
- 订单历史：标的、方向、数量、价格、状态、成交时间、拒绝原因和筛选。

用户只看到自己的账户、收藏、订单、成交和持仓。前端不得决定成交、余额或持仓；任何状态以服务端响应为准。

### MVP 实施默认与残余产品风险

下列值是本 change 为完成可运行 MVP 采用的实施默认，不冒充商业、合规或行情许可确认；可配置项必须由服务端读取，前端不得写死或覆盖：

1. 注册同时收集唯一用户名和唯一邮箱，登录接受用户名或邮箱；MVP 不发送邮箱/手机验证码。未来启用验证属于另一个身份变更。
2. 模拟账户初始资金默认 `HKD 1,000,000.0000`，由 `INITIAL_CASH_HKD` 覆盖。手续费默认 `0` basis points，并由 Phase 3 的服务端 paper-trading 配置覆盖；界面必须明确是估算值。
3. 每只证券的交易单位读取 `stocks.lot_size`，不由 React 写死；当前 mock fixture 使用 100 股仅作为测试数据。模拟撮合默认服务端即时处理，Phase 3 必须提供可覆盖的撮合延迟配置，且客户端不能决定成交时间或状态。
4. MVP 标的范围限于 mock fixture 提供的 HKEX 普通股票；ETF、REIT、窝轮、牛熊证和衍生品未建模，不得仅靠文案宣称支持。图表只保证日线 OHLCV，不实现分钟线。
5. 未取得授权前，行情 mode 必须显式默认为 mock；数据库缓存或免费公开源只能作为非授权来源并降级状态。未来直接 OMD-C 或授权供应商仍需合同、adapter、许可审查和显式配置后才能返回 `live`。

费用模型、撮合延迟和生产行情供应商仍是上线前产品/合规风险；本 MVP 的默认值只服务模拟交易和自动化验收。

## 技术落地补充

以下技术内容只实现上面的业务范围，不扩大业务权限或替换业务事实。

### 技术目标 / 非目标

目标是一个 Go 模块化单体后端、React 前端、MySQL 持久化和 Docker Compose 本地运行基线。Go 采用 `handler/controller → service/usecase → domain/model → repository → provider/adapter`；React 按 route/page、controller hook、model mapping、feature component 和 API client 拆分。每个生产代码文件小于 500 行。

非目标是微服务拆分、真实撮合、真实券商连接、网页抓取伪装成官方 feed、在浏览器保存 provider 凭据，以及将 Node/Python 免费行情库直接嵌入 Go 业务层。

### 架构合同

```text
Browser
  └─ frontend:3000 (React + UX Securities shell)
       └─ backend:8080 (Go MVC modular monolith)
            ├─ MySQL:3306
            └─ MarketDataProvider (mock / public adapter / future HKEX)
phpMyAdmin:8081 (local profile only)
```

Go 后端模块职责：

- `auth`：注册、登录、session、密码哈希和当前用户。
- `market`：标的搜索、报价、日线、provider fallback、数据状态。
- `watchlist`：用户收藏关系和排序。
- `papertrading`：订单校验、幂等、撮合、成交、账户、持仓。
- `http`：路由、认证中间件、错误映射和 request id。
- `runtime`：配置、健康检查、日志、迁移入口和 provider 装配。

React 模块职责：

- `app`：路由、session provider、主题 token 和全局错误边界。
- `pages`：登录、工作台、收藏、订单、持仓和历史页面。
- `features/market`：搜索、报价卡、K 线、状态徽章。
- `features/trading`：订单表单、确认、校验错误和结果刷新。
- `features/account`：余额、持仓、订单历史。
- `shared/api` 与 `shared/model`：API client、DTO mapping、缓存和状态类型。

### 行情 provider 合同

内部证券代码统一为六位数字，例如 `00700`。adapter 负责将其映射为腾讯 `hk00700`、`stock-api` `HK00700` 或 `easyquotation` `00700`。provider 输出：

```text
symbol, display_name, last_price, bid, ask, change,
change_percent, volume, turnover, as_of, source, data_status
```

默认顺序为配置的授权 provider → 免费公开 provider → mock。免费公开源只能返回 `unofficial`、`delayed` 或 `stale`；mock 返回 `simulated`；只有未来完成许可的 OMD-C provider 才能返回 `live`。超时、限流、字段缺失和时间戳过期必须留下脱敏日志和 request id。

### API Contract

内部 API 前缀为 `/api/v1`：

| 方法 | 路径 | 目的 |
|---|---|---|
| POST | `/auth/register` | 注册并创建模拟账户 |
| POST | `/auth/login` | 创建 session |
| POST | `/auth/logout` | 撤销 session |
| GET | `/me` | 当前用户 |
| GET | `/markets/symbols?query=` | 港股搜索 |
| GET | `/markets/quotes/{symbol}` | 标准化报价 |
| GET | `/markets/candles/{symbol}` | 日线 OHLCV |
| GET/POST/DELETE | `/watchlist`、`/watchlist/{symbol}` | 收藏 |
| GET/POST | `/orders` | 订单历史与创建模拟订单 |
| POST | `/orders/{id}/cancel` | 取消待处理订单 |
| GET | `/positions` | 持仓 |
| GET | `/account` | 余额和账户状态 |

错误响应统一为 `{ code, message, details?, request_id }`；不得返回 SQL、堆栈、密码、完整 token 或 provider 原始载荷。订单请求必须带幂等键；金额、成交状态、余额、持仓和用户 id 都是服务端字段，客户端不能提交覆盖。

### 数据模型和 migration（MVP 简单宽表）

按用户要求，MVP 使用传统宽表，先保证容易开发、容易用 phpMyAdmin 查看、容易排错。只保留 8 张核心表，不拆 `executions`、`position_lots`、outbox 或审计明细表：

字段类型、约束、关系和 migration 文件名的可执行草案见本 change 的 [`db-schema.md`](db-schema.md)。

| 表 | 用途 | 主要字段 |
|---|---|---|
| `users` | 登录用户 | `id`、`username`、`email`、`password_hash`、`status`、`created_at`、`updated_at` |
| `sessions` | 登录会话 | `id`、`user_id`、`token_hash`、`expires_at`、`revoked_at`、`created_at` |
| `stocks` | 港股标的和最新行情宽表 | `id`、`symbol`、`name`、`market`、`lot_size`、`status`、`last_price`、`bid`、`ask`、`change`、`change_percent`、`volume`、`turnover`、`quote_as_of`、`quote_source`、`data_status` |
| `stock_daily_quotes` | 日线 OHLCV | `id`、`stock_id`、`trade_date`、`open`、`high`、`low`、`close`、`volume`、`turnover` |
| `watchlists` | 用户收藏 | `id`、`user_id`、`stock_id`、`sort_order`、`created_at` |
| `accounts` | 用户模拟账户 | `id`、`user_id`、`cash_balance`、`available_cash`、`initial_cash`、`currency`、`status`、`updated_at` |
| `orders` | 订单、成交和历史宽表 | `id`、`user_id`、`account_id`、`stock_id`、`side`、`order_type`、`quantity`、`limit_price`、`filled_quantity`、`filled_price`、`estimated_amount`、`status`、`idempotency_key`、`reject_reason`、`created_at`、`filled_at`、`cancelled_at` |
| `positions` | 用户当前持仓宽表 | `id`、`user_id`、`stock_id`、`quantity`、`available_quantity`、`average_cost`、`market_price`、`unrealized_pnl`、`market_value`、`updated_at` |

简化原则：

- 只做必要 hash：`users.password_hash` 用 Argon2id 或 bcrypt；`sessions.token_hash` 用 SHA-256。用户名、邮箱、股票代码、订单号、幂等键和业务字段保持可查询明文，不额外 hash。
- `orders` 直接保存模拟成交字段；MVP 不单独建立 `executions` 表。订单状态 `pending/filled/cancelled/rejected` 和成交字段组成订单历史。
- `stocks` 保存最新报价，`stock_daily_quotes` 保存日线历史；不拆 `market_quotes` 和 `market_candles` 两套过度抽象的表。
- `positions` 只保留当前汇总持仓，不做 lot 级成本核算。成交后在同一事务更新 `orders`、`accounts`、`positions`。
- 唯一约束：`users.username`、`users.email`、`stocks.symbol`、`watchlists(user_id, stock_id)`、`accounts.user_id`、`orders(user_id, idempotency_key)`。
- 必要索引：`sessions(token_hash)`、`stock_daily_quotes(stock_id, trade_date)`、`orders(user_id, created_at)`、`orders(user_id, status)`、`positions(user_id, stock_id)`。

迁移顺序为 `001_users_sessions_accounts.sql`（`users`、`sessions`、`accounts`）→ `002_stocks_daily_quotes_watchlists.sql`（`stocks`、`stock_daily_quotes`、`watchlists`）→ `003_orders_positions.sql`（`orders`、`positions`）。`accounts` 在 Phase 1 创建，Phase 3 复用，不重复迁移其初始结构。每个迁移必须有 dry-run、apply、验证 SQL 和 rollback；测试 fixture 只能通过 seed migration 或测试事务写入。DB 专家路径 `/Users/henry/.codex/plugins/database-design/` 和 `/Users/henry/.codex/plugins/database-migrations/` 在本环境未找到，实施前需重新检查；当前设计以本项目 AGENTS、PRD 和 MySQL 约束为事实源，并在 `results/db-change-reports/` 记录缺失专家输入和残余风险。

### 后端库规划专家输入

按 `henry-propose` 要求，计划读取以下数据库专家资料：`/Users/henry/.codex/plugins/database-design/agents/database-architect.md`、`sql-pro.md`、`/Users/henry/.codex/plugins/database-migrations/agents/database-admin.md`、`database-optimizer.md` 和 `commands/sql-migrations.md`。这些路径在当前环境均未找到，因此本 change 暂不声称已获得专家结论；实施开始前必须重新检查。专家输入可用于校准表边界、3NF、索引、迁移顺序、回滚、验证 SQL 和性能风险，但不能覆盖本项目用户隔离、模拟交易、迁移专用写入和免费行情边界。每个 phase 的 handoff 与 `results/db-change-reports/` 必须记录实际读取到的建议、采纳/拒绝理由和剩余风险。

### Security / Privacy / Threat Model

- 仅密码和 session token 使用 hash；密码使用强哈希，session token 只存 SHA-256 hash 或安全 cookie，不写日志。普通业务字段不做 hash，方便 MVP 查询和 phpMyAdmin 排错。
- 所有收藏、订单、持仓和账户查询按 session actor 过滤。
- 登录、注册、下单、取消订单和行情请求需限流，并带脱敏 request/order id。
- Provider key 只来自环境变量；前端 bundle 和截图不得包含密钥。
- 订单金额、价格、数量和状态由服务端校验；模拟撮合不接受客户端成交结果。
- 日志不得记录密码、完整 token、数据库密码或原始供应商报文。

### 事务 / 幂等 / Worker / Runtime

创建订单在幂等键唯一约束下写入；撮合成交在一个 MySQL 事务中锁定账户/持仓行，直接更新 `orders.filled_quantity/filled_price/status`、扣减或增加 `accounts` 余额，并更新 `positions`。行情同步只更新 `stocks` 和 `stock_daily_quotes`，不得写用户资金表。Docker healthcheck 等待 MySQL 后再启动后端迁移；phpMyAdmin 默认只在本地 profile 暴露。

### 前端和视觉验证合同

实现必须直接使用 Logo 资源和本设计的 token。页面至少在 `1440x900`、`1280x800` 和 `390x844` 检查登录、工作台、收藏、订单、持仓和历史。按用户最新直接指令，本项目本次执行以 in-app browser 作为 UI 验收主路径；不得将 IAB 证据标记为 Chrome/Computer Use。视觉 PASS 要求无横向溢出、遮挡、文字截断、重复入口、错误状态缺失和品牌 token 漂移；Logo 比例、蓝色主色、焦点状态和深浅表面必须可见且一致。

### 技术实施顺序和证据合同

1. Phase 1：品牌壳、认证、数据库迁移和 API 错误合同，证明 Docker 冷启动与用户隔离。
2. Phase 2：行情 provider、搜索、报价、日线、收藏和工作台视觉，证明来源状态、降级和响应式布局。
3. Phase 3：模拟订单、成交、余额、持仓、订单历史和最终联动，证明事务、幂等、截图和完整回归。

每个 phase 必须先 RED、再最小 GREEN、再定向测试、用户指定的 IAB 视觉检查、独立 review 和主 agent 验收。截图、报告、handoff 和 review 缺任一项，phase 必须保持 `needs_rework` 或 `blocked`。

### 验证

规划阶段执行 `openspec validate ux-securities-hkex-trading-terminal --strict` 和 `openspec status --change ux-securities-hkex-trading-terminal`。实施阶段还需执行 Go/React 定向测试、API smoke、MySQL migration dry-run/apply/rollback、Docker 冷启动、phpMyAdmin 表检查、IAB 截图和 `git diff --check`（仅在仓库初始化 Git 后）。
