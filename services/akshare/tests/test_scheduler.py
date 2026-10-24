import hashlib
import hmac
import json
import logging
import unittest
from datetime import date, datetime, timezone
from unittest.mock import Mock

from services.akshare.scheduler import (
    MIN_SPOT_INTERVAL_SECONDS,
    Backoff,
    Config,
    Scheduler,
    build_payload,
    normalize_spot_rows,
    sign_webhook,
)


class SchedulerContractTests(unittest.TestCase):
    def test_normalizes_spot_rows_to_backend_contract(self):
        rows = normalize_spot_rows([
            {
                "代码": "700",
                "名称": "腾讯控股",
                "最新价": 380.2,
                "涨跌额": 2.1,
                "涨跌幅": 0.5543,
                "成交量": 1200000,
                "成交额": 456240000,
            }
        ], collected_at=datetime(2026, 10, 24, 8, 0, tzinfo=timezone.utc))

        self.assertEqual(rows[0]["symbol"], "00700")
        self.assertEqual(rows[0]["display_name"], "腾讯控股")
        self.assertEqual(rows[0]["source"], "akshare")
        self.assertEqual(rows[0]["data_status"], "delayed")
        self.assertEqual(rows[0]["last_price"], "380.2000")

    def test_build_payload_contains_quotes_and_candles(self):
        payload = build_payload(
            quotes=[{"symbol": "00700"}],
            candles=[{"symbol": "00700", "trade_date": "2026-10-23"}],
        )
        self.assertEqual(payload["version"], 1)
        self.assertEqual(len(payload["quotes"]), 1)
        self.assertEqual(len(payload["candles"]), 1)

    def test_normalizes_history_rows_without_adjustment(self):
        from services.akshare.scheduler import normalize_history_rows

        rows = normalize_history_rows([
            {"日期": date(2026, 10, 23), "开盘": 376, "最高": 382.2, "最低": 375, "收盘": 380.2, "成交量": 1000, "成交额": 379000}
        ], "700", datetime(2026, 10, 24, 8, 0, tzinfo=timezone.utc))
        self.assertEqual(rows[0]["symbol"], "00700")
        self.assertEqual(rows[0]["adjustment"], "none")
        self.assertEqual(rows[0]["data_status"], "unofficial")

    def test_spot_interval_cannot_be_shorter_than_fifteen_minutes(self):
        with self.assertRaises(ValueError):
            Config.from_env({"AKSHARE_POLL_INTERVAL": "899"})
        self.assertEqual(
            Config.from_env({"AKSHARE_POLL_INTERVAL": "900"}).poll_interval_seconds,
            MIN_SPOT_INTERVAL_SECONDS,
        )

    def test_hmac_matches_go_webhook_vector(self):
        body = b'{"version":1,"quotes":[],"candles":[]}'
        signature = sign_webhook("synthetic-test-secret", "1700000000", "delivery-123", body)
        expected = hmac.new(
            b"synthetic-test-secret",
            b"1700000000.delivery-123." + body,
            hashlib.sha256,
        ).hexdigest()
        self.assertEqual(signature, expected)

    def test_backoff_is_exponential_with_bounded_jitter(self):
        backoff = Backoff(base_seconds=2, max_seconds=30, jitter=lambda: 0.25)
        self.assertEqual(backoff.delay(0), 2.25)
        self.assertEqual(backoff.delay(1), 4.25)
        self.assertEqual(backoff.delay(10), 30.25)

    def test_logs_redact_secret_and_webhook_url(self):
        logger = logging.getLogger("akshare-test")
        logger.handlers.clear()
        handler = logging.Handler()
        handler.emit = Mock()
        logger.addHandler(handler)
        scheduler = Scheduler(
            Config(webhook_url="http://user:password@example.test/hook", webhook_secret="secret-value"),
            fetch_spot=lambda: [],
            send_payload=lambda payload: None,
            logger=logger,
        )
        scheduler.log_error("request failed", token="secret-value", url=scheduler.config.webhook_url)
        message = handler.emit.call_args.args[0].getMessage()
        self.assertNotIn("secret-value", message)
        self.assertNotIn("password", message)
        self.assertIn("[REDACTED]", message)

    def test_health_check_is_healthy_only_after_successful_delivery(self):
        scheduler = Scheduler(
            Config(webhook_url="http://backend:8080/internal", webhook_secret="secret"),
            fetch_spot=lambda: [],
            send_payload=lambda payload: None,
        )
        self.assertFalse(scheduler.health().healthy)
        scheduler.run_once(now=datetime(2026, 10, 24, 8, 0, tzinfo=timezone.utc))
        self.assertTrue(scheduler.health().healthy)


if __name__ == "__main__":
    unittest.main()
