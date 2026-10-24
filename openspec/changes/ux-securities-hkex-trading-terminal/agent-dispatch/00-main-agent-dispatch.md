# 主 agent 调度合同

## 目标

把 UX Securities 港股行情与模拟交易终端按三个 phase 交付为可复核的 OpenSpec change。主 agent 只负责读取事实源、派发 phase、验收证据、合并结论和返工；不替代 phase agent 大包实现。

## 事实源顺序

当前用户请求与 Logo → 本 change 的 proposal/design/tasks/specs → `AGENTS.md` 与 `agent.md` → `docs/PRD.md` → `docs/research/hkex-free-data.md` → 官方 HKEX 资料。外部示例只做校准。视觉只使用 UX Securities Logo 和本 change token。

## 执行规则

1. 严格按 Phase 1 → Phase 2 → Phase 3 顺序执行；后续 phase 不绕过前一 phase 的 PASS。
2. 每个 phase 固定三个 agent：开发代码、TDD/e2e、查收 review。TDD agent 必须先 RED，开发 agent 做最小 GREEN，review agent 独立复核。
3. UI phase 使用用户最新指定的 in-app browser 做真实 localhost 检查和截图；没有截图、report、handoff 或 review recommendation 不能 PASS，IAB 证据不得标记为 Chrome/Computer Use。
4. Phase PASS 后，主 agent 必须确认所有证据落盘，更新 `tasks.md` checkbox、phase packet checkbox、handoff recommendation 与 `orchestration-state.md`，然后终止该 phase 的三个 agent，再启动下一 phase。
5. 免费公开行情只能用于开发、演示和模拟交易；没有授权时不能标记 `live` 或连接真实撮合。

## 中间 Phase 控制

详细状态机和每个 phase 的 gate 以 [`phase-controller.md`](phase-controller.md) 为准。主 agent 必须保证同一时间只有一个 phase 处于 `running/validating`，并按 `pending → running → validating → passed → terminated` 推进；失败只能回当前 phase `rework`，不能跳阶段。下一 phase 必须等待上一 phase 的三个 child agent 已终止。

## 启动与恢复

执行入口是 change root 的 `prompt.md`。从 `orchestration-state.md` 的 `next_action` 恢复；先读取 `prompt.md`、`agent-dispatch/prompts/` 最新 dated snapshot、`original-execution-prompt.md`、当前 phase packet、最近 handoff/review 和 results report，再执行未完成 checkbox。若发生上下文压缩、重启、handoff 或文件冲突，必须重新读取整个 change 并恢复磁盘状态，不得以聊天上下文代替文件证据。

## 最终验收

运行 `npx openspec validate ux-securities-hkex-trading-terminal --strict`、`openspec status --change ux-securities-hkex-trading-terminal`、定向测试、Docker 冷启动、DB 验证、用户指定的 IAB 浏览器回归和行数检查；只有 `tasks.md` 全部 checkbox 为 `[x]` 才能完成。在 Git 仓库初始化前，`git diff --check` 只能记录环境限制，不能伪造 PASS。
