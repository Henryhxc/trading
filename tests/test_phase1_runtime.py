"""Runtime Phase 1 auth checks against an already-running local Compose stack."""

import http.cookiejar
import json
import os
import secrets
import subprocess
import unittest
import urllib.error
import urllib.request
from hashlib import sha256


BASE_URL = os.environ.get("UX_BASE_URL", "http://localhost:8182")
COMPOSE_PROJECT = os.environ.get(
    "UX_COMPOSE_PROJECT", "ux-securities-hkex-trading-terminal"
)
COMPOSE_FILE = os.environ.get("UX_COMPOSE_FILE", "deploy/docker-compose.yml")
PASSWORD = "Phase1-QA-only-Strong!2026"


def api_client():
    jar = http.cookiejar.CookieJar()
    opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(jar))
    return opener, jar


def request(opener, method, path, body=None):
    data = None if body is None else json.dumps(body).encode("utf-8")
    headers = {"Content-Type": "application/json"} if data is not None else {}
    req = urllib.request.Request(BASE_URL + path, data=data, headers=headers, method=method)
    try:
        response = opener.open(req, timeout=10)
        status, payload, response_headers = (
            response.status,
            response.read(),
            response.headers,
        )
    except urllib.error.HTTPError as error:
        status, payload, response_headers = error.code, error.read(), error.headers
    decoded = json.loads(payload) if payload else None
    return status, decoded, response_headers


def mysql(sql):
    container = subprocess.check_output(
        [
            "docker",
            "compose",
            "-p",
            COMPOSE_PROJECT,
            "-f",
            COMPOSE_FILE,
            "ps",
            "-q",
            "mysql",
        ],
        text=True,
    ).strip()
    if not container:
        raise RuntimeError("mysql service container was not found")
    command = (
        'MYSQL_PWD="$MYSQL_ROOT_PASSWORD" mysql -uroot '
        '"$MYSQL_DATABASE" --batch --skip-column-names'
    )
    return subprocess.check_output(
        ["docker", "exec", "-i", container, "sh", "-c", command],
        input=sql.encode("utf-8"),
        text=False,
    ).decode("utf-8").strip()


def register(opener, username, email):
    return request(
        opener,
        "POST",
        "/api/v1/auth/register",
        {"username": username, "email": email, "password": PASSWORD},
    )


def cookie_value(jar):
    cookie = next((item for item in jar if item.name == "ux_session"), None)
    if cookie is None:
        raise AssertionError("auth response did not set ux_session")
    return cookie


class Phase1RuntimeAuthTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.suffix = secrets.token_hex(6)
        cls.username_a = "qa" + cls.suffix
        cls.email_a = f"{cls.username_a}@example.test"
        cls.username_b = "qb" + cls.suffix
        cls.email_b = f"{cls.username_b}@example.test"

    def test_auth_lifecycle_and_session_guards(self):
        anonymous, _ = api_client()
        status, payload, _ = request(anonymous, "GET", "/api/v1/me")
        self.assertEqual((status, payload["code"]), (401, "unauthenticated"))

        client_a, jar_a = api_client()
        status, payload, headers = register(client_a, self.username_a, self.email_a)
        self.assertEqual(status, 201)
        user_a = payload["user"]
        self.assertEqual((user_a["username"], user_a["email"]), (self.username_a, self.email_a))
        self.assertIn("ux_session=", headers.get("Set-Cookie", ""))
        cookie_a = cookie_value(jar_a)
        self.assertTrue(cookie_a.has_nonstandard_attr("HttpOnly"))
        self.assertEqual(cookie_a.path, "/")
        self.assertEqual(cookie_a._rest.get("SameSite"), "Lax")
        self.assertFalse(cookie_a.secure)

        status, me, _ = request(client_a, "GET", "/api/v1/me")
        self.assertEqual(status, 200)
        self.assertEqual(me["user"], user_a)

        status, payload, _ = request(
            anonymous,
            "POST",
            "/api/v1/auth/login",
            {"identity": self.username_a, "password": "wrong-password!"},
        )
        self.assertEqual((status, payload["code"], payload["message"]),
                         (401, "invalid_credentials", "Username or password is incorrect"))
        status, payload, _ = request(
            anonymous,
            "POST",
            "/api/v1/auth/login",
            {"identity": f"missing-{self.suffix}", "password": PASSWORD},
        )
        self.assertEqual((status, payload["code"]), (401, "invalid_credentials"))

        status, payload, _ = register(api_client()[0], self.username_a, f"other-{self.suffix}@example.test")
        self.assertEqual((status, payload["code"]), (409, "identity_unavailable"))
        status, payload, _ = register(api_client()[0], f"other{self.suffix}", self.email_a)
        self.assertEqual((status, payload["code"]), (409, "identity_unavailable"))

        client_login, jar_login = api_client()
        status, logged_in, _ = request(
            client_login,
            "POST",
            "/api/v1/auth/login",
            {"identity": self.email_a, "password": PASSWORD},
        )
        self.assertEqual(status, 200)
        self.assertEqual(logged_in["user"], user_a)
        cookie_login = cookie_value(jar_login)
        self.assertNotEqual(cookie_login.value, cookie_a.value)
        status, payload, _ = request(client_login, "POST", "/api/v1/auth/logout", {})
        self.assertEqual(status, 204)
        status, payload, _ = request(client_login, "GET", "/api/v1/me")
        self.assertEqual((status, payload["code"]), (401, "unauthenticated"))

        client_b, jar_b = api_client()
        status, payload, _ = register(client_b, self.username_b, self.email_b)
        self.assertEqual(status, 201)
        user_b = payload["user"]
        self.assertNotEqual(user_a["id"], user_b["id"])
        cookie_b = cookie_value(jar_b)

        expired_hash = sha256(cookie_b.value.encode("utf-8")).hexdigest()
        update_result = mysql(
            "UPDATE sessions SET expires_at = UTC_TIMESTAMP(3) - INTERVAL 1 SECOND "
            f"WHERE user_id = {int(user_b['id'])} AND token_hash = '{expired_hash}';\n"
            "SELECT ROW_COUNT();"
        ).splitlines()[-1]
        self.assertEqual(update_result, "1")
        status, payload, _ = request(client_b, "GET", "/api/v1/me")
        self.assertEqual((status, payload["code"]), (401, "unauthenticated"))

        sql = (
            "SELECT COUNT(*), COUNT(DISTINCT user_id), "
            "SUM(user_id IN ("
            f"{int(user_a['id'])},{int(user_b['id'])}"
            ")), SUM(initial_cash = 1000000.0000), "
            "SUM(cash_balance = 1000000.0000), SUM(available_cash = 1000000.0000) "
            f"FROM accounts WHERE user_id IN ({int(user_a['id'])},{int(user_b['id'])});\n"
            "SELECT COUNT(*) FROM users WHERE id IN ("
            f"{int(user_a['id'])},{int(user_b['id'])});"
        )
        rows = mysql(sql).splitlines()
        self.assertEqual(rows[0], "2\t2\t2\t2\t2\t2")
        self.assertEqual(rows[1], "2")


if __name__ == "__main__":
    unittest.main(verbosity=2)
