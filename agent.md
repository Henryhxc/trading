# Agent 规则入口

本项目的智能体协作规则、技术约束、HKEX 行情边界、项目档案规划和验证要求统一维护在 [AGENTS.md](AGENTS.md)。

使用本文件名读取项目规则的工具应继续读取 `AGENTS.md`，避免维护两份可能不一致的规则。

## 免费港股行情调用摘要

- MVP 默认使用 `MockMarketDataProvider`。
- 原型阶段可参考 [stock-api](https://github.com/zhangxiangliang/stock-api)（腾讯/新浪，代码如 `HK00700`）和 [easyquotation](https://github.com/shidenggui/easyquotation)（腾讯港股，代码如 `00700`）；它们是免费公开源适配参考，不是 HKEX 官方实时 API。
- 所有外部行情必须由 Go 后端 `MarketDataProvider` 适配，React 只调用内部 `/api/v1/markets/*`。
- 内部代码统一六位数字（如 `00700`），由 adapter 映射到上游格式 `hk00700`、`HK00700` 或 `00700`。
- 返回值必须带 `source`、`as_of` 和 `data_status`；公开源只能标记 `unofficial`/`delayed`/`stale`，模拟源标记 `simulated`，不得冒充 `live`。
- 免费行情只用于模拟交易；真实实时行情另接已授权的 HKEX OMD-C 或信息供应商。

原型调用示例（只用于验证上游，不直接放进 Go 业务层）：

```ts
import { stocks } from "stock-api";

const quote = await stocks.auto.getStock("HK00700");
```

```python
import easyquotation

quotation = easyquotation.use("hkquote")
quote = quotation.real(["00700"])
```
