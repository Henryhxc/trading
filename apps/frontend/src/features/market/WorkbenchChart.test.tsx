import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { WorkbenchPage } from "../../pages/WorkbenchPage";
import { marketApi, type Candle, type CandleSeries } from "../../shared/api/market";
import { tradingApi } from "../../shared/api/trading";

const graph = vi.hoisted(() => ({ candles: vi.fn(), volume: vi.fn(), remove: vi.fn() }));
vi.mock("lightweight-charts", () => ({
  CandlestickSeries: "Candlestick", HistogramSeries: "Histogram", ColorType: { Solid: "solid" },
  createChart: () => ({
    addSeries: (type: string) => ({ setData: type === "Candlestick" ? graph.candles : graph.volume, priceScale: () => ({ applyOptions: vi.fn() }) }),
    timeScale: () => ({ fitContent: vi.fn() }), applyOptions: vi.fn(),
    subscribeCrosshairMove: vi.fn(), unsubscribeCrosshairMove: vi.fn(), remove: graph.remove,
  }),
}));
vi.mock("../../app/SessionProvider", () => ({ useSession: () => ({ user: { username: "Chart tester" }, clearUser: vi.fn() }) }));
vi.mock("../../pages/TradingPanel", () => ({ TradingPanel: () => null }));
vi.mock("../../shared/api/market", () => ({ marketApi: { quote: vi.fn(), candles: vi.fn(), watchlist: vi.fn(), search: vi.fn() } }));
vi.mock("../../shared/api/trading", () => ({ tradingApi: { account: vi.fn(), positions: vi.fn() } }));

const bar: Candle = { trade_date: "2026-10-23T00:00:00Z", open: 10, high: 12, low: 9, close: 11, volume: 100, turnover: 1100, source: "akshare", data_status: "unofficial", collected_at: "2026-10-24T00:00:00Z" };
const series: CandleSeries = { symbol: "00700", interval: "1d", items: [bar], source: "akshare", data_status: "unofficial", as_of: "2026-10-24T00:00:00Z" };

describe("Workbench candle behavior", () => {
  beforeEach(() => {
    sessionStorage.clear();
    vi.mocked(marketApi.search).mockResolvedValue({ items: [] });
    vi.mocked(tradingApi.account).mockResolvedValue(null);
    vi.mocked(tradingApi.positions).mockResolvedValue({ items: [] });
    vi.mocked(marketApi.quote).mockResolvedValue({ symbol: "00700", display_name: "Tencent Holdings", last_price: 11, change: 1, change_percent: 10, bid: 10, ask: 12, volume: 100, turnover: 1100, as_of: "2026-10-24T00:00:00Z", source: "akshare", data_status: "delayed" });
    vi.mocked(marketApi.watchlist).mockResolvedValue({ items: [] });
    vi.mocked(marketApi.candles).mockResolvedValue(series);
  });

  it("offers all native minute and calendar periods and requests the selected period", async () => {
    render(<WorkbenchPage />);
    await screen.findByText("Tencent Holdings");
    for (const label of ["1分", "5分", "15分", "30分", "60分", "日", "周", "月", "年"]) expect(screen.getByRole("button", { name: label })).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "5分" }));
    await waitFor(() => expect(marketApi.candles).toHaveBeenLastCalledWith("00700", "5m", expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/), expect.stringMatching(/^\d{4}-\d{2}-\d{2}$/)));
  });

  it("renders actual OHLC data and green/red volume, not a close-only bar", async () => {
    render(<WorkbenchPage />);
    await waitFor(() => expect(graph.candles).toHaveBeenCalledWith([{ time: "2026-10-23", open: 10, high: 12, low: 9, close: 11 }]));
    expect(graph.volume).toHaveBeenCalledWith([{ time: "2026-10-23", value: 100, color: "#22c55e80" }]);
  });

  it("keeps the quote visible on candle failure and supports retry", async () => {
    vi.mocked(marketApi.candles).mockRejectedValueOnce(new Error("History offline"));
    render(<WorkbenchPage />);
    expect(await screen.findByText("Tencent Holdings")).toBeTruthy();
    expect(await screen.findByText(/History offline/)).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "刷新行情" }));
    await waitFor(() => expect(graph.candles).toHaveBeenCalled());
  });

  it("shows missing native minutes explicitly without substituting daily bars", async () => {
    vi.mocked(marketApi.candles).mockImplementation(async (symbol, interval) => ({ ...series, symbol, interval, items: [], source: "none", data_status: "unavailable" }));
    render(<WorkbenchPage />);
    fireEvent.click(await screen.findByRole("button", { name: "1分" }));
    expect(await screen.findByText(/暂无原生分钟行情/)).toBeTruthy();
  });
});
