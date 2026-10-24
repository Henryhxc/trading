import { useCallback, useEffect, useRef, useState } from "react";
import { type Account, type Order, type Position, tradingApi } from "../../shared/api/trading";
import { aggregateProfit } from "./profit";
import { PortfolioAllocationChart, ProfitCalendar } from "./ProfitCalendar";
import { money } from "./usePortfolio";

type Props = { account: Account | null; positions: Position[]; loading: boolean; error: string; refresh: () => Promise<void>; onSelect: (symbol: string) => void };

export function PortfolioPanel({ account, positions, loading, error, refresh, onSelect }: Props) {
  const [profitPeriod, setProfitPeriod] = useState<"day" | "month" | "year">("day");
  const [orders, setOrders] = useState<Order[]>([]);
  const request = useRef(0);
  const loadOrderAnalytics = useCallback(async () => {
    const current = ++request.current;
    try { const result = await tradingApi.orders(); if (current === request.current) setOrders(result?.items || []); }
    catch { if (current === request.current) setOrders([]); }
  }, []);
  useEffect(() => { void loadOrderAnalytics(); return () => { request.current++; }; }, [loadOrderAnalytics]);
  async function reload() { await refresh(); }
  const dailyProfit = aggregateProfit(orders, "day");
  const monthlyProfit = aggregateProfit(orders, "month");
  const yearlyProfit = aggregateProfit(orders, "year");
  return <section className="portfolio-panel" aria-label="My portfolio">
    <div className="portfolio-heading"><div><p className="eyebrow">YOUR PAPER ACCOUNT</p><h1>My positions</h1></div><button className="quiet-button" onClick={() => void reload()}>Refresh portfolio</button></div>
    {error && <p className="error-message" role="alert">{error}<button className="quiet-button" onClick={() => void refresh()}>Retry portfolio</button></p>}
    {loading && !account && <p className="panel-state" role="status">Loading portfolio…</p>}
    <div className="portfolio-summary"><div><span>Available cash</span><strong>{money(account?.available_cash)}</strong></div><div><span>Cash balance</span><strong>{money(account?.cash_balance)}</strong></div><div><span>Initial cash</span><strong>{money(account?.initial_cash)}</strong></div><div><span>Account status</span><strong>{account?.status || "Unavailable"}</strong></div></div>
    <section className="data-card"><div className="panel-title"><h2>Open positions</h2><span className="muted-count">{positions.length} securities</span></div>
      {positions.length === 0 ? <p className="empty-state" role="status">{loading ? "Loading positions…" : error ? "Positions are unavailable. Retry portfolio." : "No open positions yet. Filled paper orders will appear here."}</p> : <div className="table-wrap"><table><thead><tr><th>Security</th><th>Qty / available</th><th>Avg cost</th><th>Market value</th><th>P&amp;L</th><th>Quote status</th></tr></thead><tbody>{positions.map(position => <tr key={position.symbol}><td><button className="text-button" onClick={() => onSelect(position.symbol)}>{position.symbol}</button><small>{position.display_name}</small></td><td>{position.quantity.toLocaleString()} / {position.available_quantity.toLocaleString()}</td><td>{money(position.average_cost)}</td><td>{money(position.market_value)}</td><td className={position.unrealized_pnl == null ? "neutral" : position.unrealized_pnl >= 0 ? "up" : "down"}>{money(position.unrealized_pnl)}</td><td>{position.data_status}</td></tr>)}</tbody></table></div>}
    </section>
    <PortfolioAllocationChart account={account} positions={positions} />
    <nav className="profit-period-navigation" aria-label="Profit period navigation" role="tablist"><span>Profit period</span><div>{([['day', '日'], ['month', '月'], ['year', '年']] as const).map(([value, label]) => <button key={value} role="tab" aria-selected={profitPeriod === value} className={profitPeriod === value ? "active" : ""} onClick={() => setProfitPeriod(value)}>{label}</button>)}</div></nav>
    {profitPeriod === "day" && <ProfitCalendar title="Daily P/L" rows={dailyProfit} granularity="day" />}
    {profitPeriod === "month" && <ProfitCalendar title="Monthly P/L" rows={monthlyProfit} granularity="month" />}
    {profitPeriod === "year" && <ProfitCalendar title="Yearly P/L" rows={yearlyProfit} granularity="year" />}
  </section>;
}
