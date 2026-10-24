# Candlestick 独立复核

- 复核日期：2026-10-24（Asia/Hong_Kong）
- 范围：本轮多周期 K 线、公开历史行情适配、入库与 API/前端合同，以及已报告 P2 修复。
- 角色：独立复核；未参与本轮生产代码修改。本报告只新增复核文档，不修改任务勾选或阶段 gate。
- 基线限制：工作目录没有 `.git`，无法执行相对固定 commit 的 diff 审查或 `git diff --check`。本次为基于当前源码、合同和证据的范围内 Standards / Spec 复核。

## 结论

**代码与已复核测试：PASS_WITH_NOTES。** 在本报告范围内，没有尚未解决、已确认的 P1/P2 发现。先前发现的迁移重入、未知证券错误映射、默认 composed mock 的原生分钟证券校验、legacy 响应合同问题均已修复并复核。

**整体阶段及项目 gate：仍为 REWORK / NOT FULLY VERIFIED。** 本结论不等于三视口浏览器验收、完整部署验收或真实 AkShare 全链路通过。Tencent/Yahoo 公开历史数据成功不能替代原有真实 AkShare gate。

## Standards 与 Spec 复核

| 维度 | 结论 | 已核查内容 |
| --- | --- | --- |
| Standards | 无未解决的已确认 P1/P2 | 职责拆分、相关新增文件不超过 500 行、前端通过内部 API 访问行情、参数化 SQL、明确事务边界、提交后发布事件、供应商来源不被改写为 HKEX 官方实时行情。 |
| Spec | 范围内代码合同通过，运行 gate 待补 | 九个周期、默认日线、六个日历月范围、真实原生分钟数据、缺失分钟数据为空而非合成、未知证券 404、未知 turnover 为 null、实际覆盖区间和来源/状态展示。 |

### 已关闭的发现

1. **Migration 004 重入**：`stock_intraday_quotes` 使用 `CREATE TABLE IF NOT EXISTS`。MySQL 集成证据覆盖表存在、迁移 ledger 条目缺失后的再次迁移，不再仅依赖正常首次迁移路径。
2. **未知证券 candle 错误映射**：HTTP handler 将 `domain.ErrNotFound` 映射为 404；原生分钟 repository 查询前确认 `stocks` 中证券存在。
3. **默认 composed mock 的分钟证券身份校验**：`Provider.NativeCandles` 在没有 native adapter 时通过 `Quote` 确认证券。已知证券返回空原生序列，未知证券返回 `ErrNotFound`，不生成 mock 分钟柱。新增测试使用真实 composed provider，并覆盖 `1m/5m/15m/30m/60m`；对应 RED/GREEN 证据已检查。
4. **Legacy candle 响应合同**：OpenAPI 以 `oneOf` 区分显式 interval 响应与省略 interval 的旧版日线响应。前者必须含 `source/data_status/as_of`；后者限定 `interval=1d` 且禁止上述 metadata。与当前 handler 行为一致。
5. **Public 模式 Python 回归测试有效性**：测试显式统计 AkShare history 调用并断言为零，避免“抛异常后被 fallback 吞掉”造成假通过。AkShare spot 调度与 public history 分支分别核查，不把 history 零调用解释为整个调度器禁用 AkShare。

## 验证来源与边界

### 复核者独立执行

| 命令 | 结果 | 边界 |
| --- | --- | --- |
| `cd apps/backend && go test ./internal/service/market ./internal/httpapi ./internal/integration/marketdata -count=1` | PASS；三个包均通过 | 最终 P2 修复后重新执行；不等于独立重跑完整 Go suite 或 vet。 |
| `python3 -m unittest discover -s services/akshare/tests -v` | PASS；22/22 | 最终测试修复后重新执行，含 public history 零 AkShare 调用断言。 |
| `cd apps/frontend && npx vitest run src/features/market/WorkbenchChart.test.tsx src/features/market/useCandleSeries.test.tsx src/features/market/chart-data.test.ts` | PASS；3 文件 / 8 测试 | 同轮先前执行，之后前端源码未变；不是浏览器视觉或交互验收。 |

OpenAPI `oneOf` 为人工源码合同检查。可选的 `jsonschema` 依赖探测失败（环境未安装），**未执行 JSON Schema 自动校验**，不得把这项写为 PASS。

### 仅审阅实施方保存的证据

以下路径相对当前 change 目录 `openspec/changes/ux-securities-hkex-trading-terminal/`。它们属于证据审阅，不是复核者再次执行：

- `results/candlestick/composed-mock-native-red.txt`、`composed-mock-native-green.txt`：五个分钟周期未知证券在修复前失败、修复后通过。
- `results/candlestick/mysql-green.txt`：3 项隔离数据库测试通过；覆盖提交与 provenance、旧批次不覆盖新批次、失败回滚、原生分钟存储/唯一性、迁移恢复、null 持久化。
- `results/candlestick/go-final-green.txt`：完整 Go 包测试通过记录。
- `results/candlestick/openspec-strict.txt`：严格验证输出 change valid。
- `results/candlestick/docker-deploy.txt`：重建/启动 backend、frontend、akshare，记录 MySQL/backend healthy；该文件本身不能证明所有容器持续健康或全部用户运行流程通过。
- `results/candlestick/native-upstream-probe.txt`：真实 AkShare minute 请求仍为 `RemoteDisconnected`；Yahoo 各分钟周期探测成功。探测成功不等于所有周期已经完整入库。
- `results/candlestick/mysql-runtime-import.txt`：真实公开数据已开始入库的快照，详情见下一节。

复核者没有独立重跑完整 Go suite、vet、前端 build、Docker 部署、破坏性 MySQL 集成准备或浏览器验收；不得将实施方证据改写为复核者执行结果。

## 已见运行证据快照

最后审阅的 `mysql-runtime-import.txt` 显示：

| 证券 | 周期 | 来源 | 条数 | 实际覆盖 |
| --- | --- | --- | --- | --- |
| 00700 | 1d | tencent | 125 | 2026-04-24 至 2026-10-23 |
| 00700 | 1m | yahoo | 1655 | 2026-10-16 01:30Z 至 2026-10-23 08:08Z |
| 00700 | 5m | yahoo | 620 | 2026-08-27 01:30Z 至 2026-09-09 02:50Z |

该快照采集时导入仍在进行，不能推断全部五个分钟周期已完整入库。复核者仅审阅此文件，没有独立发出运行数据库查询。真实 AkShare 请求失败的记录不能作为真实 AkShare 入库成功证据；也没有通过本报告独立确认浏览器中 125 条日线的显示。

## 仍需补齐的 gate

- 原始 Phase 2 登录后行情搜索、收藏等三视口验收及视觉/合同独立对照：`tasks.md` 的 `2.9/2.10/2.12`。
- 原始 Phase 3 完整模拟交易、持仓和订单历史三视口流程及独立 review：`3.9/3.10/3.12`。
- 带真实 AkShare 样本的市场数据库报告，以及 scheduler → HMAC → MySQL → REST → 已认证 WS 的实际链路；失败、重放、回滚、stale 和重启证据：`2.11/2.15`。
- 本轮 candlestick 用户可见验收、最终证据联动及任务勾选：`5.1–5.4` 仍应依对应证据判断。
- 全局与最终 gate：`0.5/4.1/4.2` 不能因本次范围内代码审查通过而自动完成。

本报告不修改 `agent-dispatch/orchestration-state.md` 的 `rework` 状态，不替代上述 gate。

## 已知限制

公开 Tencent/Yahoo 数据继续作为非官方来源；分钟实际覆盖可能短于请求的六个月，应以返回的实际范围为准。未知 turnover 保留 null，不根据价格与成交量估算。AkShare 上游断连与本次缺少的浏览器/认证 WS/完整重启实证仍需诚实记录，不能通过源码测试推断通过。没有执行真实券商连接或真实下单。

## 补充复核：400 行 webhook 批量写入

本节在初版报告后追加，范围为 `market_candle_batch.go`、其调用方与新增 deadline 测试；不是对完整运行链路的重新验收。

### 问题证据与当前结论

已审阅 `results/candlestick/delivery-deadline-red.txt`：优化前真实隔离库的 400 柱事务在约 10 秒时出现 `context deadline exceeded`，证明逐行写入存在超过 sender 15 秒单次请求预算的实际风险。该失败不能以此前的单条写入测试通过覆盖。

**补充结论：批量优化源码及已审阅隔离库证据为 PASS_WITH_NOTES，未发现此优化引入的已确认 P1/P2。** 最终 `delivery-deadline-green.txt` 的五项测试全部 PASS，总耗时 17.703 秒；其中 400 行插入与精确重放共 245.842792ms，满足同一 10 秒 context 的预算。此为实施方测试证据审阅，复核者未并行运行任何破坏性数据库测试。原始三视口及真实 AkShare gate 仍不升级。

### 已核查边界

- **乱序与幂等**：按 `symbol/interval/source/data_status/collected_at` 分组；每组最多 400 行。已有行的 OHLCV、NULL turnover、来源和状态仅在传入 `collected_at` 更晚时更新；`collected_at` 最后赋值，保持各字段判断使用同一个原存储时间。唯一约束继续区分证券、周期和柱时间；没有将公开源改名为 AkShare。
- **混合周期**：空 interval 先归一为 `1d`；分钟与日线使用各自固定表/列清单；不同 interval 不进入同一条 bulk SQL。每证券复用一次 metadata 查询结果，不覆盖已存在的 spot quote/provenance。新增混合测试覆盖两个证券、daily/1m/5m、两个收集时间组，已在最终隔离库证据中 PASS。
- **绑定与 NULL**：SQL 拼接仅使用常量表名、列名和占位符；所有市场值绑定参数。缺失 turnover 继续以 SQL NULL 写入，没有估算或回填零。
- **事务回滚**：metadata、quote、所有 candle group 仍共享同一个事务；任意 group 失败立即返回，defer rollback；仅 commit 成功后返回事件。除原有 quote 后 candle 失败案例外，新增混合测试使最后一个 group 失败，并确认先前已执行 group 的 `stocks/stock_daily_quotes/stock_intraday_quotes` 三表均无泄漏；最终证据 PASS。
- **事件**：每个 `symbol:interval` 聚合最多一个 `candles.updated`；仅实际 `RowsAffected > 0` 的 group 生成事件，在同批多个已变更 group 中保留较新的收集时间。服务仅在 repository 成功返回后发布，因此失败/commit 失败不发布本批事件。混合测试确认成功时六个唯一 period 事件且取较新收集时间，精确重放零事件；最终证据 PASS。

### 本轮独立执行与待补项

独立执行 `env -u INGESTION_TEST_DSN -u PHASE2_MIGRATION_TEST_DSN -u PHASE1_MIGRATION_TEST_DSN go test ./internal/service/market ./internal/repository/mysqlrepo -count=1`：两个包 PASS。该执行显式禁用对应数据库 DSN，数据库集成测试为跳过，不能声称本轮隔离库 GREEN。

新增 deadline 测试源码要求同一 10 秒 context 内完成 400 个 1m 柱的插入及精确重放、确认 400 行与一次 period event、重放零事件；最终 GREEN 输出已审阅。该文件还覆盖混合组原子性/事件、原有 provenance 与乱序保护、quote 回滚、原生分钟唯一性/旧批次保护/NULL。混合测试修正前曾把 migration 预置的三条 stock 误判为事务泄漏；失败输出保存在 `mixed-group-test-seed-failure.txt`，测试现于迁移后明确清空 seed 再执行，修正后的完整五项测试通过。这是测试前置条件修正，不是以删除失败证据替代修复。

混合测试在 repository 层直接调用，包含同一天日线的两个收集版本，用于验证 SQL 的先后更新边界；service 仍会拒绝同一个 webhook 内的重复日柱，不能把 repository 测试解释为 API 允许重复。

另审阅 `backend-batch-deploy.txt`、`browser-auto-refresh-snapshot.txt` 与 `browser-60m-1311x1076.jpg`：部署记录包含 batch 后端镜像构建/替换和 scheduler 重启；snapshot 记录浏览器 00700 的 60m 875 根、来源 yahoo、采集时间 2026-10-24 16:21:15、读取时间 16:24:34。复核者直接查看了该 1311×1076 截图，确认 60 分钟选中态、875 根、Yahoo/UNOFFICIAL、蜡烛与成交量、实际覆盖范围、缺失半年保证提示及来源说明可见。这些是对实施方运行产物的复核，不是复核者自行操作浏览器、原始三视口验收或完整 Docker 验收。

最终再次回读更新后的 `mysql-runtime-import-final.txt`，确认三证券均已有 125 根 Tencent 日线，日期为 2026-04-24 至 2026-10-23；全部五个 Yahoo 分钟周期均已入库：

| 证券 | 1m | 5m | 15m | 30m | 60m |
| --- | ---: | ---: | ---: | ---: | ---: |
| 00005 | 1627 | 2678 | 920 | 480 | 875 |
| 00700 | 1655 | 2679 | 920 | 480 | 875 |
| 09988 | 1655 | 2678 | 920 | 480 | 875 |

快照所有分组收集时间均为 `2026-10-24 08:21:15.612` UTC（香港 16:21:15.612），NULL turnover 共 19,797 条，重复键 0。该完整快照替代此前导入中快照，证明此次公开源回填不再停在 00700 的部分周期；仍不能替代真实 AkShare 成功、REST/认证 WS 全链路、原三视口及最终项目 gate。
