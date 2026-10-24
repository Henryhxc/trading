"""Phase 2 contract rework guards.

These checks are intentionally added before the implementation fix.  They
cover the response statuses emitted by the market/watchlist handlers and the
declared validation rules for daily-candle query parameters.
"""

import pathlib
import unittest

import yaml


ROOT = pathlib.Path(__file__).resolve().parents[1]
OPENAPI_PATH = ROOT / "docs/api/openapi.yaml"
HANDLER_PATH = ROOT / "apps/backend/internal/httpapi/market_handler.go"


class Phase2ContractReworkRed(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.spec = yaml.safe_load(OPENAPI_PATH.read_text())
        cls.handler = HANDLER_PATH.read_text()

    def test_market_and_watchlist_error_responses_match_handlers(self):
        expected = {
            ("/api/v1/markets/symbols", "get"): {"400", "401", "503"},
            ("/api/v1/markets/quotes/{symbol}", "get"): {"401", "404", "503"},
            ("/api/v1/markets/candles/{symbol}", "get"): {"400", "401", "404"},
            ("/api/v1/watchlist", "get"): {"401", "500"},
            ("/api/v1/watchlist", "post"): {"200", "201", "400", "401", "404"},
            ("/api/v1/watchlist/order", "put"): {"204", "400", "401", "500"},
            ("/api/v1/watchlist/{symbol}", "delete"): {"204", "401", "500"},
        }
        for (path, method), statuses in expected.items():
            with self.subTest(path=path, method=method):
                responses = self.spec["paths"][path][method]["responses"]
                self.assertTrue(
                    statuses.issubset(responses),
                    f"{method.upper()} {path} omits statuses {statuses - set(responses)}",
                )

    def test_candle_query_validation_is_explicit(self):
        body = self.handler[self.handler.index("func (s *server) getDailyCandles"):]
        self.assertIn("time.Parse(\"2006-01-02\"", body)
        self.assertIn("strconv.Atoi", body)
        self.assertIn("limit < 1", body)
        self.assertIn("limit > 365", body)
        self.assertIn("from.After(to)", body)
        self.assertIn('StatusBadRequest, "invalid_input"', body)


if __name__ == "__main__":
    unittest.main(verbosity=2)
