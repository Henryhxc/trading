import re
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MIGRATIONS = ROOT / "db" / "migrations"
CANONICAL_UP = MIGRATIONS / "002_stocks_daily_quotes_watchlists.sql"
CANONICAL_DOWN = MIGRATIONS / "002_stocks_daily_quotes_watchlists.down.sql"


class Phase2MigrationContractGuard(unittest.TestCase):
    """Protect the accepted three-stage migration numbering contract."""

    def read_required(self, path: Path) -> str:
        if not path.is_file():
            self.fail(f"required canonical migration is missing: {path.name}")
        return path.read_text(encoding="utf-8")

    def test_phase2_is_one_canonical_002_migration(self):
        self.assertTrue(
            CANONICAL_UP.is_file(),
            "Phase 2 must use 002_stocks_daily_quotes_watchlists.sql",
        )
        self.assertTrue(
            CANONICAL_DOWN.is_file(),
            "Phase 2 must have a matching canonical down migration",
        )
        self.assertFalse(
            (MIGRATIONS / "002_stocks_daily_quotes.sql").exists(),
            "the legacy split 002 migration must not remain in the release directory",
        )
        self.assertFalse(
            (MIGRATIONS / "003_watchlists.sql").exists(),
            "003 is reserved for orders/positions, not Phase 2 watchlists",
        )
        self.assertFalse(
            (MIGRATIONS / "003_watchlists.down.sql").exists(),
            "the legacy split watchlist rollback must not remain",
        )

    def test_canonical_up_owns_all_phase2_tables_and_reconciles_legacy_ledger(self):
        sql = self.read_required(CANONICAL_UP)
        for table in ("stocks", "stock_daily_quotes", "watchlists"):
            self.assertRegex(
                sql,
                rf"(?i)CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+{table}\b",
                f"canonical Phase 2 migration must create {table}",
            )
        for legacy_version in (
            "002_stocks_daily_quotes.sql",
            "003_watchlists.sql",
        ):
            self.assertIn(
                legacy_version,
                sql,
                f"canonical apply must reconcile legacy ledger row {legacy_version}",
            )

    def test_canonical_down_reverses_dependencies_and_clears_old_and_new_ledger_rows(self):
        sql = self.read_required(CANONICAL_DOWN)
        drops = [
            re.search(rf"(?i)DROP\s+TABLE\s+IF\s+EXISTS\s+{table}\b", sql)
            for table in ("watchlists", "stock_daily_quotes", "stocks")
        ]
        self.assertTrue(all(drops), "canonical rollback must drop every Phase 2 table")
        self.assertLess(drops[0].start(), drops[1].start())
        self.assertLess(drops[1].start(), drops[2].start())
        for version in (
            "002_stocks_daily_quotes_watchlists.sql",
            "002_stocks_daily_quotes.sql",
            "003_watchlists.sql",
        ):
            self.assertIn(version, sql, f"rollback must clear ledger row {version}")

    def test_only_orders_positions_may_own_the_003_slot(self):
        phase3_up = sorted(
            path.name
            for path in MIGRATIONS.glob("003_*.sql")
            if not path.name.endswith(".down.sql")
        )
        self.assertIn(
            phase3_up,
            ([], ["003_orders_positions.sql"]),
            "003 is either not implemented yet or is exactly orders/positions",
        )


if __name__ == "__main__":
    unittest.main()
