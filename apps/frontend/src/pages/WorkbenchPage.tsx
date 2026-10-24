import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useSession } from "../app/SessionProvider";
import { navigate } from "../app/navigation";
import { marketApi, type Quote, type SymbolItem, type WatchlistItem } from "../shared/api/market";
import { MarketSidebar } from "../features/market/MarketSidebar";
import { QuotePanel } from "../features/market/QuotePanel";
import { PortfolioPanel } from "../features/trading/PortfolioPanel";
import { usePortfolio } from "../features/trading/usePortfolio";
import { TradingPanel } from "./TradingPanel";

type WorkbenchView = "market" | "favorites" | "positions";

export function WorkbenchPage() {
  const session = useSession();
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
      if (request === directoryRequest.current) setDirectoryError(error instanceof Error ? error.message : "Directory is unavailable");
    } finally { if (request === directoryRequest.current) setDirectoryLoading(false); }
  }, []);
  const loadWatchlist = useCallback(async () => {
    try { const result = await marketApi.watchlist(); setWatchlist(result?.items || []); setFavoriteError(""); }
    catch { setFavoriteError("Favorites could not be loaded. Retry directory to reconnect."); }
  }, []);
  const loadSymbol = useCallback(async (symbol: string) => {
    const request = ++quoteRequest.current;
    setLoading(true); setQuote(null); setMessage("");
    try { const result = await marketApi.quote(symbol); if (request === quoteRequest.current) setQuote(result); }
    catch (error) { if (request === quoteRequest.current) setMessage(error instanceof Error ? error.message : "Market data is unavailable"); }
    finally { if (request === quoteRequest.current) setLoading(false); }
  }, []);
  useEffect(() => {
    void loadDirectory(); void loadWatchlist();
    const timer = setInterval(() => { void loadDirectory(); void loadWatchlist(); }, 60_000);
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
  function switchView(next: WorkbenchView) { setView(next); setQuery(""); if (next === "positions") void portfolio.refresh(); }
  function select(symbol: string) { setActive(symbol); }
  function onScroll() {
    if (view === "favorites" && watchlistRef.current) sessionStorage.setItem("ux_watchlist_scroll_anchor", String(watchlistRef.current.scrollTop));
  }
  async function toggleFavorite() {
    try { if (isFavorite) await marketApi.removeWatchlist(active); else await marketApi.addWatchlist(active); await loadWatchlist(); }
    catch { setMessage("Favorite could not be updated"); }
  }
  async function moveFavorite(index: number, direction: -1 | 1) {
    const target = index + direction;
    if (target < 0 || target >= watchlist.length) return;
    const next = [...watchlist]; [next[index], next[target]] = [next[target], next[index]];
    setWatchlist(next);
    try { await marketApi.reorderWatchlist(next.map(item => item.symbol)); }
    catch { setMessage("Favorite order could not be updated"); await loadWatchlist(); }
  }
  async function signOut() {
    await fetch("/api/v1/auth/logout", { method: "POST", credentials: "include" });
    session.clearUser(); navigate("/login");
  }
  async function refreshAfterOrder() { await Promise.all([portfolio.refresh(), loadWatchlist(), loadDirectory()]); }
  return <main className="workbench-shell">
    <header className="workbench-header">
      <div className="terminal-brand"><span className="logo-plaque"><img src="/ux-securities-logo.png" alt="UX Securities Limited" /></span><span className="terminal-caption">HK EQUITIES<span>Personal trading terminal</span></span></div>
      <div className="header-actions"><span className="session-mode">PAPER ACCOUNT</span><span className="session-name">{session.user?.username}</span><button className="quiet-button" onClick={() => void signOut()}>Sign out</button></div>
    </header>
    <nav className="workbench-navigation" aria-label="Workspace sections" role="tablist">
      <button role="tab" aria-selected={view === "market"} className={view === "market" ? "active" : ""} onClick={() => switchView("market")}>Market <span>{items.length.toLocaleString()}</span></button>
      <button role="tab" aria-selected={view === "favorites"} className={view === "favorites" ? "active" : ""} onClick={() => switchView("favorites")}>My favorites <span>{watchlist.length}</span></button>
      <button role="tab" aria-selected={view === "positions"} className={view === "positions" ? "active" : ""} onClick={() => switchView("positions")}>My positions <span>{portfolio.positions.length}</span></button>
      <p className="navigation-note">HKD · Unofficial market data · Simulation only</p>
    </nav>
    {view === "positions" ? <PortfolioPanel {...portfolio} onSelect={symbol => { select(symbol); switchView("market"); }} /> : <section className="workbench-grid">
      <MarketSidebar view={view} query={query} setQuery={setQuery} items={items} watchlist={watchlist} owned={owned} active={active} loading={directoryLoading} error={directoryError || favoriteError} retry={() => { void loadDirectory(); void loadWatchlist(); }} select={select} move={(index, direction) => void moveFavorite(index, direction)} scrollRef={watchlistRef} onScroll={onScroll} />
      <QuotePanel active={active} item={items.find(item => item.symbol === active)} quote={quote} loading={loading} error={message} favorite={isFavorite} toggleFavorite={() => void toggleFavorite()} />
      <TradingPanel quote={quote} account={portfolio.account} accountError={portfolio.error} onRetry={() => void portfolio.refresh()} onOrderCreated={refreshAfterOrder} />
    </section>}
  </main>;
}
