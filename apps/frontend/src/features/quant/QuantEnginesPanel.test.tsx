import { fireEvent, render, screen } from "../../test-utils";
import { describe, expect, it } from "vitest";
import { QuantEnginesPanel } from "./QuantEnginesPanel";

describe("QuantEnginesPanel prototype", () => {
  it("shows multiple engines and lets investors inspect a different lens", () => {
    render(<QuantEnginesPanel />);

    expect(screen.getByRole("heading", { name: "Multiple quant engines" })).toBeTruthy();
    expect(screen.getAllByRole("tab")).toHaveLength(4);
    fireEvent.click(screen.getByRole("tab", { name: /Multi-factor score/i }));
    expect(screen.getAllByText("Quality + value lead").length).toBeGreaterThan(0);
    expect(screen.getByText("Compare independent assumptions, not just headline return")).toBeTruthy();
  });

  it("keeps at least one comparison engine selected", () => {
    render(<QuantEnginesPanel />);

    const momentum = screen.getByRole("checkbox", { name: "Compare Momentum" }) as HTMLInputElement;
    const factors = screen.getByRole("checkbox", { name: "Compare Factors" }) as HTMLInputElement;
    expect(momentum.checked).toBe(true);
    expect(factors.checked).toBe(true);
    fireEvent.click(momentum);
    fireEvent.click(factors);
    expect(factors.checked).toBe(true);
  });
});
