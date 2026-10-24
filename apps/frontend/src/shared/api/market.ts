import { apiRequest } from "./client";

export type DataStatus = "simulated" | "unofficial" | "delayed" | "stale" | "unavailable" | "live";
export type SymbolItem = { symbol: string; display_name: string; quote?: Quote };
export type Quote = {
  symbol: string; display_name: string; last_price: number | null; bid: number | null; ask: number | null;
  change: number | null; change_percent: number | null; volume: number; turnover: number;
  as_of: string; source: string; data_status: DataStatus;
};
export type CandleInterval = "1m" | "5m" | "15m" | "30m" | "60m" | "1d" | "1w" | "1M" | "1y";
export type Candle = { trade_date: string; timestamp?: string; open: number; high: number; low: number; close: number; volume: number; turnover: number | null; source: string; data_status: DataStatus; collected_at: string };
export type CandleSeries = { symbol: string; interval: CandleInterval; items: Candle[]; source: string; data_status: DataStatus; as_of: string | null };
export type WatchlistItem = SymbolItem & { sort_order: number; quote: Quote };

export const marketApi = {
  search: (query: string) => apiRequest<{ items: SymbolItem[] }>(`/markets/symbols?query=${encodeURIComponent(query)}`),
  quote: (symbol: string) => apiRequest<Quote>(`/markets/quotes/${symbol}`),
  candles: (symbol: string, interval: CandleInterval, from: string, to: string) => apiRequest<CandleSeries>(`/markets/candles/${encodeURIComponent(symbol)}?${new URLSearchParams({ interval, from, to, limit: "5000" })}`),
  watchlist: () => apiRequest<{ items: WatchlistItem[] }>("/watchlist"),
  addWatchlist: (symbol: string) => apiRequest<WatchlistItem>("/watchlist", { symbol }),
  reorderWatchlist: (symbols: string[]) => apiRequest<null>("/watchlist/order", { symbols }, "PUT"),
  removeWatchlist: (symbol: string) => apiRequest<null>(`/watchlist/${symbol}`, undefined, "DELETE"),
};
