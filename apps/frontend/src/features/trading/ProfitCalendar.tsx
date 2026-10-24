import { useEffect, useMemo, useState } from "react";
import type { ProfitPeriod, ProfitRow } from "./profit";

const money = (value: number) => `${value >= 0 ? "+" : "-"}HKD ${Math.abs(value).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
const percent = (value: number) => `${value >= 0 ? "+" : ""}${value.toFixed(2)}%`;
const dateKey = (year: number, month: number, day: number) => `${year}-${String(month + 1).padStart(2, "0")}-${String(day).padStart(2, "0")}`;

function initialPeriod(rows: ProfitRow[], periodType: ProfitPeriod) {
  const period = rows[0]?.period;
  if (periodType === "day") return period ? period.slice(0, 7) : new Date().toISOString().slice(0, 7);
  return period ? period.slice(0, 4) : String(new Date().getUTCFullYear());
}

function valueClass(pnl: number) { return pnl > 0 ? "up" : pnl < 0 ? "down" : "neutral"; }

export function ProfitCalendar({ title, rows, granularity = "day", monthly }: { title: string; rows: ProfitRow[]; granularity?: ProfitPeriod; monthly?: boolean }) {
  const periodType = monthly ? "month" : granularity;
  const [period, setPeriod] = useState(() => initialPeriod(rows, periodType));
  useEffect(() => { if (rows.length > 0) setPeriod(initialPeriod(rows, periodType)); }, [rows, periodType]);
  const byPeriod = useMemo(() => new Map(rows.map(row => [row.period, row])), [rows]);
  const [year, month] = period.split("-").map(Number);
  const yearValue = Number.isFinite(year) ? year : new Date().getUTCFullYear();
  const monthValue = Number.isFinite(month) ? month - 1 : 0;
  const label = periodType === "day" ? `${yearValue}-${String(monthValue + 1).padStart(2, "0")}` : String(yearValue);

  function move(delta: number) {
    if (periodType !== "day") setPeriod(String(yearValue + delta));
    else {
      const next = new Date(Date.UTC(yearValue, monthValue + delta, 1));
      setPeriod(next.toISOString().slice(0, 7));
    }
  }

  const cells = periodType === "year"
    ? [{ key: String(yearValue), label: String(yearValue), row: byPeriod.get(String(yearValue)) }]
    : periodType === "month"
    ? Array.from({ length: 12 }, (_, index) => {
      const key = `${yearValue}-${String(index + 1).padStart(2, "0")}`;
      return { key, label: new Date(Date.UTC(yearValue, index, 1)).toLocaleDateString("en", { month: "short", timeZone: "UTC" }), row: byPeriod.get(key) };
    })
    : (() => {
      const count = new Date(Date.UTC(yearValue, monthValue + 1, 0)).getUTCDate();
      const leading = new Date(Date.UTC(yearValue, monthValue, 1)).getUTCDay();
      return [...Array.from({ length: leading }, (_, index) => ({ key: `empty-${index}`, label: "", row: undefined })), ...Array.from({ length: count }, (_, index) => {
        const day = index + 1;
        const key = dateKey(yearValue, monthValue, day);
        return { key, label: String(day), row: byPeriod.get(key) };
      })];
    })();

  return <section className="data-card profit-card">
    <div className="panel-title"><div><h2>{title}</h2><p className="panel-caption">Realized P/L from filled paper orders</p></div><span className="muted-count">{rows.length} periods</span></div>
    <div className="calendar-toolbar"><button className="quiet-button" aria-label={`Previous ${periodType === "day" ? "month" : "year"}`} onClick={() => move(-1)}>‹</button><strong>{label}</strong><button className="quiet-button" aria-label={`Next ${periodType === "day" ? "month" : "year"}`} onClick={() => move(1)}>›</button></div>
    {periodType === "day" && <div className="calendar-weekdays" aria-hidden="true">{["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map(day => <span key={day}>{day}</span>)}</div>}
    <div className={`profit-calendar ${periodType !== "day" ? "monthly-calendar" : ""} ${periodType === "year" ? "yearly-calendar" : ""}`} role="grid" aria-label={`${periodType === "year" ? "Yearly" : periodType === "month" ? "Monthly" : "Daily"} realized P/L calendar`}>
      {cells.map(cell => cell.row ? <div key={cell.key} className={`profit-cell ${valueClass(cell.row.pnl)}`} role="gridcell" aria-label={`${cell.key} ${cell.row.pnl >= 0 ? "profit" : "loss"} ${money(cell.row.pnl)} ${percent(cell.row.returnPercent)}`}><span>{cell.label}</span><strong>{money(cell.row.pnl)}</strong><small>{percent(cell.row.returnPercent)}</small></div> : <div key={cell.key} className="profit-cell empty" role="gridcell" aria-label={cell.key.startsWith("empty-") ? "Empty calendar cell" : `${cell.key} no realized P/L`}><span>{cell.label}</span>{cell.label && <small>No P/L</small>}</div>)}
    </div>
    {rows.length === 0 && <p className="empty-state" role="status">No realized P/L records yet.</p>}
  </section>;
}

export function PortfolioAllocationChart({ account, positions }: { account: { available_cash: number } | null; positions: Array<{ display_name: string; market_value: number | null }> }) {
  const entries = positions.map(position => ({ label: position.display_name, value: Math.max(0, position.market_value || 0) })).filter(entry => entry.value > 0);
  const cash = Math.max(0, account?.available_cash || 0);
  const totals = [...entries, { label: "Available cash", value: cash }];
  const total = totals.reduce((sum, entry) => sum + entry.value, 0);
  let cursor = 0;
  const stops = totals.map((entry, index) => { const start = total ? cursor / total * 100 : 0; cursor += entry.value; const end = total ? cursor / total * 100 : 100; return `${["#5478ed", "#40d794", "#f3c76b", "#c58cff"][index % 4]} ${start}% ${end}%`; });
  return <section className="data-card allocation-card"><div className="panel-title"><div><h2>Portfolio allocation</h2><p className="panel-caption">Mock holdings and cash mix from your paper account</p></div><span className="data-badge simulated">DB MOCK</span></div><div className="allocation-content"><div className="allocation-pie" role="img" aria-label="Portfolio allocation pie chart" style={{ background: total ? `conic-gradient(${stops.join(", ")})` : "#26374e" }}><span>{total ? "100%" : "—"}</span></div><div className="allocation-legend">{totals.map((entry, index) => <div key={entry.label}><i className={`allocation-swatch swatch-${index % 4}`} /><span>{entry.label}</span><strong>{total ? `${Math.round(entry.value / total * 100)}%` : "0%"}</strong></div>)}</div></div></section>;
}
