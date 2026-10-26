import { useState } from "react";
import type { CandleInterval } from "../../shared/api/market";
import { CandlestickChart } from "./CandlestickChart";
import { barLabel, halfYearBounds, periods } from "./chart-data";
import { useCandleSeries } from "./useCandleSeries";
import "./market-chart.css";
import { useI18n } from "../../app/I18nProvider";

export function MarketChart({ symbol }: { symbol: string }) {
  const { language, t } = useI18n();
  const [interval, setInterval] = useState<CandleInterval>("1d");
  const { series, loading, error, checked, refresh } = useCandleSeries(symbol, interval);
  const bounds = halfYearBounds();
  const items = series?.items || [];
  const minutes = interval.endsWith("m");
  return <section className="market-chart" aria-label={t("Candlestick chart")}>
    <div className="candle-toolbar"><div><h3>{t("Candlestick chart")} <small>{symbol} · HKD</small></h3><p>{t("Last six months")} · {bounds.from} — {bounds.to}</p></div><button type="button" aria-label={t("Refresh market data")} onClick={refresh} disabled={loading}>{loading ? t("Refreshing…") : `↻ ${t("Refresh")}`}</button></div>
    <div className="candle-periods" role="group" aria-label={t("Candlestick interval")}>{periods.map((period) => <button type="button" key={period.value} aria-pressed={interval === period.value} className={interval === period.value ? "selected" : ""} onClick={() => setInterval(period.value)}>{language === "zh" ? period.label : t(period.value.endsWith("m") ? "{n} min" : period.value === "1d" ? "Daily" : period.value === "1w" ? "Weekly" : period.value === "1M" ? "Monthly" : "Yearly", { n: period.value.slice(0, -1) })}</button>)}</div>
    <div className="candle-meta"><span>{series?.source || "none"} <b className={`candle-status ${series?.data_status || "unavailable"}`}>{t(series?.data_status || "unavailable")}</b></span><span>{t("{count} bars · refresh every 60 seconds", { count: items.length.toLocaleString() })}</span></div>
    {error && <p className="candle-error" role="alert">{t(error)} {items.length > 0 && t("Showing cached data from the last read. Refresh to retry.")}</p>}
    {loading && items.length === 0 ? <p className="candle-state" role="status">{t(minutes ? "Loading {kind} market data…" : "Loading {kind} market data…", { kind: minutes ? t("minute") : t("historical") })}</p> : items.length > 0 ? <CandlestickChart key={`${symbol}:${interval}`} items={items} interval={interval} /> : !error && <p className="candle-state" role="status">{t(minutes ? "No native minute data is available. The upstream does not provide this interval; daily bars are not substituted." : "No historical market data is available for this date range. Refresh later.")}</p>}
    {items.length > 0 && <p className="candle-coverage">{t("Coverage")}: {barLabel(items[0], interval)} — {barLabel(items.at(-1)!, interval)} · {t(minutes ? "Hong Kong time" : "Daily aggregation within range")}</p>}
    {minutes && <p className="candle-note">{t("Minute history is limited upstream. One-minute data is usually available for only the last five trading days and may not cover six months.")}</p>}
    <div className="candle-footer"><span>{t(series?.data_status === "simulated" ? "Simulated data · for demonstration only" : "Unofficial market data · for simulation only")}{series?.as_of && ` · ${t("collected")} ${new Date(series.as_of).toLocaleString(language === "zh" ? "zh-HK" : "en-HK", { timeZone: "Asia/Hong_Kong", hour12: false })}`}{checked && ` · ${t("read")} ${checked.toLocaleTimeString(language === "zh" ? "zh-HK" : "en-HK", { timeZone: "Asia/Hong_Kong", hour12: false })}`}</span><a href="https://www.tradingview.com/" target="_blank" rel="noreferrer">TradingView Lightweight Charts™</a></div>
    <p className="candle-attribution">Copyright (с) 2025 TradingView, Inc.</p>
  </section>;
}
