# Phase 3：模拟订单、成交、账户、持仓和历史

## 目标与范围

实现模拟限价订单、幂等、服务端校验、原子撮合、账户余额、持仓、成交和订单历史，并用完整浏览器流程证明联动；不连接真实撮合系统。

## 开发代码 agent

- 读取 Phase 1/2 PASS 证据、订单/portfolio spec、本 packet 和 `design.md` 事务合同。
- 按 `tasks.md` 3.1–3.7 实现最小 GREEN；客户端不得决定成交、余额或持仓，所有写入按事务边界执行。
- 复用 Phase 1 已创建的 accounts 表；只新增 orders/positions migration、索引、rollback、并发保护和 DB report；成交字段直接写入 orders，不新增 executions/position_lots；每个生产代码文件少于 500 行。
- 自检订单状态、幂等、过期行情拒绝、用户隔离与错误脱敏，写 `phase3-code-handoff.md`。

## TDD/e2e 测试 agent

- 先写余额/持仓/数量/价格校验、幂等、状态机、并发和事务回滚 RED，保存 RED 后跑 targeted GREEN、API smoke 与 DB 断言。
- 使用用户指定的 in-app browser 完成登录 → 搜索 → 查看行情 → 收藏 → 下单 → 模拟成交 → 持仓/历史完整流程，在三种视口截图到 `results/screenshots/phase3-paper-trading/`。
- 写 `phase3-tdd-handoff.md`，报告成功、拒绝、取消、过期行情和空态证据。

## 查收 review agent

- 独立检查订单状态可追溯性、余额/持仓原子性、重复提交、用户隔离、错误文案、视觉截图和 500 行限制。
- 确认没有真实订单出口、provider 凭据没有进 bundle/日志，DB report 含验证 SQL 与 rollback。
- 写 `phase3-review.md`；任何流程断裂、截图缺失、视觉遮挡或数据不一致都必须返工。

## 证据与完成打钩 gate

- 证据：`results/screenshots/phase3-paper-trading/`、交易 DB report、两个 handoff、review、E2E/API/事务输出和最终 Docker smoke。
- 主 agent 复核全链路与 strict validation 后，才可勾选 `tasks.md` 3.1–3.12、更新最终 state 为 `passed` 并终止三个 agent。
