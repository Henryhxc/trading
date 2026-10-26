import { fireEvent, render, screen } from "../../test-utils";
import { describe, expect, it } from "vitest";
import { FinancingPanel } from "./FinancingPanel";

const positions = [{ symbol: "00700", display_name: "Tencent Holdings", quantity: 100, available_quantity: 100, average_cost: 300, market_price: 320, market_value: 32000, unrealized_pnl: 2000, data_status: "simulated" }];
const account = { currency: "HKD" as const, cash_balance: 50000, available_cash: 50000, initial_cash: 50000, status: "active" as const };

describe("FinancingPanel prototype", () => {
  it("estimates financing interest and stages a simulated request", () => {
    render(<FinancingPanel account={account} positions={positions} />);

    expect(screen.getByRole("heading", { name: "Financing simulator" })).toBeTruthy();
    expect(screen.getByText("Maximum illustrative loan").parentElement?.textContent).toMatch(/HKD 9,600/);
    fireEvent.change(screen.getByLabelText("Loan amount"), { target: { value: "5000" } });
    expect(screen.getByText("Estimated monthly interest").parentElement?.textContent).toMatch(/HKD 37\.50/);
    fireEvent.click(screen.getByRole("button", { name: "Stage financing request" }));
    expect(screen.getByText(/financing request staged/i).getAttribute("role")).toBe("status");
  });

  it("shows margin risk warnings for a downside price scenario", () => {
    render(<FinancingPanel account={account} positions={positions} />);

    fireEvent.click(screen.getByRole("tab", { name: "Margin buying" }));
    fireEvent.change(screen.getByLabelText("Loan amount"), { target: { value: "9600" } });
    fireEvent.change(screen.getByLabelText("Simulated market drop"), { target: { value: "60" } });
    expect(screen.getByRole("alert").textContent).toMatch(/maintenance|margin call|liquidation/i);
  });
});
