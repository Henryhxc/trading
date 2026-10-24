import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PortfolioPanel } from "./PortfolioPanel";
import { tradingApi } from "../../shared/api/trading";

vi.mock("../../shared/api/trading", () => ({
  tradingApi: { orders: vi.fn(), cancelOrder: vi.fn() },
}));

describe("PortfolioPanel analytics", () => {
  beforeEach(() => {
    vi.mocked(tradingApi.orders).mockResolvedValue({
      items: [
        { symbol: "00700", side: "buy", order_type: "limit", quantity: 100, filled_price: 100, filled_at: "2026-05-05T08:00:00.000Z", status: "filled", limit_price: 100, id: 1, filled_quantity: 100, reject_reason: null, quote_as_of: null, created_at: "2026-05-05T08:00:00.000Z" },
        { symbol: "00700", side: "sell", order_type: "limit", quantity: 50, filled_price: 110, filled_at: "2026-05-06T08:00:00.000Z", status: "filled", limit_price: 110, id: 2, filled_quantity: 50, reject_reason: null, quote_as_of: null, created_at: "2026-05-06T08:00:00.000Z" },
      ],
    });
  });

  it("shows daily and monthly realized profit tables", async () => {
    render(<PortfolioPanel account={{ currency: "HKD", cash_balance: 1000, available_cash: 1000, initial_cash: 1000, status: "active" }} positions={[]} loading={false} error="" refresh={vi.fn(async () => undefined)} onSelect={vi.fn()} />);

    expect(await screen.findByRole("heading", { name: "Daily P/L" })).toBeTruthy();
    expect(screen.getByRole("heading", { name: "Monthly P/L" })).toBeTruthy();
    expect(screen.getAllByText("2026-05-06").length).toBeGreaterThan(0);
    expect(screen.getAllByText("+HKD 500.00").length).toBeGreaterThan(0);
    expect(screen.getAllByText("+10.00%").length).toBeGreaterThan(0);
  });
});
