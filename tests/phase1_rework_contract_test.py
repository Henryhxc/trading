import pathlib
import stat
import unittest


ROOT = pathlib.Path(__file__).resolve().parents[1]


class Phase1ReworkContractTest(unittest.TestCase):
    def test_frontend_has_route_session_and_error_boundary_modules(self):
        main = (ROOT / "apps/frontend/src/main.tsx").read_text()
        source = "\n".join(
            path.read_text()
            for path in (ROOT / "apps/frontend/src").rglob("*.tsx")
        )
        self.assertIn("AppRouter", main)
        self.assertIn("SessionProvider", main)
        self.assertIn("ErrorBoundary", main)
        self.assertIn('"loading"', source)
        self.assertIn('"error"', source)
        self.assertIn("Retry", source)

    def test_openapi_email_and_logout_match_runtime_contract(self):
        api = (ROOT / "docs/api/openapi.yaml").read_text()
        migration = (ROOT / "db/migrations/001_users_sessions_accounts.sql").read_text()
        self.assertIn("email: { type: string, format: email, maxLength: 128 }", api)
        logout = api.split("  /api/v1/auth/logout:", 1)[1].split("  /api/v1/me:", 1)[0]
        self.assertIn("security: []", logout)
        self.assertNotIn("'401':", logout)
        self.assertIn("email VARCHAR(128)", migration)

    def test_openapi_documents_bcrypt_password_bytes_and_auth_server_errors(self):
        api = (ROOT / "docs/api/openapi.yaml").read_text()
        password = api.split("    Password:\n", 1)[1].split("    User:\n", 1)[0]
        normalized_password = " ".join(password.split())
        self.assertIn("minLength: 12", password)
        self.assertIn("maxLength: 72", password)
        self.assertIn("UTF-8 bytes", password)
        self.assertIn("JSON Schema minLength/maxLength count Unicode characters", normalized_password)
        register = api.split("  /api/v1/auth/register:\n", 1)[1].split("  /api/v1/auth/login:\n", 1)[0]
        login = api.split("  /api/v1/auth/login:\n", 1)[1].split("  /api/v1/auth/logout:\n", 1)[0]
        self.assertIn("$ref: '#/components/schemas/Password'", register)
        self.assertIn("$ref: '#/components/schemas/Password'", login)
        self.assertIn("'500': { $ref: '#/components/responses/InternalServerError' }", register)
        self.assertIn("'500': { $ref: '#/components/responses/InternalServerError' }", login)
        self.assertIn("    InternalServerError:\n", api)

    def test_logo_source_is_readable_by_unprivileged_nginx_worker(self):
        logo = ROOT / "docs/assets/ux-securities-logo.png"
        self.assertNotEqual(stat.S_IMODE(logo.stat().st_mode) & 0o044, 0)


if __name__ == "__main__":
    unittest.main()
