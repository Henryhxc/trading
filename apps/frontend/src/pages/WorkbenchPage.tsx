import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSession } from "../app/SessionProvider";
import { navigate } from "../app/navigation";
import { marketApi, type Quote, type SymbolItem, type WatchlistItem } from "../shared/api/market";
import { MarketSidebar } from "../features/market/MarketSidebar";
import { QuotePanel } from "../features/market/QuotePanel";
import { PortfolioPanel } from "../features/trading/PortfolioPanel";
import { OrderHistoryPanel } from "../features/trading/OrderHistoryPanel";
import { FinancingPanel } from "../features/trading/FinancingPanel";
import { CashMovementsPanel } from "../features/funding/CashMovementsPanel";
import { usePortfolio } from "../features/trading/usePortfolio";
import { cashApi, type CashMovement } from "../shared/api/cash";
import { TradingPanel } from "./TradingPanel";
import { LanguageToggle, useI18n } from "../app/I18nProvider";
import { QuantAssistant } from "../features/assistant/QuantAssistant";
import { QuantEnginesPanel } from "../features/quant/QuantEnginesPanel";

type WorkbenchView = "market" | "favorites" | "positions" | "orders" | "financing" | "quant-engines" | "cash" | "deposit" | "withdraw" | "cash-history" | "exchange";

export function WorkbenchPage() {
  const session = useSession();
  const { t } = useI18n();
  const portfolio = usePortfolio();
  const [view, setView] = useState<WorkbenchView>("market");
  const [query, setQuery] = useState("");
  const [items, setItems] = useState<SymbolItem[]>([]);
  const [watchlist, setWatchlist] = useState<WatchlistItem[]>([]);
  const [directoryLoading, setDirectoryLoading] = useState(true);
  const [directoryError, setDirectoryError] = useState("");
  const [favoriteError, setFavoriteError] = useState("");
  const [quote, setQuote] = useState<Quote | null>(null);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [cashMovements, setCashMovements] = useState<CashMovement[]>([]);
  const [cashLoading, setCashLoading] = useState(false);
  const [cashError, setCashError] = useState("");
  const [active, setActive] = useState(() => sessionStorage.getItem("ux_active_symbol") || "00700");
  const quoteRequest = useRef(0);
  const directoryRequest = useRef(0);
  const watchlistRef = useRef<HTMLDivElement>(null);

  const loadDirectory = useCallback(async () => {
    const request = ++directoryRequest.current;
    setDirectoryLoading(true); setDirectoryError("");
    try {
      const result = await marketApi.search("");
      if (request === directoryRequest.current) setItems(result?.items || []);
    } catch (error) {
      if (request === directoryRequest.current) setDirectoryError(t(error instanceof Error ? error.message : "Directory is unavailable"));
    } finally { if (request === directoryRequest.current) setDirectoryLoading(false); }
  }, []);
  const loadWatchlist = useCallback(async () => {
    try { const result = await marketApi.watchlist(); setWatchlist(result?.items || []); setFavoriteError(""); }
    catch { setFavoriteError(t("Favorites could not be loaded. Retry directory to reconnect.")); }
  }, []);
  const loadSymbol = useCallback(async (symbol: string) => {
    const request = ++quoteRequest.current;
    setLoading(true); setQuote(null); setMessage("");
    try { const result = await marketApi.quote(symbol); if (request === quoteRequest.current) setQuote(result); }
    catch (error) { if (request === quoteRequest.current) setMessage(t(error instanceof Error ? error.message : "Market data is unavailable")); }
    finally { if (request === quoteRequest.current) setLoading(false); }
  }, []);
  const loadCashMovements = useCallback(async () => {
    setCashLoading(true); setCashError("");
    try { const result = await cashApi.list(); setCashMovements(result?.items || []); }
    catch (error) { setCashError(t(error instanceof Error ? error.message : "Cash mock data is unavailable")); }
    finally { setCashLoading(false); }
  }, []);
  useEffect(() => {
    void loadDirectory(); void loadWatchlist(); void loadCashMovements();
    const timer = setInterval(() => { void loadDirectory(); void loadWatchlist(); void loadCashMovements(); }, 60_000);
    return () => { clearInterval(timer); directoryRequest.current++; };
  }, [loadDirectory, loadWatchlist]);
  useEffect(() => {
    sessionStorage.setItem("ux_active_symbol", active);
    void loadSymbol(active);
    return () => { quoteRequest.current++; };
  }, [active, loadSymbol]);
  useEffect(() => {
    const anchor = Number(sessionStorage.getItem("ux_watchlist_scroll_anchor"));
    if (view !== "favorites" || !Number.isFinite(anchor)) return;
    requestAnimationFrame(() => { if (watchlistRef.current) watchlistRef.current.scrollTop = anchor; });
  }, [watchlist.length, view]);
  const owned = useMemo(() => new Set(portfolio.positions.map(item => item.symbol)), [portfolio.positions]);
  const isFavorite = watchlist.some(item => item.symbol === active);
  function switchView(next: WorkbenchView) { setView(next); setQuery(""); if (next === "positions" || next === "financing" || next === "cash") void portfolio.refresh(); if (next === "cash" || next === "deposit" || next === "withdraw" || next === "cash-history") void loadCashMovements(); }
  function select(symbol: string) { setActive(symbol); }
  function onScroll() {
    if (view === "favorites" && watchlistRef.current) sessionStorage.setItem("ux_watchlist_scroll_anchor", String(watchlistRef.current.scrollTop));
  }
  async function toggleFavorite() {
    try { if (isFavorite) await marketApi.removeWatchlist(active); else await marketApi.addWatchlist(active); await loadWatchlist(); }
    catch { setMessage(t("Favorite could not be updated")); }
  }
  async function moveFavorite(index: number, direction: -1 | 1) {
    const target = index + direction;
    if (target < 0 || target >= watchlist.length) return;
    const next = [...watchlist]; [next[index], next[target]] = [next[target], next[index]];
    setWatchlist(next);
    try { await marketApi.reorderWatchlist(next.map(item => item.symbol)); }
    catch { setMessage(t("Favorite order could not be updated")); await loadWatchlist(); }
  }
  async function signOut() {
    await fetch("/api/v1/auth/logout", { method: "POST", credentials: "include" });
    session.clearUser(); navigate("/login");
  }
  async function refreshAfterOrder() { await Promise.all([portfolio.refresh(), loadWatchlist(), loadDirectory()]); }
  return <main className="workbench-shell">
    <header className="workbench-header">
      <div className="terminal-brand"><span className="logo-plaque"><img src="/usmart-logo.png" alt="USMART 盈立證券" /></span><span className="terminal-caption">{t("HK EQUITIES")}<span>{t("Personal trading terminal")}</span></span></div>
      <div className="header-actions"><LanguageToggle /><span className="session-mode">{t("PAPER ACCOUNT")}</span><span className="session-name">{session.user?.username}</span><button className="quiet-button" onClick={() => void signOut()}>{t("Sign out")}</button></div>
    </header>
    <nav className="workbench-navigation" aria-label={t("Workspace sections")} role="tablist">
      <button role="tab" aria-selected={view === "market"} className={view === "market" ? "active" : ""} onClick={() => switchView("market")}>{t("Market")} <span>{items.length.toLocaleString()}</span></button>
      <button role="tab" aria-selected={view === "favorites"} className={view === "favorites" ? "active" : ""} onClick={() => switchView("favorites")}>{t("My favorites")} <span>{watchlist.length}</span></button>
      <button role="tab" aria-selected={view === "positions"} className={view === "positions" ? "active" : ""} onClick={() => switchView("positions")}>{t("My positions")} <span>{portfolio.positions.length}</span></button>
      <button role="tab" aria-selected={view === "orders"} className={view === "orders" ? "active" : ""} onClick={() => switchView("orders")}>{t("Order history")}</button>
      <button role="tab" aria-selected={view === "financing"} className={view === "financing" ? "active" : ""} onClick={() => switchView("financing")}>{t("Financing")}</button>
      <button role="tab" aria-selected={view === "quant-engines"} className={view === "quant-engines" ? "active" : ""} onClick={() => switchView("quant-engines")}>{t("Quant engines")}</button>
      <button role="tab" aria-selected={view === "cash"} className={view === "cash" ? "active" : ""} onClick={() => switchView("cash")}>{t("Cash overview")}</button>
      <button role="tab" aria-selected={view === "deposit"} className={view === "deposit" ? "active" : ""} onClick={() => switchView("deposit")}>{t("Deposit")}</button>
      <button role="tab" aria-selected={view === "withdraw"} className={view === "withdraw" ? "active" : ""} onClick={() => switchView("withdraw")}>{t("Withdraw")}</button>
      <button role="tab" aria-selected={view === "cash-history"} className={view === "cash-history" ? "active" : ""} onClick={() => switchView("cash-history")}>{t("Cash history")}</button>
      <button role="tab" aria-selected={view === "exchange"} className={view === "exchange" ? "active" : ""} onClick={() => switchView("exchange")}>{t("Exchange")}</button>
      <p className="navigation-note">{t("HKD · Unofficial market data · Simulation only")}</p>
    </nav>
    {view === "positions" ? <PortfolioPanel {...portfolio} onSelect={symbol => { select(symbol); switchView("market"); }} /> : view === "orders" ? <OrderHistoryPanel /> : view === "financing" ? <FinancingPanel account={portfolio.account} positions={portfolio.positions} /> : view === "quant-engines" ? <QuantEnginesPanel /> : view === "cash" || view === "deposit" || view === "withdraw" || view === "cash-history" || view === "exchange" ? <CashMovementsPanel mode={view === "cash" ? "overview" : view === "cash-history" ? "history" : view} movements={cashMovements} account={portfolio.account} loading={cashLoading} error={cashError} /> : <section className="workbench-grid">
      <MarketSidebar view={view === "favorites" ? "favorites" : "market"} query={query} setQuery={setQuery} items={items} watchlist={watchlist} owned={owned} active={active} loading={directoryLoading} error={directoryError || favoriteError} retry={() => { void loadDirectory(); void loadWatchlist(); }} select={select} move={(index, direction) => void moveFavorite(index, direction)} scrollRef={watchlistRef} onScroll={onScroll} />
      <QuotePanel active={active} item={items.find(item => item.symbol === active)} quote={quote} loading={loading} error={message} favorite={isFavorite} toggleFavorite={() => void toggleFavorite()} />
      <TradingPanel quote={quote} account={portfolio.account} accountError={portfolio.error} onRetry={() => void portfolio.refresh()} onOrderCreated={refreshAfterOrder} />
    </section>}
    <QuantAssistant />
  </main>;
}
