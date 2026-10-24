# 多周期 K 线合同（2026-10-24 用户追加范围）

用户已要求安装开源 K 线、默认过去半年历史、分钟/日/周/月/年切换及持续更新。本合同替代旧文档中“仅日线、不含分钟线”的限制；不扩大真实交易或杠杆权限。

## 体验和数据

- 使用 TradingView Lightweight Charts（Apache-2.0），保留 NOTICE/归属链接。显示真实 OHLC 蜡烛、成交量、十字光标、缩放与拖动；上涨绿、下跌红，同时显示带符号涨跌百分比。
- 周期为 `1m/5m/15m/30m/60m/1d/1w/1M/1y`，默认 `1d`，请求过去六个日历月。各分钟周期使用上游原生分钟数据；周/月/年由不复权日线按香港日历聚合。聚合包含区间内实际已有记录，不能补造缺失交易日。
- 六个月是请求窗口，不是上游覆盖承诺。界面显示实际首末时间、条数、来源、采集时间、状态；分钟历史较短或请求失败须明确提示。空序列不得自动换成日线或 mock。
- React 仅调用内部 API。图表每 60 秒重新读取当前周期，手动刷新立即读取；切换证券/周期后旧请求结果不能覆盖当前选择。数据出错保留明确错误和重试，不能隐藏为成功。

## API 与存储（先于应用实现）

- `GET /api/v1/markets/candles/{symbol}?interval=...&from=YYYY-MM-DD&to=YYYY-MM-DD&limit=...`。无 interval 保持旧日线兼容；显式 interval 最大 5000 根，时间窗口不超过 366 天。响应含 `symbol, interval, items, source, data_status, as_of`；items 保留 OHLCV/turnover，分钟额外有 RFC3339 `timestamp`，日线保留 `trade_date`。
- 新增 migration `004_intraday_candles.sql`：`stock_intraday_quotes`，唯一键 `(stock_id, interval_code, bar_time)`，UTC `bar_time`，原生周期、OHLCV、来源、状态和采集时间。不改用户/资金/订单数据。原八表限制因分钟存储新增一表而被本次用户需求修订。
- webhook v1 的 candle 增加可选 `interval` 与 `timestamp`，缺省 interval 为 `1d`；分钟只接受上述原生分钟周期。同批重复证券/周期/时间拒绝。更新按 collected_at 条件 upsert，旧采集不得覆盖新记录。报价/日线/分钟仍在一个事务提交，成功后再发布 `candles.updated`。
- 每个 webhook 最多 400 条历史记录，repository 按证券/周期/来源/状态/采集时间分组，用绑定参数的多值 INSERT 批量写入；每事务每证券只解析一次 metadata。修复逐条写入导致超过 15 秒发送预算、后续周期只导入一部分的问题。多组任一失败全部回滚，事件按证券/周期去重，精确重放不广播。隔离库性能用例要求 400 条插入及幂等重放在 10 秒截止时间内完成。
- rollback 仅删除 `stock_intraday_quotes`，分钟历史会丢失，可从上游重新采集；原日线及业务数据不变。默认保留导入的分钟记录，未来清理需独立显式操作。

## 采集与失败分支

- Python AkShare 辅助容器无宿主端口。默认配置标的 `00700,00005,09988`，日线自动计算六个月回补窗口；分钟调用 `stock_hk_hist_min_em`。spot、日线、各分钟周期独立尝试，spot 失败不能阻塞可用历史导入。
- 日线/分钟每 15 分钟轮询，签名 webhook 分块发送，有限重试和脱敏日志；上游失败不得写合成行情。分钟最近数据范围受上游限制，不能承诺六个月分钟历史。
- 数据只作开发/模拟用途，保留实际 `source`，历史 `unofficial`，分钟可动态变 `stale`，永不标记 HKEX 官方 `live`。mock mode 的日线始终显示 simulated，分钟无数据明确 unavailable。
- AkShare 各周期失败时，可由同一 Python 辅助服务使用独立公开源适配器降级：腾讯不复权日线标记 `source=tencent`，东方财富原生分钟标记 `source=eastmoney`；不得冒充 AkShare。公开请求设置超时、有限重试，按标的与周期独立降级。
- 腾讯日线和 Yahoo 原生分钟未提供成交额，签名输入与 REST `turnover` 使用 `null`，migration `005_nullable_daily_turnover.sql` 允许日线及分钟成交额为空。周/月/年任一组成日缺失成交额时，聚合成交额也为 null；不填零、不估算。其他来源成交额仍要求上游真实值。回滚 005 在存在 null 时会失败，必须先补齐真实成交额，禁止以零覆盖缺失值。
- 独立 Yahoo Finance chart 适配器标记 `source=yahoo`，只接受原生 OHLCV 与 Unix 时间戳；分钟成交额未提供保存 null。1m 请求最近 5 个交易日；5m/15m/30m 受最多约 60 日限制，60m 尝试半年窗口，界面只承诺实际返回范围。OHLC 不完整的记录直接跳过，禁止插值。
- `HISTORY_PROVIDER=akshare` 保留 AkShare 历史优先及公开源降级；`public` 直接使用腾讯日线、Yahoo 原生分钟，Yahoo 失败才尝试东方财富原生分钟。当前 Compose 默认 `public`，避免已确认的 AkShare/Eastmoney 断连阻塞历史。spot 报价仍独立尝试 AkShare；历史成功不代表 spot 成功，旧报价保持 stale。所有适配器仍在同一无业务端口的辅助容器，Go 来源字段不会被改名。

## 文件与验证

- Go domain/service/repository/httpapi 分别管理合同、聚合/状态、SQL 和参数/错误映射；Python 采集、规范化和调度分文件。
- React `features/market/CandlestickChart.tsx`、`MarketChart.tsx`、`chart-data.ts` 管理渲染、查询生命周期和数据转换；`shared/api/market.ts` 复用共享合同。
- RED→GREEN：Go 验证周期/日期/OHLC 聚合及分钟不伪装；MySQL 临时测试库验证唯一键/乱序/upsert/回滚；Python 验证半年前日期、时区和独立失败；前端行为测试验证周期、请求竞态和错误/空态。
- 实验命令：`go test ./... -count=1`、`go vet ./...`、`python3 -m unittest discover -s services/akshare/tests`、`npm run test:frontend`、`npm run build`；浏览器检查 1440x900、1280x800、390x844。证据写入 change 的 `results/candlestick/`。
