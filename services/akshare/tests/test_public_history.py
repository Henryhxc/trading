import unittest
from datetime import datetime, timezone
from services.akshare import public_history


class PublicHistoryTests(unittest.TestCase):
    def test_tencent_unadjusted_rows_keep_missing_turnover_null(self):
        observed = []
        def fetch(url, params):
            observed.append(params)
            return {"data": {"hk00700": {"day": [["2026-10-23", "10", "11", "12", "9", "100"]]}}}
        rows = public_history.fetch_history("00700", "1d", "20260424", "20261024", datetime.now(timezone.utc), fetch)
        self.assertEqual(rows[0]["source"], "tencent")
        self.assertIsNone(rows[0]["turnover"])
        self.assertEqual(rows[0]["adjustment"], "none")
        self.assertTrue(observed[0]["param"].endswith(","))

    def test_never_substitutes_adjusted_rows_for_unadjusted_history(self):
        fetch = lambda *args: {"data": {"hk00700": {"qfqday": [["2026-10-23", "10", "11", "12", "9", "100"]]}}}
        with self.assertRaises(ValueError):
            public_history.fetch_history("00700", "1d", "20260424", "20261024", datetime.now(timezone.utc), fetch)

    def test_minute_rows_keep_native_period_time_and_source(self):
        fetch = lambda *args: {"data": {"trends": ["2026-10-23 09:31,10,11,12,9,100,1100,11"]}}
        rows = public_history.fetch_history("00700", "1m", "20260424", "20261024", datetime.now(timezone.utc), fetch)
        self.assertEqual(rows[0]["timestamp"], "2026-10-23T01:31:00Z")
        self.assertEqual(rows[0]["source"], "eastmoney")
        self.assertEqual(rows[0]["interval"], "1m")

    def test_invalid_native_ohlc_is_not_reconstructed_or_inserted(self):
        fetch = lambda *args: {"data": {"trends": ["2026-10-23 09:31,0,11,12,9,100,1100,11"]}}
        with self.assertRaises(ValueError):
            public_history.fetch_history("00700", "1m", "20260424", "20261024", datetime.now(timezone.utc), fetch)


if __name__ == "__main__":
    unittest.main()
