"""Focused RED contract for Phase 2 market outage and search empty-state fixes."""

import pathlib
import re
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]
DOMAIN = ROOT / "apps/backend/internal/domain/market/quote.go"
HANDLER = ROOT / "apps/backend/internal/httpapi/market_handler.go"
WORKBENCH = ROOT / "apps/frontend/src/pages/WorkbenchPage.tsx"


class Phase2ErrorEmptyReworkRed(unittest.TestCase):
    def test_market_domain_exposes_shared_unavailable_error_and_handlers_map_it(self):
        domain = DOMAIN.read_text()
        handler = HANDLER.read_text()
        self.assertIn("ErrUnavailable", domain)
        self.assertRegex(handler, r"errors\.Is\(err,\s*market\.ErrUnavailable\)")
        self.assertRegex(
            handler,
            r"getQuote[\s\S]{0,1800}StatusServiceUnavailable[\s\S]{0,500}StatusNotFound",
        )
        self.assertRegex(
            handler,
            r"getDailyCandles[\s\S]{0,2200}StatusServiceUnavailable[\s\S]{0,500}StatusNotFound",
        )

    def test_search_success_empty_state_is_distinct_from_transport_error(self):
        source = WORKBENCH.read_text()
        self.assertRegex(source, r"search(State|Status)")
        self.assertRegex(source, r"results\.length\s*===\s*0")
        self.assertRegex(source, r"No securities (?:match|found)")
        self.assertRegex(source, r"search(Error|Failure)")
        self.assertRegex(source, r"search(State|Status)\s*===\s*[\"']error[\"']")


if __name__ == "__main__":
    unittest.main(verbosity=2)
