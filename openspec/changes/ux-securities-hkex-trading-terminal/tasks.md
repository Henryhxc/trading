## 0. 来源收敛与全局门禁

- [x] 0.1 阅读并记录 `AGENTS.md`、`agent.md`、`docs/PRD.md`、`docs/research/hkex-free-data.md` 与本 change 全部 spec；将事实源顺序写入 phase handoff。
- [x] 0.2 固化 Logo 到 `docs/assets/ux-securities-logo.png` 或前端资源目录，复核 `#1D41A9`、`#2857E2`、`#0B0B0D`、`#FFFFFF` token，并保存来源截图或色值报告。
- [x] 0.3 在 `docs/design/`、`docs/api/`、`docs/evidence/`、`apps/frontend`、`apps/backend`、`deploy`、`db/migrations` 建立项目档案；每个生产代码文件保持少于 500 行。
- [x] 0.4 在实现前确认注册身份、模拟初始资金/费用/交易单位/撮合延迟、标的范围和授权实时 provider；原仅日线范围已由第 5 节用户追加要求替代，历史来源按实际 provider 标记。未确认项必须保持配置化并记录风险，不能偷偷写死。
- [ ] 0.5 全局门禁：每个 phase 按 RED → 最小 GREEN → 用户指定的 in-app browser 视觉检查 → 独立 review → 主 agent 验收；任一证据缺失或截图出现遮挡/溢出/错误状态缺失即 `needs_rework`。

## 1. Phase 1：品牌壳、认证、数据库和 Docker 基础

### Frontend slice

- [x] 1.1 初始化 React 路由、会话状态、API client、错误边界和 UX Securities token；登录/注册页面加载真实 Logo 资源。
- [x] 1.2 实现登录、注册、退出和未登录路由保护的空态、加载态、校验错误和服务端错误映射。
- [x] 1.3 维护前端目录档案与组件职责，确保单文件不超过 500 行，并补充 `docs/design/develop-frontend-design.md`。

### Backend slice

- [x] 1.4 初始化 Go MVC 模块（handler/controller、service/usecase、domain、repository、runtime），实现统一错误响应与 request id。
- [x] 1.5 创建 users、sessions、accounts 的 MySQL migration、索引、唯一约束、seed/test fixture、验证 SQL 和 rollback；实施前重新检查数据库专家路径，并记录接受/拒绝的建议。
- [x] 1.6 实现注册、登录、退出、`/me`、session 过期和密码强哈希；所有受保护查询按 session actor 过滤。
- [x] 1.7 编写 `deploy/docker-compose.yml`，包含 frontend、backend、mysql 和本地 phpMyAdmin，默认端口 `3000/8080/3306/8081`、健康检查和依赖顺序。

### E2E / QA slice

- [x] 1.8 TDD/e2e agent 先写认证与用户隔离 RED 测试，运行并记录失败，再完成针对性 GREEN、API smoke 和 migration dry-run/apply/rollback。
- [x] 1.9 使用用户指定的 in-app browser 打开登录/注册 localhost，在 `1440x900`、`1280x800`、`390x844` 截图到 `results/screenshots/phase1-foundation/`，检查 Logo 比例、键盘焦点、错误文案和无横向溢出。
- [x] 1.10 查收 review agent 独立检查 diff、测试输出、Docker 冷启动、MySQL 健康和 phpMyAdmin 访问，并写 `agent-dispatch/reviews/phase1-review.md`。
- [x] 1.11 生成 `results/db-change-reports/<timestamp>-phase1-auth.md`，记录 migration、target DB、验证 SQL、rollback、保护范围、失败/返工与残余风险。
- [x] 1.12 完成打钩 gate：主 agent 复核截图、report、handoff、review 后，才更新本 phase tasks、phase packet、handoff recommendation 和 orchestration state 为 `passed/terminated`。

## 2. Phase 2：行情 provider、搜索、日线、收藏和工作台

### Frontend slice

- [x] 2.1 实现行情工作台布局：搜索/市场列表、报价卡、日线图、收藏入口、数据来源徽章与错误/空态；仅调用内部 `/api/v1/markets/*`。
- [x] 2.2 接入 UX Securities 品牌 token 与交易语义涨跌色，完成桌面高密度布局和窄屏单列/抽屉；保存 active symbol 与筛选后的 scroll anchor。
- [x] 2.3 实现收藏添加、删除、排序、重复幂等反馈和行情不可用状态；补充前端模型映射与 `docs/design/develop-frontend-design.md`。

### Backend slice

- [x] 2.4 定义 `MarketDataProvider`、标准 Quote/Candle/WebSocket DTO、显式 `mock/akshare` mode 和动态数据状态合同；默认 mock，AkShare 仅可标记 `unofficial/delayed/stale/unavailable`。
- [x] 2.5 实现简单宽表 `stocks`、`stock_daily_quotes`、`watchlists` migration、索引、保留策略和 rollback；以条件 upsert 保证旧报价不覆盖新报价、日线重复写入幂等。原八表限制由第 5 节修订，新增分钟历史表。
- [x] 2.6 实现搜索、报价、日线、收藏 API 与用户隔离；实现 HMAC webhook 验证、验收窗口内重放保护、单事务写库、提交后广播、认证 WebSocket、动态过期和 REST 断线恢复。
- [x] 2.7 将 AkShare 持续轮询、历史回补、签名 webhook 和 WebSocket 约定记录到 `docs/api/hkex-market-data.md`：明确 Python 调度器独立于 Go 业务层且无公网端口；规范 `00700` 五字符代码、DTO、HMAC headers、15 分钟延迟/全市场分页、来源/采集时间、校验、退避、幂等、提交后广播和非官方状态，并核对上游条款与字段变化。

### E2E / QA slice

- [x] 2.8 先写调度规范化、HMAC/时间戳/重放、事务 upsert、回滚不广播、提交后广播、动态过期、WebSocket session/origin、重连恢复、provider 状态、收藏幂等和用户隔离 RED；保留真实失败输出后完成 targeted GREEN 与 API contract smoke。
- [ ] 2.9 使用用户指定的 in-app browser 打开真实 localhost 工作台，在 `1440x900`、`1280x800`、`390x844` 覆盖搜索、报价、日线、收藏、空态、WebSocket 实时更新、断线后 REST 恢复、stale/unavailable，截图到 `results/screenshots/phase2-market-workbench/`。
- [ ] 2.10 查收 review agent 独立比较设计合同与实现截图，检查品牌色/Logo、AkShare 非官方来源、提交后推送、WebSocket 鉴权、遮挡、滚动锚点、错误提示和网络失败回退，写 review report。
- [ ] 2.11 生成行情/收藏 DB change report，包含验证 SQL、真实 AkShare 样本、事务/乱序/幂等、数据保留和 rollback；记录无 outbox 造成的广播窗口及免费源仅用于开发/模拟的残余风险。
- [ ] 2.12 完成打钩 gate：主 agent 复核 TDD/review 截图与报告后，更新 phase packet、tasks、handoff 和 orchestration state；否则返回责任 agent 返工。
- [x] 2.13 更新共享 OpenAPI 合同，使日线序列响应包含来源与数据状态，并记录内部 webhook 与 WebSocket 合同；确认 AkShare 数据标记 `akshare` 且绝不为 `live`，React 只访问内部 API。
- [x] 2.14 实现无公网端口的 AkShare Python 调度容器：配置标的/回补日期、默认且最短 15 分钟 spot 轮询、字段规范化、HMAC 签名、超时、带抖动退避、健康检查和脱敏日志；固定 `akshare==1.19.3` 并补齐单元测试与依赖锁定。
- [ ] 2.15 重建 Docker 并以真实 AkShare 拉取证明调度器 → 签名 webhook → MySQL 行情行 → REST → 已认证 WebSocket 全链路；证明无效签名、重放、事务回滚、调度停止变 stale、重启幂等且调度器无宿主机端口。

## 3. Phase 3：模拟订单、成交、账户、持仓和历史联动

### Frontend slice

- [x] 3.1 实现买入/卖出限价单表单、估算金额、行情时间、余额/持仓提示、确认、幂等键和服务端错误映射。
- [x] 3.2 实现账户余额、持仓、未实现盈亏、估值、成交和订单历史页面；覆盖 pending/filled/cancelled/rejected、stale/unavailable 和空态。
- [x] 3.3 完成下单成功后的账户、持仓、订单历史联动刷新及响应式回归；补充前端档案和 API 字段映射。

### Backend slice

- [x] 3.4 复用 Phase 1 已创建的 `accounts` 表，新增 `orders`、`positions` migration、唯一约束、索引、rollback 和 DB change report；成交字段直接放入 `orders`，不新增 executions/position_lots；记录数据库专家路径复查结果。
- [x] 3.5 实现订单服务端校验、幂等创建、取消、状态机、模拟撮合和过期行情拒绝；客户端不得提交成交、余额或持仓结果。
- [x] 3.6 在同一 MySQL 事务内更新 `orders` 成交字段、扣减/增加余额和更新 `positions`；不新增 executions 表，覆盖余额不足、卖出超持仓、并发和事务回滚。
- [x] 3.7 实现 `/orders`、`/orders/{id}/cancel`、`/positions`、`/account` 和历史筛选 API，严格按当前用户隔离并脱敏错误。

### E2E / QA slice

- [x] 3.8 TDD/e2e agent 先写订单校验、幂等、状态机、原子成交、并发和用户隔离 RED，记录 RED 后完成 targeted GREEN、API smoke 与事务断言。
- [ ] 3.9 使用用户指定的 in-app browser 完成登录 → 搜索 → 查看行情 → 收藏 → 下单 → 模拟成交 → 持仓/历史的完整 flow，在三种视口截图到 `results/screenshots/phase3-paper-trading/`。
- [ ] 3.10 查收 review agent 独立检查订单状态、资金/持仓一致性、重复提交、错误文案、截图视觉和 500 行限制，写 `agent-dispatch/reviews/phase3-review.md`。
- [x] 3.11 生成交易 DB change report，写明 migration id、事务验证 SQL、rollback、并发/失败/返工记录和残余风险；确认未连接真实交易撮合。
- [ ] 3.12 完成最终联动 gate：主 agent 复核全部 phase evidence、Docker 冷启动、OpenSpec strict validation、代码行数和报告后，才将所有完成项与 orchestration state 标记为 `passed/terminated`。

## 5. 用户追加：多周期开源 K 线（2026-10-24）

本节以 `docs/design/candlestick-history.md` 为合同，替代 0.4 中仅日线限制；不更改原有未通过门禁。

- [ ] 5.1 安装并标注 Lightweight Charts，显示 OHLC/成交量和分钟/日/周/月/年选择，默认半年窗口。
- [ ] 5.2 扩展 API/签名 ingestion 和分钟存储 migration/rollback，按香港日历聚合周/月/年，补齐 RED/GREEN 与临时 MySQL 验证。
- [ ] 5.3 实现半年日线回补和原生分钟轮询，各数据源独立失败、分块提交、幂等更新；以真实上游数据记录导入覆盖和限制。
- [ ] 5.4 验证图表切换、自动/手动更新、错误/空态与三视口截图，并完成独立审查；上游失败保持明确未通过。

## 4. 最终完成门禁

## 6. 用户追加：完整市场浏览与协调品牌工作台（2026-10-24）

以 `docs/design/workbench-layout.md` 为合同，不改变其他未通过门禁。

- [x] 6.1 默认展示港股目录，空查询 API、真实元数据导入及无报价状态通过 RED/GREEN 验证。
- [x] 6.2 市场/收藏/持仓导航、金色收藏和青色持仓标记；持仓/历史只在专页显示。
- [x] 6.3 完成左列表、中间 K 线、右下单三栏，Logo 蓝与协调深蓝灰背景，保留九周期及模拟交易。
- [x] 6.4 自动化测试、Docker 部署和 IAB 视觉/交互证据，记录实际视口及整体未通过门禁。证据：`results/workbench-layout/verification-2026-10-24.md`（1311×1076，不替代原三视口门禁）。

## 4. 最终完成门禁（续）

- [ ] 4.1 主 agent 扫描 `tasks.md`，确认全部要求已经是 `[x]`，没有遗留 `[ ]`、`needs_rework` 或 `blocked`；未满足时继续执行或明确记录 blocker。
- [ ] 4.2 主 agent 重新运行 strict validation、Docker/浏览器关键回归并更新最终 report；只有全部证据落盘后才把本项和整体 orchestration state 标记完成。
