import { useEffect, useMemo, useRef, useState } from "react";
import { CandlestickSeries, ColorType, createChart, HistogramSeries, type IChartApi, type ISeriesApi, type Time } from "lightweight-charts";
import type { Candle, CandleInterval } from "../../shared/api/market";
import { barLabel, chartData } from "./chart-data";

type Props = { items: Candle[]; interval: CandleInterval };
export function CandlestickChart({ items, interval }: Props) {
  const host = useRef<HTMLDivElement>(null);
  const chart = useRef<IChartApi | null>(null);
  const price = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const volume = useRef<ISeriesApi<"Histogram"> | null>(null);
  const [hover, setHover] = useState<Candle | null>(null);
  const data = useMemo(() => chartData(items, interval), [items, interval]);
  const rows = useRef(data.rows);
  rows.current = data.rows;

  useEffect(() => {
    if (!host.current) return;
    const instance = createChart(host.current, {
      height: 340, width: host.current.clientWidth || 600,
      layout: { background: { type: ColorType.Solid, color: "#0d1726" }, textColor: "#aabbd1", attributionLogo: true, fontFamily: "Inter, Segoe UI, sans-serif", fontSize: 11 },
      grid: { vertLines: { color: "#1a283b" }, horzLines: { color: "#1a283b" } },
      rightPriceScale: { borderColor: "#2a3a50", scaleMargins: { top: 0.08, bottom: 0.25 } },
      timeScale: { borderColor: "#2a3a50", rightOffset: 3, minBarSpacing: 1 },
      localization: { locale: "zh-HK", timeFormatter: (time: Time) => typeof time === "number" ? new Date(time * 1000).toLocaleString("zh-HK", { timeZone: "Asia/Hong_Kong", hour12: false }) : String(time) },
    });
    const candleSeries = instance.addSeries(CandlestickSeries, { upColor: "#22c55e", downColor: "#f87171", borderVisible: false, wickUpColor: "#22c55e", wickDownColor: "#f87171" });
    const volumeSeries = instance.addSeries(HistogramSeries, { priceFormat: { type: "volume" }, priceScaleId: "volume" });
    volumeSeries.priceScale().applyOptions({ scaleMargins: { top: 0.82, bottom: 0 } });
    instance.subscribeCrosshairMove((event) => {
      const bar = rows.current.find((row) => row.time === event.time);
      setHover(bar?.candle || null);
    });
    chart.current = instance; price.current = candleSeries; volume.current = volumeSeries;
    const observer = new ResizeObserver(([entry]) => { if (entry.contentRect.width > 0) instance.applyOptions({ width: entry.contentRect.width }); });
    observer.observe(host.current);
    return () => { observer.disconnect(); instance.remove(); chart.current = null; price.current = null; volume.current = null; };
  }, []);

  useEffect(() => {
    price.current?.setData(data.candles);
    volume.current?.setData(data.volume);
    chart.current?.applyOptions({ timeScale: {
      timeVisible: interval.endsWith("m"), secondsVisible: false,
      tickMarkFormatter: (time: Time, type: number) => typeof time === "number" ? new Intl.DateTimeFormat("zh-HK", { timeZone: "Asia/Hong_Kong", ...(type >= 3 ? { hour: "2-digit", minute: "2-digit", hour12: false } : { month: "2-digit", day: "2-digit" }) }).format(new Date(time * 1000)) : null,
    } });
    setHover(null);
  }, [data, interval]);
  // Fit only on mount/period changes; polling keeps the user's zoom and scroll.
  const fitted = useRef(false);
  useEffect(() => { if (!fitted.current && data.candles.length) { chart.current?.timeScale().fitContent(); fitted.current = true; } }, [data]);

  const selected = hover || data.rows.at(-1)?.candle;
  const change = selected && selected.open > 0 ? (selected.close / selected.open - 1) * 100 : null;
  return <>
    <div className="candle-readout" aria-live="off">{selected && <><span>{barLabel(selected, interval)}</span><span>O <b>{selected.open.toFixed(2)}</b></span><span>H <b>{selected.high.toFixed(2)}</b></span><span>L <b>{selected.low.toFixed(2)}</b></span><span>C <b>{selected.close.toFixed(2)}</b></span><span className={change !== null && change >= 0 ? "candle-up" : "candle-down"}>本柱 {change === null ? "—" : `${change >= 0 ? "+" : ""}${change.toFixed(2)}%`}</span></>}</div>
    <div className="candle-canvas" ref={host} role="img" aria-label="OHLC candlestick chart with trading volume" />
  </>;
}
