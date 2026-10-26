import { useMemo, useState } from "react";
import type { Account } from "../../shared/api/trading";
import type { CashMovement } from "../../shared/api/cash";
import { useI18n } from "../../app/I18nProvider";

type Mode = "overview" | "deposit" | "withdraw" | "history" | "exchange";
type MovementTypeFilter = "all" | CashMovement["type"];
type MovementStatusFilter = "all" | CashMovement["status"];
type Props = { mode: Mode; movements: CashMovement[]; account: Account | null; loading?: boolean; error?: string };
type CurrencyCode = "HKD" | "USD" | "CNY";
type CurrencyBalance = { code: CurrencyCode; name: string; balance: number; available: number; hkdValue: number; color: string };

const cash = (value: number) => `HKD ${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
const nativeCash = (code: CurrencyCode, value: number) => `${code} ${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
const movementLabel = (type: CashMovement["type"]) => type === "deposit" ? "Deposit" : "Withdrawal";
const ratesToHKD: Record<CurrencyCode, number> = { HKD: 1, USD: 7.8, CNY: 1.08 };

function getCurrencyBalances(account: Account | null): CurrencyBalance[] {
  const hkd = account?.cash_balance || 0;
  return [
    { code: "HKD", name: "Hong Kong Dollar", balance: hkd, available: account?.available_cash || 0, hkdValue: hkd, color: "#5478ed" },
    { code: "USD", name: "US Dollar", balance: 12500, available: 12500, hkdValue: 12500 * ratesToHKD.USD, color: "#40d794" },
    { code: "CNY", name: "Renminbi", balance: 68000, available: 68000, hkdValue: 68000 * ratesToHKD.CNY, color: "#f3c76b" },
  ];
}

export function CashMovementsPanel({ mode, movements, account, loading = false, error = "" }: Props) {
  const { t } = useI18n();
  const [amount, setAmount] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [typeFilter, setTypeFilter] = useState<MovementTypeFilter>("all");
  const [statusFilter, setStatusFilter] = useState<MovementStatusFilter>("all");
  const title = t(mode === "deposit" ? "Deposit funds" : mode === "withdraw" ? "Withdraw funds" : mode === "history" ? "Cash movement history" : mode === "exchange" ? "Exchange currencies" : "Cash overview");
  const filteredMovements = useMemo(() => movements.filter(item => (typeFilter === "all" || item.type === typeFilter) && (statusFilter === "all" || item.status === statusFilter)), [movements, statusFilter, typeFilter]);
  const pendingCount = movements.filter(item => item.status === "pending").length;
  const completedDeposits = movements.filter(item => item.type === "deposit" && item.status === "completed").length;
  const completedWithdrawals = movements.filter(item => item.type === "withdraw" && item.status === "completed").length;
  const netMovement = movements.reduce((total, item) => item.status === "completed" ? total + (item.type === "deposit" ? item.amount : -item.amount) : total, 0);
  const isForm = mode === "deposit" || mode === "withdraw";

  if (loading) return <main className="prototype-page" aria-label={t("Cash prototype")}><p className="panel-state" role="status">{t("Loading cash mock data…")}</p></main>;
  if (error) return <main className="prototype-page" aria-label={t("Cash prototype")}><p className="error-message" role="alert">{t(error)}</p></main>;

  return <main className="prototype-page" aria-label={t("{title} prototype", { title })}>
    <header className="prototype-heading"><div><p className="eyebrow">{t("PAPER ACCOUNT · DATABASE MOCK")}</p><h1>{title}</h1><p className="prototype-subtitle">{t("Prototype only. No real money movement or bank connection is performed.")}</p></div><span className="data-badge simulated">{t("SIMULATION")}</span></header>
    {mode === "overview" && <Overview movements={movements} account={account} pendingCount={pendingCount} completedDeposits={completedDeposits} completedWithdrawals={completedWithdrawals} netMovement={netMovement} />}
    {mode === "history" && <History movements={filteredMovements} total={movements.length} typeFilter={typeFilter} statusFilter={statusFilter} setTypeFilter={setTypeFilter} setStatusFilter={setStatusFilter} />}
    {isForm && <FundingForm mode={mode} amount={amount} submitted={submitted} account={account} setAmount={value => { setAmount(value); setSubmitted(false); }} setSubmitted={setSubmitted} />}
    {mode === "exchange" && <ExchangeForm />}
  </main>;
}

function Overview({ movements, account, pendingCount, completedDeposits, completedWithdrawals, netMovement }: { movements: CashMovement[]; account: Account | null; pendingCount: number; completedDeposits: number; completedWithdrawals: number; netMovement: number }) {
  const { t } = useI18n();
  const balances = getCurrencyBalances(account);
  const totalHkd = balances.reduce((total, item) => total + item.hkdValue, 0) || 1;
  const pie = balances.map(item => `${item.color} ${((item.hkdValue / totalHkd) * 100).toFixed(2)}%`).join(", ");
  return <>
    <div className="prototype-metrics"><div><span>{t("Available cash")}</span><strong>{cash(account?.available_cash || 0)}</strong></div><div><span>{t("Cash balance")}</span><strong>{cash(account?.cash_balance || 0)}</strong></div><div><span>{t("Pending requests")}</span><strong>{pendingCount}</strong></div></div>
    <div className="cash-overview-grid">
      <section className="data-card cash-summary-card"><div className="panel-title"><div><h2>{t("Account snapshot")}</h2><p className="panel-caption">{t("Settled cash activity in this paper account")}</p></div><span className="muted-count">HKD</span></div><dl className="cash-account-list"><div><dt>{t("Net completed movement")}</dt><dd className={netMovement >= 0 ? "up" : "down"}>{netMovement >= 0 ? "+" : "-"}{cash(Math.abs(netMovement))}</dd></div><div><dt>{t("Completed deposits")}</dt><dd>{completedDeposits}</dd></div><div><dt>{t("Completed withdrawals")}</dt><dd>{completedWithdrawals}</dd></div><div><dt>{t("Unsettled requests")}</dt><dd>{pendingCount}</dd></div></dl></section>
      <section className="data-card cash-recent-card"><div className="panel-title"><div><h2>{t("Recent activity")}</h2><p className="panel-caption">{t("Latest records from MySQL")}</p></div><span className="muted-count">{t("{count} latest", { count: Math.min(movements.length, 3) })}</span></div><MovementList movements={movements.slice(0, 3)} /></section>
    </div>
    <div className="currency-overview-grid">
      <section className="data-card currency-balances-card"><div className="panel-title"><div><h2>{t("Currency balances")}</h2><p className="panel-caption">{t("Prototype balances converted to HKD")}</p></div><span className="muted-count">{t("{count} currencies", { count: balances.length })}</span></div><div className="currency-balances">{balances.map(item => <div className="currency-balance-row" key={item.code}><span className="currency-swatch" style={{ backgroundColor: item.color }} aria-hidden="true" /><div className="currency-balance-identity"><strong>{item.code}</strong><small>{t(item.name)}</small></div><div className="currency-balance-values"><strong>{nativeCash(item.code, item.balance)}</strong><small>{t("{value} equivalent · {available} available", { value: nativeCash("HKD", item.hkdValue), available: nativeCash(item.code, item.available) })}</small></div></div>)}</div></section>
      <section className="data-card currency-allocation-card"><div className="panel-title"><div><h2>{t("Currency allocation")}</h2><p className="panel-caption">{t("Share of total cash at mock reference rates")}</p></div><span className="muted-count">{t("HKD basis")}</span></div><div className="currency-allocation"><div className="currency-allocation-pie" role="img" aria-label={t("Currency allocation")} style={{ background: `conic-gradient(${pie})` }}><span>{nativeCash("HKD", totalHkd)}</span></div><div className="currency-allocation-legend">{balances.map(item => <div key={item.code}><span className="currency-swatch" style={{ backgroundColor: item.color }} aria-hidden="true" /><span>{t("{currency} allocation", { currency: item.code })}</span><strong>{((item.hkdValue / totalHkd) * 100).toFixed(1)}%</strong></div>)}</div></div></section>
    </div>
  </>;
}

function ExchangeForm() {
  const { t } = useI18n();
  const [from, setFrom] = useState<CurrencyCode>("HKD");
  const [to, setTo] = useState<CurrencyCode>("USD");
  const [amount, setAmount] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const numericAmount = Number(amount);
  const validAmount = Number.isFinite(numericAmount) && numericAmount > 0 && from !== to;
  const receive = validAmount ? numericAmount * ratesToHKD[from] / ratesToHKD[to] : 0;
  const options = (selected: CurrencyCode) => (["HKD", "USD", "CNY"] as CurrencyCode[]).filter(code => code !== selected);
  return <div className="exchange-layout">
    <section className="data-card exchange-form-card"><div className="panel-title"><div><p className="panel-kicker">{t("CURRENCY CONVERSION")}</p><h2>{t("Exchange currencies")}</h2></div><span className="data-badge simulated">{t("MOCK REQUEST")}</span></div><form onSubmit={event => { event.preventDefault(); if (validAmount) setSubmitted(true); }}>
      <div className="exchange-fields"><label>{t("From currency")}<select aria-label={t("From currency")} value={from} onChange={event => { setFrom(event.target.value as CurrencyCode); setSubmitted(false); }}>{["HKD", "USD", "CNY"].map(code => <option key={code}>{code}</option>)}</select></label><button className="exchange-swap" type="button" aria-label={t("Swap currencies")} onClick={() => { setFrom(to); setTo(from); setSubmitted(false); }}>⇄</button><label>{t("To currency")}<select aria-label={t("To currency")} value={to} onChange={event => { setTo(event.target.value as CurrencyCode); setSubmitted(false); }}>{options(from).map(code => <option key={code}>{code}</option>)}</select></label></div>
      <label>{t("Exchange amount")}<input aria-label={t("Exchange amount")} inputMode="decimal" type="number" min="0" step="0.01" value={amount} onChange={event => { setAmount(event.target.value); setSubmitted(false); }} placeholder="0.00" /></label>
      <div className="exchange-preview"><div><span>{t("Exchange rate")}</span><strong>1 {from} = {(ratesToHKD[from] / ratesToHKD[to]).toFixed(4)} {to}</strong></div><div><span>{t("Estimated receive {currency}", { currency: to })}</span><strong>{nativeCash(to, receive)}</strong></div><div><span>{t("Prototype fee")}</span><strong>HKD 0.00</strong></div></div>
      <button className="primary-button full-width" type="submit" disabled={!validAmount}>{t("Preview exchange")}</button>{from === to && <p className="ticket-warning" role="alert">{t("Choose two different currencies to preview a conversion.")}</p>}{submitted && <p className="success-message" role="status">{t("Exchange request staged in prototype. No real conversion was executed.")}</p>}
    </form></section>
    <aside className="data-card exchange-context"><div className="panel-title"><div><p className="panel-kicker">{t("REFERENCE RATES")}</p><h2>{t("Three supported currencies")}</h2></div><span className="muted-count">{t("SIMULATED")}</span></div><dl className="exchange-rate-list">{(["HKD", "USD", "CNY"] as CurrencyCode[]).map(code => <div key={code}><dt>{code}</dt><dd>{nativeCash("HKD", ratesToHKD[code])}</dd></div>)}</dl><p className="panel-caption">{t("Rates and balances are mock data for the prototype. This action does not alter your paper account.")}</p></aside>
  </div>;
}

function History({ movements, total, typeFilter, statusFilter, setTypeFilter, setStatusFilter }: { movements: CashMovement[]; total: number; typeFilter: MovementTypeFilter; statusFilter: MovementStatusFilter; setTypeFilter: (value: MovementTypeFilter) => void; setStatusFilter: (value: MovementStatusFilter) => void }) {
  const { t } = useI18n();
  return <section className="data-card history-card"><div className="panel-title"><div><h2>{t("Cash movement history")}</h2><p className="panel-caption">{t("Complete deposit and withdrawal ledger")}</p></div><span className="muted-count">{t("{shown} of {total} records", { shown: movements.length, total })}</span></div><div className="filter-row movement-filter-row"><label>{t("Type")}<select aria-label={t("Movement type")} value={typeFilter} onChange={event => setTypeFilter(event.target.value as MovementTypeFilter)}><option value="all">{t("All types")}</option><option value="deposit">{t("Deposits")}</option><option value="withdraw">{t("Withdrawals")}</option></select></label><label>{t("Status")}<select aria-label={t("Movement status")} value={statusFilter} onChange={event => setStatusFilter(event.target.value as MovementStatusFilter)}><option value="all">{t("All statuses")}</option><option value="completed">{t("Completed")}</option><option value="pending">{t("Pending")}</option></select></label><button className="quiet-button" type="button" onClick={() => { setTypeFilter("all"); setStatusFilter("all"); }}>{t("Clear filters")}</button></div>{movements.length > 0 ? <MovementTable movements={movements} /> : <p className="empty-state" role="status">{t("No cash movements match these filters.")}</p>}</section>;
}

function FundingForm({ mode, amount, submitted, account, setAmount, setSubmitted }: { mode: "deposit" | "withdraw"; amount: string; submitted: boolean; account: Account | null; setAmount: (value: string) => void; setSubmitted: (value: boolean) => void }) {
  const { t } = useI18n();
  const deposit = mode === "deposit";
  const validAmount = Number(amount) > 0;
  return <div className="funding-layout">
    <section className="data-card movement-form-card"><div className="panel-title"><div><p className="panel-kicker">{t(deposit ? "ADD CASH" : "MOVE CASH OUT")}</p><h2>{t(deposit ? "Deposit funds" : "Withdraw funds")}</h2></div><span className="data-badge simulated">{t("MOCK REQUEST")}</span></div><form onSubmit={event => { event.preventDefault(); if (validAmount) setSubmitted(true); }}>
      <label>{t(deposit ? "Deposit amount" : "Withdrawal amount")}<input aria-label={t(deposit ? "Deposit amount" : "Withdrawal amount")} inputMode="decimal" type="number" min="1" step="0.01" value={amount} onChange={event => setAmount(event.target.value)} placeholder="0.00" /></label>
      {deposit ? <><label>{t("Funding method")}<select aria-label={t("Funding method")}><option>{t("Bank transfer (mock)")}</option><option>{t("FPS (mock)")}</option></select></label><label>{t("Source account")}<input aria-label={t("Source account")} value={t("Demo bank account ·•••• 0210")} readOnly /></label></> : <><label>{t("Receiving account")}<select aria-label={t("Receiving account")}><option>{t("Demo bank account ·•••• 0210")}</option><option>{t("FPS beneficiary (mock)")}</option></select></label><div className="fee-preview"><span>{t("Estimated fee")}</span><strong>HKD 0.00</strong></div></>}
      <button className="primary-button full-width" type="submit" disabled={!validAmount}>{t(deposit ? "Submit deposit" : "Submit withdrawal")}</button>{submitted && <p className="success-message" role="status">{t("{type} request staged in prototype. No real transfer was sent.", { type: t(deposit ? "Deposit" : "Withdrawal") })}</p>}
    </form></section>
    <aside className="data-card funding-context"><div className="panel-title"><div><p className="panel-kicker">{t("ACCOUNT CONTEXT")}</p><h2>{t(deposit ? "Ready to fund" : "Available to withdraw")}</h2></div><span className="muted-count">HKD</span></div><div className="funding-context-value"><span>{t(deposit ? "Current cash balance" : "Available cash")}</span><strong>{cash(deposit ? account?.cash_balance || 0 : account?.available_cash || 0)}</strong></div><p className="panel-caption">{t("This prototype stages a request for review only. No bank connection or real money movement is performed.")}</p></aside>
  </div>;
}

function MovementList({ movements }: { movements: CashMovement[] }) {
  const { language, t } = useI18n();
  if (movements.length === 0) return <p className="empty-state" role="status">{t("No cash activity yet.")}</p>;
  return <ul className="cash-activity-list" aria-label={t("Recent cash activity")}>{movements.map(item => <li key={item.id}><div><strong>{t(movementLabel(item.type))}</strong><small>{new Date(item.created_at).toLocaleDateString(language === "zh" ? "zh-HK" : "en-HK", { timeZone: "Asia/Hong_Kong" })} · {t(item.method)}</small></div><div className="cash-activity-value"><strong className={item.type === "deposit" ? "up" : "down"}>{item.type === "deposit" ? "+" : "-"}{cash(item.amount)}</strong><span className={`order-status ${item.status}`}>{t(item.status)}</span></div></li>)}</ul>;
}

function MovementTable({ movements }: { movements: CashMovement[] }) {
  const { language, t } = useI18n();
  return <div className="table-wrap"><table aria-label={t("Cash movement history")}><thead><tr><th>{t("Date")}</th><th>{t("Type")}</th><th>{t("Amount")}</th><th>{t("Method")}</th><th>{t("Status")}</th></tr></thead><tbody>{movements.map(item => <tr key={item.id}><td>{new Date(item.created_at).toLocaleString(language === "zh" ? "zh-HK" : "en-HK", { timeZone: "Asia/Hong_Kong" })}</td><td className={item.type === "deposit" ? "up" : "down"}>{t(movementLabel(item.type))}</td><td>{cash(item.amount)}</td><td>{t(item.method)}</td><td><span className={`order-status ${item.status}`}>{t(item.status)}</span></td></tr>)}</tbody></table></div>;
}
