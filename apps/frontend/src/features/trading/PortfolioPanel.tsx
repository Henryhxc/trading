import { type FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { type Account, type Position, type Order, type OrderFilters, tradingApi } from "../../shared/api/trading";
import { money } from "./usePortfolio";

type Props = { account: Account | null; positions: Position[]; loading: boolean; error: string; refresh: () => Promise<void>; onSelect: (symbol: string) => void };
const date = (value: string | null) => value ? new Date(value).toLocaleString() : "—";

export function PortfolioPanel({ account, positions, loading, error, refresh, onSelect }: Props) {
  const [orders, setOrders] = useState<Order[]>([]);
  const [filters, setFilters] = useState<OrderFilters>({});
  const [applied, setApplied] = useState<OrderFilters>({});
  const [orderLoading, setOrderLoading] = useState(true);
  const [orderError, setOrderError] = useState("");
  const request = useRef(0);
  const loadOrders = useCallback(async () => {
    const current = ++request.current;
    setOrderLoading(true); setOrderError("");
    try { const result = await tradingApi.orders(applied); if (current === request.current) setOrders(result?.items || []); }
    catch (cause) { if (current === request.current) setOrderError(cause instanceof Error ? cause.message : "Order history is unavailable"); }
    finally { if (current === request.current) setOrderLoading(false); }
  }, [applied]);
  useEffect(() => { void loadOrders(); return () => { request.current++; }; }, [loadOrders]);
  async function reload() { await Promise.all([refresh(), loadOrders()]); }
  async function cancel(id: number) {
    try { await tradingApi.cancelOrder(id); await reload(); }
    catch (cause) { setOrderError(cause instanceof Error ? cause.message : "Order could not be cancelled"); }
  }
  function applyFilters(event: FormEvent) { event.preventDefault(); setApplied({ ...filters, symbol: filters.symbol?.trim() }); }
  return <section className="portfolio-panel" aria-label="My portfolio">
    <div className="portfolio-heading"><div><p className="eyebrow">YOUR PAPER ACCOUNT</p><h1>My positions</h1></div><button className="quiet-button" onClick={() => void reload()}>Refresh portfolio</button></div>
    {error && <p className="error-message" role="alert">{error}<button className="quiet-button" onClick={() => void refresh()}>Retry portfolio</button></p>}
    {loading && !account && <p className="panel-state" role="status">Loading portfolio…</p>}
    <div className="portfolio-summary"><div><span>Available cash</span><strong>{money(account?.available_cash)}</strong></div><div><span>Cash balance</span><strong>{money(account?.cash_balance)}</strong></div><div><span>Initial cash</span><strong>{money(account?.initial_cash)}</strong></div><div><span>Account status</span><strong>{account?.status || "Unavailable"}</strong></div></div>
    <section className="data-card"><div className="panel-title"><h2>Open positions</h2><span className="muted-count">{positions.length} securities</span></div>
      {positions.length === 0 ? <p className="empty-state" role="status">{loading ? "Loading positions…" : error ? "Positions are unavailable. Retry portfolio." : "No open positions yet. Filled paper orders will appear here."}</p> : <div className="table-wrap"><table><thead><tr><th>Security</th><th>Qty / available</th><th>Avg cost</th><th>Market value</th><th>P&amp;L</th><th>Quote status</th></tr></thead><tbody>{positions.map(position => <tr key={position.symbol}><td><button className="text-button" onClick={() => onSelect(position.symbol)}>{position.symbol}</button><small>{position.display_name}</small></td><td>{position.quantity.toLocaleString()} / {position.available_quantity.toLocaleString()}</td><td>{money(position.average_cost)}</td><td>{money(position.market_value)}</td><td className={position.unrealized_pnl == null ? "neutral" : position.unrealized_pnl >= 0 ? "up" : "down"}>{money(position.unrealized_pnl)}</td><td>{position.data_status}</td></tr>)}</tbody></table></div>}
    </section>
    <section className="data-card"><div className="panel-title"><h2>Order history</h2><span className="muted-count">{orders.length} orders</span></div>
      <form className="filter-row" onSubmit={applyFilters}><input aria-label="Filter orders by symbol" placeholder="Symbol" value={filters.symbol || ""} onChange={event => setFilters({ ...filters, symbol: event.target.value })} /><select aria-label="Filter orders by side" value={filters.side || ""} onChange={event => setFilters({ ...filters, side: event.target.value as OrderFilters["side"] })}><option value="">All sides</option><option value="buy">Buy</option><option value="sell">Sell</option></select><select aria-label="Filter orders by status" value={filters.status || ""} onChange={event => setFilters({ ...filters, status: event.target.value as OrderFilters["status"] })}><option value="">All statuses</option><option value="pending">Pending</option><option value="filled">Filled</option><option value="cancelled">Cancelled</option><option value="rejected">Rejected</option></select><button className="quiet-button" type="submit">Filter</button></form>
      {orderError && <p className="error-message" role="alert">{orderError}<button className="quiet-button" onClick={() => void loadOrders()}>Retry orders</button></p>}
      {orderLoading ? <p className="panel-state" role="status">Loading orders…</p> : orders.length === 0 ? <p className="empty-state" role="status">{orderError ? "Order history is unavailable." : "No orders match the current filters."}</p> : <div className="table-wrap"><table><thead><tr><th>Security</th><th>Side / quantity</th><th>Limit</th><th>Status</th><th>Created</th><th>Action</th></tr></thead><tbody>{orders.map(order => <tr key={order.id}><td><strong>{order.symbol}</strong><small>{order.order_type}</small></td><td className={order.side === "buy" ? "up" : "down"}>{order.side.toUpperCase()} {order.quantity.toLocaleString()}</td><td>{money(order.limit_price)}</td><td><span className={`order-status ${order.status}`}>{order.status}</span>{order.reject_reason && <small>{order.reject_reason}</small>}</td><td>{date(order.created_at)}</td><td>{order.status === "pending" && <button className="text-button" onClick={() => void cancel(order.id)}>Cancel</button>}</td></tr>)}</tbody></table></div>}
    </section>
  </section>;
}
