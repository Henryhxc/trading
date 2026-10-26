import unittest
from datetime import datetime, timezone
from services.akshare import scheduler
from services.akshare.collector import collect_history


class CandlestickTests(unittest.TestCase):
    def test_configured_public_history_does_not_wait_on_unavailable_akshare(self):
        akshare_calls = []
        class API:
            def stock_hk_hist(self, **kwargs):
                akshare_calls.append("daily")
                raise ConnectionError("AkShare unavailable")
            def stock_hk_hist_min_em(self, **kwargs):
                akshare_calls.append("minute")
                raise ConnectionError("AkShare unavailable")
        config = scheduler.Config.from_env({"HISTORY_PROVIDER": "public", "AKSHARE_SYMBOLS": "00700", "AKSHARE_MINUTE_PERIODS": "1m"})
        calls = []
        def fetch(symbol, interval, *args):
            calls.append(interval)
            return [{"source": "tencent" if interval == "1d" else "yahoo"}]
        rows = collect_history(API(), config, datetime.now(timezone.utc), fetch)
        self.assertEqual(akshare_calls, [], "Public history must not wait for AkShare failures")
        self.assertEqual(calls, ["1d", "1m"])
        self.assertEqual([row["source"] for row in rows], ["tencent", "yahoo"])
        with self.assertRaises(ValueError):
            scheduler.Config.from_env({"HISTORY_PROVIDER": "unknown"})

    def test_failed_akshare_uses_explicit_fallback_without_relabelling(self):
        class API:
            def stock_hk_hist(self, **kwargs):
                raise ConnectionError("offline")
        calls = []
        def fallback(symbol, interval, start, end, now):
            calls.append((symbol, interval, start, end))
            return [{"source": "tencent", "symbol": symbol}]
        bars = collect_history(API(), scheduler.Config(symbols=("00700",), minute_periods=()), datetime(2026, 10, 24, tzinfo=timezone.utc), fallback)
        self.assertEqual(bars[0]["source"], "tencent")
        self.assertEqual(calls, [("00700", "1d", "20260424", "20261024")])

    def test_large_history_is_delivered_in_bounded_chunks(self):
        delivered = []
        rows = [{"index": n} for n in range(801)]
        service = scheduler.Scheduler(scheduler.Config(), lambda: [], delivered.append, fetch_history=lambda: rows)
        service.run_once()
        self.assertEqual([len(p["candles"]) for p in delivered], [400, 400, 1])
        self.assertEqual([r["index"] for p in delivered for r in p["candles"]], list(range(801)))

    def test_one_native_period_failure_preserves_other_periods(self):
        class Frame:
            def to_dict(self, kind):
                return [{"时间": "2026-10-23 09:35:00", "开盘": 10, "最高": 12, "最低": 9, "收盘": 11, "成交量": 100, "成交额": 1100}]
        class API:
            def stock_hk_hist(self, **kwargs):
                raise ConnectionError("daily unavailable")
            def stock_hk_hist_min_em(self, **kwargs):
                if kwargs["period"] == "1":
                    raise ConnectionError("1m unavailable")
                return Frame()
        bars = collect_history(API(), scheduler.Config(symbols=("00700",), minute_periods=("1m", "5m")), datetime.now(timezone.utc))
        self.assertEqual([bar["interval"] for bar in bars], ["5m"])

    def test_failed_spot_does_not_discard_successful_history(self):
        # Catch the current coupling where spot failure prevents any backfill.
        delivered = []
        def failed_spot():
            raise ConnectionError("upstream disconnected")
        service = scheduler.Scheduler(scheduler.Config(), failed_spot, delivered.append,
                                      fetch_history=lambda: [{"symbol": "00700"}])
        service.run_once()
        self.assertEqual(delivered[0]["quotes"], [])
        self.assertEqual(delivered[0]["candles"], [{"symbol": "00700"}])

    def test_default_history_covers_previous_six_calendar_months(self):
        config = scheduler.Config.from_env({"AKSHARE_SYMBOLS": "00700"})
        bounds = getattr(config, "history_bounds", lambda now: ("", ""))
        self.assertEqual(bounds(datetime(2026, 10, 24, tzinfo=timezone.utc)), ("20260424", "20261024"))
        self.assertEqual(bounds(datetime(2026, 8, 31, tzinfo=timezone.utc)), ("20260228", "20260831"))

    def test_history_collection_is_limited_to_three_symbols(self):
        default = scheduler.Config.from_env({})
        configured = scheduler.Config.from_env({"AKSHARE_SYMBOLS": "00700,00005,09988,01810,00941"})
        self.assertEqual(default.symbols, ("00700", "09988", "01810"))
        self.assertEqual(configured.symbols, ("00700", "00005", "09988"))

    def test_history_start_cannot_expand_beyond_six_months(self):
        config = scheduler.Config.from_env({"AKSHARE_SYMBOLS": "00700", "AKSHARE_HISTORY_START": "20200101"})
        self.assertEqual(config.history_bounds(datetime(2026, 10, 24, tzinfo=timezone.utc)), ("20260424", "20261024"))

    def test_successful_history_backfill_is_not_repeated_during_quote_polling(self):
        history_calls = []
        delivered = []
        service = scheduler.Scheduler(
            scheduler.Config(symbols=("00700",)),
            lambda: [],
            delivered.append,
            fetch_history=lambda: history_calls.append("fetch") or [{"symbol": "00700"}],
        )
        start = datetime(2026, 10, 24, tzinfo=timezone.utc)
        service.run_once(start)
        service.run_once(start.replace(minute=15))
        self.assertEqual(history_calls, ["fetch"])
        service.run_once(start.replace(day=25))
        self.assertEqual(history_calls, ["fetch", "fetch"])

    def test_minute_normalization_retains_actual_time_and_native_period(self):
        normalize = getattr(scheduler, "normalize_minute_rows", lambda *args: [])
        bars = normalize([{"时间": "2026-10-23 09:35:00", "开盘": 10, "最高": 12,
                           "最低": 9, "收盘": 11, "成交量": 100, "成交额": 1100}],
                         "00700", "5m", datetime(2026, 10, 24, tzinfo=timezone.utc))
        self.assertEqual(len(bars), 1)
        self.assertEqual(bars[0]["timestamp"], "2026-10-23T01:35:00Z")
        self.assertEqual(bars[0]["interval"], "5m")
        self.assertEqual(bars[0]["data_status"], "unofficial")
        self.assertEqual(bars[0]["source"], "akshare")

    def test_duplicate_minutes_keep_most_recent_row_and_sort(self):
        normalize = getattr(scheduler, "normalize_minute_rows", lambda *args: [])
        row = {"时间": "2026-10-23 09:31:00", "开盘": 10, "最高": 12, "最低": 9,
               "收盘": 11, "成交量": 100, "成交额": 1100}
        bars = normalize([row, dict(row, 收盘=12)], "00700", "1m", datetime.now(timezone.utc))
        self.assertEqual(len(bars), 1)
        self.assertEqual(bars[0]["close"], "12.0000")


if __name__ == "__main__":
    unittest.main()
