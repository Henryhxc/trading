# UX Securities 港股交易终端一口气实施 Prompt

开启 goal mode，token_budget=500000000。

## Superpowers 要求

- 开始前加载并遵守 `$using-superpowers`、`$test-driven-development`、`$verification-before-completion`、`$subagent-driven-development`，以及当前 change 所需的 OpenSpec apply 工作流。
- 每个实现 slice 必须先 RED，再最小 GREEN，最后 refactor；没有看到预期失败的 RED test 或 source guard，不得写生产代码。
- 主 agent 是唯一调度者，child agent 只执行明确 packet，不递归派工、不跨 phase、不宣布总任务完成。
- 每个 child agent 完成后写 handoff 或 review 并立即停止；最多保持 3 个 live child agents。
- 任何 tests、API、截图、Docker、数据库或 review 不通过，都必须为责任 packet 开新的 rework child agent，重新验证后才能推进。

## 目标

在当前工作区完整实现 `ux-securities-hkex-trading-terminal`：React 前端、Go MVC 单体后端、MySQL、Docker Compose、phpMyAdmin、UX Securities Logo 与品牌色、用户认证、港股行情、收藏、模拟订单、成交、持仓和订单历史。必须真正启动 Docker，并使用 Chrome/Computer Use 打开 Docker 暴露的 localhost 完成端到端测试和截图。只有 `tasks.md` 中全部 checkbox 都已由主 agent 根据证据打成 `[x]`，才可以报告完成。

## 工作目录与变更入口

- 工作目录：`/Users/hongxincong/project/UX Security`
- Change root：`openspec/changes/ux-securities-hkex-trading-terminal/`
- 第一恢复入口：`openspec/changes/ux-securities-hkex-trading-terminal/prompt.md`
- 第二恢复入口：`openspec/changes/ux-securities-hkex-trading-terminal/agent-dispatch/prompts/` 下最新日期 prompt
- 原始 prompt：`openspec/changes/ux-securities-hkex-trading-terminal/agent-dispatch/original-execution-prompt.md`
- 阶段控制器：`openspec/changes/ux-securities-hkex-trading-terminal/agent-dispatch/phase-controller.md`

## 必读事实源

按以下顺序读取，不得用聊天记忆替代磁盘文件：

1. `AGENTS.md`、`agent.md`
2. `openspec/changes/ux-securities-hkex-trading-terminal/proposal.md`
3. `design.md`、`tasks.md`、全部 `specs/**/spec.md`
4. `agent-dispatch/00-main-agent-dispatch.md` 与当前 phase packet
5. `docs/PRD.md`、`docs/research/hkex-free-data.md`
6. `docs/assets/ux-securities-logo.png`
7. `agent-dispatch/orchestration-state.md`、最新 handoff、review、results report

若本 prompt 与 change 中的 proposal/design/spec/tasks 冲突，停止实施，先以 change 文件为准并在 self-check/change-log 记录冲突；不得自行扩大范围。

## 必须保持的业务和技术边界

- 品牌只使用 UX Securities Logo 和 design.md 的 token：`#1D41A9`、`#2857E2`、`#0B0B0D`、`#FFFFFF`、`#6B7280`。
- React 只能调用内部 `/api/v1/*`；不得从浏览器直连腾讯、新浪或其他第三方行情地址。
- 行情 provider 统一走 Go `MarketDataProvider`；免费公开源只能标记 `unofficial`、`delayed` 或 `stale`，mock 标记 `simulated`，未获授权不得标记 `live`。
- `stock-api` 与 `easyquotation` 只作为免费行情适配参考；代码映射为内部六位代码 `00700`，由后端 adapter 处理 `HK00700`、`hk00700`、`00700`。
- 第一版只做模拟交易，绝不连接真实撮合、券商真实下单或发送真实订单。
- 用户、收藏、订单、成交、账户和持仓必须按 session actor 隔离；客户端不得决定成交状态、余额或持仓。
- 数据库采用 8 张传统宽表：`users`、`sessions`、`stocks`、`stock_daily_quotes`、`watchlists`、`accounts`、`orders`、`positions`；订单直接保存成交字段，不新增 `executions` 或 `position_lots`。
- 只做必要 hash：密码 hash 和 session token hash；用户名、邮箱、股票、订单、幂等键及其他业务字段保持明文可查询。
- 每个生产代码文件少于 500 行；超过必须按职责拆分并记录。
- 默认端口：前端 `3000`、后端 `8080`、MySQL `3306`、phpMyAdmin `8081`；允许环境变量覆盖，但必须记录实际端口。

## 调度方式

严格按三个 phase 顺序执行：

1. `agent-dispatch/01-phase1-foundation.md`：品牌壳、认证、数据库、Docker。
2. `agent-dispatch/02-phase2-market-workbench.md`：行情 provider、搜索、报价、日线、收藏、工作台。
3. `agent-dispatch/03-phase3-paper-trading.md`：模拟订单、成交、账户、持仓、历史。

每个 phase 固定三个 lane：开发代码 agent、TDD/e2e 测试验证 agent、查收 review agent。主 agent 在 phase PASS 前必须读取实际 diff、测试输出、截图、report、handoff 和 review；PASS 后立即更新 `tasks.md`、phase packet 和 `orchestration-state.md`，终止这三个 agent，再启动下一 phase。

## 中间 Phase 控制器（硬门禁）

执行时必须同时读取并遵守 `agent-dispatch/phase-controller.md`。phase 状态只能按 `pending → running → validating → passed → terminated` 推进；失败时进入 `rework`，真实外部阻塞时进入 `blocked`。任一时刻只能有一个 phase 为 `running` 或 `validating`。

- Phase 1 PASS 前禁止启动 Phase 2；Phase 2 PASS 前禁止启动 Phase 3。
- 每个 phase 都必须经过 `entry-check → red-check → green-check → runtime-check → browser-check → review-check → main-acceptance → termination-check`。
- 主 agent 负责 phase 状态、checkbox 和下一阶段启动；child agent 不能跨 phase、不能打总任务勾、不能宣布完成。
- 当前 phase 失败时只对当前 phase 返工，必须新开 child agent，重新跑当前 phase 全部 gate；不能用下一 phase 的测试掩盖当前 phase 失败。
- phase PASS 后必须先终止当前三个 child agent，更新 orchestration state，再启动下一个 phase。

## 一口气执行周期

1. 检查 Docker、Node、Go、MySQL 客户端、Chrome/Computer Use 能力；把环境结果写入 orchestration state。
2. 读取全部必读文件；保存本 prompt 到 `agent-dispatch/original-execution-prompt.md`，并确认 dated snapshot 存在。
3. 启动 Phase 1 的三个 child agents，并把 state 改为 `running`；TDD agent 先提交并运行 RED，依次完成该 phase 的八个中间检查点。
4. 只有 Phase 1 通过 `runtime-check` 和 `browser-check` 后，才进入 review 和主 agent acceptance；PASS 后更新 1.x checkbox、写 state、终止三个 agent。
5. 按同一控制器启动 Phase 2，再启动 Phase 3；每次只能运行一个 phase。Phase 3 必须完成真实浏览器链路：注册/登录 → 搜索港股 → 查看报价/日线 → 收藏 → 下模拟订单 → 模拟成交 → 查看余额/持仓/订单历史。
6. 若任何命令、测试、截图、视觉、DB 或 review 失败，将当前 phase 标记 `rework`，新开 rework agent，重新跑当前 phase 全部八个检查点；不得跳到下一 phase。
7. 全部 phase `terminated` 后，运行全量测试、API smoke、Docker 冷启动、MySQL 查询、phpMyAdmin 查看、浏览器回归、生产代码行数检查和 OpenSpec strict validation。

## Docker 与浏览器硬门禁

- 使用项目 Compose 文件启动；记录 `docker compose ps`、healthcheck、迁移日志和服务 URL。
- 后端不可用、MySQL 未健康、迁移未完成、前端无法调用内部 API 或 phpMyAdmin 无法查看数据库时，不能进入 PASS。
- Chrome/Computer Use 是 UI 主验收路径。Playwright/CDP/脚本截图只能补充，不能替代真实浏览器检查。
- 至少检查 `1440x900`、`1280x800`、`390x844`；覆盖登录、行情工作台、收藏、下单、持仓、订单历史的加载、空态、错误和成功态。
- 截图必须证明 Logo 比例、UX Securities 蓝色 token、焦点状态、行情来源状态、无横向溢出、无遮挡、无文字截断、窄屏可操作。
- 任一截图缺失、浏览器无法打开、Docker 不可用或视觉问题存在，都只能是 `needs_rework`/`blocked`，不可打勾、不可报告完成。

## RED/GREEN 与验证要求

- RED 必须因目标行为缺失而失败，不能因语法、环境或拼写错误失败；保存 RED 输出。
- GREEN 只写通过当前 RED 所需的最小实现；再做 refactor。
- Phase 1 至少验证认证、用户隔离、密码哈希、session 失效、migration rollback、Docker health。
- Phase 2 至少验证 provider 字段 allowlist、免费源状态、fallback、搜索、日线、收藏幂等和第三方请求边界。
- Phase 3 至少验证订单校验、幂等、状态机、余额/持仓原子事务、并发、用户隔离和过期行情拒绝。
- 每个数据库变更都写 `results/db-change-reports/`，包含 migration id、target DB、验证 SQL、rollback、失败/返工和残余风险。
- 任何未执行的测试或未打开的页面必须标记“待执行”，不得写 PASS。

## 上下文压缩、重启或恢复

如果发生上下文压缩、模型切换、终端重启、handoff 或怀疑记忆不完整，必须暂停代码改动并按以下顺序恢复：

1. 读取本文件 `prompt.md`。
2. 读取 `agent-dispatch/prompts/` 最新 `YYYY-MM-DD-prompt*.md`。
3. 读取完整 `agent-dispatch/original-execution-prompt.md`。
4. 读取 `AGENTS.md`、`agent.md`、`design.md`、`tasks.md`、全部 specs 和 `00-main-agent-dispatch.md`。
5. 读取 `orchestration-state.md`、`self-check-and-change-log.md`、最新 handoff/review 和所有 results report。
6. 执行 `git status --short --branch`、`git diff --name-only`；如果不是 Git 仓库，记录环境限制并检查文件树与报告。
7. 读取运行中的终端/Compose 输出，确认 child agent 状态、当前 phase、当前 packet、已完成 checkbox、失败项和下一步。
8. 在 `self-check-and-change-log.md` 写恢复时间、恢复依据、当前阶段、下一步和新发现。
9. 重新执行当前 packet 的 targeted verification；不得只凭压缩前记忆宣布完成。

如果 `prompt.md`、dated snapshot、original prompt、tasks 或 orchestration state 内容不一致，停止并先同步这些文件；同步前不允许写生产代码或勾选任务。

## 完成条件

只有同时满足以下条件才可最终报告完成：

- `tasks.md` 的所有 `- [ ]` 都已变成 `- [x]`，没有遗漏、跳过或以文字代替的任务。
- 三个 phase 都有开发 handoff、TDD/e2e handoff、独立 review，且结论为 `PASS` 或 `PASS_WITH_NOTES`。
- 所有指定截图、report、API smoke、DB report、Docker 日志和浏览器端到端证据已落盘。
- `openspec validate ux-securities-hkex-trading-terminal --strict` 通过。
- 主 agent 已独立复跑关键验证，并在 `orchestration-state.md` 标记三个 phase `passed`、child agents `terminated`、整体 `complete`。
- 仍有任何 `[ ]`、`needs_rework`、`blocked`、缺失截图、未验证数据库、未启动 Docker 或未完成浏览器链路时，必须继续执行或明确报告 blocker，不能声称完成。

## 最终报告格式

最终报告必须列出：实现摘要、Docker 服务和实际 URL、测试命令与结果、Chrome/Computer Use 截图目录、DB change reports、OpenSpec validation、所有 tasks 已打勾的证据、剩余风险。不要把待执行项目写成通过。
