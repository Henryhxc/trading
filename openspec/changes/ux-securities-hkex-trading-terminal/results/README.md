# 证据目录

本目录保存 phase 测试输出、Chrome/Computer Use 截图、视觉比较、API smoke、Docker 验证和数据库变更报告。截图与 report 是 phase PASS 的硬证据；未执行项目必须写 `待执行`，不得标为 PASS。

目录约定：

- `screenshots/prototype-reference/`：实现前的参考图或 Logo 基线。
- `screenshots/phase1-foundation/`：认证与品牌壳。
- `screenshots/phase2-market-workbench/`：行情、日线、收藏与响应式。
- `screenshots/phase3-paper-trading/`：下单、成交、持仓和历史联动。
- `db-change-reports/`：每次 migration 的 target DB、验证 SQL、rollback、失败与残余风险。
