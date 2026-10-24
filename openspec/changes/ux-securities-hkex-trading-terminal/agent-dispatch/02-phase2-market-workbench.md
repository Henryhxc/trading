# Phase 2：行情 provider、搜索、日线、收藏和工作台

## 目标与范围

建立统一 `MarketDataProvider`、mock/免费公开 adapter 边界、搜索/报价/日线/收藏 API 和品牌化行情工作台，明确 `unofficial/delayed/stale/unavailable/simulated` 状态。

## 开发代码 agent

- 读取前一 phase 的 PASS 证据、本 packet、行情/收藏 spec、`agent.md` 免费行情约定和 `docs/research/hkex-free-data.md`。
- 按 `tasks.md` 2.1–2.7 实现最小 GREEN；React 不直连第三方行情源，Go adapter 负责代码映射和降级。
- 写行情/收藏 migration、验证 SQL、rollback 与 DB report，记录数据库专家路径复查结果。
- 自检用户隔离、来源状态、Logo/token、响应式布局，写 `phase2-code-handoff.md`。

## TDD/e2e 测试 agent

- 先写 provider 状态映射、fallback、搜索、收藏幂等和权限 RED，记录 RED 后跑 targeted GREEN 与 API contract smoke。
- 使用用户指定的 in-app browser 打开真实 localhost 工作台，在 `1440x900`、`1280x800`、`390x844` 检查搜索、报价、日线、收藏、stale/unavailable、滚动锚点；截图到 `results/screenshots/phase2-market-workbench/`。
- 写 `phase2-tdd-handoff.md`，附 provider fixture、网络失败和截图 report。

## 查收 review agent

- 独立比较设计合同、spec、API、截图和实现 diff；检查 Logo、品牌色、来源徽章、空态、错误提示、横向溢出和第三方请求边界。
- 验证免费行情没有标记 `live`，并核对 DB report、保留策略和 rollback。
- 写 `phase2-review.md`，有任何视觉或契约偏差即 `needs_rework`。

## 证据与完成打钩 gate

- 证据：`results/screenshots/phase2-market-workbench/`、行情/收藏 DB report、两个 handoff、review、API/test 输出。
- 主 agent 复核后，才可勾选 `tasks.md` 2.1–2.12 对应完成项，更新 phase state 为 `passed` 并终止三个 agent。
