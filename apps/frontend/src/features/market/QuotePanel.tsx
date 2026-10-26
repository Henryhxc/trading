import { type Quote, type SymbolItem } from "../../shared/api/market";
import { MarketChart } from "./MarketChart";
import { formatPercent, formatPrice } from "./MarketSidebar";
import { useI18n } from "../../app/I18nProvider";

type Props = { active: string; item?: SymbolItem; quote: Quote | null; loading: boolean; error: string; favorite: boolean; toggleFavorite: () => void };
export function QuotePanel({ active, item, quote, loading, error, favorite, toggleFavorite }: Props) {
  const { language, t } = useI18n();
  const change = quote?.change;
  return <section className="quote-panel" aria-label={t("Quote and chart")}>
    <div className="quote-heading"><div><p className="eyebrow">{active} · HKEX · HKD</p><h2>{quote?.display_name || item?.display_name || active}</h2><p className="quote-source">{quote ? `${quote.source} · ${new Date(quote.as_of).toLocaleString(language === "zh" ? "zh-HK" : "en-HK", { timeZone: "Asia/Hong_Kong" })}` : t("No cached spot quote")}</p></div><button className={`favorite-button ${favorite ? "saved" : ""}`} aria-label={t(favorite ? "Remove from favorites" : "Add to favorites")} onClick={toggleFavorite}>{favorite ? "★" : "☆"}</button></div>
    {error && <p className="error-message" role="alert">{t(error)}</p>}
    {loading ? <p className="panel-state" role="status">{t("Loading market data…")}</p> : <>
      <div className="quote-value"><strong>{formatPrice(quote?.last_price)}</strong><span className={change == null || change === 0 ? "neutral" : change > 0 ? "up" : "down"}>{change == null ? "—" : `${change > 0 ? "+" : ""}${change.toFixed(2)}`} ({formatPercent(quote?.change_percent)})</span><span className={`data-badge ${quote?.data_status || "unavailable"}`}>{t(quote?.data_status || "unavailable")}</span></div>
      <div className="quote-stats"><span>{t("Bid")} <b>{formatPrice(quote?.bid)}</b></span><span>{t("Ask")} <b>{formatPrice(quote?.ask)}</b></span><span>{t("Volume")} <b>{quote?.volume.toLocaleString() || "—"}</b></span><span>{t("Turnover")} <b>{quote ? quote.turnover.toLocaleString() : "—"}</b></span></div>
    </>}
    <MarketChart symbol={active} />
  </section>;
}
