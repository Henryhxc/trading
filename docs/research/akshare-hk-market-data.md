# AkShare 港股行情接入研究

访问日期：2026-10-24（Asia/Hong_Kong）

## 结论

AkShare `1.19.3` 可作为本项目的**非官方、延迟行情采集依赖**，但不能被描述成 HKEX 官方或实时行情源：

- 历史日线使用 `ak.stock_hk_hist(symbol, period, start_date, end_date, adjust)`；本项目应固定 `period="daily"`，且在现有日线表没有复权维度时固定 `adjust=""`（不复权）。官方文档称它按日更新，返回单一证券的历史行情。[官方文档：历史行情数据-东财](https://akshare.akfamily.xyz/data/stock/stock.html#id76)
- 持续行情采集使用 `ak.stock_hk_spot_em()`；若产品范围被明确限制为港股主板，可改用 `ak.stock_hk_main_board_spot_em()` 以缩小全量分页范围。两个函数都不接收证券代码，都会抓取整个目标市场；官方文档明确两者的数据有 **15 分钟延迟**。[官方文档：所有港股现货](https://akshare.akfamily.xyz/data/stock/stock.html#id71)；[官方文档：港股主板现货](https://akshare.akfamily.xyz/data/stock/stock.html#id72)
- 这两个港股接口在当前发行版中是同步 HTTP 拉取函数，不提供港股 webhook、订阅回调或 WebSocket 推送合同。项目要求的“签名 webhook 写 Go/MySQL、提交后经认证 WebSocket 推送”必须由本项目实现；WebSocket 可以在数据库提交后立即通知，但其中的 AkShare 报价仍是约 15 分钟延迟数据，不能因此标为 `live`。证据见固定版本源码：[spot 实现](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1225-L1392)、[history 实现](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1395-L1464) 和[发行依赖清单](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/pyproject.toml#L5-L58)。
- AkShare 调用这些公开接口不要求用户 API key；函数签名没有 key/token 参数，源码直接请求东方财富域名。它仍要求 Python 运行环境和对相应外部域名的网络访问。该结论不等于上游授予生产、商业使用或再分发许可。
- 进入本项目的数据必须写 `source="akshare"`，并把上游标为 `upstream="eastmoney"`（若 DTO/审计记录保留该字段）。`data_status` 对 AkShare 只允许 `unofficial`、`delayed`、`stale`，绝不能写 `live`。成功采集且仍在新鲜度窗口内的 spot 数据应为 `delayed`；超时未更新后转为 `stale`；历史回补可标为 `unofficial`。

## 1. 固定版本与运行要求

本研究固定并检查了 `akshare==1.19.3`：

| 项目 | 已核实事实 | 第一方证据 |
| --- | --- | --- |
| 当前发行版 | `1.19.3`，Git tag `release-v1.19.3`，commit `2ea08ff44512a9b0c008cd4ff77463b940959c7e` | [GitHub release](https://github.com/akfamily/akshare/releases/tag/release-v1.19.3)、[GitHub tag commit](https://github.com/akfamily/akshare/commit/2ea08ff44512a9b0c008cd4ff77463b940959c7e)、[PyPI 1.19.3](https://pypi.org/project/akshare/1.19.3/) |
| 发布时间 | PyPI 文件上传于 2026-10-23；GitHub release 同日发布 | [PyPI 版本 JSON](https://pypi.org/pypi/akshare/1.19.3/json)、[GitHub release](https://github.com/akfamily/akshare/releases/tag/release-v1.19.3) |
| Python | `requires-python = ">=3.11"`；发行 classifiers 列出 Python 3.11–3.15 | [固定版本 pyproject.toml](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/pyproject.toml#L5-L40)、[PyPI 元数据](https://pypi.org/pypi/akshare/1.19.3/json) |
| 软件许可 | AkShare 软件为 MIT License | [固定版本 LICENSE](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/LICENSE)、[pyproject.toml](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/pyproject.toml#L5-L11) |
| wheel SHA-256 | `a93cbe67acf2042e922bc05a3a817ee51406e21855b8da98e23fa60b18325b12` | [PyPI 版本 JSON](https://pypi.org/pypi/akshare/1.19.3/json) |
| sdist SHA-256 | `c63a17211ea018b9921befd07da1b2c548026cf2a6d7b7b1e7d01a5e0f41ad00` | [PyPI 版本 JSON](https://pypi.org/pypi/akshare/1.19.3/json) |

实施时应使用独立 Python 3.11+ 调度器环境，精确固定 `akshare==1.19.3` 并校验发行物 hash；不要把 Python/AkShare 嵌入 Go 业务层。升级 AkShare 必须先重新跑字段合同测试和一组固定证券的 smoke test，因为项目 README 明确声明接口可能因不可控因素被移除。[固定版本 README Statement](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/README.md#L184-L192)

## 2. 历史日线接口

### 2.1 函数名与签名

当前发行版源码签名为：

```python
stock_hk_hist(
    symbol: str = "00593",
    period: str = "daily",
    start_date: str = "19700101",
    end_date: str = "22220101",
    adjust: str = "",
) -> pandas.DataFrame
```

参数合同：

| 参数 | 官方取值/格式 | 本项目选择 |
| --- | --- | --- |
| `symbol` | 港股代码，例如 `00593` | adapter 输出 AkShare 所需的五字符、前导零代码，例如 `00700` |
| `period` | `daily`、`weekly`、`monthly` | 固定 `daily` |
| `start_date` | `YYYYMMDD` | 导入窗口开始日 |
| `end_date` | `YYYYMMDD` | 导入窗口结束日，包含边界由 pandas 日期切片行为决定 |
| `adjust` | `""` 不复权、`qfq` 前复权、`hfq` 后复权 | 现有表固定 `""`；不得把不同复权口径写进同一 `(stock_id, trade_date)` 唯一键 |

第一方依据：[官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id76)；[固定版本函数源码与参数映射](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1395-L1429)。

一个容易忽略的实现事实是：`start_date` / `end_date` 没有被传给东方财富请求。源码每次用固定的 `end=20500000`、`lmt=1000000` 拉取完整历史，然后在 pandas 中按日期切片。因此缩小日期窗口不会等比例降低上游传输量，运行时不能按大量证券高频重复调用 history。[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1420-L1451)

### 2.2 返回列、单位与实际类型

正常非空返回固定 11 列：

| AkShare 列 | 文档类型 | 单位 | 项目用途 |
| --- | --- | --- | --- |
| `日期` | `object` | 交易日 | `trade_date` |
| `开盘` | `float64` | HKD | `open` |
| `收盘` | `float64` | HKD | `close` |
| `最高` | `float64` | HKD | `high` |
| `最低` | `float64` | HKD | `low` |
| `成交量` | 文档通常为 `int64`（后复权示例写 `int32`） | 股 | `volume` |
| `成交额` | `float64` | HKD | `turnover` |
| `振幅` | `float64` | % | 当前表不保存 |
| `涨跌幅` | `float64` | % | 当前日线表不保存 |
| `涨跌额` | `float64` | HKD | 当前日线表不保存 |
| `换手率` | `float64` | % | 当前表不保存 |

第一方依据：[官方输出合同](https://akshare.akfamily.xyz/data/stock/stock.html#id76)；[固定版本列名与转换源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1430-L1464)。

类型不能只按文档名盲信：源码用 `pd.to_numeric(..., errors="coerce")`，无效值会变成 `NaN`，实际 pandas dtype 也可能因缺失值从整数提升为浮点；日期用 `pd.to_datetime(..., errors="coerce").dt.date`，无效日期可变成空值。导入器必须逐行验证并使用十进制定点转换，不能把 Python 浮点直接当数据库金额合同。

可能出现的缺列/失败形态：

- 上游返回的 `klines` 为空，或日期切片后为空时，函数明确返回**没有预设列的空 `DataFrame`**。导入器必须把它当“无数据”，不能继续按列取值。[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1430-L1451)
- 若 `data` 为 `null`、`klines` 缺失、JSON 非预期或每行字段数改变，当前实现会在字典索引、JSON 解析或固定列赋名处抛异常；它没有 schema fallback。[同一源码定位](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1430-L1447)
- history 使用一次 `requests.get(..., timeout=15)`，没有该函数内重试，也没有显式 `raise_for_status()`；网络、HTTP、JSON 和字段错误都应由调度器捕获并按任务失败处理。[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1430-L1432)
- 不支持的 `period` 或 `adjust` 会在字典索引处失败；调用前必须白名单校验，不能把任意字符串传入。[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1418-L1426)

## 3. 证券代码格式

AkShare 官方文档和源码示例使用五字符、前导零的港股代码，例如 `00593`；调用时不带 `HK`、交易所后缀或东方财富的 `116.` 前缀。源码自己组装 `secid = "116." + symbol`，但不做代码长度或数字校验。[官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id76)；[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1395-L1426)

本项目文档称内部代码“六位”，但给出的 `00700` 实际是五字符；当前 OpenAPI 又接受 5–6 位，数据库 `stocks.symbol` 是 `CHAR(6)`，现有数据也使用 `00700`。因此实施前必须把该文字歧义归一到一个可测试合同。就现有磁盘事实和 AkShare 输入而言，安全映射是：

1. 从 AkShare spot 的 `代码` 接受且只接受五位十进制字符串，保留前导零，例如 `00700`。
2. 写入现有 `stocks.symbol` 时保留 `00700`，不要再次左补零变成 `000700`，否则会与现有标的键不一致。
3. 若 API 暂时仍接受六位输入，只允许六位值的首字符为 `0`，移除恰好一个首零后再映射到 AkShare；其他六位值应拒绝，不能静默截断。
4. webhook DTO 和 Go 入站校验必须再次验证代码白名单；不可只依赖 Python 调度器。

这里的 AkShare 外部格式是第一方事实；内部格式歧义来自项目当前合同，需由项目在 OpenAPI/设计/实现中统一，不应归咎于 AkShare。

## 4. 持续轮询所用现货接口

### 4.1 `stock_hk_spot_em()`

官方文档将其定义为“所有港股的实时行情数据”，同时明确说明**数据有 15 分钟延时**，并称一次返回最近交易日所有港股。函数没有输入参数。[官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id71)

返回 12 列：`序号`、`代码`、`名称`、`最新价`、`涨跌额`、`涨跌幅`、`今开`、`最高`、`最低`、`昨收`、`成交量`、`成交额`。价格/成交额单位为 HKD，成交量单位为股，涨跌幅单位为百分比。源码将各数值列用 `errors="coerce"` 转为数值。[官方输出合同](https://akshare.akfamily.xyz/data/stock/stock.html#id71)；[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1225-L1307)

该返回没有以下字段：上游报价时间、买一、卖一、每手股数、交易状态、停牌/退市状态。因此：

- `bid` / `ask` 必须为 `NULL` 或保留上一次有权威来源的值，绝不能用 `last_price` 伪造；
- `quote_as_of` 只能记录本项目的 `collected_at`，并在字段语义中明确它是采集时间，不是交易所事件时间；
- 不得根据价格为空或成交量为零擅自推断 `suspended` / `delisted`；
- `lot_size` 不能从此接口得到，应由独立静态资料维护。

### 4.2 `stock_hk_main_board_spot_em()`

该函数同样没有输入参数、字段相同、延迟 15 分钟，但范围只覆盖港股主板。[官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id72)；[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1310-L1392)

若 MVP 明确只支持主板，它比全港股函数更适合轮询；若业务需要 ETF、REIT、窝轮、牛熊证等非主板筛选范围，就不能用主板函数替代全市场函数。选择必须来自产品证券范围，不能为了减少请求而悄悄缩窄产品范围。

### 4.3 分页、重试与频率风险

两个 spot 函数都把每页大小设为 100，并通过 `fetch_paginated_data` 拉完所有页。该 helper：

- 首次和每页请求的默认 timeout 为 15 秒；
- HTTP 请求最多尝试 3 次，采用指数退避和随机抖动；
- 每个后续分页之间额外随机等待 0.5–1.5 秒；
- 假定响应存在 `data.diff`、`data.total` 和排序字段 `f3`；字段缺失会失败。

第一方依据：[spot 参数](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1232-L1246)；[分页 helper](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/utils/func.py#L18-L60)；[HTTP retry helper](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/utils/request.py#L15-L64)。

官方文档没有为这两个函数公布一个可依赖的数字频率配额。因此不能声称“每分钟 N 次安全”，也不能把内置重试当成上游 SLA。由于官方已经说明数据延迟 15 分钟，默认轮询间隔不应短于 15 分钟；交易时段使用单实例调度、全局锁、随机抖动和指数退避，休市时停止或大幅降频。连续调度的含义是按计划重复采集，不是实时流。

## 5. API key、数据来源与许可边界

### API key

`stock_hk_hist`、`stock_hk_spot_em`、`stock_hk_main_board_spot_em` 的官方签名没有 API key 参数；源码直接向东方财富公开 HTTP 端点发 GET 请求，也没有读取用户凭据环境变量。因此部署不需要 AkShare API key。第一方依据：[history 源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1395-L1432)、[spot 源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1225-L1246)。

源码中的固定 `ut` 请求参数是 AkShare 随代码发布的上游请求常量，不是项目用户申请或应写入 secrets 的 API key。它仍可能被上游更改，不能视为稳定授权凭据。

### 数据来源与许可

AkShare 文档与 docstring 明确把这三项接口标为东方财富数据，源码请求 `push2.eastmoney.com` / `push2his.eastmoney.com`。[官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id71)；[固定版本源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1225-L1246)

MIT License 覆盖 AkShare 软件代码，不自动授予东方财富行情数据的生产使用、商业使用或再分发权。AkShare 自己的 README 还声明数据只用于学术研究、仅供参考、不构成投资建议，并提醒数据风险及接口可能移除。[固定版本 LICENSE](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/LICENSE)；[固定版本 README Statement](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/README.md#L184-L192)

因此本项目只能把 AkShare 作为原型/模拟交易的非官方来源；上线前仍须单独审查上游服务条款和行情许可。没有 API key 不代表没有许可约束。

## 6. webhook 与 WebSocket 核实

在固定发行版 `1.19.3` 中，这三个港股函数的实现路径均是 Python 调用方主动执行、同步返回 `DataFrame`：history 使用 `requests.get`，spot 使用同步分页 HTTP helper。函数没有 callback、subscription、webhook URL、WebSocket connection 或 async iterator 参数。[spot 源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1225-L1392)；[history 源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1395-L1464)

固定版本的运行依赖清单也没有 WebSocket client/server 依赖；源码包中 `websocket` 字样只存在于旧版本 changelog 文本（曾涉及其他接口），并不构成当前港股接口的推送能力。[固定版本依赖清单](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/pyproject.toml#L41-L58)；[历史 changelog 示例](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/__init__.py#L810-L842)

所以项目边界应是：

```text
AkShare/Eastmoney（约 15 分钟延迟）
  -> Python scheduler 主动轮询与标准化
  -> 出站 HMAC 签名 HTTPS webhook（仅内网 Go ingest endpoint）
  -> Go 二次校验并在 MySQL 事务中幂等 upsert
  -> COMMIT 成功后发布内部事件
  -> Go 的认证 WebSocket 向已登录客户端推送
```

HMAC、重放窗口、幂等键、Go webhook handler、事务提交后事件和认证 WebSocket 都是本项目自己的安全/交付机制，不是 AkShare 功能。调度器无需开放公网入站端口；它只需访问上游并向 Go 发出站请求。

## 7. 映射到 `stocks` 与 `stock_daily_quotes`

### 7.1 spot -> `stocks`

| AkShare / 调度元数据 | `stocks` | 规则 |
| --- | --- | --- |
| `代码` | `symbol` | 严格代码映射，当前保存 `00700` 形式 |
| `名称` | `name` | 非空、长度校验；名称变化允许更新 |
| 常量 | `market` | `HKEX`；这不表示数据源是 HKEX 官方 feed |
| 无 | `lot_size` | 不从 AkShare spot 推断；保留静态资料值 |
| 无 | `status` | 不从价格/成交量推断；保留已有值 |
| `最新价` | `last_price` | HKD，有限、非负、定点化 |
| 无 | `bid`, `ask` | `NULL` 或不更新；禁止用最新价伪造 |
| `涨跌额` | `change` | HKD，允许负数 |
| `涨跌幅` | `change_percent` | 百分比数值，不再乘/除 100，接口与 UI 应统一单位 |
| `成交量` | `volume` | 股，有限且非负 |
| `成交额` | `turnover` | HKD，有限且非负 |
| scheduler `collected_at` | `quote_as_of` | UTC 采集时间；不是交易所事件时间 |
| 常量 | `quote_source` | 必须为 `akshare` |
| scheduler 计算 | `data_status` | 成功且新鲜为 `delayed`；超阈值为 `stale`；不能为 `live` |
| scheduler `collected_at` | `updated_at` | DB 事务写入时间/采集时间按项目统一语义 |

spot 同时返回当日 `今开/最高/最低/昨收`，但当前 `stocks` 没有这些字段；它们不应被错误塞进日线表，因为盘中当日 K 线未必收盘且接口没有事件时间。需要盘中 OHLC 时应另定合同。

### 7.2 history -> `stock_daily_quotes`

| AkShare | `stock_daily_quotes` | 规则 |
| --- | --- | --- |
| 当前 symbol 对应的 `stocks.id` | `stock_id` | 先按规范化 symbol 查/建 stock，再使用 FK |
| `日期` | `trade_date` | 合法日期，幂等键 `(stock_id, trade_date)` |
| `开盘` | `open` | HKD 定点数 |
| `最高` | `high` | HKD 定点数 |
| `最低` | `low` | HKD 定点数 |
| `收盘` | `close` | HKD 定点数 |
| `成交量` | `volume` | 股；缺失保留 `NULL`，不伪造 0 |
| `成交额` | `turnover` | HKD；缺失保留 `NULL`，不伪造 0 |

写入前至少校验：日期在请求窗口内；OHLC 均非空、有限、非负；`low <= open/close <= high`；volume/turnover 若存在则非负；同一 payload 不含重复证券日；复权口径固定为不复权。upsert 只在整批验证成功后进入事务，重复导入相同证券日必须幂等。

当前 [`stock_daily_quotes` 迁移](../../db/migrations/002_stocks_daily_quotes_watchlists.sql) 没有 `source`、`data_status`、`collected_at` 或 `adjustment` 列。因此它**无法单独证明每条日线来自 AkShare，也无法区分复权口径**。若验收要求对每条日线保证 `source="akshare"` 和非官方状态，实施必须二选一：

1. 给日线表增加 `source`、`data_status`、`collected_at`、`adjustment`，并把约束/唯一键与选定口径匹配；或
2. 增加不可变的 import batch/provenance 表，让每条日线可追溯到一个 `source="akshare"` 的采集批次。

只在 `stocks.quote_source` 写 `akshare` 不能证明历史表每条记录的来源。现阶段最小、最不含糊的方案是历史表固定 `adjustment="none"`，并保存行级 provenance。

### 7.3 状态规则

对 `source="akshare"` 的入站 webhook，Go 端必须拒绝 `live`、`simulated` 或未知状态，即使 Python 误发。建议状态机：

- spot 成功且 `now - collected_at` 未超过项目的新鲜度阈值：`delayed`；
- 未能按计划更新、最后成功值超过阈值：`stale`；
- 历史回补/导入记录：`unofficial`；
- 网络失败、限流、空数据或 schema 失败：不写伪造行情、不广播空报价；保留最后成功值并在超过阈值后标为 `stale`，采集任务错误另行记录。

阈值是项目策略，不是 AkShare 的上游事实；应至少覆盖 15 分钟固有延迟、轮询间隔和合理 grace period，并在配置/测试中固定。

## 8. 风险与守护措施

| 风险 | 第一方证据 | 必须的项目行为 |
| --- | --- | --- |
| 数据先天延迟 | 文档明确 spot 延迟 15 分钟 | 只标 `delayed/stale`，UI 展示来源与时间，模拟订单不得宣称按实时价成交 |
| 全市场分页成本 | spot 无 symbol 参数、每页 100、拉取所有页 | 单实例、至少 15 分钟间隔、抖动、主板范围只有在产品允许时才采用 |
| 无公开数字频控/SLA | 接口文档没有公布额度；README 提醒接口可移除 | 自有退避/熔断/监控，不承诺稳定实时服务 |
| history 无内置 retry | 单次 `requests.get` | 调度层有限重试、指数退避；失败批次不写 DB |
| 空数据 | history 可返回无列空 DataFrame | 视为无数据或失败，绝不转为零值行情 |
| 字段/响应结构变化 | 源码硬编码 JSON 路径和列数 | exact schema contract test；缺列/多列/非数值立即隔离并转 stale |
| 无上游事件时间 | spot 输出没有 timestamp | 保存 `collected_at` 并明确语义；不伪造 exchange timestamp |
| 浮点/NaN | 源码 `to_numeric(errors="coerce")` | `Decimal`/定点校验，拒绝 NaN/Inf，金额按 DB 精度量化 |
| 许可不等同于免 key | MIT 仅覆盖软件；README 限定研究用途 | 上线/分发前完成上游条款与许可审查 |
| WebSocket 造成“实时”误解 | AkShare 无港股推送合同且 spot 延迟 15 分钟 | WebSocket 文案仅表示“提交后及时通知”，数据状态始终随上游延迟展示 |

## 9. 一次性导入与运行时调度边界

### 一次性历史回补

- 使用 `stock_hk_hist(..., period="daily", adjust="")`，按证券串行或低并发执行；日期范围用于本地筛选，不会缩小上游完整历史请求。
- 先生成并校验完整批次，再以幂等 upsert 写入 MySQL；失败证券可单独重试，但不能让部分无效行混入成功批次。
- 写完后由现有 Go `/api/v1/markets/candles/*` 从 MySQL 提供，不让 React 直接访问 AkShare/东方财富。
- 这是离线导入，不要求 WebSocket；若要通知 UI，可只发布“历史已更新”事件，不把几千根日线逐根广播。

### 持续延迟行情调度

- Python scheduler 在交易时段每 15 分钟或更慢调用一次 `stock_hk_spot_em()`，标准化后只发送变化/关注范围内的记录到内部签名 webhook。
- Go 必须在入站处验证 HMAC、时间戳/重放窗口、payload version、幂等键、代码、数值、`source` 和 `data_status`；验证通过后在事务中更新 `stocks`。
- 只有 DB commit 成功后才向认证 WebSocket 发布；rollback、重复幂等请求或无变化不应产生错误广播。
- 调度器不开放公网业务端口，不持有用户 session，不直接写用户订单/资金/持仓表，也不触发真实订单。
- history 只用于初次回补和每日收盘后补洞，不应当作盘中 spot 的替代轮询接口。

这两个模式可以共用标准化与签名客户端代码，但必须有不同 job、超时、幂等键和状态语义。AkShare 仅负责拉取并返回 DataFrame；调度、webhook、MySQL 事务与 WebSocket 都属于本项目。

## 第一方来源索引

- [AkShare 股票数据官方文档：港股 spot/history](https://akshare.akfamily.xyz/data/stock/stock.html#id70)
- [AkShare 1.19.3 GitHub release](https://github.com/akfamily/akshare/releases/tag/release-v1.19.3)
- [AkShare 1.19.3 PyPI](https://pypi.org/project/akshare/1.19.3/)
- [PyPI 1.19.3 JSON 元数据](https://pypi.org/pypi/akshare/1.19.3/json)
- [固定 commit：港股 spot/history 源码](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1225-L1464)
- [固定 commit：分页 helper](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/utils/func.py#L18-L60)
- [固定 commit：HTTP retry helper](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/utils/request.py#L15-L64)
- [固定 commit：Python/依赖/许可元数据](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/pyproject.toml#L5-L58)
- [固定 commit：README Statement](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/README.md#L184-L192)
- [固定 commit：MIT License](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/LICENSE)
