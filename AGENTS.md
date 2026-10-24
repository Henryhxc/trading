# UX Securities 项目智能体协作规则

## 项目目标

本项目规划并实现一个面向港股行情的个人交易终端，产品品牌使用用户提供的 **UX Securities Limited** Logo。附带的 Futubull 截图只作为信息架构、交易工作台密度和交互分区的参考，不作为素材、文案或功能清单的逐像素复制依据。

## 已确认的技术约束

- 前端使用 React；后端使用 Go。
- 后端采用 MVC 组织方式。业务代码仍需按职责拆分，不能把所有 handler、SQL 和业务判断塞进一个文件。
- 使用 MySQL 保存用户、收藏、模拟订单、持仓和订单历史；数据库通过 phpMyAdmin 查看。
- 使用 Docker Compose 单体式部署。三项核心服务为 `frontend`、`backend`、`mysql`；phpMyAdmin 是可选的管理服务和辅助端口，不改变业务服务边界。
- 建议默认端口：前端 `3000`、后端 `8080`、MySQL `3306`、phpMyAdmin `8081`。端口可由环境变量覆盖，但文档、Compose 和本地说明必须保持一致。
- 任一生产代码文件不超过 500 行。超过前必须按职责拆分；不得通过压缩、生成代码或把大量逻辑放到单个配置文件来规避限制。
- 视觉基线仅使用用户提供的 UX Securities Logo，以及清晰的金融终端深色工作台和可读的交易数据层级。

## HKEX 数据边界

HKEX 官方资料提供的是授权的实时行情数据馈送，不是一个面向普通应用的免费 REST 行情 API：

- 证券市场使用 HKEX Orion Market Data Platform（OMD-C），产品包括股票、ETF、REIT、债券、窝轮、牛熊证等；数据层级包括 SS、SP、SF。
- 衍生品市场使用 OMD-D，覆盖股指及个股期货/期权、利率、汇率、黄金和商品等；数据层级包括 DS、DP、DF。
- IIS 提供发行人公告和交易相关消息。
- HKEX Data Marketplace 提供历史逐笔、历史完整盘口和静态证券资料，可用于回测和开发测试。
- OMD 页面提供二进制协议、开发者指南、连接指南及授权/许可资料。实施时不得抓取 HKEX 网页报价冒充正式行情接入，也不得把行情许可写成下单权限。
- MVP 默认使用“行情适配器 + 模拟撮合/模拟订单”完成产品闭环。真实下单必须另接有权限的券商或交易参与者 API，并通过独立的 execution adapter；在获得供应商合同、接口文档和用户授权前不得发送真实订单。

官方事实源（访问日期：2026-10-24）：

- [HKEX Real-time Datafeeds](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Overview/Real_time-Datafeeds?sc_lang=en)
- [HKEX OMD-C](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Securities-Market-OMD-C?sc_lang=en)
- [HKEX OMD-D](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Derivatives-Market-OMD-D?sc_lang=en)
- [HKEX Data Marketplace](https://www.hkex.com.hk/Services/Market-Data-Services/Historical-Data-Services/HKEX-Data-Marketplace?sc_lang=en)

## 免费港股行情调用约定

MVP 允许使用公开行情源完成开发和模拟交易，但必须把它们当作非官方适配器，不得宣称已接入 HKEX 官方实时行情：

- 默认 provider 是 `MockMarketDataProvider`，用于开发、演示和自动化测试。
- 可选的免费原型来源是 [stock-api](https://github.com/zhangxiangliang/stock-api) 所使用的腾讯/新浪公开行情接口，或 [easyquotation](https://github.com/shidenggui/easyquotation) 所使用的腾讯港股接口。两者只作为数据源适配参考，受上游限流、字段变化和服务条款约束。
- Go 后端通过 `MarketDataProvider` 接口统一调用；React 前端只能调用内部 `/api/v1/markets/*`，不得直接请求腾讯、新浪或第三方行情地址。不得为了接入它们增加额外业务端口。
- 内部证券代码统一为六位数字，例如 `00700`；外部映射由 adapter 负责：腾讯使用 `hk00700`，`stock-api` 使用 `HK00700`，`easyquotation` 使用 `00700`。
- provider 至少返回 `symbol`、`display_name`、`last_price`、`change`、`change_percent`、`volume`、`turnover`、`as_of`、`source` 和 `data_status`。免费公开源只能标记为 `unofficial`、`delayed` 或 `stale`，不能标记为 `live`；模拟源标记为 `simulated`。
- 建议降级顺序为：配置的授权 provider → 免费公开 provider → mock。超时、限流、字段解析失败和时间戳过期必须记录脱敏日志，并返回可解释的 `data_status`。
- `stock-api` 是 Node.js/TypeScript 库，`easyquotation` 是 Python 库；不得把它们直接嵌进 Go 业务层。正式实现应在 Go 中写独立 adapter，或仅在本地开发工具中调用它们。
- 免费行情只服务模拟订单，不能触发真实下单。真实实时行情需另行接入已授权的 HKEX OMD-C 或信息供应商，并完成许可审查。

## 业务范围

第一版必须支持：注册/登录、港股搜索和行情展示、关注/收藏、模拟买卖、模拟持仓、订单历史、账户余额和基本风险校验。真实经纪商开户、入金、出金、真实交易、杠杆、融资融券、期权行权和监管报表不属于第一版。

订单、持仓和成交必须按用户隔离。模拟订单状态至少包含 `pending`、`filled`、`cancelled`、`rejected`；任何行情延迟、行情缺失、余额不足、数量不合法或市场关闭都必须能在 API 和界面中解释。

## 项目档案与实施文档

设计阶段必须维护以下项目档案；路径和职责不能只写在聊天里：

```text
docs/
├─ PRD.md
├─ design/
│  ├─ design.md
│  ├─ develop-frontend-design.md
│  ├─ develop-backend-design.md
│  ├─ develop-design-task-linkage.md
│  └─ project-archive.md
├─ api/
│  ├─ openapi.yaml
│  └─ hkex-market-data.md
└─ evidence/
   ├─ README.md
   ├─ screenshots/
   └─ reports/
apps/
├─ frontend/
└─ backend/
deploy/
└─ docker-compose.yml
db/
└─ migrations/
```

`design.md` 是业务语义母版，先写范围、角色、流程、状态、权限、数据可见性和验收标准，再写技术补充。前后端设计必须分别列出目录树、具体文件职责、接口字段、失败分支、测试命令和证据路径。涉及数据库时先完成数据合同、迁移和回滚设计，再进入依赖它的应用实现。

## 代码与 API 规则

- 共享 API 合同优先写入 `docs/api/openapi.yaml`，前端和后端不得各自发明字段。
- Go 侧按 `controller/handler → service/usecase → model/domain → repository → integration` 拆分；React 侧按页面、功能组件、数据访问和状态模块拆分。具体文件以 `design.md` 最终清单为准。
- 数据库写入必须有明确事务边界、唯一约束和失败后的状态说明。密码只存强哈希，session/token 不写日志。
- 外部行情供应商的凭据只能来自环境变量或密钥管理，不写入仓库、截图、PRD、测试输出或前端 bundle。
- 日志必须去敏并带可关联的 request/order 标识；禁止记录密码、完整 token、数据库密码和原始供应商载荷。
- 读取、测试和文档检查可以自动执行；未经明确授权不得连接真实券商、发送真实订单或修改生产数据库。

## 验证要求

- 文档阶段：检查需求追溯、官方来源链接、目录清单、API 字段一致性和 `git diff --check`。
- 实施阶段：行为变更遵循 RED → 最小 GREEN → 定向测试 → 集成/E2E → 独立审查。
- UI 验证至少覆盖登录、行情搜索、收藏、下单、持仓和订单历史的加载/空态/错误/成功状态；截图使用与参考图匹配的视口，并记录差异和返工结论。
- Docker 验证必须覆盖冷启动、MySQL 健康检查、迁移、后端健康接口、前端访问、phpMyAdmin 查看和服务间网络连接。
- 未执行的测试必须标为“待执行”，不能写成 PASS。
