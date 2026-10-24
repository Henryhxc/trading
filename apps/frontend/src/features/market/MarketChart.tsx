import { useState } from "react";
import type { CandleInterval } from "../../shared/api/market";
import { CandlestickChart } from "./CandlestickChart";
import { barLabel, halfYearBounds, periods } from "./chart-data";
import { useCandleSeries } from "./useCandleSeries";
import "./market-chart.css";

export function MarketChart({ symbol }: { symbol: string }) {
  const [interval, setInterval] = useState<CandleInterval>("1d");
  const { series, loading, error, checked, refresh } = useCandleSeries(symbol, interval);
  const bounds = halfYearBounds();
  const items = series?.items || [];
  const minutes = interval.endsWith("m");
  return <section className="market-chart" aria-label="K 线行情">
    <div className="candle-toolbar"><div><h3>K 线行情 <small>{symbol} · HKD</small></h3><p>近半年 · {bounds.from} — {bounds.to}</p></div><button type="button" aria-label="刷新行情" onClick={refresh} disabled={loading}>{loading ? "更新中…" : "↻ 刷新"}</button></div>
    <div className="candle-periods" role="group" aria-label="K 线周期">{periods.map((period) => <button type="button" key={period.value} aria-pressed={interval === period.value} className={interval === period.value ? "selected" : ""} onClick={() => setInterval(period.value)}>{period.label}</button>)}</div>
    <div className="candle-meta"><span>{series?.source || "none"} <b className={`candle-status ${series?.data_status || "unavailable"}`}>{series?.data_status || "unavailable"}</b></span><span>{items.length.toLocaleString()} 根 · 每 60 秒读取</span></div>
    {error && <p className="candle-error" role="alert">{error} {items.length > 0 && "显示上次读取的缓存，请刷新重试。"}</p>}
    {loading && items.length === 0 ? <p className="candle-state" role="status">正在载入{minutes ? "分钟" : "历史"}行情…</p> : items.length > 0 ? <CandlestickChart key={`${symbol}:${interval}`} items={items} interval={interval} /> : !error && <p className="candle-state" role="status">{minutes ? "暂无原生分钟行情。上游暂未提供该周期，不会用日线替代。" : "暂无该日期范围的历史行情，请稍后刷新。"}</p>}
    {items.length > 0 && <p className="candle-coverage">实际覆盖 {barLabel(items[0], interval)} — {barLabel(items.at(-1)!, interval)} · {minutes ? "香港时间" : "区间内日线聚合"}</p>}
    {minutes && <p className="candle-note">分钟历史受上游限制；1 分钟通常仅近 5 个交易日，不保证覆盖半年。</p>}
    <div className="candle-footer"><span>{series?.data_status === "simulated" ? "模拟数据 · 仅用于演示" : "非官方行情 · 仅用于模拟"}{series?.as_of && ` · 采集 ${new Date(series.as_of).toLocaleString("zh-HK", { timeZone: "Asia/Hong_Kong", hour12: false })}`}{checked && ` · 读取 ${checked.toLocaleTimeString("zh-HK", { timeZone: "Asia/Hong_Kong", hour12: false })}`}</span><a href="https://www.tradingview.com/" target="_blank" rel="noreferrer">TradingView Lightweight Charts™</a></div>
    <p className="candle-attribution">Copyright (с) 2025 TradingView, Inc.</p>
  </section>;
}
