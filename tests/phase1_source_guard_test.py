from pathlib import Path
import re
import unittest


ROOT = Path(__file__).resolve().parents[1]
SOURCE_ROOTS = (ROOT / "apps", ROOT / "deploy")
MIGRATIONS_ROOT = ROOT / "db" / "migrations"


def read_project_sources(roots):
    suffixes = {".go", ".sql", ".yaml", ".yml", ".tsx", ".ts", ".jsx", ".js"}
    files = []
    for root in roots:
        if root.exists():
            files.extend(path for path in root.rglob("*") if path.is_file() and path.suffix in suffixes)
    return {path: path.read_text(encoding="utf-8", errors="replace") for path in files}


class Phase1SourceGuardTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.sources = read_project_sources(SOURCE_ROOTS)
        cls.text = "\n".join(cls.sources.values())
        cls.migrations = read_project_sources((MIGRATIONS_ROOT,))
        cls.migration_text = "\n".join(cls.migrations.values())

    def require(self, expression, text, missing):
        self.assertRegex(text, expression, f"Missing Phase 1 behavior: {missing}")

    def test_auth_routes_are_wired(self):
        self.require(r"/auth/register", self.text, "HTTP route for POST /api/v1/auth/register")

    def test_login_route_is_wired(self):
        self.require(r"/auth/login", self.text, "HTTP route for POST /api/v1/auth/login")

    def test_logout_route_is_wired(self):
        self.require(r"/auth/logout", self.text, "HTTP route for POST /api/v1/auth/logout")

    def test_current_user_route_is_wired(self):
        self.require(r"/me", self.text, "authenticated current-user route GET /api/v1/me")

    def test_session_expiry_is_enforced(self):
        self.require(
            r"(?i)(expires_at|expiresat|expiry|expiration)",
            self.text + "\n" + self.migration_text,
            "session expiry field or check",
        )

    def test_logout_revokes_sessions(self):
        self.require(
            r"(?i)(revoked_at|revokedat|revoke|revoked)",
            self.text + "\n" + self.migration_text,
            "server-side session revocation on logout",
        )

    def test_passwords_use_a_strong_hash(self):
        self.require(
            r"(?i)(bcrypt|argon2id|argon2)",
            self.text,
            "bcrypt or Argon2 password hashing",
        )

    def test_protected_handlers_receive_the_session_actor(self):
        self.require(
            r"(?i)(user_id|userid|actor_id|actorid)",
            self.text,
            "authenticated actor identity available to protected handlers",
        )

    def test_protected_data_queries_are_actor_scoped(self):
        self.require(
            r"(?i)(user_id\s*=\s*\?|user_id\s*==|user_id\s*=\s*actor|actor.*user_id|user_id.*actor)",
            self.text,
            "repository/data access constrained by the authenticated user",
        )

    def test_users_table_migration_exists(self):
        self.require(r"(?i)create\s+table\s+.*users", self.migration_text, "users table migration")

    def test_sessions_table_migration_exists(self):
        self.require(r"(?i)create\s+table\s+.*sessions", self.migration_text, "sessions table migration")

    def test_accounts_table_migration_exists(self):
        self.require(r"(?i)create\s+table\s+.*accounts", self.migration_text, "accounts table migration")

    def test_auth_migration_rollback_scripts_exist(self):
        sql_files = list(MIGRATIONS_ROOT.rglob("*.sql")) if MIGRATIONS_ROOT.exists() else []
        rollback_files = [
            path
            for path in sql_files
            if re.search(r"(?i)(\.down\.sql$|rollback|down/)", str(path.relative_to(MIGRATIONS_ROOT)))
        ]
        self.assertTrue(rollback_files, "Missing Phase 1 behavior: reversible auth migration rollback scripts")
        rollback_text = "\n".join(path.read_text(encoding="utf-8", errors="replace") for path in rollback_files)
        for table in ("users", "sessions", "accounts"):
            with self.subTest(table=table):
                self.require(rf"(?i)drop\s+table\s+(if\s+exists\s+)?`?{table}`?", rollback_text,
                             f"rollback drops {table} table")

    def test_health_endpoint_is_wired(self):
        self.require(r"(?i)(/health|/healthz|/readyz)", self.text, "backend health endpoint")


class ConciseResult(unittest.TextTestResult):
    def addFailure(self, test, err):
        unittest.TestResult.addFailure(self, test, err)
        self.stream.write(f"FAIL {test._testMethodName}: {err[1]}\n")

    def addError(self, test, err):
        unittest.TestResult.addError(self, test, err)
        self.stream.write(f"ERROR {test._testMethodName}: {err[1]}\n")

    def printErrors(self):
        pass


if __name__ == "__main__":
    suite = unittest.defaultTestLoader.loadTestsFromTestCase(Phase1SourceGuardTests)
    result = unittest.TextTestRunner(verbosity=0, resultclass=ConciseResult).run(suite)
    raise SystemExit(not result.wasSuccessful())
