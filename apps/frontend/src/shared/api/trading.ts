import { apiRequest } from "./client";

export type OrderSide = "buy" | "sell";
export type OrderStatus = "pending" | "filled" | "cancelled" | "rejected";
export type Account = { currency: "HKD"; cash_balance: number; available_cash: number; initial_cash: number; status: "active" | "locked" };
export type Position = { symbol: string; display_name: string; quantity: number; available_quantity: number; average_cost: number; market_price: number | null; unrealized_pnl: number | null; market_value: number | null; data_status: string };
export type Order = { id: number; symbol: string; side: OrderSide; order_type: "limit"; quantity: number; limit_price: number; estimated_amount?: number; filled_quantity: number; filled_price: number | null; status: OrderStatus; reject_reason: string | null; quote_as_of: string | null; created_at: string; filled_at: string | null };
export type OrderFilters = { symbol?: string; side?: OrderSide | ""; status?: OrderStatus | "" };

function query(filters: OrderFilters) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) if (value) params.set(key, value);
  const suffix = params.toString();
  return suffix ? `?${suffix}` : "";
}

export const tradingApi = {
  account: () => apiRequest<Account>("/account"),
  positions: () => apiRequest<{ items: Position[] }>("/positions"),
  orders: (filters: OrderFilters = {}) => apiRequest<{ items: Order[] }>(`/orders${query(filters)}`),
  createOrder: (input: { symbol: string; side: OrderSide; order_type: "limit"; quantity: number; limit_price: number }) => {
    const clientOrderID = crypto.randomUUID();
    return apiRequest<Order>("/orders", { ...input, client_order_id: clientOrderID }, "POST", { "Idempotency-Key": clientOrderID });
  },
  cancelOrder: (id: number) => apiRequest<Order>(`/orders/${id}/cancel`, undefined, "POST"),
};
