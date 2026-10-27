import { useMemo, useState, type CSSProperties } from "react";
import { useI18n } from "../../app/I18nProvider";

type EngineKey = "momentum" | "factor" | "mean-reversion" | "risk-parity";

type EngineDefinition = {
  key: EngineKey;
  name: string;
  shortName: string;
  description: string;
  investorUse: string;
  risk: string;
  signal: string;
  returnPct: number;
  drawdownPct: number;
  winRate: number;
  turnover: string;
};

const ENGINE_DEFINITIONS: EngineDefinition[] = [
  { key: "momentum", name: "Trend momentum", shortName: "Momentum", description: "Ranks relative strength and confirms trend persistence across the selected horizon.", investorUse: "Helps investors avoid fighting a strong trend and define a rules-based entry window.", risk: "Crowded trades can reverse quickly after a regime change.", signal: "Strong upward trend", returnPct: 18.4, drawdownPct: -12.8, winRate: 61, turnover: "2.4x" },
  { key: "factor", name: "Multi-factor score", shortName: "Factors", description: "Combines value, quality, growth and volatility signals into one explainable score.", investorUse: "Helps investors compare several stocks consistently instead of relying on one headline or ratio.", risk: "Factor leadership can rotate, leaving yesterday's winners behind.", signal: "Quality + value lead", returnPct: 15.7, drawdownPct: -9.6, winRate: 58, turnover: "1.7x" },
  { key: "mean-reversion", name: "Mean reversion", shortName: "Reversion", description: "Looks for stretched price and volume conditions that may normalize toward a reference range.", investorUse: "Helps investors frame a pullback plan with an invalidation level rather than chasing a move.", risk: "A real breakdown can stay oversold longer than expected.", signal: "Oversold rebound setup", returnPct: 11.2, drawdownPct: -15.1, winRate: 54, turnover: "3.1x" },
  { key: "risk-parity", name: "Risk parity allocation", shortName: "Risk parity", description: "Balances contribution to portfolio risk across equities, cash and defensive sleeves.", investorUse: "Helps investors size positions around loss tolerance and keep one idea from dominating the account.", risk: "Correlation spikes can reduce diversification when protection is needed most.", signal: "Balanced risk budget", returnPct: 9.8, drawdownPct: -7.4, winRate: 56, turnover: "0.9x" },
];

const SIGNAL_STEPS = [72, 58, 84, 67, 49, 76, 63, 88, 70, 55, 81, 73];

export function QuantEnginesPanel() {
  const { t } = useI18n();
  const [activeKey, setActiveKey] = useState<EngineKey>("momentum");
  const [compareKeys, setCompareKeys] = useState<EngineKey[]>(["momentum", "factor"]);
  const [symbol, setSymbol] = useState("00700");
  const [horizon, setHorizon] = useState("6M");
  const [riskBudget, setRiskBudget] = useState(10);
  const [includeCosts, setIncludeCosts] = useState(true);
  const [hasRun, setHasRun] = useState(true);

  const active = ENGINE_DEFINITIONS.find(engine => engine.key === activeKey) || ENGINE_DEFINITIONS[0];
  const compared = useMemo(() => ENGINE_DEFINITIONS.filter(engine => compareKeys.includes(engine.key)), [compareKeys]);

  function toggleCompare(key: EngineKey) {
    setCompareKeys(current => current.includes(key)
      ? current.length === 1 ? current : current.filter(item => item !== key)
      : [...current, key]);
  }

  return <section className="quant-engines-page">
    <div className="prototype-heading">
      <div><p className="eyebrow">{t("QUANT ENGINE LAB · PROTOTYPE")}</p><h1>{t("Multiple quant engines")}</h1><p className="panel-caption quant-engine-lede">{t("Compare explainable signals, portfolio risk, and investor use cases before any paper order is considered.")}</p></div>
      <span className="data-badge simulated">{t("SIMULATED")}</span>
    </div>
    <div className="quant-engine-intro"><strong>{t("How this helps investors")}</strong><span>{t("Use several independent lenses, inspect the assumptions, then decide whether a signal fits your time horizon and loss tolerance.")}</span></div>
    <div className="quant-engine-layout">
      <aside className="data-card quant-engine-directory" aria-label={t("Quant engine directory")}>
        <div className="panel-title"><div><h2>{t("Engine directory")}</h2><p className="panel-caption">{t("Select one lens to inspect")}</p></div><span className="muted-count">{ENGINE_DEFINITIONS.length}</span></div>
        <div className="quant-engine-list" role="tablist" aria-label={t("Quant engine directory")}>
          {ENGINE_DEFINITIONS.map(engine => <button key={engine.key} role="tab" aria-selected={engine.key === activeKey} className={`quant-engine-option ${engine.key === activeKey ? "active" : ""}`} onClick={() => setActiveKey(engine.key)}>
            <span className={`quant-engine-dot ${engine.key}`} aria-hidden="true" /><span><strong>{t(engine.name)}</strong><small>{t(engine.signal)}</small></span><b>{engine.returnPct.toFixed(1)}%</b>
          </button>)}
        </div>
        <div className="quant-engine-help"><strong>{t("Investor lens")}</strong><p>{t(active.investorUse)}</p><small>{t("Risk to check")}: {t(active.risk)}</small></div>
      </aside>
      <div className="quant-engine-workspace">
        <section className="data-card quant-engine-controls">
          <div className="panel-title"><div><h2>{t("Backtest setup")}</h2><p className="panel-caption">{t("Illustrative data only · no order is sent")}</p></div><span className="data-badge delayed">{t("UNOFFICIAL DATA")}</span></div>
          <div className="quant-engine-control-grid">
            <label>{t("Security")}<select value={symbol} onChange={event => setSymbol(event.target.value)}><option value="00700">00700 · TENCENT</option><option value="09988">09988 · Alibaba</option><option value="01810">01810 · Xiaomi</option></select></label>
            <label>{t("Lookback") }<select value={horizon} onChange={event => setHorizon(event.target.value)}><option value="3M">{t("3 months")}</option><option value="6M">{t("6 months")}</option><option value="1Y">{t("1 year")}</option></select></label>
            <label>{t("Risk budget")}<input type="range" min="4" max="20" step="1" value={riskBudget} onChange={event => setRiskBudget(Number(event.target.value))} /><strong>{riskBudget}%</strong></label>
            <label className="quant-engine-checkbox"><input type="checkbox" checked={includeCosts} onChange={event => setIncludeCosts(event.target.checked)} />{t("Include fees and slippage")}</label>
          </div>
          <button className="primary-button quant-engine-run" type="button" onClick={() => setHasRun(true)}>{t("Run comparison")}</button>
        </section>
        <section className="data-card quant-engine-result">
          <div className="panel-title"><div><p className="eyebrow">{t(active.name)}</p><h2>{t("Signal and risk readout")}</h2><p className="panel-caption">{symbol} · {horizon} · {includeCosts ? t("costs included") : t("before costs")}</p></div><span className="data-badge simulated">{hasRun ? t("READY") : t("DRAFT")}</span></div>
          <div className="quant-engine-metrics"><div><span>{t("Illustrative return")}</span><strong className="up">+{active.returnPct.toFixed(1)}%</strong></div><div><span>{t("Max drawdown")}</span><strong className="down">{active.drawdownPct.toFixed(1)}%</strong></div><div><span>{t("Win rate")}</span><strong>{active.winRate}%</strong></div><div><span>{t("Portfolio turnover")}</span><strong>{active.turnover}</strong></div></div>
          <div className="quant-engine-chart" aria-label={t("Illustrative signal strength chart")}><div className="quant-engine-chart-axis"><span>100</span><span>50</span><span>0</span></div><div className="quant-engine-bars">{SIGNAL_STEPS.map((value, index) => <span key={`${value}-${index}`} style={{ "--signal": `${value}%` } as CSSProperties} className={value > 70 ? "strong" : ""} />)}</div></div>
          <div className="quant-engine-readout"><div><strong>{t("Current signal")}</strong><span>{t(active.signal)}</span></div><div><strong>{t("Decision guardrail")}</strong><span>{t("Require two confirmations and review the drawdown before sizing a position.")}</span></div></div>
        </section>
        <section className="data-card quant-engine-compare">
          <div className="panel-title"><div><h2>{t("Engine comparison")}</h2><p className="panel-caption">{t("Compare independent assumptions, not just headline return")}</p></div><span className="muted-count">{compared.length} {t("selected")}</span></div>
          <div className="quant-engine-table-wrap"><table><thead><tr><th>{t("Engine")}</th><th>{t("Signal")}</th><th>{t("Return")}</th><th>{t("Drawdown")}</th><th>{t("Investor help")}</th><th>{t("Compare")}</th></tr></thead><tbody>{ENGINE_DEFINITIONS.map(engine => <tr key={engine.key}><td><strong>{t(engine.shortName)}</strong><small>{t(engine.name)}</small></td><td>{t(engine.signal)}</td><td className="up">+{engine.returnPct.toFixed(1)}%</td><td className="down">{engine.drawdownPct.toFixed(1)}%</td><td>{t(engine.investorUse)}</td><td><label className="quant-engine-checkbox"><input type="checkbox" checked={compareKeys.includes(engine.key)} onChange={() => toggleCompare(engine.key)} aria-label={`${t("Compare")} ${t(engine.shortName)}`} /></label></td></tr>)}</tbody></table></div>
        </section>
      </div>
    </div>
  </section>;
}
