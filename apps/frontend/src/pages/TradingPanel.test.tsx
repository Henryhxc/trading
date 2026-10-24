import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TradingPanel } from "./TradingPanel";
import { tradingApi } from "../shared/api/trading";

vi.mock("../shared/api/trading", () => ({ tradingApi: { createOrder: vi.fn() } }));

const quote = { symbol: "00005", display_name: "HSBC Holdings", last_price: 63.45, bid: 63.4, ask: 63.5, change: 0.5, change_percent: 0.8, volume: 100, turnover: 6345, as_of: "2026-10-24T00:00:00Z", source: "mock", data_status: "simulated" as const };
const account = { currency: "HKD" as const, cash_balance: 815300, available_cash: 815300, initial_cash: 1000000, status: "active" as const };

function renderTicket() {
  return render(<TradingPanel quote={quote} account={account} accountError="" onRetry={vi.fn()} onOrderCreated={vi.fn(async () => undefined)} />);
}

describe("TradingPanel buying power", () => {
  beforeEach(() => vi.clearAllMocks());

  it("shows the maximum whole-lot buy quantity from available cash and price", () => {
    renderTicket();

    const ticket = screen.getByRole("complementary", { name: "Order ticket" });
    expect(within(ticket).getByText("12,800 shares / 128 lots")).toBeTruthy();
    expect(within(ticket).getByText(/HKD 812,160\.00/)).toBeTruthy();
  });

  it("blocks a buy quantity above the cash buying limit with a shortfall warning", () => {
    renderTicket();
    const ticket = screen.getByRole("complementary", { name: "Order ticket" });
    fireEvent.change(within(ticket).getByLabelText("Quantity"), { target: { value: "12900" } });
    expect(within(ticket).getByRole("alert").textContent).toMatch(/current buying limit of 12,800 shares/i);
    expect(within(ticket).getByRole("alert").textContent).toMatch(/shortfall: HKD 3,205\.00/i);
    expect((within(ticket).getByRole("button", { name: "Submit paper order" }) as HTMLButtonElement).disabled).toBe(true);
    expect(tradingApi.createOrder).not.toHaveBeenCalled();
  });

  it("keeps the cash limit enforced when the form is submitted directly", () => {
    renderTicket();
    const ticket = screen.getByRole("complementary", { name: "Order ticket" });
    fireEvent.change(within(ticket).getByLabelText("Quantity"), { target: { value: "12900" } });
    fireEvent.submit(ticket.querySelector("form") as HTMLFormElement);
    expect(tradingApi.createOrder).not.toHaveBeenCalled();
    expect(within(ticket).getAllByRole("alert").some(alert => /current buying limit of 12,800 shares/i.test(alert.textContent || ""))).toBe(true);
  });
});
