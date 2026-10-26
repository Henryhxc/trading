import { type RefObject } from "react";
import { type SymbolItem, type WatchlistItem } from "../../shared/api/market";
import { useI18n } from "../../app/I18nProvider";

type Props = {
  view: "market" | "favorites"; query: string; setQuery: (value: string) => void;
  items: SymbolItem[]; watchlist: WatchlistItem[]; owned: Set<string>; active: string;
  loading: boolean; error: string; retry: () => void; select: (symbol: string) => void;
  move: (index: number, direction: -1 | 1) => void; scrollRef: RefObject<HTMLDivElement | null>; onScroll: () => void;
};
export const formatPrice = (value: number | null | undefined) => value == null ? "—" : value.toFixed(2);
export const formatPercent = (value: number | null | undefined) => value == null ? "—" : `${value > 0 ? "+" : ""}${value.toFixed(2)}%`;

export function MarketSidebar({ view, query, setQuery, items, watchlist, owned, active, loading, error, retry, select, move, scrollRef, onScroll }: Props) {
  const { t } = useI18n();
  const favorites = new Map(watchlist.map(item => [item.symbol, item]));
  const all = view === "favorites" ? watchlist : items;
  const filter = query.trim().toLocaleLowerCase();
  const visible = all.filter(item => `${item.symbol} ${item.display_name}`.toLocaleLowerCase().includes(filter));
  return <aside className="market-sidebar" aria-label={t("Market watch")}>
    <div className="sidebar-heading"><p className="eyebrow">{t("HK EQUITIES")}</p><h1>{t(view === "market" ? "Market overview" : "My favorites")}</h1></div>
    <div className="search-row"><span aria-hidden="true">⌕</span><input type="search" aria-label={t("Search symbols")} placeholder={t("Filter code or company")} value={query} onChange={event => setQuery(event.target.value)} />{query && <button type="button" aria-label={t("Clear search")} onClick={() => setQuery("")}>×</button>}</div>
    <div className="directory-heading"><span>{t(view === "market" ? "All HKD equities" : "Saved securities")} <b>{visible.length.toLocaleString()}</b></span><span>{t("Last / Change")}</span></div>
    <div className="directory-legend"><span className="favorite-marker">★ {t("Favorite")}</span><span className="position-marker">◆ {t("Position")}</span></div>
    {error && <p className="error-message" role="alert">{t(error)}<button className="quiet-button" onClick={retry}>{t("Retry directory")}</button></p>}
    <div className="securities-list" role="region" aria-label={t("Securities directory")} ref={scrollRef} onScroll={onScroll}>
      {loading && items.length === 0 && view === "market" ? <p className="panel-state" role="status">{t("Loading securities…")}</p> : visible.map((item, index) => {
        const quote = favorites.get(item.symbol)?.quote || item.quote;
        const change = quote?.change_percent;
        const tone = change == null || change === 0 ? "neutral" : change > 0 ? "up" : "down";
        return <div className={`security-row ${item.symbol === active ? "active" : ""}`} key={item.symbol}>
          <button className="security-select" aria-pressed={item.symbol === active} onClick={() => select(item.symbol)}>
            <span className="security-identity"><strong>{item.symbol}</strong><span className="security-name">{item.display_name}</span><span className="security-markers">{favorites.has(item.symbol) && <small className="favorite-marker">★ {t("Favorite")}</small>}{owned.has(item.symbol) && <small className="position-marker">◆ {t("Position")}</small>}</span></span>
            <span className="security-price"><strong>{formatPrice(quote?.last_price)}</strong><span className={tone}>{formatPercent(change)}</span><small>{t(quote?.data_status || "No quote")}</small></span>
          </button>
          {view === "favorites" && !filter && <span className="watch-order-actions"><button aria-label={t("Move {symbol} up", { symbol: item.symbol })} disabled={index === 0} onClick={() => move(index, -1)}>↑</button><button aria-label={t("Move {symbol} down", { symbol: item.symbol })} disabled={index === visible.length - 1} onClick={() => move(index, 1)}>↓</button></span>}
        </div>;
      })}
      {!loading && !error && visible.length === 0 && <p className="empty-state" role="status">{filter ? t("No securities match {query}.", { query: query.trim() }) : t("No favorites yet. Select a market security and add its star.")}</p>}
    </div>
    <p className="directory-footnote">{t("Static securities directory · HKD equities. Unavailable prices stay blank.")}<br />{t("HKEX file declares 26 Oct 2026, later than this import on 24 Oct; directory metadata is not a current trading-status snapshot.")}</p>
  </aside>;
}
