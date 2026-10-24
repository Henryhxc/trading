# 数据库变更报告

每个 migration 或会改变持久化边界的 phase 必须新增一份 Markdown 报告，至少包含：migration id、改了什么、为什么、target DB、表/索引/约束、验证 SQL、dry-run/apply/rollback、测试 fixture、失败与返工记录、用户隔离影响和残余风险。实施前若数据库专家文件仍缺失，必须在报告中保留该事实并说明替代审查。
