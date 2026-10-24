## Purpose

为港股工作台提供可追溯的标的搜索、标准化报价和日线数据，并明确区分模拟、非官方、延迟、过期和授权实时数据。

## ADDED Requirements

### Requirement: 行情必须返回统一字段和来源状态

系统 MUST 将外部行情转换为统一字段：`symbol`、`display_name`、`last_price`、`bid`、`ask`、`change`、`change_percent`、`volume`、`turnover`、`as_of`、`source` 和 `data_status`。

#### Scenario: 返回有效行情

- **WHEN** provider 返回完整且未超过允许时限的报价
- **THEN** 系统返回统一报价字段
- **AND** 返回来源和数据时间

#### Scenario: 免费公开源返回行情

- **WHEN** 报价来自腾讯、新浪或其他未授权的免费公开源
- **THEN** `data_status` MUST 是 `unofficial`、`delayed` 或 `stale`
- **AND** 系统 MUST NOT 标记为 `live`

### Requirement: 用户可以搜索港股并查看日线

系统 MUST 支持按代码或名称搜索港股，并支持查看标的最新报价和日线 OHLCV；内部代码统一为 `00700` 形式的五字符前导零代码，兼容输入和外部代码映射由 provider adapter 负责。

#### Scenario: 搜索和查看标的

- **WHEN** 用户输入港股代码或名称并选择结果
- **THEN** 系统显示标的名称、代码、最新价、涨跌、成交量、成交额、数据时间和状态
- **AND** 用户可以打开该标的日线图

#### Scenario: 展示 AkShare 同步的历史日线

- **WHEN** 用户查看已导入 AkShare 历史数据的标的日线图
- **THEN** 系统显示日线 OHLCV，并标明来源为 AkShare、数据状态为 `unofficial`
- **AND** 这些历史数据仅供开发/模拟用途，不得被称为 HKEX 官方实时行情或用于真实交易

#### Scenario: 找不到标的

- **WHEN** 搜索没有匹配标的
- **THEN** 系统返回空结果状态和可理解提示

### Requirement: 行情故障必须可解释并安全降级

系统 MUST 处理 provider 超时、限流、断线、字段缺失、时间戳过期和序列错误，并将状态降级为 `stale`、`unavailable` 或 `simulated`。

#### Scenario: 授权源不可用

- **WHEN** 配置的行情源超时或拒绝请求
- **THEN** 系统按配置顺序尝试可用的免费源或 mock provider
- **AND** 响应包含实际来源和数据状态

### Requirement: AkShare 调度必须持续、非官方且可安全降级

系统 MUST 将 AkShare 作为与 Go 业务层隔离的持续轮询调度器，用于开发/模拟场景的 spot 更新和历史日线回补；不得把 Python 库嵌入 Go 业务层，不得把 AkShare 称为 HKEX 官方实时源或将其数据标记为 `live`。

#### Scenario: 调度并写入 AkShare 行情

- **WHEN** 调度器从 AkShare 获得配置标的的 spot 或日线数据
- **THEN** 调度器 MUST 将代码、报价或 OHLCV、来源和采集时间规范化后发送到内部 ingestion 接口
- **AND** 数据保留 AkShare 非官方来源标记，重复导入同一证券和交易日必须幂等
- **AND** spot MUST 标记为 `delayed`，历史日线 MUST 保存 `source=akshare`、`data_status=unofficial`、采集时间和不复权口径
- **AND** 上游不可用、字段变化或数据校验失败时必须留下可解释的脱敏记录，不得静默写入不完整数据

#### Scenario: 调度停止后行情动态过期

- **WHEN** AkShare 更新停止并超过配置的过期阈值
- **THEN** 报价读取和推送状态 MUST 动态降级为 `stale` 或 `unavailable`
- **AND** 已保存的旧状态不得让客户端继续误认为数据及时

#### Scenario: 调度器失败并恢复

- **WHEN** AkShare 超时、限流或网络中断
- **THEN** 调度器 MUST 使用有上限的退避继续重试，并保持上次已提交数据可通过 REST 读取
- **AND** 恢复后 MUST 继续更新，不要求人工重启

#### Scenario: 调度频率尊重上游延迟和全量分页成本

- **WHEN** 调度器配置 AkShare spot 周期
- **THEN** 默认和最短周期 MUST 为 15 分钟，且每个周期只执行单实例全市场抓取后筛选配置标的
- **AND** WebSocket 的提交后及时推送不得被描述为上游实时行情

#### Scenario: 行情过期时提交模拟订单

- **WHEN** 订单使用的行情超过允许时限
- **THEN** 系统拒绝依赖该价格的模拟撮合
- **AND** 返回行情过期原因

### Requirement: 内部行情 webhook 必须验签、幂等并原子写库

系统 MUST 在写库前验证 webhook 的 HMAC、时间戳、delivery id 和 payload，并在单一事务中更新最新报价和日线；任何失败都不得产生部分写入。

#### Scenario: 有效签名成功写入

- **WHEN** 内部 webhook 收到时间窗口内、签名有效且字段合法的 AkShare 批次
- **THEN** 系统 MUST 事务性更新 `stocks` 和适用的 `stock_daily_quotes` 数据
- **AND** 只有提交成功后才可对外发布对应更新事件

#### Scenario: 无效或过期签名被拒绝

- **WHEN** webhook 签名错误、时间戳过期、delivery id 缺失或 payload 不合法
- **THEN** 系统 MUST 在写库前拒绝请求并返回可解释但不泄密的错误
- **AND** 不得发送 WebSocket 更新

#### Scenario: 重放和乱序不会覆盖新数据

- **WHEN** 相同 delivery id 在验收窗口内被重放，或较旧报价在较新报价后到达
- **THEN** 系统 MUST 拒绝重放或将其作为无副作用的幂等请求处理
- **AND** 较旧数据不得覆盖已提交的较新报价，也不得重复广播

#### Scenario: 事务失败不广播

- **WHEN** 批次中任一必需写入导致事务回滚
- **THEN** 系统 MUST 保持该批次不产生部分数据
- **AND** 系统 MUST NOT 广播该批次的更新

### Requirement: 已认证客户端必须及时收到提交后的行情更新

系统 MUST 在现有后端端口提供 session 认证的 WebSocket 行情流，并保留 REST 快照作为首次加载和断线恢复路径。

#### Scenario: 登录用户收到提交后的更新

- **WHEN** 有有效 session 的客户端订阅行情流且 webhook 事务成功提交
- **THEN** 客户端 MUST 收到包含规范化行情、来源、数据状态和数据时间的更新
- **AND** 客户端随后通过 REST 读取时 MUST 能看到不早于该事件的已提交数据

#### Scenario: 未登录连接被拒绝

- **WHEN** 无有效 session 或来源不被允许的客户端尝试升级 WebSocket
- **THEN** 系统 MUST 拒绝连接且不得泄露行情流或认证细节

#### Scenario: 断线重连恢复快照

- **WHEN** WebSocket 断开后重新连接
- **THEN** 客户端 MUST 使用有上限的退避重连，并通过 REST 重新获取最新快照后继续处理事件
- **AND** 漏失的内存广播不得导致页面永久停留在旧报价
