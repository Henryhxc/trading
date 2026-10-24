import { describe, expect, it } from "vitest";
import { chartData, halfYearBounds } from "./chart-data";
import type { Candle } from "../../shared/api/market";

describe("candle calendar validation", () => {
  it("clamps the six-month start to the target month end in Hong Kong", () => {
    expect(halfYearBounds(new Date("2026-08-31T01:00:00Z"))).toEqual({ from: "2026-02-28", to: "2026-08-31" });
    expect(halfYearBounds(new Date("2026-10-23T20:00:00Z"))).toEqual({ from: "2026-04-24", to: "2026-10-24" });
  });
  it("rejects impossible calendar dates before passing data to the chart", () => {
    const bar: Candle = { trade_date: "2026-02-31T00:00:00Z", open: 10, high: 12, low: 9, close: 11, volume: 10, turnover: 110, source: "akshare", data_status: "unofficial", collected_at: "2026-10-24T00:00:00Z" };
    expect(() => chartData([bar], "1d")).toThrow(/日期/);
  });
});
