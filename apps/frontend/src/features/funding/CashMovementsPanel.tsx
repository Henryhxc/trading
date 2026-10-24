import { useMemo, useState } from "react";
import type { Account } from "../../shared/api/trading";
import type { CashMovement } from "../../shared/api/cash";

type Mode = "overview" | "deposit" | "withdraw" | "history" | "exchange";
type MovementTypeFilter = "all" | CashMovement["type"];
type MovementStatusFilter = "all" | CashMovement["status"];
type Props = { mode: Mode; movements: CashMovement[]; account: Account | null; loading?: boolean; error?: string };
type CurrencyCode = "HKD" | "USD" | "CNY";
type CurrencyBalance = { code: CurrencyCode; name: string; balance: number; available: number; hkdValue: number; color: string };

const cash = (value: number) => `HKD ${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
const nativeCash = (code: CurrencyCode, value: number) => `${code} ${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
const date = (value: string) => new Date(value).toLocaleDateString();
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
  const [amount, setAmount] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const [typeFilter, setTypeFilter] = useState<MovementTypeFilter>("all");
  const [statusFilter, setStatusFilter] = useState<MovementStatusFilter>("all");
  const title = mode === "deposit" ? "Deposit funds" : mode === "withdraw" ? "Withdraw funds" : mode === "history" ? "Cash movement history" : mode === "exchange" ? "Exchange currencies" : "Cash overview";
  const filteredMovements = useMemo(() => movements.filter(item => (typeFilter === "all" || item.type === typeFilter) && (statusFilter === "all" || item.status === statusFilter)), [movements, statusFilter, typeFilter]);
  const pendingCount = movements.filter(item => item.status === "pending").length;
  const completedDeposits = movements.filter(item => item.type === "deposit" && item.status === "completed").length;
  const completedWithdrawals = movements.filter(item => item.type === "withdraw" && item.status === "completed").length;
  const netMovement = movements.reduce((total, item) => item.status === "completed" ? total + (item.type === "deposit" ? item.amount : -item.amount) : total, 0);
  const isForm = mode === "deposit" || mode === "withdraw";

  if (loading) return <main className="prototype-page" aria-label="Cash prototype"><p className="panel-state" role="status">Loading cash mock data…</p></main>;
  if (error) return <main className="prototype-page" aria-label="Cash prototype"><p className="error-message" role="alert">{error}</p></main>;

  return <main className="prototype-page" aria-label={`${title} prototype`}>
    <header className="prototype-heading"><div><p className="eyebrow">PAPER ACCOUNT · DATABASE MOCK</p><h1>{title}</h1><p className="prototype-subtitle">Prototype only. No real money movement or bank connection is performed.</p></div><span className="data-badge simulated">SIMULATION</span></header>
    {mode === "overview" && <Overview movements={movements} account={account} pendingCount={pendingCount} completedDeposits={completedDeposits} completedWithdrawals={completedWithdrawals} netMovement={netMovement} />}
    {mode === "history" && <History movements={filteredMovements} total={movements.length} typeFilter={typeFilter} statusFilter={statusFilter} setTypeFilter={setTypeFilter} setStatusFilter={setStatusFilter} />}
    {isForm && <FundingForm mode={mode} amount={amount} submitted={submitted} account={account} setAmount={value => { setAmount(value); setSubmitted(false); }} setSubmitted={setSubmitted} />}
    {mode === "exchange" && <ExchangeForm />}
  </main>;
}

function Overview({ movements, account, pendingCount, completedDeposits, completedWithdrawals, netMovement }: { movements: CashMovement[]; account: Account | null; pendingCount: number; completedDeposits: number; completedWithdrawals: number; netMovement: number }) {
  const balances = getCurrencyBalances(account);
  const totalHkd = balances.reduce((total, item) => total + item.hkdValue, 0) || 1;
  const pie = balances.map(item => `${item.color} ${((item.hkdValue / totalHkd) * 100).toFixed(2)}%`).join(", ");
  return <>
    <div className="prototype-metrics"><div><span>Available cash</span><strong>{cash(account?.available_cash || 0)}</strong></div><div><span>Cash balance</span><strong>{cash(account?.cash_balance || 0)}</strong></div><div><span>Pending requests</span><strong>{pendingCount}</strong></div></div>
    <div className="cash-overview-grid">
      <section className="data-card cash-summary-card"><div className="panel-title"><div><h2>Account snapshot</h2><p className="panel-caption">Settled cash activity in this paper account</p></div><span className="muted-count">HKD</span></div><dl className="cash-account-list"><div><dt>Net completed movement</dt><dd className={netMovement >= 0 ? "up" : "down"}>{netMovement >= 0 ? "+" : "-"}{cash(Math.abs(netMovement))}</dd></div><div><dt>Completed deposits</dt><dd>{completedDeposits}</dd></div><div><dt>Completed withdrawals</dt><dd>{completedWithdrawals}</dd></div><div><dt>Unsettled requests</dt><dd>{pendingCount}</dd></div></dl></section>
      <section className="data-card cash-recent-card"><div className="panel-title"><div><h2>Recent activity</h2><p className="panel-caption">Latest records from MySQL</p></div><span className="muted-count">{Math.min(movements.length, 3)} latest</span></div><MovementList movements={movements.slice(0, 3)} /></section>
    </div>
    <div className="currency-overview-grid">
      <section className="data-card currency-balances-card"><div className="panel-title"><div><h2>Currency balances</h2><p className="panel-caption">Prototype balances converted to HKD</p></div><span className="muted-count">{balances.length} currencies</span></div><div className="currency-balances">{balances.map(item => <div className="currency-balance-row" key={item.code}><span className="currency-swatch" style={{ backgroundColor: item.color }} aria-hidden="true" /><div className="currency-balance-identity"><strong>{item.code}</strong><small>{item.name}</small></div><div className="currency-balance-values"><strong>{nativeCash(item.code, item.balance)}</strong><small>{nativeCash("HKD", item.hkdValue)} equivalent · {nativeCash(item.code, item.available)} available</small></div></div>)}</div></section>
      <section className="data-card currency-allocation-card"><div className="panel-title"><div><h2>Currency allocation</h2><p className="panel-caption">Share of total cash at mock reference rates</p></div><span className="muted-count">HKD basis</span></div><div className="currency-allocation"><div className="currency-allocation-pie" role="img" aria-label="Currency allocation" style={{ background: `conic-gradient(${pie})` }}><span>{nativeCash("HKD", totalHkd)}</span></div><div className="currency-allocation-legend">{balances.map(item => <div key={item.code}><span className="currency-swatch" style={{ backgroundColor: item.color }} aria-hidden="true" /><span>{item.code} allocation</span><strong>{((item.hkdValue / totalHkd) * 100).toFixed(1)}%</strong></div>)}</div></div></section>
    </div>
  </>;
}

function ExchangeForm() {
  const [from, setFrom] = useState<CurrencyCode>("HKD");
  const [to, setTo] = useState<CurrencyCode>("USD");
  const [amount, setAmount] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const numericAmount = Number(amount);
  const validAmount = Number.isFinite(numericAmount) && numericAmount > 0 && from !== to;
  const receive = validAmount ? numericAmount * ratesToHKD[from] / ratesToHKD[to] : 0;
  const options = (selected: CurrencyCode) => (["HKD", "USD", "CNY"] as CurrencyCode[]).filter(code => code !== selected);
  return <div className="exchange-layout">
    <section className="data-card exchange-form-card"><div className="panel-title"><div><p className="panel-kicker">CURRENCY CONVERSION</p><h2>Exchange currencies</h2></div><span className="data-badge simulated">MOCK REQUEST</span></div><form onSubmit={event => { event.preventDefault(); if (validAmount) setSubmitted(true); }}>
      <div className="exchange-fields"><label>From currency<select aria-label="From currency" value={from} onChange={event => { setFrom(event.target.value as CurrencyCode); setSubmitted(false); }}>{["HKD", "USD", "CNY"].map(code => <option key={code}>{code}</option>)}</select></label><button className="exchange-swap" type="button" aria-label="Swap currencies" onClick={() => { setFrom(to); setTo(from); setSubmitted(false); }}>⇄</button><label>To currency<select aria-label="To currency" value={to} onChange={event => { setTo(event.target.value as CurrencyCode); setSubmitted(false); }}>{options(from).map(code => <option key={code}>{code}</option>)}</select></label></div>
      <label>Exchange amount<input aria-label="Exchange amount" inputMode="decimal" type="number" min="0" step="0.01" value={amount} onChange={event => { setAmount(event.target.value); setSubmitted(false); }} placeholder="0.00" /></label>
      <div className="exchange-preview"><div><span>Exchange rate</span><strong>1 {from} = {(ratesToHKD[from] / ratesToHKD[to]).toFixed(4)} {to}</strong></div><div><span>Estimated receive {to}</span><strong>{nativeCash(to, receive)}</strong></div><div><span>Prototype fee</span><strong>HKD 0.00</strong></div></div>
      <button className="primary-button full-width" type="submit" disabled={!validAmount}>Preview exchange</button>{from === to && <p className="ticket-warning" role="alert">Choose two different currencies to preview a conversion.</p>}{submitted && <p className="success-message" role="status">Exchange request staged in prototype. No real conversion was executed.</p>}
    </form></section>
    <aside className="data-card exchange-context"><div className="panel-title"><div><p className="panel-kicker">REFERENCE RATES</p><h2>Three supported currencies</h2></div><span className="muted-count">SIMULATED</span></div><dl className="exchange-rate-list">{(["HKD", "USD", "CNY"] as CurrencyCode[]).map(code => <div key={code}><dt>{code}</dt><dd>{nativeCash("HKD", ratesToHKD[code])}</dd></div>)}</dl><p className="panel-caption">Rates and balances are mock data for the prototype. This action does not alter your paper account.</p></aside>
  </div>;
}

function History({ movements, total, typeFilter, statusFilter, setTypeFilter, setStatusFilter }: { movements: CashMovement[]; total: number; typeFilter: MovementTypeFilter; statusFilter: MovementStatusFilter; setTypeFilter: (value: MovementTypeFilter) => void; setStatusFilter: (value: MovementStatusFilter) => void }) {
  return <section className="data-card history-card"><div className="panel-title"><div><h2>Cash movement history</h2><p className="panel-caption">Complete deposit and withdrawal ledger</p></div><span className="muted-count">{movements.length} of {total} records</span></div><div className="filter-row movement-filter-row"><label>Type<select aria-label="Movement type" value={typeFilter} onChange={event => setTypeFilter(event.target.value as MovementTypeFilter)}><option value="all">All types</option><option value="deposit">Deposits</option><option value="withdraw">Withdrawals</option></select></label><label>Status<select aria-label="Movement status" value={statusFilter} onChange={event => setStatusFilter(event.target.value as MovementStatusFilter)}><option value="all">All statuses</option><option value="completed">Completed</option><option value="pending">Pending</option></select></label><button className="quiet-button" type="button" onClick={() => { setTypeFilter("all"); setStatusFilter("all"); }}>Clear filters</button></div>{movements.length > 0 ? <MovementTable movements={movements} /> : <p className="empty-state" role="status">No cash movements match these filters.</p>}</section>;
}

function FundingForm({ mode, amount, submitted, account, setAmount, setSubmitted }: { mode: "deposit" | "withdraw"; amount: string; submitted: boolean; account: Account | null; setAmount: (value: string) => void; setSubmitted: (value: boolean) => void }) {
  const deposit = mode === "deposit";
  const validAmount = Number(amount) > 0;
  return <div className="funding-layout">
    <section className="data-card movement-form-card"><div className="panel-title"><div><p className="panel-kicker">{deposit ? "ADD CASH" : "MOVE CASH OUT"}</p><h2>{deposit ? "Deposit funds" : "Withdraw funds"}</h2></div><span className="data-badge simulated">MOCK REQUEST</span></div><form onSubmit={event => { event.preventDefault(); if (validAmount) setSubmitted(true); }}>
      <label>{deposit ? "Deposit amount" : "Withdrawal amount"}<input aria-label={deposit ? "Deposit amount" : "Withdrawal amount"} inputMode="decimal" type="number" min="1" step="0.01" value={amount} onChange={event => setAmount(event.target.value)} placeholder="0.00" /></label>
      {deposit ? <><label>Funding method<select aria-label="Funding method"><option>Bank transfer (mock)</option><option>FPS (mock)</option></select></label><label>Source account<input aria-label="Source account" value="Demo bank account ·•••• 0210" readOnly /></label></> : <><label>Receiving account<select aria-label="Receiving account"><option>Demo bank account ·•••• 0210</option><option>FPS beneficiary (mock)</option></select></label><div className="fee-preview"><span>Estimated fee</span><strong>HKD 0.00</strong></div></>}
      <button className="primary-button full-width" type="submit" disabled={!validAmount}>Submit {deposit ? "deposit" : "withdrawal"}</button>{submitted && <p className="success-message" role="status">{deposit ? "Deposit" : "Withdrawal"} request staged in prototype. No real transfer was sent.</p>}
    </form></section>
    <aside className="data-card funding-context"><div className="panel-title"><div><p className="panel-kicker">ACCOUNT CONTEXT</p><h2>{deposit ? "Ready to fund" : "Available to withdraw"}</h2></div><span className="muted-count">HKD</span></div><div className="funding-context-value"><span>{deposit ? "Current cash balance" : "Available cash"}</span><strong>{cash(deposit ? account?.cash_balance || 0 : account?.available_cash || 0)}</strong></div><p className="panel-caption">This prototype stages a request for review only. No bank connection or real money movement is performed.</p></aside>
  </div>;
}

function MovementList({ movements }: { movements: CashMovement[] }) {
  if (movements.length === 0) return <p className="empty-state" role="status">No cash activity yet.</p>;
  return <ul className="cash-activity-list" aria-label="Recent cash activity">{movements.map(item => <li key={item.id}><div><strong>{movementLabel(item.type)}</strong><small>{date(item.created_at)} · {item.method}</small></div><div className="cash-activity-value"><strong className={item.type === "deposit" ? "up" : "down"}>{item.type === "deposit" ? "+" : "-"}{cash(item.amount)}</strong><span className={`order-status ${item.status}`}>{item.status === "completed" ? "Completed" : "Pending"}</span></div></li>)}</ul>;
}

function MovementTable({ movements }: { movements: CashMovement[] }) {
  return <div className="table-wrap"><table aria-label="Cash movement history"><thead><tr><th>Date</th><th>Type</th><th>Amount</th><th>Method</th><th>Status</th></tr></thead><tbody>{movements.map(item => <tr key={item.id}><td>{date(item.created_at)}</td><td className={item.type === "deposit" ? "up" : "down"}>{movementLabel(item.type)}</td><td>{cash(item.amount)}</td><td>{item.method}</td><td><span className={`order-status ${item.status}`}>{item.status === "completed" ? "Completed" : "Pending"}</span></td></tr>)}</tbody></table></div>;
}
