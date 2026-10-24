import { type FormEvent, useEffect, useState } from "react";
import { ApiError } from "../shared/api/client";
import { type Quote } from "../shared/api/market";
import { type Account, type OrderSide, tradingApi } from "../shared/api/trading";
import { money } from "../features/trading/usePortfolio";

type Props = { quote: Quote | null; account: Account | null; accountError: string; onRetry: () => void; onOrderCreated: () => Promise<void> };

export function TradingPanel({ quote, account, accountError, onRetry, onOrderCreated }: Props) {
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
      setMessage("Choose a security, whole-share quantity, and valid limit price."); return;
    }
    setSubmitting(true);
    try {
      const order = await tradingApi.createOrder({ symbol: quote.symbol, side, order_type: "limit", quantity: parsedQuantity, limit_price: parsedPrice });
      if (!order) throw new ApiError("Order response is unavailable", 502);
      setSuccess(`Order ${order.status}. ${order.reject_reason || "Server response received."}`);
      await onOrderCreated();
    } catch (error) { setMessage(error instanceof ApiError ? error.message : "Order could not be submitted"); }
    finally { setSubmitting(false); }
  }
  const estimate = Number(quantity) > 0 && Number(limitPrice) > 0 ? Number(quantity) * Number(limitPrice) : null;
  return <aside className="order-ticket" aria-label="Order ticket">
    <div className="panel-title"><div><p className="eyebrow">PAPER TRADING</p><h2>Place paper order</h2></div><span className="data-badge simulated">SIM</span></div>
    <p className="order-context">{quote ? <><strong>{quote.symbol}</strong> {quote.display_name}</> : "Select a security with a quote"}</p>
    <form onSubmit={submit} className="order-form">
      <div className="side-toggle" role="group" aria-label="Order side"><button type="button" aria-pressed={side === "buy"} className={side === "buy" ? "selected buy" : "buy"} onClick={() => setSide("buy")}>Buy</button><button type="button" aria-pressed={side === "sell"} className={side === "sell" ? "selected sell" : "sell"} onClick={() => setSide("sell")}>Sell</button></div>
      <label>Order type<select value="limit" onChange={() => {}}><option value="limit">Limit order</option></select></label>
      <label>Limit price (HKD)<input type="number" min="0.0001" step="0.0001" value={limitPrice} onChange={event => setLimitPrice(event.target.value)} /></label>
      <label>Quantity<input type="number" min="1" step="1" value={quantity} onChange={event => setQuantity(event.target.value)} /></label>
      <div className="ticket-cash"><span>Available cash</span><strong>{money(account?.available_cash)}</strong></div>
      {accountError && <p className="error-message" role="alert">{accountError}<button className="quiet-button" type="button" onClick={onRetry}>Retry account</button></p>}
      <p className="estimate"><span>Estimated amount</span><strong>{money(estimate)}</strong></p>
      {quote?.data_status === "stale" && <p className="ticket-warning">Quote is stale. The server may reject this order until data refreshes.</p>}
      {message && <p className="error-message" role="alert">{message}</p>}{success && <p className="success-message" role="status">{success}</p>}
      <button className={`primary-button full-width submit-${side}`} disabled={submitting || !quote || account?.status === "locked"}>{submitting ? "Submitting…" : "Submit paper order"}</button>
      <p className="form-note">Simulation only. Server checks lot size, market status, balance and quote freshness before matching.</p>
    </form>
    <div className="ticket-footer"><span className="data-badge simulated">NO REAL EXECUTION</span><p>Paper trades use a virtual HKD balance.</p></div>
  </aside>;
}
