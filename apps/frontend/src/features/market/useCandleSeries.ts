import { useEffect, useRef, useState } from "react";
import { marketApi, type CandleInterval, type CandleSeries } from "../../shared/api/market";
import { chartData, halfYearBounds } from "./chart-data";

type State = { key: string; series: CandleSeries | null; loading: boolean; error: string; checked: Date | null };

export function useCandleSeries(symbol: string, interval: CandleInterval) {
  const key = `${symbol}:${interval}`;
  const [state, setState] = useState<State>({ key, series: null, loading: true, error: "", checked: null });
  const [revision, setRevision] = useState(0);
  const generation = useRef(0);
  useEffect(() => {
    let active = true;
    const request = async () => {
      const id = ++generation.current;
      setState((last) => ({ key, series: last.key === key ? last.series : null, loading: true, error: "", checked: last.key === key ? last.checked : null }));
      try {
        const bounds = halfYearBounds();
        const series = await marketApi.candles(symbol, interval, bounds.from, bounds.to);
        if (!series || series.symbol !== symbol || series.interval !== interval) throw new Error("行情周期响应不匹配，请重试。");
        chartData(series.items, interval);
        if (active && generation.current === id) setState({ key, series, loading: false, error: "", checked: new Date() });
      } catch (error) {
        if (active && generation.current === id) setState((last) => ({ ...last, loading: false, error: error instanceof Error ? error.message : "行情暂时不可用，请重试。" }));
      }
    };
    void request();
    const timer = window.setInterval(() => void request(), 60_000);
    return () => { active = false; window.clearInterval(timer); };
  }, [key, symbol, interval, revision]);
  return { ...(state.key === key ? state : { key, series: null, loading: true, error: "", checked: null }), refresh: () => setRevision((value) => value + 1) };
}
