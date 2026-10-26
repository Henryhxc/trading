import { useCallback, useEffect, useRef, useState } from "react";
import { type Account, type Order, type Position, tradingApi } from "../../shared/api/trading";
import { aggregateProfit } from "./profit";
import { PortfolioAllocationChart, ProfitCalendar } from "./ProfitCalendar";
import { money } from "./usePortfolio";
import { useI18n } from "../../app/I18nProvider";

type Props = { account: Account | null; positions: Position[]; loading: boolean; error: string; refresh: () => Promise<void>; onSelect: (symbol: string) => void };

export function PortfolioPanel({ account, positions, loading, error, refresh, onSelect }: Props) {
  const { t } = useI18n();
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
  return <section className="portfolio-panel" aria-label={t("My portfolio")}>
    <div className="portfolio-heading"><div><p className="eyebrow">{t("YOUR PAPER ACCOUNT")}</p><h1>{t("My positions")}</h1></div><button className="quiet-button" onClick={() => void reload()}>{t("Refresh portfolio")}</button></div>
    {error && <p className="error-message" role="alert">{t(error)}<button className="quiet-button" onClick={() => void refresh()}>{t("Retry portfolio")}</button></p>}
    {loading && !account && <p className="panel-state" role="status">{t("Loading portfolio…")}</p>}
    <div className="portfolio-summary"><div><span>{t("Available cash")}</span><strong>{money(account?.available_cash)}</strong></div><div><span>{t("Cash balance")}</span><strong>{money(account?.cash_balance)}</strong></div><div><span>{t("Initial cash")}</span><strong>{money(account?.initial_cash)}</strong></div><div><span>{t("Account status")}</span><strong>{t(account?.status || "Unavailable")}</strong></div></div>
    <section className="data-card"><div className="panel-title"><h2>{t("Open positions")}</h2><span className="muted-count">{t("{count} securities", { count: positions.length })}</span></div>
      {positions.length === 0 ? <p className="empty-state" role="status">{t(loading ? "Loading positions…" : error ? "Positions are unavailable. Retry portfolio." : "No open positions yet. Filled paper orders will appear here.")}</p> : <div className="table-wrap"><table><thead><tr><th>{t("Security")}</th><th>{t("Qty / available")}</th><th>{t("Avg cost")}</th><th>{t("Market value")}</th><th>{t("P&L")}</th><th>{t("Quote status")}</th></tr></thead><tbody>{positions.map(position => <tr key={position.symbol}><td><button className="text-button" onClick={() => onSelect(position.symbol)}>{position.symbol}</button><small>{position.display_name}</small></td><td>{position.quantity.toLocaleString()} / {position.available_quantity.toLocaleString()}</td><td>{money(position.average_cost)}</td><td>{money(position.market_value)}</td><td className={position.unrealized_pnl == null ? "neutral" : position.unrealized_pnl >= 0 ? "up" : "down"}>{money(position.unrealized_pnl)}</td><td>{t(position.data_status)}</td></tr>)}</tbody></table></div>}
    </section>
    <PortfolioAllocationChart account={account} positions={positions} />
    <nav className="profit-period-navigation" aria-label={t("Profit period navigation")} role="tablist"><span>{t("Profit period")}</span><div>{(["day", "month", "year"] as const).map(value => <button key={value} role="tab" aria-selected={profitPeriod === value} className={profitPeriod === value ? "active" : ""} onClick={() => setProfitPeriod(value)}>{t(value)}</button>)}</div></nav>
    {profitPeriod === "day" && <ProfitCalendar title="Daily P/L" rows={dailyProfit} granularity="day" />}
    {profitPeriod === "month" && <ProfitCalendar title="Monthly P/L" rows={monthlyProfit} granularity="month" />}
    {profitPeriod === "year" && <ProfitCalendar title="Yearly P/L" rows={yearlyProfit} granularity="year" />}
  </section>;
}
