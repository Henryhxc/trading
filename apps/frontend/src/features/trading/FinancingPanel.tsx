import { useMemo, useState } from "react";
import type { Account, Position } from "../../shared/api/trading";
import { money } from "./usePortfolio";
import { useI18n } from "../../app/I18nProvider";

type Mode = "pledge" | "margin";
type Props = { account: Account | null; positions: Position[] };
const HAIRCUT = 0.5;
const ADVANCE_RATE = 0.6;
const ANNUAL_RATE = 0.09;
const MAINTENANCE_LINE = 0.3;
const LIQUIDATION_LINE = 0.2;
const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

export function FinancingPanel({ account, positions }: Props) {
  const { t } = useI18n();
  const [mode, setMode] = useState<Mode>("pledge");
  const [loanAmount, setLoanAmount] = useState("");
  const [marketDrop, setMarketDrop] = useState("0");
  const [staged, setStaged] = useState(false);
  const collateral = useMemo(() => positions.reduce((sum, position) => sum + Math.max(0, position.market_value || 0), 0), [positions]);
  const adjustedCollateral = collateral * HAIRCUT;
  const maxLoan = adjustedCollateral * ADVANCE_RATE;
  const requestedLoan = Number(loanAmount);
  const loanIsValid = Number.isFinite(requestedLoan) && requestedLoan > 0 && requestedLoan <= maxLoan;
  const monthlyInterest = loanIsValid ? requestedLoan * ANNUAL_RATE / 12 : 0;
  const drop = Math.min(100, Math.max(0, Number(marketDrop) || 0)) / 100;
  const stressedCollateral = collateral * (1 - drop);
  const stressedEquity = stressedCollateral - requestedLoan;
  const riskRatio = stressedCollateral > 0 ? stressedEquity / stressedCollateral : 0;
  const riskMessage = t(requestedLoan <= 0 ? "Enter a loan amount to run a margin stress scenario." : riskRatio <= LIQUIDATION_LINE ? "Illustrative liquidation threshold reached. Actual broker rules and execution may differ." : riskRatio <= MAINTENANCE_LINE ? "Illustrative maintenance margin call. Add eligible collateral or repay part of the loan." : "Within the illustrative maintenance range. Prices and collateral values can change quickly.");

  return <main className="prototype-page financing-page" aria-label={t("Financing simulator prototype")}>
    <header className="prototype-heading"><div><p className="eyebrow">{t("PAPER ACCOUNT · SIMULATION ONLY")}</p><h1>{t("Financing simulator")}</h1><p className="prototype-subtitle">{t("Compare stock-backed borrowing with margin buying power using illustrative assumptions.")}</p></div><span className="data-badge simulated">{t("NO REAL CREDIT")}</span></header>
    <div className="financing-notice" role="note"><strong>{t("Prototype only")}</strong><span>{t("No broker connection, credit assessment, loan drawdown, interest posting, or real trade is performed. Rates and collateral rules below are examples, not a broker offer.")}</span></div>
    <nav className="financing-mode-navigation" aria-label={t("Financing type")} role="tablist"><button role="tab" aria-selected={mode === "pledge"} className={mode === "pledge" ? "active" : ""} onClick={() => { setMode("pledge"); setStaged(false); }}>{t("Stock-backed loan")}</button><button role="tab" aria-selected={mode === "margin"} className={mode === "margin" ? "active" : ""} onClick={() => { setMode("margin"); setStaged(false); }}>{t("Margin buying")}</button></nav>
    <section className="financing-summary" aria-label={t("Illustrative collateral summary")}><div><span>{t("Current holdings value")}</span><strong>{money(collateral)}</strong></div><div><span>{t("Adjusted collateral value · {rate}", { rate: formatPercent(HAIRCUT) })}</span><strong>{money(adjustedCollateral)}</strong></div><div><span>{t("Maximum illustrative loan")}</span><strong>{money(maxLoan)}</strong></div><div><span>{t("Available cash · not used as collateral")}</span><strong>{money(account?.available_cash)}</strong></div></section>
    <div className="financing-layout">
      <section className="data-card financing-form-card"><div className="panel-title"><div><p className="panel-kicker">{t(mode === "pledge" ? "PLEDGE EXISTING SHARES" : "BUYING POWER SIMULATION")}</p><h2>{t(mode === "pledge" ? "Stock-backed loan" : "Margin buying")}</h2><p className="panel-caption">{t(mode === "pledge" ? "Estimate cash borrowing against eligible paper holdings." : "Estimate the effect of a simulated loan used alongside your own cash.")}</p></div><span className="data-badge simulated">{t("MOCK")}</span></div>
        <form onSubmit={event => { event.preventDefault(); if (loanIsValid) setStaged(true); }}>
          <label>{t("Simulated loan amount")}<input aria-label={t("Loan amount")} type="number" inputMode="decimal" min="0.01" max={maxLoan} step="0.01" value={loanAmount} onChange={event => { setLoanAmount(event.target.value); setStaged(false); }} placeholder={t("Enter amount in HKD")} /></label>
          {!loanIsValid && requestedLoan > maxLoan && <p className="ticket-warning" role="alert">{t("Amount exceeds the illustrative collateral limit of {amount}.", { amount: money(maxLoan) })}</p>}
          <div className={`financing-interest ${mode === "margin" ? "margin-interest" : ""}`}><div><span>{t("Illustrative annual rate")}</span><strong>{formatPercent(ANNUAL_RATE)}</strong></div><div><span>{t("Estimated monthly interest")}</span><strong>{money(monthlyInterest)}</strong></div><div><span>{t("Estimated 30-day interest")}</span><strong>{money(loanIsValid ? requestedLoan * ANNUAL_RATE * 30 / 365 : 0)}</strong></div>{mode === "margin" && <div><span>{t("Estimated buying funds · cash + loan")}</span><strong>{money((account?.available_cash || 0) + (loanIsValid ? requestedLoan : 0))}</strong></div>}</div>
          <button className="primary-button full-width" type="submit" disabled={!loanIsValid}>{t("Stage financing request")}</button>
          {staged && <p className="success-message" role="status">{t("Financing request staged in prototype. No credit check or loan drawdown was made.")}</p>}
        </form>
      </section>
      <aside className="data-card financing-context"><div className="panel-title"><div><p className="panel-kicker">{t("COLLATERAL BREAKDOWN")}</p><h2>{t("Eligible paper positions")}</h2></div><span className="muted-count">{t("{count} holdings", { count: positions.length })}</span></div>
        {positions.length === 0 ? <p className="empty-state" role="status">{t("No positions are available to estimate collateral.")}</p> : <ul className="collateral-list">{positions.map(position => <li key={position.symbol}><div><strong>{position.symbol} · {position.display_name}</strong><small>{t("Market value {amount} · illustrative haircut {rate}", { amount: money(position.market_value), rate: formatPercent(HAIRCUT) })}</small></div><strong>{money(Math.max(0, position.market_value || 0) * HAIRCUT)}</strong></li>)}</ul>}
        <p className="panel-caption">{t("Estimated limit = total market value × {haircut} haircut × {advance} advance rate. Actual eligibility, concentration limits, currency conversion, and fees vary by broker.", { haircut: formatPercent(HAIRCUT), advance: formatPercent(ADVANCE_RATE) })}</p>
      </aside>
    </div>
    <section className="data-card margin-stress-card"><div className="panel-title"><div><p className="panel-kicker">{t("DOWNSIDE SCENARIO")}</p><h2>{t("Margin risk monitor")}</h2><p className="panel-caption">{t("Illustrative equity ratio = stressed holdings value less loan, divided by stressed holdings value.")}</p></div><span className="muted-count">{t("Assumptions only")}</span></div>
      <div className="margin-stress-layout"><label>{t("Simulated market drop")} <strong>{drop.toFixed(0)}%</strong><input aria-label={t("Simulated market drop")} type="range" min="0" max="80" step="5" value={marketDrop} onChange={event => setMarketDrop(event.target.value)} /></label><div className="margin-stress-metrics"><div><span>{t("Stressed holdings value")}</span><strong>{money(stressedCollateral)}</strong></div><div><span>{t("Equity ratio")}</span><strong className={riskRatio <= MAINTENANCE_LINE ? "down" : "up"}>{formatPercent(riskRatio)}</strong></div><div><span>{t("Maintenance / liquidation example")}</span><strong>{formatPercent(MAINTENANCE_LINE)} / {formatPercent(LIQUIDATION_LINE)}</strong></div></div></div>
      <p className={`margin-risk-message ${riskRatio <= LIQUIDATION_LINE ? "critical" : riskRatio <= MAINTENANCE_LINE ? "warning" : "normal"}`} role={requestedLoan > 0 && riskRatio <= MAINTENANCE_LINE ? "alert" : "status"}>{riskMessage}</p>
    </section>
  </main>;
}
