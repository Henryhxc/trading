export type IntentId = "portfolio" | "macro" | "industry" | "market" | "quant" | "risk";

const MEMORY_KEY = "ux-quant-assistant-memory-v1";

function browserStorage(): Storage | undefined {
  try { return window.localStorage; } catch { return undefined; }
}

const intentRules: Array<[IntentId, RegExp]> = [
  ["portfolio", /portfolio|position|holding|exposure|持仓|仓位|组合|敞口/i],
  ["macro", /macro|rate|election|inflation|economy|利率|宏观|选举|通胀|经济/i],
  ["industry", /industry|sector|company|spacex|starlink|semiconductor|产业|行业|公司|半导体|星链/i],
  ["market", /market|price|quote|candle|trend|成交量|行情|价格|股价|趋势|k线/i],
  ["quant", /quant|factor|backtest|strategy|signal|量化|因子|回测|策略|信号/i],
  ["risk", /risk|drawdown|volatility|loss|回撤|波动|亏损|风险/i],
];

const followUpPattern = /^(那|然后|这个|它|再呢|为什么|如何|how about|what about|and then|why|what next)/i;

export function classifyIntents(prompt: string, previous: IntentId[] = []): IntentId[] {
  const matched = intentRules.filter(([, pattern]) => pattern.test(prompt)).map(([intent]) => intent);
  if (matched.length) return matched;
  if (previous.length && followUpPattern.test(prompt.trim())) return previous;
  return ["market"];
}

export function loadLongTermMemory(storage: Storage | undefined = browserStorage()): string[] {
  if (!storage) return [];
  try {
    const parsed: unknown = JSON.parse(storage.getItem(MEMORY_KEY) || "[]");
    return Array.isArray(parsed) ? parsed.filter((item): item is string => typeof item === "string") : [];
  } catch {
    return [];
  }
}

export function saveLongTermMemory(value: string, storage: Storage | undefined = browserStorage()): string[] {
  const memory = loadLongTermMemory(storage);
  const trimmed = value.trim().slice(0, 120);
  if (!trimmed || memory.some(item => item.toLocaleLowerCase() === trimmed.toLocaleLowerCase())) return memory;
  const next = [...memory, trimmed].slice(-10);
  storage?.setItem(MEMORY_KEY, JSON.stringify(next));
  return next;
}

export function removeLongTermMemory(value: string, storage: Storage | undefined = browserStorage()): string[] {
  const next = loadLongTermMemory(storage).filter(item => item !== value);
  storage?.setItem(MEMORY_KEY, JSON.stringify(next));
  return next;
}
