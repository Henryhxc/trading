## Purpose

提供使用 UX Securities Logo 和品牌 token 的 React 工作台，并以 Docker Compose 启动前端、Go 后端、MySQL 与本地 phpMyAdmin。

## ADDED Requirements

### Requirement: 页面必须遵守 UX Securities 品牌壳

系统 MUST 使用用户提供的 UX Securities Logo、主蓝 `#1D41A9`、强调蓝 `#2857E2`、近黑文字和白色表面作为可追溯视觉基线；Logo MUST 保持比例、清晰度和原始颜色。

#### Scenario: 登录和工作台渲染

- **WHEN** 用户打开登录页或行情工作台
- **THEN** 页面显示 UX Securities Logo 和统一品牌 token
- **AND** 交互焦点、hover、错误和涨跌状态均有可读的非颜色提示

#### Scenario: 窄屏访问

- **WHEN** 用户以 390x844 等窄屏打开工作台
- **THEN** 页面使用单列或抽屉布局
- **AND** 不出现横向溢出、遮挡或不可操作控件

### Requirement: 系统必须提供单体 Docker 运行入口

系统 MUST 提供 Docker Compose 配置，默认暴露前端 `3000`、Go API `8080`、MySQL `3306` 和本地 phpMyAdmin `8081`，并通过健康检查与依赖顺序完成冷启动。

#### Scenario: 冷启动

- **WHEN** 开发者执行 Compose 启动命令
- **THEN** MySQL 健康后后端完成迁移并提供健康接口
- **AND** 前端可访问内部 API，phpMyAdmin 可查看本地数据库

#### Scenario: 配置覆盖

- **WHEN** 开发者使用环境变量覆盖端口或 provider
- **THEN** 服务使用覆盖值启动
- **AND** 文档仍说明默认值与安全的本地管理范围

### Requirement: 生产代码文件必须受大小约束

系统 MUST 保证每个生产代码文件少于 500 行；超过职责边界时必须拆分并在设计档案中记录。

#### Scenario: 代码审查

- **WHEN** phase review 检查新增或修改的生产代码
- **THEN** 报告列出超过限制的文件（若有）
- **AND** 超限 phase 不能通过
