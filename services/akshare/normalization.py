"""Vendor rows to the internal OHLC contract. Never manufacture missing bars."""
from __future__ import annotations

import math
import re
from datetime import date, datetime, timezone
from typing import Any, Iterable, Mapping
from zoneinfo import ZoneInfo

HK = ZoneInfo("Asia/Hong_Kong")
MINUTE_PERIODS = ("1m", "5m", "15m", "30m", "60m")


def value(row: Mapping[str, Any], *keys: str, default: Any = None) -> Any:
    for key in keys:
        if key in row:
            return row[key]
    return default


def number(raw: Any, places: int = 4) -> str:
    parsed = float(raw)
    if not math.isfinite(parsed):
        raise ValueError("non-finite numeric field")
    return f"{parsed:.{places}f}"


def normalize_symbol(raw: Any) -> str:
    digits = re.sub(r"\D", "", str(raw or ""))
    if not 1 <= len(digits) <= 5:
        raise ValueError("invalid Hong Kong symbol")
    return digits.zfill(5)


def utc(stamp: datetime) -> str:
    return stamp.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def normalize_spot_rows(rows: Iterable[Mapping[str, Any]], collected_at: datetime) -> list[dict[str, Any]]:
    result = []
    for row in rows:
        result.append({
            "symbol": normalize_symbol(value(row, "代码", "symbol")),
            "display_name": str(value(row, "名称", "display_name", default="")).strip(),
            "last_price": number(value(row, "最新价", "last_price")), "bid": None, "ask": None,
            "change": number(value(row, "涨跌额", "change")),
            "change_percent": number(value(row, "涨跌幅", "change_percent")),
            "volume": number(value(row, "成交量", "volume"), 0),
            "turnover": number(value(row, "成交额", "turnover")),
            "collected_at": utc(collected_at), "source": "akshare", "data_status": "delayed",
        })
    return result


def normalize_history_rows(rows: Iterable[Mapping[str, Any]], symbol: str, collected_at: datetime) -> list[dict[str, Any]]:
    result = {}
    for row in rows:
        raw = value(row, "日期", "trade_date")
        day = raw.date() if isinstance(raw, datetime) else raw
        day = day.isoformat() if isinstance(day, date) else str(day)[:10]
        date.fromisoformat(day)
        result[day] = {
            "symbol": normalize_symbol(symbol), "trade_date": day,
            "open": number(value(row, "开盘", "open")), "high": number(value(row, "最高", "high")),
            "low": number(value(row, "最低", "low")), "close": number(value(row, "收盘", "close")),
            "volume": number(value(row, "成交量", "volume"), 0),
            "turnover": number(value(row, "成交额", "turnover")),
            "source": "akshare", "data_status": "unofficial", "collected_at": utc(collected_at),
            "adjustment": "none",
        }
    return [result[key] for key in sorted(result)]


def normalize_minute_rows(rows: Iterable[Mapping[str, Any]], symbol: str, interval: str, collected_at: datetime) -> list[dict[str, Any]]:
    if interval not in MINUTE_PERIODS:
        raise ValueError("unsupported native minute interval")
    result = {}
    for row in rows:
        raw = value(row, "时间", "timestamp")
        stamp = raw if isinstance(raw, datetime) else datetime.fromisoformat(str(raw))
        if stamp.tzinfo is None:
            stamp = stamp.replace(tzinfo=HK)
        day = stamp.astimezone(HK).date().isoformat()
        mapped = dict(row, trade_date=day)
        bar = normalize_history_rows([mapped], symbol, collected_at)[0]
        bar.update(interval=interval, timestamp=utc(stamp))
        result[bar["timestamp"]] = bar
    return [result[key] for key in sorted(result)]
