import { render, screen } from "../../test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { tradingApi, type Order } from "../../shared/api/trading";
import { OrderHistoryPanel } from "./OrderHistoryPanel";

vi.mock("../../shared/api/trading", () => ({ tradingApi: { orders: vi.fn() } }));

describe("OrderHistoryPanel", () => {
  beforeEach(() => {
    vi.mocked(tradingApi.orders).mockResolvedValue({
      items: JSON.parse('[{"id":1,"symbol":"00700","name":"Tencent","side":"buy","quantity":100,"limit_price":380,"estimated_amount":38000,"status":"filled","filled_quantity":100,"filled_price":380,"created_at":"2026-10-24T01:00:00Z","quote_as_of":"2026-10-24T01:00:00Z"}]') as Order[],
    });
  });

  it("renders API orders without an order_type as limit orders instead of crashing", async () => {
    render(<OrderHistoryPanel />);

    expect(await screen.findByRole("heading", { name: "Order history" })).toBeTruthy();
    expect(await screen.findByText("Limit order")).toBeTruthy();
    expect(screen.getByText("00700")).toBeTruthy();
  });
});
