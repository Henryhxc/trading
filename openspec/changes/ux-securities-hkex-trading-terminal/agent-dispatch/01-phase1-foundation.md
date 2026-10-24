# Phase 1：品牌壳、认证、数据库和 Docker

## 目标与范围

建立 UX Securities 品牌壳、React 路由、Go MVC 基础、认证/会话、users/sessions/accounts migration、统一错误和 Compose 冷启动。输出必须证明用户隔离与本地可运行。

## 开发代码 agent

- 读取 `AGENTS.md`、`agent.md`、`docs/PRD.md`、`design.md`、认证/品牌 spec 和本 packet。
- 按 `tasks.md` 1.1–1.7 实现最小 GREEN；生产代码文件少于 500 行。
- 实施前重新检查数据库专家路径；若仍缺失，在 handoff 与 DB report 记录替代审查和风险。
- 自检代码、API、迁移、Docker 与 Logo 资源，写 `agent-dispatch/handoffs/phase1-code-handoff.md`。

## TDD/e2e 测试 agent

- 先为注册、登录、退出、过期 session、用户隔离和健康检查写 RED，并保存 RED 输出。
- 运行 targeted GREEN、API smoke、migration dry-run/apply/rollback；使用用户指定的 in-app browser 打开真实 localhost 登录/注册页，在 `1440x900`、`1280x800`、`390x844` 截图到 `results/screenshots/phase1-foundation/`。
- 检查 Logo 比例、焦点、错误文案、无溢出，并写 `agent-dispatch/handoffs/phase1-tdd-handoff.md`。

## 查收 review agent

- 独立读取 diff、测试、Docker 日志、MySQL 健康、phpMyAdmin 页面和 phase1 截图。
- 检查 session 安全、用户隔离、密码/密钥脱敏、迁移 rollback、默认端口和 500 行约束。
- 写 `agent-dispatch/reviews/phase1-review.md`；截图/证据缺失或可见缺陷时结论必须为 `needs_rework`。

## 证据与完成打钩 gate

- 证据：`results/screenshots/phase1-foundation/`、`results/db-change-reports/`、两个 handoff、review 和命令输出。
- 主 agent 复核后，才可勾选 `tasks.md` 1.1–1.12 对应完成项，把本 packet、handoff recommendation、phase state 更新为 `passed`，并把三个 agent 记为 `terminated`。

## Main-agent acceptance record

- 2026-10-24：主 agent reviewed `agent-dispatch/reviews/phase1-review-post-rework.md` (`PASS`), reran backend/frontend/contract/OpenSpec/runtime checks, and recorded `results/phase1-main-agent-acceptance-2026-10-24.md` (`PASS_WITH_NOTES`). Tasks `1.1`–`1.12` are checked. IAB evidence is explicitly not labeled Chrome/Computer Use; task `0.4` remains open pending product decisions.
