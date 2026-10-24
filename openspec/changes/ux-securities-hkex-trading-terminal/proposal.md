## Why

当前项目只有产品级 PRD 和智能体规则，尚未形成一份可以直接交给实现团队的 OpenSpec change。需要把 UX Securities 的品牌 Logo、公司主题色、React + Go + MySQL + Docker 单体架构，以及港股行情、账户、收藏、模拟订单、持仓和订单历史收敛为可验收的实施合同。

## What Changes

- 建立 UX Securities 品牌化港股交易终端的业务、视觉和技术设计合同。
- 使用用户提供的 UX Securities Logo；从 Logo 提取深蓝主色、亮蓝强调色、黑色文字和白色背景，形成前端 token 和视觉验收基线。
- 规划 React 前端、Go MVC 后端、MySQL、phpMyAdmin 和 Docker Compose 的单体部署边界，保持前端 `3000`、后端 `8080`、MySQL `3306`、phpMyAdmin `8081` 端口约定。
- 规划登录、港股搜索与行情、收藏、模拟下单、模拟撮合、持仓、订单历史和账户余额的完整用户流程。
- 规划 `MockMarketDataProvider`、AkShare 持续轮询调度器和未来授权 HKEX OMD-C 适配器的边界；AkShare 通过签名内部 webhook 写入行情缓存，并在数据库提交后通过认证 WebSocket 推送给前端，免费或非官方来源不能冒充 HKEX 官方实时行情。
- 生成按 phase 调度的开发、TDD/e2e、查收 review 任务包，并建立截图、API、数据库、Docker 和返工证据门禁。

## Capabilities

### New Capabilities

- `identity/user-auth`: 用户注册、登录、退出、会话保护和账户隔离。
- `market-data/hkex-quotes`: 港股标的搜索、报价、日线数据、行情来源和数据状态。
- `watchlist`: 用户收藏的新增、删除、幂等和排序。
- `paper-trading/orders`: 模拟限价订单、校验、幂等、状态转换和取消。
- `paper-trading/portfolio`: 模拟成交、余额、持仓和订单历史的一致性查询。
- `brand-shell/deployment`: UX Securities 品牌化 React 工作台和 Docker 单体运行入口。

### Modified Capabilities

无。当前仓库没有已存在的 OpenSpec capability 规范；本 change 创建上述初始行为合同。

## Impact

- 新增 `openspec/changes/ux-securities-hkex-trading-terminal/` 下的业务提案、设计、规格、任务和 phase 调度文档。
- 未来实现将影响 `apps/frontend`、`apps/backend`、`db/migrations`、`deploy/docker-compose.yml`、共享 API 合同和 `docs/design` 档案。
- Docker Compose 增加无公网端口的 AkShare Python 调度服务；沿用后端 `8080` 提供内部 webhook 与客户端 WebSocket，不增加业务端口或新数据表。
- 需要使用 MySQL 迁移、会话/密码哈希、模拟撮合事务、行情 provider adapter 和 Docker 健康检查；不连接 HKEX 撮合系统，不发送真实订单。
- 本 change 的视觉事实源仅为 UX Securities Logo 和本 change 的品牌 token，不复用其他公司的品牌规范、素材或业务内容。
