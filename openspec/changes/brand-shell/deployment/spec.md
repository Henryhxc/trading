## Purpose

建立 UX Securities 自有品牌的交易工作台和可重复运行的 Docker 单体入口，使用户能从登录进入行情、收藏、模拟交易和资产页面。

## ADDED Requirements

### Requirement: 界面必须使用 UX Securities 品牌合同

前端 MUST 使用用户提供的 UX Securities Logo，并使用从 Logo 提取的品牌 token：深蓝主色约 `#1D41A9`、亮蓝强调色约 `#2857E2`、近黑文字、白色背景和中性灰边框；不得引入其他公司的 Logo、品牌色或素材。

#### Scenario: 打开登录和工作台

- **WHEN** 用户打开登录页或行情工作台
- **THEN** 页面显示 UX Securities Logo 和统一品牌色
- **AND** 主按钮、选中状态、链接和焦点状态使用品牌 token

### Requirement: 单体部署必须提供固定服务入口

系统 MUST 以 Docker Compose 启动 React 前端、Go 后端和 MySQL；phpMyAdmin 作为本地管理服务提供。默认端口 MUST 为前端 `3000`、后端 `8080`、MySQL `3306`、phpMyAdmin `8081`，并允许通过环境变量覆盖。

#### Scenario: 冷启动服务

- **WHEN** 用户执行项目规定的 Docker Compose 启动命令
- **THEN** MySQL 通过健康检查后运行迁移
- **AND** 后端健康接口、前端入口和 phpMyAdmin 管理入口可访问

#### Scenario: 后端不可用

- **WHEN** 前端无法连接后端
- **THEN** 页面显示可理解的服务不可用状态
- **AND** 不暴露数据库密码、堆栈或 provider 凭据

### Requirement: 工作台必须覆盖关键用户状态

前端 MUST 为登录、行情搜索、收藏、下单、持仓和订单历史提供加载、空态、错误和成功状态，并在桌面和窄屏视口保持可用且无横向溢出。

#### Scenario: 窄屏打开工作台

- **WHEN** 用户在窄屏视口打开行情工作台
- **THEN** 导航、行情详情和订单入口按单列或抽屉布局呈现
- **AND** 关键控件可见、可点击且不被遮挡

