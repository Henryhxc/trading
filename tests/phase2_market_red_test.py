"""Phase 2 market/workbench RED contract.

This suite intentionally fails while the Phase 2 market and watchlist slices are
absent. It is a source-level contract only; no external quote provider is called.
"""

import pathlib
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]
BACKEND = ROOT / "apps/backend/internal"
OPENAPI = (ROOT / "docs/api/openapi.yaml").read_text()


class Phase2MarketContractRed(unittest.TestCase):
    def assert_source_contains(self, relative, *needles):
        path = BACKEND / relative
        self.assertTrue(path.exists(), f"missing Phase 2 source file: {path}")
        source = path.read_text()
        for needle in needles:
            self.assertIn(needle, source, f"{needle!r} missing from {path}")

    def test_provider_interface_and_status_allowlist_exist(self):
        self.assert_source_contains(
            "domain/market/quote.go",
            "type MarketDataProvider interface",
            "DataStatus",
            'DataStatusSimulated',
            'DataStatusUnofficial',
            'DataStatusDelayed',
            'DataStatusStale',
            'DataStatusUnavailable',
        )

    def test_provider_fallback_is_explicit_and_never_live_for_free_sources(self):
        self.assert_source_contains(
            "integration/marketdata/provider.go",
            "authorized",
            "unofficial",
            "mock",
            "DataStatusLive",
            "DataStatusStale",
            "DataStatusUnavailable",
        )
        self.assertNotIn('data_status: "live"', (BACKEND / "integration/marketdata/provider.go").read_text())

    def test_market_routes_cover_search_quote_and_candles(self):
        for route in (
            "/api/v1/markets/symbols",
            "/api/v1/markets/quotes/{symbol}",
            "/api/v1/markets/candles/{symbol}",
        ):
            self.assertIn(route, OPENAPI)
        self.assert_source_contains(
            "httpapi/market_handler.go",
            "searchSymbols",
            "getQuote",
            "getDailyCandles",
            "data_status",
        )

    def test_watchlist_is_idempotent_and_actor_scoped(self):
        self.assert_source_contains(
            "repository/mysqlrepo/watchlist.go",
            "userID",
            "ON DUPLICATE KEY UPDATE",
            "DeleteWatchlist",
        )
        self.assert_source_contains(
            "domain/watchlist/item.go",
            "UserID",
            "Symbol",
        )
        self.assertIn("/api/v1/watchlist", OPENAPI)
        self.assertIn("uniqueItems: true", OPENAPI)

    def test_market_and_watchlist_tests_exist_before_green(self):
        expected = (
            "service/market/service_test.go",
            "service/watchlist/service_test.go",
            "repository/mysqlrepo/watchlist_test.go",
        )
        for relative in expected:
            self.assertTrue((BACKEND / relative).exists(), f"missing RED/GREEN test: {relative}")


if __name__ == "__main__":
    unittest.main(verbosity=2)
