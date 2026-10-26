import { type FormEvent, useCallback, useEffect, useRef, useState } from "react";
import { type Order, type OrderFilters, tradingApi } from "../../shared/api/trading";
import { money } from "./usePortfolio";
import { useI18n } from "../../app/I18nProvider";

const date = (value: string | null, language: string) => value ? new Date(value).toLocaleString(language === "zh" ? "zh-HK" : "en-HK", { timeZone: "Asia/Hong_Kong" }) : "—";

export function OrderHistoryPanel() {
  const { language, t } = useI18n();
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
    catch (cause) { if (current === request.current) setError(cause instanceof Error ? cause.message : t("Order history is unavailable")); }
    finally { if (current === request.current) setLoading(false); }
  }, [applied]);
  useEffect(() => { void loadOrders(); return () => { request.current++; }; }, [loadOrders]);
  function applyFilters(event: FormEvent) { event.preventDefault(); setApplied({ ...filters, symbol: filters.symbol?.trim() }); }
  async function cancel(id: number) {
    try { await tradingApi.cancelOrder(id); await loadOrders(); }
    catch (cause) { setError(cause instanceof Error ? cause.message : t("Order could not be cancelled")); }
  }
  return <main className="prototype-page order-history-page" aria-label={t("Order history prototype")}>
    <header className="prototype-heading"><div><p className="eyebrow">{t("PAPER ACCOUNT · ORDER MANAGEMENT")}</p><h1>{t("Order history")}</h1><p className="prototype-subtitle">{t("Review, filter, and cancel pending paper orders.")}</p></div><span className="data-badge simulated">{t("SIMULATION")}</span></header>
    <section className="data-card order-history-card"><div className="panel-title"><div><h2>{t("All orders")}</h2><p className="panel-caption">{t("Orders are isolated to the signed-in paper account.")}</p></div><span className="muted-count">{t("{count} orders", { count: orders.length })}</span></div>
      <form className="filter-row" onSubmit={applyFilters}><input aria-label={t("Filter orders by symbol")} placeholder={t("Symbol")} value={filters.symbol || ""} onChange={event => setFilters({ ...filters, symbol: event.target.value })} /><select aria-label={t("Filter orders by side")} value={filters.side || ""} onChange={event => setFilters({ ...filters, side: event.target.value as OrderFilters["side"] })}><option value="">{t("All sides")}</option><option value="buy">{t("Buy")}</option><option value="sell">{t("Sell")}</option></select><select aria-label={t("Filter orders by status")} value={filters.status || ""} onChange={event => setFilters({ ...filters, status: event.target.value as OrderFilters["status"] })}><option value="">{t("All statuses")}</option><option value="pending">{t("Pending")}</option><option value="filled">{t("Filled")}</option><option value="cancelled">{t("Cancelled")}</option><option value="rejected">{t("Rejected")}</option></select><button className="quiet-button" type="submit">{t("Filter")}</button></form>
      {error && <p className="error-message" role="alert">{t(error)}<button className="quiet-button" onClick={() => void loadOrders()}>{t("Retry orders")}</button></p>}
      {loading ? <p className="panel-state" role="status">{t("Loading orders…")}</p> : orders.length === 0 ? <p className="empty-state" role="status">{t(error ? "Order history is unavailable." : "No orders match the current filters.")}</p> : <div className="table-wrap"><table><thead><tr><th>{t("Security")}</th><th>{t("Side / quantity")}</th><th>{t("Limit")}</th><th>{t("Status")}</th><th>{t("Created")}</th><th>{t("Actions")}</th></tr></thead><tbody>{orders.map(order => <tr key={order.id}><td><strong>{order.symbol}</strong><small>{t(order.order_type || "limit")}</small></td><td className={order.side === "buy" ? "up" : "down"}>{t(order.side)} {order.quantity.toLocaleString()}</td><td>{money(order.limit_price)}</td><td><span className={`order-status ${order.status}`}>{t(order.status)}</span>{order.reject_reason && <small>{t(order.reject_reason)}</small>}</td><td>{date(order.created_at, language)}</td><td>{order.status === "pending" && <button className="text-button" onClick={() => void cancel(order.id)}>{t("Cancel")}</button>}</td></tr>)}</tbody></table></div>}
    </section>
  </main>;
}
