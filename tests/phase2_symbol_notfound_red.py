#!/usr/bin/env python3
"""RED/source guards for Phase 2 unknown-symbol semantics."""

from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parents[1]
DOMAIN = (ROOT / "apps/backend/internal/domain/market/quote.go").read_text()
REPO = (ROOT / "apps/backend/internal/repository/mysqlrepo/market.go").read_text()
PROVIDER = (ROOT / "apps/backend/internal/integration/marketdata/provider.go").read_text()
OPENAPI = (ROOT / "docs/api/openapi.yaml").read_text()


class Phase2SymbolNotFoundRed(unittest.TestCase):
    def test_typed_not_found_and_repository_translation_exist(self):
        self.assertIn("ErrNotFound", DOMAIN)
        self.assertRegex(REPO, r"sql\.ErrNoRows[\s\S]{0,180}market\.ErrNotFound")

    def test_provider_and_mock_preserve_not_found(self):
        self.assertRegex(PROVIDER, r"errors\.Is\(err,\s*market\.ErrNotFound\)")
        self.assertRegex(PROVIDER, r"return market\.Quote\{\},\s*market\.ErrNotFound")
        self.assertRegex(PROVIDER, r"return nil,\s*market\.ErrNotFound")

    def test_candles_document_unavailable_response(self):
        candles = OPENAPI.split("/api/v1/markets/candles/{symbol}", 1)[1].split("/api/v1/watchlist", 1)[0]
        self.assertIn("'503': { $ref: '#/components/responses/Unavailable' }", candles)


if __name__ == "__main__":
    unittest.main()
