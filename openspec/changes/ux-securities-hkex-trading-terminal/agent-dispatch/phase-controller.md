# Phase Controller：阶段控制器

本文件是三个 phase 的中间控制层。主 agent 必须按本文件控制状态，不能因为开发代码暂时可运行就跳过 gate。

## 状态机

```text
pending -> running -> validating -> passed -> terminated
                         |             |
                         v             v
                       rework       next phase running
                         |
                         v
                      validating
                         |
                         v
                       blocked
```

允许的状态值只有：`pending`、`running`、`validating`、`rework`、`passed`、`blocked`、`terminated`。

## 控制规则

1. 任一时刻只能有一个 phase 为 `running` 或 `validating`。
2. `phase N+1` 只有在 `phase N = passed` 且 phase N 的三个 child agent 都已 `terminated` 后才能启动。
3. 主 agent 每次切换状态前必须更新 `orchestration-state.md`，记录时间、证据、下一步和责任 agent。
4. child agent 不得修改全局 phase 状态，不得打总任务 checkbox，不得开启下一 phase。
5. `rework` 必须新开 child agent；不得复用已完成的 agent。返工完成后回到 `validating`，重新跑该 phase 全部 gate。
6. `blocked` 只用于真实外部阻塞；blocked 时不得把 phase 或 tasks 标成完成。
7. 主 agent 只有在看到开发 handoff、TDD/e2e handoff、独立 review、测试输出、截图、Docker/API/DB evidence 后，才能把 phase 从 `validating` 改为 `passed`。

## 浏览器表面决策

原始执行 prompt 指定 Chrome/Computer Use；用户后续直接指定使用 in-app browser。本次执行按最新用户指令使用 IAB 完成浏览器验收，不再把 Chrome/Computer Use 作为必选表面。证据必须标记为 IAB，不得声称执行过 CUA。该表面选择不降低视口、交互状态、截图和独立复核要求。

## Phase 1 gate：foundation

进入条件：规划文件已读，环境检查已记录，三个 phase agent 尚未启动。

必须通过：认证页面与 Logo 可访问；users/sessions/accounts migration、session 失效和用户隔离通过 RED/GREEN 与 API smoke；Docker Compose 真启动且 MySQL healthcheck、后端 migration、`/health`、前端、phpMyAdmin 可访问；IAB 在 `1440x900`、`1280x800`、`390x844` 保存登录/注册截图；handoff、review、DB report 全部存在。

退出动作：主 agent 更新 `tasks.md` 的 1.1–1.12、写 phase1 PASS 记录、终止三个 agent，然后才能把 Phase 2 从 `pending` 改为 `running`。

## Phase 2 gate：market-workbench

进入条件：Phase 1 已 `passed` 且三个 Phase 1 agent 已 `terminated`。

必须通过：provider contract、mock/fallback、免费行情状态、搜索、报价、日线、收藏 API 和 targeted tests；React 只调用内部 API；IAB 三种视口截图齐全；行情/收藏 DB report、handoff、review 全部存在，免费源没有标记为 `live`。

退出动作：主 agent 更新 `tasks.md` 的 2.1–2.12、写 phase2 PASS 记录、终止三个 agent，然后才能把 Phase 3 从 `pending` 改为 `running`。

## Phase 3 gate：paper-trading

进入条件：Phase 2 已 `passed` 且三个 Phase 2 agent 已 `terminated`。

必须通过：订单校验、幂等、状态机、取消、过期行情拒绝和用户隔离；成交、余额、持仓和订单历史单事务一致并有并发/回滚断言；IAB 完成注册/登录 → 搜索 → 收藏 → 下模拟订单 → 模拟成交 → 持仓/历史全链路；三个视口截图、交易 DB report、handoff、review、Docker final smoke 和 OpenSpec strict validation 全部存在。

退出动作：主 agent 更新 `tasks.md` 的 3.1–3.12、4.1–4.2，确认无任何 `[ ]`、`needs_rework` 或 `blocked`，将三个 agent 标记 `terminated`，再把整体状态改为 `complete`。

## 每个 phase 的固定中间检查点

每个 phase 都必须按以下顺序执行并写入 handoff/review：`entry-check` → `red-check` → `green-check` → `runtime-check` → `browser-check` → `review-check` → `main-acceptance` → `termination-check`。
