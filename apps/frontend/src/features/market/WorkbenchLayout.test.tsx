import { render, screen, fireEvent, within, waitFor } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import { WorkbenchPage } from "../../pages/WorkbenchPage";
import { marketApi, type Quote } from "../../shared/api/market";
import { tradingApi } from "../../shared/api/trading";

// Canvas has no rendering context in jsdom; keep all workspace and trading components real.
vi.mock("lightweight-charts", () => ({
  CandlestickSeries: "Candle", HistogramSeries: "Volume", ColorType: { Solid: "solid" },
  createChart: () => ({ addSeries: () => ({ setData: vi.fn(), priceScale: () => ({ applyOptions: vi.fn() }) }),
    timeScale: () => ({ fitContent: vi.fn() }), applyOptions: vi.fn(), subscribeCrosshairMove: vi.fn(), unsubscribeCrosshairMove: vi.fn(), remove: vi.fn() }),
}));
vi.mock("../../app/SessionProvider", () => ({ useSession: () => ({ user: { username: "Layout tester" }, clearUser: vi.fn() }) }));
vi.mock("../../shared/api/market", () => ({ marketApi: { search: vi.fn(), quote: vi.fn(), candles: vi.fn(), watchlist: vi.fn(), addWatchlist: vi.fn(), removeWatchlist: vi.fn(), reorderWatchlist: vi.fn() } }));
vi.mock("../../shared/api/trading", () => ({ tradingApi: { account: vi.fn(), positions: vi.fn(), orders: vi.fn(), createOrder: vi.fn(), cancelOrder: vi.fn() } }));
const quote: Quote = { symbol: "00700", display_name: "Tencent Holdings", last_price: 11, bid: 10, ask: 12, change: 1, change_percent: 10, volume: 100, turnover: 1100, as_of: "2026-10-24T00:00:00Z", source: "akshare", data_status: "stale" };

beforeEach(() => {
  vi.clearAllMocks(); sessionStorage.clear();
  vi.mocked(marketApi.search).mockResolvedValue({ items: [
    { symbol: "00005", display_name: "HSBC Holdings" }, { symbol: "00700", display_name: "Tencent Holdings" },
    { symbol: "09988", display_name: "Alibaba Group" },
  ] });
  vi.mocked(marketApi.quote).mockResolvedValue(quote);
  vi.mocked(marketApi.watchlist).mockResolvedValue({ items: [{ symbol: "00700", display_name: "Tencent Holdings", sort_order: 0, quote }] });
  vi.mocked(marketApi.candles).mockResolvedValue({ symbol: "00700", interval: "1d", items: [], source: "none", data_status: "unavailable", as_of: null });
  vi.mocked(tradingApi.account).mockResolvedValue({ currency: "HKD", cash_balance: 1000, available_cash: 1000, initial_cash: 1000, status: "active" });
  vi.mocked(tradingApi.positions).mockResolvedValue({ items: [{ symbol: "00005", display_name: "HSBC Holdings", quantity: 400, available_quantity: 400, average_cost: 10, market_price: 11, market_value: 4400, unrealized_pnl: 400, data_status: "stale" }] });
  vi.mocked(tradingApi.orders).mockResolvedValue({ items: [] });
});

it("shows the market without a search and distinguishes favorites from owned securities", async () => {
  render(<WorkbenchPage />);
  const list = await screen.findByRole("region", { name: "Securities directory" });
  expect(await within(list).findByText("Alibaba Group")).toBeTruthy();
  const owned = within(list).getByRole("button", { name: /00005.*HSBC Holdings/ });
  const saved = within(list).getByRole("button", { name: /00700.*Tencent Holdings/ });
  expect(within(owned).getByText(/◆ Position/)).toBeTruthy();
  expect(within(saved).getByText(/★ Favorite/)).toBeTruthy();
  expect(within(owned).queryByText(/★ Favorite/)).toBeNull();
  expect(within(list).getByText("+10.00%")).toBeTruthy();
});

it("keeps the order ticket beside the chart and full portfolio only in My positions", async () => {
  render(<WorkbenchPage />);
  const ticket = await screen.findByRole("complementary", { name: "Order ticket" });
  expect(within(ticket).getByRole("heading", { name: "Place paper order" })).toBeTruthy();
  expect(screen.queryByRole("heading", { name: "Open positions" })).toBeNull();
  expect(screen.queryByRole("heading", { name: "Order history" })).toBeNull();
  fireEvent.click(screen.getByRole("tab", { name: /My positions/ }));
  expect(await screen.findByRole("heading", { name: "Open positions" })).toBeTruthy();
  expect(screen.getByRole("heading", { name: "Order history" })).toBeTruthy();
  expect(screen.queryByRole("complementary", { name: "Order ticket" })).toBeNull();
});

it("filters the visible list and restores all securities when cleared", async () => {
  render(<WorkbenchPage />);
  const list = await screen.findByRole("region", { name: "Securities directory" });
  await within(list).findByText("Alibaba Group");
  fireEvent.change(screen.getByRole("searchbox", { name: "Search symbols" }), { target: { value: "09988" } });
  expect(within(list).queryByText("HSBC Holdings")).toBeNull();
  expect(within(list).getByText("Alibaba Group")).toBeTruthy();
  fireEvent.change(screen.getByRole("searchbox", { name: "Search symbols" }), { target: { value: "" } });
  expect(within(list).getByText("HSBC Holdings")).toBeTruthy();
});

it("shows a directory error and retries without hiding the chart", async () => {
  vi.mocked(marketApi.search).mockRejectedValueOnce(new Error("Directory offline"));
  render(<WorkbenchPage />);
  expect(await screen.findByText(/Directory offline/)).toBeTruthy();
  expect(screen.getByRole("heading", { name: /K 线行情/ })).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "Retry directory" }));
  await waitFor(() => expect(screen.getByRole("button", { name: /09988.*Alibaba Group/ })).toBeTruthy());
});

it("refreshes cash and ownership markers after submitting an editable paper order", async () => {
  vi.mocked(tradingApi.createOrder).mockResolvedValue({ id: 42, symbol: "00700", side: "buy", order_type: "limit", quantity: 100, limit_price: 11, filled_quantity: 100, filled_price: 11, status: "filled", reject_reason: null, quote_as_of: quote.as_of, created_at: quote.as_of, filled_at: quote.as_of });
  render(<WorkbenchPage />);
  const ticket = await screen.findByRole("complementary", { name: "Order ticket" });
  await within(ticket).findByText("HKD 1,000.00");
  vi.mocked(tradingApi.account).mockResolvedValue({ currency: "HKD", cash_balance: 900, available_cash: 900, initial_cash: 1000, status: "active" });
  vi.mocked(tradingApi.positions).mockResolvedValue({ items: [{ symbol: "00700", display_name: "Tencent Holdings", quantity: 100, available_quantity: 100, average_cost: 11, market_price: 11, market_value: 1100, unrealized_pnl: 0, data_status: "stale" }] });
  fireEvent.change(within(ticket).getByLabelText("Quantity"), { target: { value: "100" } });
  fireEvent.change(within(ticket).getByLabelText("Limit price (HKD)"), { target: { value: "11" } });
  fireEvent.click(within(ticket).getByRole("button", { name: "Submit paper order" }));
  expect(await within(ticket).findByText(/Order filled/)).toBeTruthy();
  expect(await within(ticket).findByText("HKD 900.00")).toBeTruthy();
  const saved = screen.getByRole("button", { name: /00700.*Tencent Holdings/ });
  expect(within(saved).getByText(/◆ Position/)).toBeTruthy();
});
