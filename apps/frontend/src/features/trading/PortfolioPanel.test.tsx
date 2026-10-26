import { render, screen, waitFor } from "../../test-utils";
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

  it("switches realized profit calendar between day, month, and year", async () => {
    render(<PortfolioPanel account={{ currency: "HKD", cash_balance: 1000, available_cash: 1000, initial_cash: 1000, status: "active" }} positions={[]} loading={false} error="" refresh={vi.fn(async () => undefined)} onSelect={vi.fn()} />);

    expect(await screen.findByRole("heading", { name: "Daily P/L" })).toBeTruthy();
    expect(screen.getByRole("tab", { name: "Day" }).getAttribute("aria-selected")).toBe("true");
    expect(screen.getByRole("grid", { name: "Daily realized P/L calendar" })).toBeTruthy();
    await waitFor(() => expect(screen.getByRole("gridcell", { name: /2026-05-06.*profit/i })).toBeTruthy());
    expect(screen.getAllByText("+HKD 500.00").length).toBeGreaterThan(0);
    expect(screen.getAllByText("+10.00%").length).toBeGreaterThan(0);

    screen.getByRole("tab", { name: "Month" }).click();
    expect(await screen.findByRole("heading", { name: "Monthly P/L" })).toBeTruthy();
    expect(screen.getByRole("grid", { name: "Monthly realized P/L calendar" })).toBeTruthy();
    screen.getByRole("tab", { name: "Year" }).click();
    expect(await screen.findByRole("heading", { name: "Yearly P/L" })).toBeTruthy();
    expect(screen.getByRole("grid", { name: "Yearly realized P/L calendar" })).toBeTruthy();
  });

  it("shows an accessible portfolio allocation pie chart", async () => {
    render(<PortfolioPanel account={{ currency: "HKD", cash_balance: 1000, available_cash: 1000, initial_cash: 1000, status: "active" }} positions={[{ symbol: "00700", display_name: "Tencent Holdings", quantity: 100, available_quantity: 100, average_cost: 300, market_price: 320, market_value: 32000, unrealized_pnl: 2000, data_status: "simulated" }]} loading={false} error="" refresh={vi.fn(async () => undefined)} onSelect={vi.fn()} />);

    expect(await screen.findByRole("img", { name: /portfolio allocation/i })).toBeTruthy();
    expect(screen.getAllByText("Tencent Holdings")[1]?.parentElement?.textContent).toMatch(/97%/);
  });

  it("keeps order history out of the positions page", async () => {
    render(<PortfolioPanel account={{ currency: "HKD", cash_balance: 1000, available_cash: 1000, initial_cash: 1000, status: "active" }} positions={[]} loading={false} error="" refresh={vi.fn(async () => undefined)} onSelect={vi.fn()} />);

    expect(await screen.findByRole("heading", { name: "My positions" })).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Order history" })).toBeNull();
  });
});
