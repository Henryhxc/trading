import { type FormEvent, useEffect, useState } from "react";
import { ApiError } from "../shared/api/client";
import { type Quote } from "../shared/api/market";
import { type Account, type OrderSide, tradingApi } from "../shared/api/trading";
import { money } from "../features/trading/usePortfolio";
import { useI18n } from "../app/I18nProvider";

type Props = { quote: Quote | null; account: Account | null; accountError: string; onRetry: () => void; onOrderCreated: () => Promise<void> };
const LOT_SIZE = 100;

export function TradingPanel({ quote, account, accountError, onRetry, onOrderCreated }: Props) {
  const { t } = useI18n();
  const [side, setSide] = useState<OrderSide>("buy");
  const [quantity, setQuantity] = useState("100");
  const [limitPrice, setLimitPrice] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState("");
  const [success, setSuccess] = useState("");
  useEffect(() => {
    setLimitPrice(quote?.last_price == null ? "" : quote.last_price.toFixed(2));
    setMessage(""); setSuccess("");
  }, [quote?.symbol, quote?.last_price]);
  async function submit(event: FormEvent) {
    event.preventDefault(); setMessage(""); setSuccess("");
    const parsedQuantity = Number(quantity); const parsedPrice = Number(limitPrice);
    if (!quote || !Number.isInteger(parsedQuantity) || parsedQuantity < 1 || !Number.isFinite(parsedPrice) || parsedPrice <= 0) {
      setMessage(t("Choose a security, whole-share quantity, and valid limit price.")); return;
    }
    if (side === "buy" && account?.available_cash != null) {
      const maxQuantityForOrder = Math.floor(account.available_cash / parsedPrice / LOT_SIZE) * LOT_SIZE;
      if (parsedQuantity > maxQuantityForOrder) {
        const shortfall = parsedQuantity * parsedPrice - account.available_cash;
        setMessage(t("Quantity exceeds the current buying limit of {quantity} shares. Shortfall: {amount}.", { quantity: maxQuantityForOrder.toLocaleString(), amount: money(shortfall) }));
        return;
      }
    }
    setSubmitting(true);
    try {
      const order = await tradingApi.createOrder({ symbol: quote.symbol, side, order_type: "limit", quantity: parsedQuantity, limit_price: parsedPrice });
      if (!order) throw new ApiError(t("Order response is unavailable"), 502);
      setSuccess(t("Order {status}. {message}", { status: t(order.status), message: t(order.reject_reason || "Server response received.") }));
      await onOrderCreated();
    } catch (error) { setMessage(t(error instanceof ApiError ? error.message : "Order could not be submitted")); }
    finally { setSubmitting(false); }
  }
  const estimate = Number(quantity) > 0 && Number(limitPrice) > 0 ? Number(quantity) * Number(limitPrice) : null;
  const parsedQuantity = Number(quantity);
  const parsedPrice = Number(limitPrice);
  const availableCash = account?.available_cash ?? null;
  const maxBuyQuantity = availableCash != null && parsedPrice > 0 ? Math.floor(availableCash / parsedPrice / LOT_SIZE) * LOT_SIZE : null;
  const maxBuyAmount = maxBuyQuantity != null ? maxBuyQuantity * parsedPrice : null;
  const buyShortfall = side === "buy" && estimate != null && availableCash != null && estimate > availableCash ? estimate - availableCash : 0;
  const exceedsBuyingPower = side === "buy" && maxBuyQuantity != null && Number.isInteger(parsedQuantity) && parsedQuantity > maxBuyQuantity;
  const invalidQuantity = !Number.isInteger(parsedQuantity) || parsedQuantity < 1;
  const blockedByBuyingPower = exceedsBuyingPower || buyShortfall > 0;
  return <aside className="order-ticket" aria-label={t("Order ticket")}>
    <div className="panel-title"><div><p className="eyebrow">{t("PAPER TRADING")}</p><h2>{t("Place paper order")}</h2></div><span className="data-badge simulated">SIM</span></div>
    <p className="order-context">{quote ? <><strong>{quote.symbol}</strong> {quote.display_name}</> : t("Select a security with a quote")}</p>
    <form onSubmit={submit} className="order-form">
      <div className="side-toggle" role="group" aria-label={t("Order side")}><button type="button" aria-pressed={side === "buy"} className={side === "buy" ? "selected buy" : "buy"} onClick={() => setSide("buy")}>{t("Buy")}</button><button type="button" aria-pressed={side === "sell"} className={side === "sell" ? "selected sell" : "sell"} onClick={() => setSide("sell")}>{t("Sell")}</button></div>
      <label>{t("Order type")}<select value="limit" onChange={() => {}}><option value="limit">{t("Limit order")}</option></select></label>
      <label>{t("Limit price (HKD)")}<input type="number" min="0.0001" step="0.0001" value={limitPrice} onChange={event => setLimitPrice(event.target.value)} /></label>
      <label>{t("Quantity")}<input type="number" min="1" step="1" value={quantity} onChange={event => setQuantity(event.target.value)} /></label>
      <div className="ticket-cash"><span>{t("Available cash")}</span><strong>{money(account?.available_cash)}</strong></div>
      {side === "buy" && maxBuyQuantity != null && <div className="buying-power"><span>{t("Max buy quantity")}</span><strong>{t("{shares} shares / {lots} lots", { shares: maxBuyQuantity.toLocaleString(), lots: (maxBuyQuantity / LOT_SIZE).toLocaleString() })}</strong><small>{t("Maximum estimated amount: {amount}. Based on HKD {price} × {lot} shares per lot", { amount: money(maxBuyAmount), price: parsedPrice.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }), lot: LOT_SIZE })}</small></div>}
      {accountError && <p className="error-message" role="alert">{t(accountError)}<button className="quiet-button" type="button" onClick={onRetry}>{t("Retry account")}</button></p>}
      <p className="estimate"><span>{t("Estimated amount")}</span><strong>{money(estimate)}</strong></p>
      {side === "buy" && blockedByBuyingPower && <p className="error-message buying-power-warning" role="alert">{t(invalidQuantity ? "Enter a whole-share quantity." : exceedsBuyingPower ? "Quantity exceeds the current buying limit of {quantity} shares. Shortfall: {amount}." : "Quantity exceeds buying power by {amount}.", { quantity: maxBuyQuantity?.toLocaleString() || "0", amount: money(buyShortfall) })}</p>}
      {quote?.data_status === "stale" && <p className="ticket-warning">{t("Quote is stale. The server may reject this order until data refreshes.")}</p>}
      {message && <p className="error-message" role="alert">{message}</p>}{success && <p className="success-message" role="status">{success}</p>}
      <button className={`primary-button full-width submit-${side}`} disabled={submitting || !quote || account?.status === "locked" || blockedByBuyingPower}>{t(submitting ? "Submitting…" : "Submit paper order")}</button>
      <p className="form-note">{t("Simulation only. Server checks lot size, market status, balance and quote freshness before matching.")}</p>
    </form>
    <div className="ticket-footer"><span className="data-badge simulated">{t("NO REAL EXECUTION")}</span><p>{t("Paper trades use a virtual HKD balance.")}</p></div>
  </aside>;
}
