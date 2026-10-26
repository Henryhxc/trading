import { useState, type FormEvent } from "react";
import { useI18n } from "../../app/I18nProvider";
import { classifyIntents, loadLongTermMemory, removeLongTermMemory, saveLongTermMemory, type IntentId } from "./assistantMemory";

type AgentId = "portfolio" | "macro" | "left-side" | "quant" | "risk" | "technical";
type Analysis = { agent: AgentId; text: string };
type Round = { id: number; prompt: string; analyses: Analysis[]; synthesis: string; intents: IntentId[]; memory: string[] };
type QuantAssistantProps = { memoryStorage?: Storage };

const agents: { id: AgentId; name: string; specialty: string; response: string }[] = [
  { id: "portfolio", name: "Portfolio analyst", specialty: "Position & exposure", response: "Portfolio lens: separate a thematic watchlist from an actual holding. SpaceX is privately held, so this scenario is not a directly tradable listed share." },
  { id: "macro", name: "Macro strategist", specialty: "Growth drivers & risks", response: "Macro lens: Starlink's long-run case depends on broadband demand, launch capacity, spectrum access, regulation, competition, and the capital needed to scale. These drivers can conflict; growth is not a guaranteed return." },
  { id: "left-side", name: "Left-side trader", specialty: "Early entries & thesis risk", response: "Left-side lens: entering before a confirmed reversal can improve the entry price, but a selloff may continue. Treat each add as a predefined paper-trade tranche with a thesis invalidation point, not an automatic dip-buy." },
  { id: "quant", name: "Quant strategist", specialty: "Factors & market regime", response: "Quant pullback plan: use a broad-market trend filter, volatility-normalized drawdown bands, and staged paper entries only after the signal stabilizes. Define maximum portfolio risk and an exit condition first; backtest across regimes with gap risk, fees, and slippage. This is a framework, not a live signal." },
  { id: "risk", name: "Risk monitor", specialty: "Volatility & drawdown", response: "Risk lens: a further hypothetical 40% drop can follow an earlier 40% decline. Avoid averaging down without limit; size for the full adverse scenario and remember private-company valuations may be infrequent and non-transparent." },
  { id: "technical", name: "Technical analyst", specialty: "Price structure & levels", response: "Technical lens: wait for a defined trend or reversal condition and volume confirmation; a large drawdown alone does not establish a floor. No live SpaceX price series is available in this mock." },
];

const initialRounds: Round[] = [
  {
    id: 1,
    prompt: "How should I evaluate a hypothetical SpaceX position and the Starlink trend?",
    intents: ["portfolio", "industry", "market"],
    memory: [],
    analyses: [
      { agent: "portfolio", text: "Portfolio lens: separate a thematic watchlist from an actual holding. SpaceX is privately held, so this scenario is not a directly tradable listed share." },
      { agent: "macro", text: "Macro lens: Starlink's long-run case depends on broadband demand, launch capacity, spectrum access, regulation, competition, and the capital needed to scale. These drivers can conflict; growth is not a guaranteed return." },
      { agent: "technical", text: "Technical lens: wait for a defined trend or reversal condition and volume confirmation; a large drawdown alone does not establish a floor. No live SpaceX price series is available in this mock." },
    ],
    synthesis: "Synthesis: treat Starlink as a long-horizon business thesis with execution, funding, regulatory, and competition risks. Because SpaceX is private, this case is for portfolio-thinking practice only, not a direct stock trade.",
  },
  {
    id: 2,
    prompt: "If the theme suffers a hypothetical 40% selloff, compare a left-side entry with a quant pullback plan.",
    intents: ["industry", "quant", "risk"],
    memory: [],
    analyses: [
      { agent: "left-side", text: "Left-side lens: entering before a confirmed reversal can improve the entry price, but a selloff may continue. Use predefined paper-trade tranches and a thesis invalidation point; do not auto-buy every dip." },
      { agent: "quant", text: "Quant pullback plan: an illustrative 100-unit proxy falling to 60 and then 36 is down 64% from its start. A rule-based paper test could require a broad-market trend filter, a volatility-normalized pullback band, and a stable reversal signal before staged entries. The proxy is synthetic, not actual SpaceX market data." },
      { agent: "risk", text: "Risk lens: after a 40% loss, recovering to the starting value requires a 66.7% gain; after a cumulative 64% loss, it requires about 177.8%. A drawdown limit and thesis invalidation rule matter more than lowering average cost." },
      { agent: "macro", text: "Macro lens: re-check whether demand, funding costs, launch cadence, spectrum access, or competition changed. A lower mark is not automatically better value if the business thesis weakened." },
    ],
    synthesis: "Synthesis: the left-side approach accepts more timing risk for earlier exposure; the quant framework waits for measurable conditions and caps risk. Compare both in a paper backtest. Neither is a current recommendation or an executable signal.",
  },
];

function runParallelReview(id: number, prompt: string, focus: AgentId, previousIntents: IntentId[], memory: string[]): Round {
  const orderedAgents = [...agents].sort((a, b) => Number(b.id === focus) - Number(a.id === focus));
  const intents = classifyIntents(prompt, previousIntents);
  const drawdown = intents.includes("risk") || /(40%|sell.?off|跌|亏损|loss)/i.test(prompt);
  const macro = intents.includes("macro") || /(starlink|频谱|demand)/i.test(prompt);
  const portfolio = intents.includes("portfolio");
  const scenario = drawdown ? drawdownAnalysis : macro ? macroAnalysis : portfolio ? portfolioAnalysis : null;
  const analyses = orderedAgents.map(agent => ({ agent: agent.id, text: scenario?.[agent.id] || agent.response }));
  const synthesis = drawdown ? "Drawdown synthesis: percentage losses compound asymmetrically. Do not treat a 40% fall as a buy signal; compare rules against a no-trade baseline and include liquidity, gaps, fees, and slippage." : macro ? "Macro synthesis: separate operating traction from equity value. Track demand, revenue per user, churn, capacity, launch economics, capex, spectrum rights, and regulatory access without assuming unreported financials." : portfolio ? "Portfolio synthesis: SpaceX is not a listed security and has no public quote in this terminal. Use this case to test concentration limits, not to infer a real position size or order." : "Harness synthesis: this is a scripted SpaceX/Starlink teaching case, not a live research system. Verify current facts before relying on any conclusion.";
  return { id, prompt, analyses, synthesis, intents, memory };
}

const drawdownAnalysis: Record<AgentId, string> = {
  portfolio: "Portfolio lens: SpaceX is private, so the terminal cannot read an account holding or calculate real exposure. Tag this as an unlisted-theme scenario and cap its paper risk separately from actual HK shares.",
  macro: "Macro lens: test whether the selloff comes from funding costs, launch delays, spectrum access, regulation, or competition. A valuation shock and a broken operating thesis are different cases.",
  "left-side": "Left-side lens: divide a hypothetical paper allocation into four equal stages and add only while the thesis checklist remains intact. Staging reduces timing concentration but cannot prevent further losses.",
  quant: "Drawdown math: A hypothetical 100-unit investment falls to 60, then 36. That is a cumulative loss of 64%, not 80%, and recovery from 36 to 100 requires a 177.8% gain. This is an arithmetic example, not actual SpaceX market data.",
  risk: "Risk sizing example: with a hypothetical $10,000 portfolio, a 1% risk budget is $100. If a paper-trade stop is 10% below entry, position value is capped at $1,000 before fees and gap risk ($100 / 10%). This is a sizing illustration, not a trade instruction.",
  technical: "Technical lens: a percentage drop alone is not a reversal. For a listed instrument, test whether price forms a base, reclaims a chosen trend level, and confirms with volume; this case has no real SpaceX candles to evaluate.",
};

const macroAnalysis: Record<AgentId, string> = {
  portfolio: "Portfolio lens: separate exposure to satellite connectivity from owning SpaceX. Starlink is part of a private company, not a separately listed share; any listed proxy adds tracking and valuation-basis risk.",
  macro: "Macro lens: assess broadband demand alongside spectrum rights, country approvals, launch cadence, terminal affordability, and competition. Higher rates pressure long-duration valuations and increase the cost of constellation investment.",
  "left-side": "Left-side lens: a favorable long-term story does not define an entry. Write down the operating evidence that would justify a paper position and the evidence that would invalidate it before price weakness arrives.",
  quant: "Quant lens: track observable proxies rather than inventing private-company financials: launch cadence, service availability, pricing changes, regulatory approvals, and disclosed subscriber milestones. Return 'insufficient data' where metrics are not reported.",
  risk: "Risk lens: launch disruption can affect capacity, service quality, revenue timing, and capital needs together. Stress these linked risks jointly rather than applying one isolated downside percentage.",
  technical: "Technical lens: there is no public SpaceX share-price series for a valid chart signal. Do not substitute another company's chart and call it SpaceX technical analysis.",
};

const portfolioAnalysis: Record<AgentId, string> = {
  portfolio: "Portfolio lens: there is no SpaceX ticker or public quote to match against this account. This is a hypothetical theme allocation, not a claim that the user owns SpaceX.",
  macro: "Macro lens: distinguish portfolio concentration from business conviction. A Starlink thesis does not remove funding, spectrum, regulation, execution, or private-valuation risks.",
  "left-side": "Left-side lens: do not build a paper position around a guessed private valuation. Define observation criteria and entry conditions first; without a reference price, risk sizing is unreliable.",
  quant: "Quant lens: without a public price series, volatility estimate, and investable instrument, position sizing cannot be calculated honestly. The model should say 'not enough data' rather than inventing a quantity.",
  risk: "Risk lens: map indirect exposure separately, but do not assume technology or aerospace holdings move one-for-one with Starlink. Use a conservative correlation range in scenario analysis.",
  technical: "Technical lens: no SpaceX OHLC candles are available. Do not manufacture support, resistance, or entry levels for a private security.",
};

export function QuantAssistant({ memoryStorage }: QuantAssistantProps) {
  const { t } = useI18n();
  const [open, setOpen] = useState(false);
  const [agentId, setAgentId] = useState<AgentId>("quant");
  const [question, setQuestion] = useState("");
  const [memoryDraft, setMemoryDraft] = useState("");
  const [memory, setMemory] = useState(() => loadLongTermMemory(memoryStorage));
  const [rounds, setRounds] = useState<Round[]>(initialRounds);
  const agent = agents.find(item => item.id === agentId) || agents[0];

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const prompt = question.trim();
    if (!prompt) return;
    setRounds(current => [...current, runParallelReview(Date.now(), prompt, agent.id, current.at(-1)?.intents || [], memory)]);
    setQuestion("");
  }

  function remember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMemory(saveLongTermMemory(memoryDraft, memoryStorage));
    setMemoryDraft("");
  }

  function forget(value: string) {
    setMemory(removeLongTermMemory(value, memoryStorage));
  }

  return <div className="quant-assistant">
    {open && <section className="quant-assistant-panel" role="dialog" aria-label={t("Quant assistants")}>
      <header className="quant-assistant-header">
        <div><p className="panel-kicker">{t("PAPER RESEARCH · MULTI-AGENT HARNESS")}</p><h2>{t("Quant assistants")}</h2></div>
        <button type="button" className="quant-assistant-close" aria-label={t("Close assistant")} onClick={() => setOpen(false)}>×</button>
      </header>
      <div className="quant-agent-list" role="group" aria-label={t("Choose a lead analyst")}>
        {agents.map(item => <button type="button" key={item.id} className={item.id === agentId ? "active" : ""} aria-pressed={item.id === agentId} onClick={() => setAgentId(item.id)}>
          <strong>{t(item.name)}</strong><small>{t(item.specialty)}</small>
        </button>)}
      </div>
      <div className="quant-chat-context"><span className="quant-agent-mark">Q</span><p><strong>{t(agent.name)} · {t("Lead analyst")}</strong><small>{t("All 6 analysts are dispatched in parallel for each question")}</small></p></div>
      <details className="quant-memory">
        <summary>{t("Long-term memory")} <span>{memory.length}</span></summary>
        <p>{t("Only preferences you add are saved on this device. Account and position data are not stored here.")}</p>
        <form onSubmit={remember}>
          <label className="visually-hidden" htmlFor="quant-memory-draft">{t("Memory preference")}</label>
          <input id="quant-memory-draft" value={memoryDraft} maxLength={120} onChange={event => setMemoryDraft(event.target.value)} placeholder={t("Memory preference")} />
          <button type="submit" disabled={!memoryDraft.trim()}>{t("Remember")}</button>
        </form>
        {memory.length > 0 ? <ul>{memory.map(item => <li key={item}><span>{item}</span><button type="button" aria-label={t("Forget {item}", { item })} onClick={() => forget(item)}>×</button></li>)}</ul> : <small className="quant-memory-empty">{t("No saved preferences")}</small>}
      </details>
      <div className="quant-chat-messages" aria-live="polite">
        {rounds.map((round, index) => <section className="quant-round" key={round.id} aria-label={t("Analysis round {number}", { number: index + 1 })}>
          <div className="quant-round-heading"><span>{t("ROUND {number}", { number: index + 1 })}</span><small>{t("Parallel analysis · {count} agents", { count: round.analyses.length })}</small></div>
          {round.intents.length > 0 && <div className="quant-intents" aria-label={t("Detected intents")}>{round.intents.map(intent => <span key={intent}>{t(intent[0].toUpperCase() + intent.slice(1))}</span>)}</div>}
          <article className="quant-message user"><p>{t(round.prompt)}</p></article>
          <div className="quant-parallel-results">
            {round.analyses.map(result => {
              const resultAgent = agents.find(item => item.id === result.agent);
              return <article className="quant-message assistant" key={result.agent}>
                <strong className="quant-message-agent">{t(resultAgent?.name || "Quant assistant")}</strong>
                <p>{t(result.text)}</p><small>{t("Mock response · educational analysis only")}</small>
              </article>;
            })}
          </div>
          <article className="quant-synthesis"><strong>{t("Harness synthesis")}</strong><p>{t(round.synthesis)}</p></article>
          {round.memory.length > 0 && <p className="quant-memory-applied"><strong>{t("Applied memory")}</strong> · {round.memory.join(" · ")}</p>}
        </section>)}
      </div>
      <form className="quant-chat-form" onSubmit={submit}>
        <label className="visually-hidden" htmlFor="quant-assistant-question">{t("Ask a question")}</label>
        <textarea id="quant-assistant-question" aria-label={t("Ask a question")} value={question} onChange={event => setQuestion(event.target.value)} placeholder={t("Ask a question")} rows={2} />
        <button type="submit" className="quant-send" aria-label={t("Send message")} disabled={!question.trim()}>↑</button>
      </form>
      <p className="quant-assistant-disclaimer">{t("Reasoning summary only; no private model thought process is shown. SpaceX is private; Starlink is not a standalone listed stock. Educational simulation, not investment advice.")}</p>
    </section>}
    <button type="button" className={`quant-assistant-launcher${open ? " is-open" : ""}`} aria-label={t(open ? "Close quant assistants" : "Open quant assistants")} aria-expanded={open} onClick={() => setOpen(value => !value)}>
      <span aria-hidden="true">Q</span><i />
    </button>
  </div>;
}
