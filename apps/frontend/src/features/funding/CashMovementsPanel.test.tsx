import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { CashMovementsPanel } from "./CashMovementsPanel";

const movements = [{ id: 1, type: "deposit" as const, amount: 10000, currency: "HKD" as const, method: "Bank transfer (mock)", status: "completed" as const, created_at: "2026-05-01T08:00:00.000Z", prototype: true }];
const historyMovements = [
  ...movements,
  { id: 2, type: "withdraw" as const, amount: 1200, currency: "HKD" as const, method: "FPS beneficiary (mock)", status: "pending" as const, created_at: "2026-05-02T08:00:00.000Z", prototype: true },
];
const account = { currency: "HKD" as const, cash_balance: 10000, available_cash: 10000, initial_cash: 10000, status: "active" as const };

describe("CashMovementsPanel", () => {
  it("keeps deposit, withdrawal, and history as distinct prototype screens", () => {
    render(<CashMovementsPanel mode="deposit" movements={movements} account={account} />);

    expect(screen.getByRole("heading", { name: "Deposit funds", level: 1 })).toBeTruthy();
    expect(screen.getByText(/Prototype only/i)).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Deposit amount"), { target: { value: "2500" } });
    fireEvent.click(screen.getByRole("button", { name: "Submit deposit" }));
    expect(screen.getByRole("status").textContent).toMatch(/deposit request staged/i);
  });

  it("renders withdrawal controls without sharing the deposit form", () => {
    render(<CashMovementsPanel mode="withdraw" movements={movements} account={account} />);

    expect(screen.getByRole("heading", { name: "Withdraw funds", level: 1 })).toBeTruthy();
    expect(screen.getByLabelText("Withdrawal amount")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Submit deposit" })).toBeNull();
  });

  it("shows mock movement history as a separate page", () => {
    render(<CashMovementsPanel mode="history" movements={historyMovements} account={account} />);

    expect(screen.getByRole("heading", { name: "Cash movement history", level: 1 })).toBeTruthy();
    expect(screen.getByRole("table", { name: "Cash movement history" })).toBeTruthy();
    expect(screen.getByText("Completed", { selector: "span" })).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Movement type"), { target: { value: "withdraw" } });
    expect(screen.getByText("FPS beneficiary (mock)")).toBeTruthy();
    expect(screen.queryByText("Bank transfer (mock)")).toBeNull();
  });

  it("keeps overview focused on account snapshot and recent activity", () => {
    render(<CashMovementsPanel mode="overview" movements={historyMovements} account={account} />);

    expect(screen.getByRole("heading", { name: "Account snapshot" })).toBeTruthy();
    expect(screen.getByRole("heading", { name: "Recent activity" })).toBeTruthy();
    expect(screen.queryByRole("table", { name: "Cash movement history" })).toBeNull();
    expect(screen.getByRole("heading", { name: "Currency balances" })).toBeTruthy();
    expect(screen.getByText("USD")).toBeTruthy();
    expect(screen.getByText("CNY")).toBeTruthy();
    expect(screen.getByRole("img", { name: /currency allocation/i })).toBeTruthy();
  });

  it("supports a simulated multi-currency exchange with a live conversion preview", () => {
    render(<CashMovementsPanel mode="exchange" movements={movements} account={account} />);

    expect(screen.getByRole("heading", { name: "Exchange currencies", level: 1 })).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Exchange amount"), { target: { value: "1000" } });
    expect(screen.getByText(/Estimated receive.*USD/i)).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Preview exchange" }));
    expect(screen.getByRole("status").textContent).toMatch(/exchange request staged/i);
  });
});
