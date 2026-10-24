import type { CandlestickData, HistogramData, Time, UTCTimestamp } from "lightweight-charts";
import type { Candle, CandleInterval } from "../../shared/api/market";

export const periods: { value: CandleInterval; label: string }[] = [
  { value: "1m", label: "1分" }, { value: "5m", label: "5分" }, { value: "15m", label: "15分" },
  { value: "30m", label: "30分" }, { value: "60m", label: "60分" }, { value: "1d", label: "日" },
  { value: "1w", label: "周" }, { value: "1M", label: "月" }, { value: "1y", label: "年" },
];

export function halfYearBounds(now = new Date()) {
  const hkDay = new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Hong_Kong", year: "numeric", month: "2-digit", day: "2-digit" }).format(now);
  const [year, month, day] = hkDay.split("-").map(Number);
  const start = new Date(Date.UTC(year, month - 7, 1));
  const lastDay = new Date(Date.UTC(start.getUTCFullYear(), start.getUTCMonth() + 1, 0)).getUTCDate();
  start.setUTCDate(Math.min(day, lastDay));
  return { from: start.toISOString().slice(0, 10), to: hkDay };
}

export function chartData(items: Candle[], interval: CandleInterval) {
  const minutes = interval.endsWith("m");
  const unique = new Map<string | number, { candle: Candle; time: Time }>();
  for (const candle of items) {
    if (![candle.open, candle.high, candle.low, candle.close, candle.volume].every(Number.isFinite) || candle.volume < 0 || candle.low < 0 || candle.low > Math.min(candle.open, candle.close) || candle.high < Math.max(candle.open, candle.close)) throw new Error("行情 OHLC 数据不完整，请重试。");
    let time: Time;
    if (minutes) {
      const stamp = candle.timestamp ? Date.parse(candle.timestamp) : NaN;
      if (!Number.isFinite(stamp)) throw new Error("缺少原生分钟时间，无法绘图。");
      time = Math.floor(stamp / 1000) as UTCTimestamp;
    } else {
      time = candle.trade_date.slice(0, 10);
      const stamp = Date.parse(`${time}T00:00:00Z`);
      if (!/^\d{4}-\d{2}-\d{2}$/.test(time) || !Number.isFinite(stamp) || new Date(stamp).toISOString().slice(0, 10) !== time) throw new Error("行情日期无效。");
    }
    unique.set(time, { candle, time });
  }
  const sorted = [...unique.values()].sort((a, b) => a.time < b.time ? -1 : a.time > b.time ? 1 : 0);
  const candles: CandlestickData<Time>[] = sorted.map(({ candle, time }) => ({ time, open: candle.open, high: candle.high, low: candle.low, close: candle.close }));
  const volume: HistogramData<Time>[] = sorted.map(({ candle, time }) => ({ time, value: candle.volume, color: candle.close >= candle.open ? "#22c55e80" : "#f8717180" }));
  return { candles, volume, rows: sorted };
}

export function barLabel(bar: Candle, interval: CandleInterval) {
  if (!interval.endsWith("m")) return bar.trade_date.slice(0, 10);
  return new Intl.DateTimeFormat("zh-HK", { timeZone: "Asia/Hong_Kong", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit", hour12: false }).format(new Date(bar.timestamp!));
}
