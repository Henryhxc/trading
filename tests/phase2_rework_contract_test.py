import pathlib
import re
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]
HANDLER = ROOT / "apps/backend/internal/httpapi/market_handler.go"
WATCHLIST_REPO = ROOT / "apps/backend/internal/repository/mysqlrepo/watchlist.go"
MARKET_API = ROOT / "apps/frontend/src/shared/api/market.ts"
WORKBENCH = ROOT / "apps/frontend/src/pages/WorkbenchPage.tsx"


class Phase2ReworkContractTests(unittest.TestCase):
    def test_empty_search_is_rejected_before_provider_call(self):
        source = HANDLER.read_text()
        self.assertRegex(
            source,
            r"searchSymbols[\s\S]{0,1200}TrimSpace\(r\.URL\.Query\(\)\.Get\(\"query\"\)\)[\s\S]{0,700}StatusBadRequest",
        )

    def test_duplicate_add_exposes_created_result_and_handler_uses_200(self):
        repo = WATCHLIST_REPO.read_text()
        handler = HANDLER.read_text()
        self.assertRegex(repo, r"AddWatchlist\([\s\S]{0,2400}(bool|created|RowsAffected)")
        self.assertRegex(handler, r"addWatchlist[\s\S]{0,1800}StatusOK[\s\S]{0,500}StatusCreated")

    def test_workbench_persists_reorder_and_filtered_scroll_anchor(self):
        api = MARKET_API.read_text()
        page = WORKBENCH.read_text()
        self.assertRegex(api, r"reorderWatchlist|reorder:\s*\(")
        self.assertRegex(page, r"ux_watchlist_scroll_anchor|scrollAnchor|scrollTop")
        self.assertRegex(page, r"watchlist/order|reorderWatchlist")
        self.assertRegex(page, r"sort_order|moveFavorite|reorder")


if __name__ == "__main__":
    unittest.main()
