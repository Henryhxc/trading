import unittest
from datetime import datetime, timezone
from services.akshare import yahoo_history


class YahooHistoryTests(unittest.TestCase):
    def test_native_minutes_keep_hk_timestamp_and_missing_amount(self):
        observed = []
        def fetch(url, params):
            observed.append((url, params))
            return {"chart": {"result": [{"timestamp": [1792719000, 1792719060], "indicators": {"quote": [{"open": [10, None], "close": [11, None], "high": [12, None], "low": [9, None], "volume": [100, 0]}]}}]}}
        rows = yahoo_history.fetch_history("00700", "1m", "20260424", "20261024", datetime(2026, 10, 24, tzinfo=timezone.utc), fetch)
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows[0]["source"], "yahoo")
        self.assertIsNone(rows[0]["turnover"])
        self.assertEqual(rows[0]["interval"], "1m")
        self.assertTrue(rows[0]["timestamp"].endswith("Z"))
        self.assertIn("0700.HK", observed[0][0])
        self.assertEqual(observed[0][1]["interval"], "1m")
        self.assertEqual(observed[0][1]["range"], "5d")

    def test_provider_failure_or_invalid_ohlc_does_not_make_bars(self):
        with self.assertRaises(ValueError):
            yahoo_history.fetch_history("00700", "5m", "20260424", "20261024", datetime.now(timezone.utc), lambda *args: {"chart": {"result": None, "error": {"code": "Not Found"}}})


if __name__ == "__main__":
    unittest.main()
