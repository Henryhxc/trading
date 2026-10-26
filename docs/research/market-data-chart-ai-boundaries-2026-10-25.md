# 行情来源、K 线图表与开放模型边界

核查日期：2026-10-25（Asia/Hong_Kong）
范围：本项目当前源码与仓库文档；HKEX、AkShare、TradingView Lightweight Charts、DeepSeek 官方文档/仓库。此文用于说明原型能力，不构成数据许可或法律意见。

## 摘要

- HKEX 提供授权的 OMD 实时数据馈送和历史数据产品，但在官方资料中没有面向任意个人应用的免费通用行情 REST API。实时 feed、供应商/应用服务商许可、终端用户使用权及数据再分发是不同合同事项；准确费用需按产品、用户/终端数量、用途、展示/非展示及分发模式向 HKEX 或获授权供应商询价。公开行情网页不是可抓取的 API。
- 项目使用 AkShare 做非官方历史行情采集原型。仓库截至核查日标定 AkShare 1.19.3：该港股 history 接口有日/周/月数据参数，但当前项目导入合同实际固定为未复权日线；分钟线不能从该历史日线接口推导或冒充真实分钟数据。应用目前允许分钟、日、周、月、年周期；周/月/年由日线聚合，分钟周期只读取原生分钟表，缺数据时为空并解释，不用日线或合成数据替代。
- 前端固定依赖 `lightweight-charts@5.2.1`，其职责是把传入的 OHLCV 绘制成图表，不抓行情、不提供行情数据。上游项目采用 Apache-2.0，要求页面保留 TradingView attribution notice 和链接；使用该库不等于获得 TradingView 商业 Charting Library 或 HKEX 数据许可。
- DeepSeek 官方发布过可下载权重及代码的模型/研究仓库，但每个仓库/权重都必须逐项查看许可。DeepSeek-R1、DeepSeek-V3.2-Exp 仓库显示 MIT；“开放权重/代码”不能推断训练语料、完整训练配方、全部基础设施或服务能力均开放，更不能推断团队仅凭开源材料就能复刻同等模型。V3.2-Exp 自述为实验版本。

## 1. HKEX：接口、授权与费用

HKEX 的官方产品页面介绍的是面向市场数据订阅者的实时数据馈送，包括证券市场 OMD-C、衍生品市场 OMD-D，以及不同数据层级/产品；官方另有数据许可说明、终端用户许可、供应商许可与 ASP 要求页面。它们是 feed、许可和分发体系的说明，不是向所有开发者开放的免费 REST API 文档。

官方资料没有给出一个适用于所有个人原型的统一“HKEX API 价”。费用/许可取决于具体 feed/历史数据集、直接接入还是供应商、显示或非显示用途、终端用户数量、再分发/应用服务安排及适用协议。应先确认业务模型，再向 HKEX Market Data / 获授权供应商询问适用费率、连接与实施成本、最低承诺、用户报告义务、展示与再分发权、历史数据下载/存储/回放限制。不可只看软件/API 费用，亦不可把交易所行情授权当成券商下单权限。

HKEX 网站的“证券价格”页可供人浏览报价，但这不产生网页抓取授权，也不构成稳定、受支持的应用 API。HKEX Data Marketplace 提供历史数据产品目录，数据集、格式、授权和价格须按目录及订单确认；不能概括成免费。

本项目现阶段合适表述：开发和演示使用 mock 与标明来源/状态的非官方历史数据；若要真实实时行情，先取得适用 feed 许可或选用合同覆盖本产品用途的授权数据商，再实现适配器。不要把非官方 AkShare/Eastmoney 或页面抓取标成“HKEX 官方实时”。

一手来源：

- [HKEX Real-time Datafeeds](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Overview/Real_time-Datafeeds?sc_lang=en)（产品范围及 feed 概览）
- [HKEX OMD-C](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Securities-Market-OMD-C?sc_lang=en)（证券市场数据平台）
- [HKEX OMD-D](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Derivatives-Market-OMD-D?sc_lang=en)（衍生品市场数据平台）
- [HKEX Data Licensing](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Data-Licensing?sc_lang=en)（许可入口）
- [Market Data End-User Licence](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Data-Licensing/HKEX-IS/Market-Data-End_User-Licence?sc_lang=en)（终端用户许可）
- [Market Data Vendor Licence](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Data-Licensing/HKEX-IS/Market-Data-Vendor-Licence?sc_lang=en)（供应商许可）
- [Requirements for Market Data Application Service Provider](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Data-Licensing/HKEX-IS/Market-Data-Application-Service-Provider-Licence?sc_lang=en)（ASP 要求）
- [HKEX Data Marketplace](https://www.hkex.com.hk/Services/Market-Data-Services/Historical-Data-Services/HKEX-Data-Marketplace?sc_lang=en)（历史产品市场）
- [HKEX Securities Prices: Equities](https://www.hkex.com.hk/Market-Data/Securities-Prices/Equities?sc_lang=en)（浏览报价页面，不是 REST API 合同）

不确定性：本页不是 HKEX 报价单。许可费和实际采购要求会变动，最终金额与可用权利只能以 HKEX/供应商当期书面报价和合同为准。

## 2. 本项目 AkShare 数据和周期边界

仓库现存 [AkShare 港股接入研究](akshare-hk-market-data.md) 固定核查 `akshare==1.19.3`，并链接至版本源码。AkShare 的 `stock_hk_hist` 为 Eastmoney 历史行情适配，参数周期为 `daily`、`weekly`、`monthly`，没有分钟粒度；接口无需用户 API key 不等于数据有商业/再分发授权。其 spot 函数官方说明存在约 15 分钟延迟，不能标记 live。

要区分上游接口的能力和项目现在实际采用的流程：现有项目研究与日线表合同把导入固定为 `period="daily"`、`adjust=""`，所以不能说本项目已通过 AkShare 导入了分钟线、周线或月线原始数据。周/月/年可以对已取得日线做 OHLCV 聚合，但数据覆盖范围受底层日线历史覆盖影响。前端时间按钮/图表包本身不会填补缺失数据。

当前实现核实：

| 界面周期 | 当前内部合同 | 数据行为 | 能否称为 AkShare 原生周期 |
| --- | --- | --- | --- |
| 分钟：1m/5m/15m/30m/60m | API 与 `stock_intraday_quotes.interval_code` 支持 | 只读对应周期的原生分钟 K 线；没有记录即 `unavailable`/空态，禁止用日线或伪造分钟替代 | 否；`stock_hk_hist` 不提供分钟数据 |
| 日：1d | `stock_daily_quotes` | 从已导入日线读取 | 是上游支持粒度；本项目采集合同当前使用它 |
| 周：1w | API 支持 | 后端按香港时区周一为桶，对日线合并 OHLCV | 当前是项目聚合，不是本项目导入的 AkShare 周线 |
| 月：1M | API 支持 | 后端按日线合并为自然月 OHLCV | 当前是项目聚合，不是本项目导入的 AkShare 月线 |
| 年：1y | API 支持 | 后端按日线合并为自然年 OHLCV | 项目派生周期，不是该接口的原生 period |

依据是仓库代码：`apps/frontend/src/features/market/chart-data.ts` 声明九个可选区间；`apps/backend/internal/domain/market/quote.go` 白名单；`apps/backend/internal/service/market/candles.go` 对分钟请求调用 `NativeCandleProvider`，非分钟周期从日线读入并对周/月/年聚合；`db/migrations/004_intraday_candles.sql` 另设分钟数据表。导入来源不能仅凭“图表可切周期”推断。

AkShare 一手来源：

- [AkShare 港股历史行情官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id76)（参数、周期及输出字段）
- [AkShare 港股现货官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id71)（全港股现货、延迟说明）
- [AkShare 主板现货官方文档](https://akshare.akfamily.xyz/data/stock/stock.html#id72)（主板现货、延迟说明）
- [AkShare 1.19.3 release](https://github.com/akfamily/akshare/releases/tag/release-v1.19.3)
- [AkShare 1.19.3 `stock_hist_em.py`](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/akshare/stock_feature/stock_hist_em.py#L1225-L1464)（固定版本 spot/history 实现）
- [AkShare 1.19.3 LICENSE](https://github.com/akfamily/akshare/blob/2ea08ff44512a9b0c008cd4ff77463b940959c7e/LICENSE)（软件代码 MIT；不授予其访问的数据集许可）

不确定性：免费上游可能改字段、限流、断连或删除接口；接口文档的 `weekly/monthly` 支持不证明数据覆盖、更新 SLA 或项目已采集。分钟 K 线须另找明确提供港股原生分钟 bars 且许可允许该用途的来源，或继续展示诚实的无数据状态。

## 3. Lightweight Charts：图表代码与行情数据分离

本项目 `apps/frontend/package.json` 固定依赖 `lightweight-charts@5.2.1`，K 线组件把 API 返回的 candle rows 转为 candlestick series 与 volume histogram。Lightweight Charts 是 TradingView 发布的轻量级 Web 图表库，能绘制和交互式展示调用方提供的数据；它不提供行情、证券目录、HKEX 接入、历史回填或交易执行。

官方仓库当前 README 声明 Apache License 2.0，并明确要求在用户可访问的网站/移动应用页面放置 attribution notice 和指向 TradingView 的链接；README 提到可使用 `attributionLogo` 选项满足链接要求。应按实际锁定的 5.2.1 发行包核对随包 LICENSE/NOTICE 与当版文档，保留所需声明。Apache-2.0 许可的是图表软件，不是行情数据、TradingView 商业 Charting Library 或品牌背书。

来源：

- [项目锁定依赖](../../apps/frontend/package.json)（本地版本 `5.2.1`）
- [TradingView Lightweight Charts 官方 GitHub](https://github.com/tradingview/lightweight-charts)（功能定位、版本源码和许可）
- [上游 LICENSE](https://github.com/tradingview/lightweight-charts/blob/master/LICENSE)（Apache-2.0）
- [上游 NOTICE](https://github.com/tradingview/lightweight-charts/blob/master/NOTICE)（署名文本）
- [官方文档：attributionLogo](https://tradingview.github.io/lightweight-charts/docs/api/interfaces/LayoutOptions#attributionlogo)（图表中的来源链接设置）

版本注意：GitHub `master` 文档是滚动版本；项目使用 5.2.1，发布/再分发前应以 npm 锁文件和该版本 tarball 的 LICENSE/NOTICE 为准，而不是只用 master 替代固定版本检查。

## 4. DeepSeek：开放内容与不能推断的能力

截至核查日，DeepSeek 官方 GitHub 组织公开多个模型与研究仓库。两个代表性官方仓库：DeepSeek-R1 README 列出 R1/R1-Zero 以及基于 Qwen/Llama 的蒸馏权重下载，并附论文、模型说明和用法建议；仓库 LICENSE 为 MIT。DeepSeek-V3.2-Exp README 将该发布说明为实验版本，介绍 Sparse Attention 的研究目标并显示 MIT LICENSE。官方权重仓库/下载链接使研究者可在遵守各自许可的条件下取得并运行这些检查点。

不能把“开源”过度翻译为“所有要素完全开放”：代码仓库有 MIT 标识，并不自动证明数据语料、数据许可、训练日志、完整训练数据、全部生产推理服务、商标权和第三方基座许可都属于同一许可。蒸馏模型还继承其 Qwen/Llama 基座各自的条件，需分别核对模型卡、权重许可及第三方依赖。

更不能由可下载权重推断“我们也能造出同规模/质量模型”。复现需要满足许可和隐私约束的数据、训练与评测配方、算法/工程人才、大规模算力与网络、稳定训练基础设施、失败试验预算及安全对齐流程；只读公开论文和代码无法证明这些输入已具备，也无法保证复现质量。更务实的项目路径是原型化多角色工作流，先对比托管 API 与可合法部署的现成开源权重，再做任务特定评测、检索/工具接入、风险控制和成本/延迟测算；不要把 scripted mock agent 描述为真实并行模型推理。

来源：

- [DeepSeek 官方 GitHub 组织](https://github.com/deepseek-ai)（公开项目目录）
- [DeepSeek-R1 官方仓库](https://github.com/deepseek-ai/DeepSeek-R1)（README：研究描述、可下载权重与蒸馏信息）
- [DeepSeek-R1 LICENSE](https://github.com/deepseek-ai/DeepSeek-R1/blob/main/LICENSE)（该仓库代码许可；权重/蒸馏基座仍应核对各自模型卡）
- [DeepSeek-V3 官方仓库](https://github.com/deepseek-ai/DeepSeek-V3)（V3 模型与实现说明）
- [DeepSeek-V3.2-Exp 官方仓库](https://github.com/deepseek-ai/DeepSeek-V3.2-Exp)（实验性版本说明及仓库许可）
- [DeepSeek-V3.2-Exp LICENSE](https://github.com/deepseek-ai/DeepSeek-V3.2-Exp/blob/main/LICENSE)

时间/许可注意：模型清单和许可可能变化。本文只对上面列明的一手仓库在核查日可见内容作陈述；“MIT”仅覆盖许可证文本定义的相应作品，不能替代权重文件的 model card、基座模型条款、数据权利或商业合规审查。
