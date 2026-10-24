import { act, renderHook } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import { useCandleSeries } from "./useCandleSeries";
import { marketApi, type CandleInterval, type CandleSeries } from "../../shared/api/market";

vi.mock("../../shared/api/market", () => ({ marketApi: { candles: vi.fn() } }));
const series = (interval: CandleInterval): CandleSeries => ({ symbol: "00700", interval, items: [], source: "none", data_status: "unavailable", as_of: null });
beforeEach(() => { vi.useFakeTimers(); vi.mocked(marketApi.candles).mockReset(); });

it("ignores the older request after the selected period changes", async () => {
  let resolveOld!: (value: CandleSeries) => void;
  vi.mocked(marketApi.candles).mockImplementationOnce(() => new Promise((resolve) => { resolveOld = resolve; })).mockResolvedValue(series("5m"));
  const hook = renderHook(({ interval }: { interval: CandleInterval }) => useCandleSeries("00700", interval), { initialProps: { interval: "1d" as CandleInterval } });
  await act(async () => hook.rerender({ interval: "5m" }));
  expect(hook.result.current.series?.interval).toBe("5m");
  await act(async () => resolveOld(series("1d")));
  expect(hook.result.current.series?.interval).toBe("5m");
});

it("polls every minute, preserves cache on failure, and stops after unmount", async () => {
  vi.mocked(marketApi.candles).mockResolvedValueOnce(series("1d")).mockRejectedValue(new Error("upstream offline"));
  const hook = renderHook(() => useCandleSeries("00700", "1d"));
  await act(async () => {});
  expect(hook.result.current.series?.interval).toBe("1d");
  await act(async () => vi.advanceTimersByTime(60_000));
  expect(hook.result.current.error).toBe("upstream offline");
  expect(hook.result.current.series?.interval).toBe("1d");
  expect(marketApi.candles).toHaveBeenCalledTimes(2);
  hook.unmount();
  await act(async () => vi.advanceTimersByTime(120_000));
  expect(marketApi.candles).toHaveBeenCalledTimes(2);
});
