import { type FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { type Order, type OrderFilters, tradingApi } from "../../shared/api/trading";
import { money } from "./usePortfolio";

const date = (value: string | null) => value ? new Date(value).toLocaleString() : "—";

export function OrderHistoryPanel() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [filters, setFilters] = useState<OrderFilters>({});
  const [applied, setApplied] = useState<OrderFilters>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const request = useRef(0);
  const loadOrders = useCallback(async () => {
    const current = ++request.current;
    setLoading(true); setError("");
    try { const result = await tradingApi.orders(applied); if (current === request.current) setOrders(result?.items || []); }
    catch (cause) { if (current === request.current) setError(cause instanceof Error ? cause.message : "Order history is unavailable"); }
    finally { if (current === request.current) setLoading(false); }
  }, [applied]);
  useEffect(() => { void loadOrders(); return () => { request.current++; }; }, [loadOrders]);
  function applyFilters(event: FormEvent) { event.preventDefault(); setApplied({ ...filters, symbol: filters.symbol?.trim() }); }
  async function cancel(id: number) {
    try { await tradingApi.cancelOrder(id); await loadOrders(); }
    catch (cause) { setError(cause instanceof Error ? cause.message : "Order could not be cancelled"); }
  }
  return <main className="prototype-page order-history-page" aria-label="Order history prototype">
    <header className="prototype-heading"><div><p className="eyebrow">PAPER ACCOUNT · ORDER MANAGEMENT</p><h1>Order history</h1><p className="prototype-subtitle">Review, filter, and cancel pending paper orders.</p></div><span className="data-badge simulated">SIMULATION</span></header>
    <section className="data-card order-history-card"><div className="panel-title"><div><h2>All orders</h2><p className="panel-caption">Orders are isolated to the signed-in paper account.</p></div><span className="muted-count">{orders.length} orders</span></div>
      <form className="filter-row" onSubmit={applyFilters}><input aria-label="Filter orders by symbol" placeholder="Symbol" value={filters.symbol || ""} onChange={event => setFilters({ ...filters, symbol: event.target.value })} /><select aria-label="Filter orders by side" value={filters.side || ""} onChange={event => setFilters({ ...filters, side: event.target.value as OrderFilters["side"] })}><option value="">All sides</option><option value="buy">Buy</option><option value="sell">Sell</option></select><select aria-label="Filter orders by status" value={filters.status || ""} onChange={event => setFilters({ ...filters, status: event.target.value as OrderFilters["status"] })}><option value="">All statuses</option><option value="pending">Pending</option><option value="filled">Filled</option><option value="cancelled">Cancelled</option><option value="rejected">Rejected</option></select><button className="quiet-button" type="submit">Filter</button></form>
      {error && <p className="error-message" role="alert">{error}<button className="quiet-button" onClick={() => void loadOrders()}>Retry orders</button></p>}
      {loading ? <p className="panel-state" role="status">Loading orders…</p> : orders.length === 0 ? <p className="empty-state" role="status">{error ? "Order history is unavailable." : "No orders match the current filters."}</p> : <div className="table-wrap"><table><thead><tr><th>Security</th><th>Side / quantity</th><th>Limit</th><th>Status</th><th>Created</th><th>Action</th></tr></thead><tbody>{orders.map(order => <tr key={order.id}><td><strong>{order.symbol}</strong><small>{order.order_type}</small></td><td className={order.side === "buy" ? "up" : "down"}>{order.side.toUpperCase()} {order.quantity.toLocaleString()}</td><td>{money(order.limit_price)}</td><td><span className={`order-status ${order.status}`}>{order.status}</span>{order.reject_reason && <small>{order.reject_reason}</small>}</td><td>{date(order.created_at)}</td><td>{order.status === "pending" && <button className="text-button" onClick={() => void cancel(order.id)}>Cancel</button>}</td></tr>)}</tbody></table></div>}
    </section>
  </main>;
}
