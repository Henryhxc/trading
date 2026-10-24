"""AkShare polling and signed delivery boundary for the Go market-data API."""

from __future__ import annotations

import hashlib
import calendar
import hmac
import logging
import os
import random
import re
import time
import uuid
from dataclasses import dataclass
from datetime import date, datetime, timezone
from typing import Any, Callable, Iterable, Mapping
from urllib.parse import urlsplit, urlunsplit
from zoneinfo import ZoneInfo

from .normalization import normalize_history_rows, normalize_minute_rows, normalize_spot_rows, normalize_symbol

MIN_SPOT_INTERVAL_SECONDS = 15 * 60
_SECRET_RE = re.compile(r"(?i)(secret|token|password|authorization)=?[^\s,]+")


def build_payload(quotes: list[dict[str, Any]], candles: list[dict[str, Any]]) -> dict[str, Any]:
    return {"version": 1, "quotes": quotes, "candles": candles}


def sign_webhook(secret: str, timestamp: str, delivery_id: str, body: bytes) -> str:
    message = f"{timestamp}.{delivery_id}.".encode() + body
    return hmac.new(secret.encode(), message, hashlib.sha256).hexdigest()


@dataclass(frozen=True)
class Config:
    webhook_url: str = "http://backend:8080/api/v1/internal/market-data/akshare"
    webhook_secret: str = ""
    poll_interval_seconds: int = MIN_SPOT_INTERVAL_SECONDS
    request_timeout_seconds: float = 15.0
    symbols: tuple[str, ...] = ("00700", "00005", "09988")
    history_start: str = ""
    history_end: str = ""
    minute_periods: tuple[str, ...] = ("1m", "5m", "15m", "30m", "60m")
    history_provider: str = "akshare"

    def history_bounds(self, now: datetime) -> tuple[str, str]:
        end = datetime.strptime(self.history_end, "%Y%m%d").date() if self.history_end else now.astimezone(ZoneInfo("Asia/Hong_Kong")).date()
        month_index = end.year * 12 + end.month - 1 - 6
        year, month = divmod(month_index, 12)
        start = end.replace(year=year, month=month + 1, day=min(end.day, calendar.monthrange(year, month + 1)[1]))
        return self.history_start or start.strftime("%Y%m%d"), end.strftime("%Y%m%d")

    @classmethod
    def from_env(cls, values: Mapping[str, str] | None = None) -> "Config":
        values = os.environ if values is None else values
        interval = int(values.get("AKSHARE_POLL_INTERVAL", str(MIN_SPOT_INTERVAL_SECONDS)))
        if interval < MIN_SPOT_INTERVAL_SECONDS:
            raise ValueError("AKSHARE_POLL_INTERVAL cannot be shorter than 15 minutes")
        symbols = tuple(normalize_symbol(item) for item in values.get("AKSHARE_SYMBOLS", "00700,00005,09988").split(",") if item.strip())
        periods = tuple(item.strip() for item in values.get("AKSHARE_MINUTE_PERIODS", "1m,5m,15m,30m,60m").split(",") if item.strip())
        if any(item not in ("1m", "5m", "15m", "30m", "60m") for item in periods):
            raise ValueError("unsupported minute period")
        history_provider = values.get("HISTORY_PROVIDER", "akshare")
        if history_provider not in ("akshare", "public"):
            raise ValueError("unsupported history provider")
        return cls(
            webhook_url=values.get("AKSHARE_WEBHOOK_URL", cls.webhook_url),
            webhook_secret=values.get("AKSHARE_WEBHOOK_SECRET", ""),
            poll_interval_seconds=interval,
            request_timeout_seconds=float(values.get("AKSHARE_REQUEST_TIMEOUT", "15")),
            symbols=symbols,
            history_start=values.get("AKSHARE_HISTORY_START", ""),
            history_end=values.get("AKSHARE_HISTORY_END", ""),
            minute_periods=periods,
            history_provider=history_provider,
        )


@dataclass(frozen=True)
class Health:
    healthy: bool
    last_success: datetime | None = None
    last_error: str | None = None


class Backoff:
    def __init__(self, base_seconds: float = 1.0, max_seconds: float = 60.0, jitter: Callable[[], float] | None = None):
        self.base_seconds = base_seconds
        self.max_seconds = max_seconds
        self.jitter = jitter or random.random

    def delay(self, attempt: int) -> float:
        return min(self.max_seconds, self.base_seconds * (2 ** attempt)) + self.jitter()


class Scheduler:
    def __init__(self, config: Config, fetch_spot: Callable[[], Iterable[Mapping[str, Any]]], send_payload: Callable[[dict[str, Any]], None], fetch_history: Callable[[], Iterable[Mapping[str, Any]]] | None = None, logger: logging.Logger | None = None):
        self.config = config
        self.fetch_spot = fetch_spot
        self.send_payload = send_payload
        self.fetch_history = fetch_history
        self.logger = logger or logging.getLogger("akshare.scheduler")
        self._health = Health(False)
        self.backoff = Backoff()

    def _redact(self, value: Any) -> str:
        text = str(value)
        parts = urlsplit(text)
        if parts.password:
            text = urlunsplit((parts.scheme, parts.hostname or "", parts.path, parts.query, parts.fragment))
        text = _SECRET_RE.sub(lambda match: match.group(1) + "=[REDACTED]", text)
        return text.replace(self.config.webhook_secret, "[REDACTED]") if self.config.webhook_secret else text

    def log_error(self, message: str, **fields: Any) -> None:
        safe_fields = " ".join(f"{key}={self._redact(value)}" for key, value in fields.items())
        self.logger.error("%s %s", self._redact(message), safe_fields)

    def run_once(self, now: datetime | None = None) -> None:
        now = now or datetime.now(timezone.utc)
        quotes, candles, errors = [], [], []
        for label, fetch in (("spot", lambda: normalize_spot_rows(self.fetch_spot(), now)), ("history", self.fetch_history)):
            if fetch is None:
                continue
            try:
                rows = list(fetch())
                if label == "spot":
                    quotes = rows
                else:
                    candles = rows
            except Exception as exc:
                errors.append(exc)
                self.log_error("upstream failed", kind=label, error=type(exc).__name__)
        try:
            if not quotes and not candles and errors:
                raise errors[-1]
            # Bound webhook size and database transaction duration; replay uses upserts.
            for offset in range(0, max(len(quotes), len(candles), 1), 400):
                payload = build_payload(quotes[offset:offset + 400], candles[offset:offset + 400])
                if payload["quotes"] or payload["candles"]:
                    self._deliver(payload)
            self._health = Health(True, now, None)
        except Exception as exc:
            self._health = Health(False, self._health.last_success, str(exc))
            self.log_error("poll delivery failed", error=exc)
            raise

    def _deliver(self, payload: dict[str, Any]) -> None:
        last_error: Exception | None = None
        for attempt in range(3):
            try:
                self.send_payload(payload)
                return
            except Exception as exc:
                last_error = exc
                if attempt < 2:
                    time.sleep(self.backoff.delay(attempt))
        raise last_error or RuntimeError("delivery failed")

    def health(self) -> Health:
        return self._health


def main() -> None:
    import akshare as ak
    import requests
    from .collector import collect_history
    from .public_history import fetch_history as fetch_public_history

    config = Config.from_env()
    if not config.webhook_secret:
        raise RuntimeError("AKSHARE_WEBHOOK_SECRET is required")
    session = requests.Session()

    def fetch_spot() -> Iterable[Mapping[str, Any]]:
        frame = ak.stock_hk_spot_em()
        return frame.to_dict("records")

    def fetch_history() -> Iterable[Mapping[str, Any]]:
        return collect_history(ak, config, datetime.now(timezone.utc), fallback=fetch_public_history)

    def send_payload(payload: dict[str, Any]) -> None:
        body = __import__("json").dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()
        timestamp = str(int(time.time()))
        delivery_id = str(uuid.uuid4())
        response = session.post(config.webhook_url, data=body, timeout=config.request_timeout_seconds, headers={
            "Content-Type": "application/json",
            "X-UX-Timestamp": timestamp,
            "X-UX-Delivery-ID": delivery_id,
            "X-UX-Signature": sign_webhook(config.webhook_secret, timestamp, delivery_id, body),
        })
        response.raise_for_status()

    scheduler = Scheduler(config, fetch_spot, send_payload, fetch_history=fetch_history)
    while True:
        try:
            scheduler.run_once()
        except Exception:
            pass
        time.sleep(config.poll_interval_seconds)


if __name__ == "__main__":
    main()
