import { describe, expect, it } from "vitest";
import { aggregateProfit, type ProfitOrder } from "./profit";

const filled = (overrides: Partial<ProfitOrder>): ProfitOrder => ({
  symbol: "00700",
  side: "buy",
  quantity: 100,
  filled_price: 100,
  filled_at: "2026-05-05T08:00:00.000Z",
  status: "filled",
  ...overrides,
});

describe("aggregateProfit", () => {
  it("calculates realized daily profit from weighted open lots", () => {
    const rows = aggregateProfit([
      filled({ quantity: 200, filled_price: 100 }),
      filled({ side: "sell", quantity: 50, filled_price: 110, filled_at: "2026-05-06T08:00:00.000Z" }),
      filled({ side: "sell", quantity: 50, filled_price: 90, filled_at: "2026-05-07T08:00:00.000Z" }),
    ], "day");

    expect(rows).toEqual([
      { period: "2026-05-07", pnl: -500, costBasis: 5000, returnPercent: -10 },
      { period: "2026-05-06", pnl: 500, costBasis: 5000, returnPercent: 10 },
    ]);
  });

  it("groups daily fills into calendar months", () => {
    const rows = aggregateProfit([
      filled({ quantity: 100, filled_price: 20 }),
      filled({ side: "sell", quantity: 50, filled_price: 22, filled_at: "2026-05-20T08:00:00.000Z" }),
      filled({ side: "sell", quantity: 50, filled_price: 25, filled_at: "2026-06-03T08:00:00.000Z" }),
    ], "month");

    expect(rows).toEqual([
      { period: "2026-06", pnl: 250, costBasis: 1000, returnPercent: 25 },
      { period: "2026-05", pnl: 100, costBasis: 1000, returnPercent: 10 },
    ]);
  });
});
