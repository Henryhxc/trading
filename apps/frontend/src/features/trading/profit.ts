export type ProfitOrder = {
  symbol: string;
  side: "buy" | "sell";
  quantity: number;
  filled_price: number | null;
  filled_at: string | null;
  created_at?: string;
  status: string;
};

export type ProfitRow = {
  period: string;
  pnl: number;
  costBasis: number;
  returnPercent: number;
};

export type ProfitPeriod = "day" | "month" | "year";
type Lot = { quantity: number; cost: number };
type Aggregate = { pnl: number; costBasis: number };

const round = (value: number) => Math.round((value + Number.EPSILON) * 100) / 100;

function periodFor(value: string, period: ProfitPeriod) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  if (period === "year") return date.toISOString().slice(0, 4);
  return period === "month" ? date.toISOString().slice(0, 7) : date.toISOString().slice(0, 10);
}

export function aggregateProfit(orders: ProfitOrder[], period: ProfitPeriod): ProfitRow[] {
  const lots = new Map<string, Lot>();
  const aggregates = new Map<string, Aggregate>();
  const sorted = orders
    .filter(order => order.status === "filled" && order.quantity > 0 && order.filled_price != null)
    .sort((left, right) => new Date(left.filled_at || left.created_at || 0).getTime() - new Date(right.filled_at || right.created_at || 0).getTime());

  for (const order of sorted) {
    const price = order.filled_price as number;
    const symbolLot = lots.get(order.symbol) || { quantity: 0, cost: 0 };
    if (order.side === "buy") {
      symbolLot.quantity += order.quantity;
      symbolLot.cost += order.quantity * price;
      lots.set(order.symbol, symbolLot);
      continue;
    }
    const matched = Math.min(order.quantity, symbolLot.quantity);
    const periodKey = periodFor(order.filled_at || order.created_at || "", period);
    if (matched <= 0 || !periodKey) continue;
    const averageCost = symbolLot.cost / symbolLot.quantity;
    const costBasis = matched * averageCost;
    const aggregate = aggregates.get(periodKey) || { pnl: 0, costBasis: 0 };
    aggregate.pnl += matched * (price - averageCost);
    aggregate.costBasis += costBasis;
    aggregates.set(periodKey, aggregate);
    symbolLot.quantity -= matched;
    symbolLot.cost -= costBasis;
    lots.set(order.symbol, symbolLot);
  }

  return [...aggregates.entries()]
    .sort(([left], [right]) => right.localeCompare(left))
    .map(([periodKey, value]) => ({
      period: periodKey,
      pnl: round(value.pnl),
      costBasis: round(value.costBasis),
      returnPercent: value.costBasis ? round(value.pnl / value.costBasis * 100) : 0,
    }));
}
