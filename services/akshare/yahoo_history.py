"""Native Yahoo OHLCV fallback. Missing turnover stays unknown, never estimated."""
from __future__ import annotations

import math
from datetime import datetime, timedelta, timezone

from .normalization import HK, normalize_symbol, number, utc
from .public_history import request_json, valid_ohlc


def fetch_history(symbol: str, interval: str, start: str, end: str, now: datetime, fetch=request_json) -> list[dict]:
    if interval not in ("1m", "5m", "15m", "30m", "60m"):
        raise ValueError("unsupported native minute period")
    symbol = normalize_symbol(symbol)
    first = datetime.strptime(start, "%Y%m%d").replace(tzinfo=HK)
    until = datetime.strptime(end, "%Y%m%d").replace(tzinfo=HK) + timedelta(days=1)
    params = {"interval": interval, "includePrePost": "false"}
    if interval == "1m":
        params["range"] = "5d"
    else:
        # Yahoo caps 5/15/30-minute requests to the recent 60-day window.
        earliest = max(first, now - timedelta(days=59)) if interval != "60m" else first
        params.update(period1=int(earliest.timestamp()), period2=int(min(until, now).timestamp()))
    ticker = str(int(symbol)).zfill(4) + ".HK"
    data = fetch("https://query1.finance.yahoo.com/v8/finance/chart/" + ticker, params)
    results = (data.get("chart") or {}).get("result") or []
    if not results:
        raise ValueError("native minute history unavailable")
    result = results[0]
    quote = ((result.get("indicators") or {}).get("quote") or [{}])[0]
    bars = {}
    for index, timestamp in enumerate(result.get("timestamp") or []):
        try:
            stamp = datetime.fromtimestamp(timestamp, timezone.utc)
            if not first <= stamp < until:
                continue
            values = {key: quote[key][index] for key in ("open", "high", "low", "close", "volume")}
            if any(value is None or not math.isfinite(float(value)) for value in values.values()):
                continue
            if not valid_ohlc(values) or float(values["volume"]) < 0:
                continue
            bar = {key: number(value, 0 if key == "volume" else 4) for key, value in values.items()}
            bar.update(symbol=symbol, interval=interval, timestamp=utc(stamp), trade_date=stamp.astimezone(HK).date().isoformat(),
                       source="yahoo", data_status="unofficial", collected_at=utc(now), adjustment="none", turnover=None)
            bars[bar["timestamp"]] = bar
        except (ValueError, TypeError, KeyError, IndexError, OverflowError):
            continue
    if not bars:
        raise ValueError("no valid native minute bars")
    return [bars[key] for key in sorted(bars)]
