# HKEX 港股行情的免费方案（官方资料研究）

访问日期：2026-10-24（Asia/Hong_Kong）。本笔记只使用 HKEX / HKEX Information Services（HKEX-IS）官方页面；费用页面的“Updated”日期保留在相应条目中，实际采购前应再次核对最新费率和许可合同。

## 结论先行

如果“免费”指可用于生产交易系统的实时行情，HKEX 官方资料没有显示可长期免费使用的 OMD-C/OMD-D 实时方案。生产实时行情通过 HKEX-IS 授权和 OMD 数据源提供；免费部分主要是公开网页展示、部分随 OMD 订阅附送的数据流，以及有明确费用的开发/测试服务中的技术文档和测试流程。把网页上的报价抓取当作交易数据 API，或绕过 HKEX-IS 许可分发，均不应视为官方支持的接入方式。

## 1. 可以接入哪些市场与产品

* **OMD-C（证券市场）**：SEHK 上交易的股票、权证、REIT、债务证券、ETF、单位信托/互惠基金、股本挂钩产品、衍生权证、牛熊证等。数据层级是 Securities Standard（SS，合并的 market-by-price）、Securities Premium（SP，流式 market-by-price）和 Securities FullTick（SF，流式 market-by-order）。来源：[Real-time Datafeeds](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Overview/Real_time-Datafeeds?sc_lang=en)；[OMD-C 技术/接入页](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Securities-Market-OMD-C?sc_lang=en)（页面 Updated 29 Dec 2017，访问 2026-10-24）。
* **OMD-D（衍生品市场）**：HKFE 及 SEHK 股票期权市场的股票期货、股指期货及期权、利率和固定收益期货、商品、股票期权、货币期货、黄金期货等。数据层级是 Derivatives Standard（DS）、Derivatives Premium（DP）和 Derivatives FullTick（DF）。来源：[Real-time Datafeeds](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Overview/Real_time-Datafeeds?sc_lang=en)；[OMD-D 技术/接入页](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Derivatives-Market-OMD-D?sc_lang=en)（页面 Updated 12 Oct 2026，访问 2026-10-24）。
* **发行人新闻与公告**：IIS News feed，覆盖上市公司公告、财报、招股书、配股结果和交易所新闻。来源同上 [Real-time Datafeeds](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Overview/Real_time-Datafeeds?sc_lang=en)。
* **历史数据**：HKEX Data Marketplace 提供证券市场 Historical Full Book、衍生品市场（股票期权及非股票期权）Historical Full Book、证券属性日文件和 CCASS 持股数据；交付方式为 SFTP、cloud-to-cloud 或 direct download。来源：[HKEX Data Marketplace](https://www.hkex.com.hk/Services/Market-Data-Services/Historical-Data-Services/HKEX-Data-Marketplace?sc_lang=en)（访问 2026-10-24）。页面没有把这些产品标为免费，不能假定历史数据零成本。

## 2. “免费”与实时、延迟、网页数据的边界

### 公开网页展示

HKEX 网站公开提供 [Equities](https://www.hkex.com.hk/Market-Data/Securities-Prices/Equities?sc_lang=en)、[Exchange Traded Products](https://www.hkex.com.hk/Market-Data/Securities-Prices/Exchange-Traded-Products?sc_lang=en)、[Futures & Options Prices](https://www.hkex.com.hk/Market-Data/Futures-and-Options-Prices/Equity-Index?sc_lang=en) 等查询页。这些页面适合作为人工查看或原型演示入口；HKEX 的官方实时数据产品页把可供系统接入的产品定义为 OMD-C/OMD-D feed，并未在网页报价页提供公开、免许可的交易 API。不要把网页 HTML/脚本抓取当成获授权的生产行情源。（访问 2026-10-24。）

### 延迟数据

HKEX 费率表明确列出“Delayed Data only（on a delay basis of 15 minutes or more）”这一授权类别，说明至少 15 分钟延迟可以作为独立许可/分发类别购买；它不是 OMD 实时订阅的免费替代品。证券市场及衍生品市场的市场数据供应商费率页均有该类别。来源：[证券市场 Market Data Vendors](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Securities-(Hong-Kong)/Market-Data/Market-Data-Vendors?sc_lang=en)、[衍生品 Market Data Vendors](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Listed-Derivatives/Market-Data/Market-Data-Vendors?sc_lang=en)（两页均显示 Updated 27 Feb 2015；访问 2026-10-24）。

### OMD 附送的 complimentary feeds

订阅相应 OMD 数据流时，HKEX 列出若干 complimentary feeds（例如证券的 Odd Lot Order、Conflated Broker Queue、Index、Stock Connect Market；衍生品的 Derivatives Trades、Index、Order）。这些数据流不能单独订阅，HKEX 保留日后收费权；因此不能当作独立免费行情源。来源：[证券市场 End-users 费率页](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Securities-(Hong-Kong)/Market-Data/End_users?sc_lang=en)及[衍生品 End-users 费率页](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Listed-Derivatives/Market-Data/End_users?sc_lang=en)（访问 2026-10-24）。

## 3. 官方费用信号（用于判断“免费”是否成立）

以下是官方公开费率中的代表性金额，均为 HK$，不是报价承诺：

| 场景 | 证券 OMD-C | 衍生品 OMD-D | 含义 |
|---|---:|---:|---|
| End-user licence（每季度） | SS 33,300；SP 43,200；SF 64,800 | DS 12,600；DP 16,500；DF 24,750 | 生产实时数据许可本身收费 |
| Direct connection（首套，列于许可页） | SS 45,000；SP 58,500；SF 87,600 | DS 15,000；DP 19,500；DF 29,250 | 连接费用另计；Mainland Market Data Hub 连接费为两倍 |
| 连续访问个人用户（每月） | Level 1 120；Level 2 200；SF Level 2+One 240；SF Full Book 400 | Level 1 25；Level 2 75；DP Level 2+One 90；DF Full Book 300 | 终端/用户使用费 |
| Non-display（每公司每月） | 自动交易 20,000；带可交易产品的衍生数据 20,000；其他 400 | 自动交易 10,000；带可交易产品的衍生数据 10,000；其他 150 | 算法交易/内部计算并非天然免费 |

来源：[证券 End-users](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Securities-(Hong-Kong)/Market-Data/End_users?sc_lang=en)、[衍生品 End-users](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Listed-Derivatives/Market-Data/End_users?sc_lang=en)（均显示 Updated 27 Feb 2015；访问 2026-10-24）。供应商向客户再分发时还要申请 Vendor Licence 并支付 redistribution、connection、subscriber 等费用；代表性费率见[证券供应商页](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Securities-(Hong-Kong)/Market-Data/Market-Data-Vendors?sc_lang=en)和[衍生品供应商页](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Listed-Derivatives/Market-Data/Market-Data-Vendors?sc_lang=en)。

HKEX 费用总览明确把 Market Data Vendors、End-users 等作为独立费用类别：[Fees Overview](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Overview?sc_lang=en)（页面 Updated 22 Feb 2023；访问 2026-10-24）。

## 4. 开发、测试与接入方式

* **直接接 OMD**：HKEX 说明 OMD 是低延迟、统一消息格式的平台；客户可直接从 OMD 接收，也可通过 information vendor 间接接收。OMD-C/OMD-D 页面分别提供 on-boarding checklist、Readiness Test 文档、binary protocol interface specification、Developer Guide 和 Connectivity Guide。来源：[Market Data Infrastructure](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure?sc_lang=en)、[OMD-C](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Securities-Market-OMD-C?sc_lang=en)、[OMD-D](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Derivatives-Market-OMD-D?sc_lang=en)（访问 2026-10-24）。
* **通过供应商**：若只是先做产品原型，可以购买已有 information vendor 的延迟或实时服务；此时授权和费用由供应商与 HKEX-IS 的许可链共同约束。HKEX 官方基础设施页明确支持“directly from OMD or indirectly via information vendors”。
* **ISV 测试环境不是免费**：HKEX 的 Independent Software Vendor Test Service 价目写明，证券或衍生品市场分别收取一次性注册费 HK$10,000；测试费证券每日 HK$1,500 / 每月 HK$21,000，衍生品每日 HK$500 / 每月 HK$7,000。已是 HKEX licensed IV/EU 或其许可协议 Schedule 2 所列关联公司的 ISV，一次性注册费可获豁免，但测试费仍按政策收取。来源：[ISV Test Service Price (Fee Schedule)](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Testing-Services/Independent-Software-Vendor/cPrice-(Fee-Schedule)?sc_lang=en)、[ISV FAQ](https://www.hkex.com.hk/Global/Exchange/FAQ/Market-Data/Independent-Software-Vendor?sc_lang=en)（访问 2026-10-24）。
* **技术资料可公开下载**：OMD-C/OMD-D 的开发者指南、接口规范和连接指南在官方页面公开，但下载资料不等于获得生产数据许可或交易资格；实际接入仍需 HKEX-IS 申请、连接、自检/Readiness Test 和相应许可。

## 5. 对项目选型的建议

1. **零预算验证 UI/策略逻辑**：先用 HKEX 公开网页手工导出的样例/自建 mock 数据，或使用已获授权供应商提供的延迟数据；文档中明确标注“演示/延迟”，不要承诺实时交易。
2. **需要真实实时行情但不做数据分发**：按 End-user 申请 OMD-C（港股）或 OMD-D（期货/期权），选择 SS/SP/SF 或 DS/DP/DF 的深度；预算至少覆盖许可、连接和用户/非展示费用。
3. **要向客户提供行情**：按 Vendor Licence 路径评估 redistribution、subscriber、connection、延迟类别和非展示费用；不能把一份 End-user 许可扩展成公共行情服务。
4. **只开发 feed handler**：使用公开 OMD 技术资料；如需官方测试，按 ISV Test Service 预算，或先向 HKEX-IS/信息供应商确认是否已有可复用测试资格。

## 来源索引（全部为 HKEX 官方域名）

* [Real-time Data Services / Real-time Datafeeds](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Overview/Real_time-Datafeeds?sc_lang=en)
* [OMD Infrastructure](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure?sc_lang=en)
* [OMD-C](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Securities-Market-OMD-C?sc_lang=en)
* [OMD-D](https://www.hkex.com.hk/Services/Market-Data-Services/Infrastructure/HKEX-Orion-Market-Data-Platform-Derivatives-Market-OMD-D?sc_lang=en)
* [HKEX Data Marketplace](https://www.hkex.com.hk/Services/Market-Data-Services/Historical-Data-Services/HKEX-Data-Marketplace?sc_lang=en)
* [Fees Overview](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Overview?sc_lang=en)
* [Securities Market Data Vendors](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Securities-(Hong-Kong)/Market-Data/Market-Data-Vendors?sc_lang=en) / [End-users](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Securities-(Hong-Kong)/Market-Data/End_users?sc_lang=en)
* [Derivatives Market Data Vendors](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Listed-Derivatives/Market-Data/Market-Data-Vendors?sc_lang=en) / [End-users](https://www.hkex.com.hk/Services/Rules-and-Forms-and-Fees/Fees/Listed-Derivatives/Market-Data/End_users?sc_lang=en)
* [ISV Test Service Fee Schedule](https://www.hkex.com.hk/Services/Market-Data-Services/Real-Time-Data-Services/Testing-Services/Independent-Software-Vendor/cPrice-(Fee-Schedule)?sc_lang=en) / [ISV FAQ](https://www.hkex.com.hk/Global/Exchange/FAQ/Market-Data/Independent-Software-Vendor?sc_lang=en)
