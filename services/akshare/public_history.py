"""Explicit public-source fallback; no adjusted/synthetic replacement bars."""
from __future__ import annotations

import json
import logging
import time
from datetime import datetime
from urllib.parse import urlencode
from urllib.request import Request, urlopen

from .normalization import normalize_history_rows, normalize_minute_rows, normalize_symbol, number, utc


def request_json(url: str, params: dict) -> dict:
    for attempt in range(3):
        try:
            request = Request(url + "?" + urlencode(params), headers={"User-Agent": "Mozilla/5.0"})
            with urlopen(request, timeout=15) as response:
                return json.load(response)
        except Exception:
            if attempt == 2:
                raise
            time.sleep(1 + attempt)
    raise RuntimeError("public history unavailable")


def valid_ohlc(row: dict) -> bool:
    try:
        return 0 < float(row["low"]) <= min(float(row["open"]), float(row["close"])) <= max(float(row["open"]), float(row["close"])) <= float(row["high"])
    except (ValueError, KeyError, TypeError):
        return False


def fetch_history(symbol: str, interval: str, start: str, end: str, now: datetime, fetch=request_json) -> list[dict]:
    if interval == "1d":
        return fetch_primary_history(symbol, interval, start, end, now, fetch)
    try:
        from .yahoo_history import fetch_history as fetch_yahoo
        return fetch_yahoo(symbol, interval, start, end, now, fetch)
    except Exception:
        return fetch_primary_history(symbol, interval, start, end, now, fetch)


def fetch_primary_history(symbol: str, interval: str, start: str, end: str, now: datetime, fetch=request_json) -> list[dict]:
    symbol = normalize_symbol(symbol)
    first = datetime.strptime(start, "%Y%m%d").date().isoformat()
    last = datetime.strptime(end, "%Y%m%d").date().isoformat()
    if interval == "1d":
        data = fetch("https://web.ifzq.gtimg.cn/appstock/app/fqkline/get", {"param": f"hk{symbol},day,{first},{last},366,"})
        # Only the unadjusted `day` key is accepted. Tencent supplies no amount.
        rows = (data.get("data") or {}).get("hk" + symbol, {}).get("day", [])
        bars = []
        for row in rows:
            if len(row) < 6 or not first <= row[0] <= last:
                continue
            bar = {"symbol": symbol, "trade_date": row[0], "interval": "1d", "open": number(row[1]), "close": number(row[2]),
                   "high": number(row[3]), "low": number(row[4]), "volume": number(row[5], 0), "turnover": None,
                   "source": "tencent", "data_status": "unofficial", "collected_at": utc(now), "adjustment": "none"}
            datetime.strptime(row[0], "%Y-%m-%d")
            if valid_ohlc(bar):
                bars.append(bar)
    else:
        if interval not in ("1m", "5m", "15m", "30m", "60m"):
            raise ValueError("unsupported native minute period")
        params = {"secid": "116." + symbol, "fields1": "f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13"}
        if interval == "1m":
            params.update(ndays=5, fields2="f51,f52,f53,f54,f55,f56,f57,f58")
            data = fetch("https://push2his.eastmoney.com/api/qt/stock/trends2/get", params)
            rows = (data.get("data") or {}).get("trends", [])
        else:
            params.update(klt=interval[:-1], fqt=0, beg=start, end=end, lmt=5000, fields2="f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61")
            data = fetch("https://push2his.eastmoney.com/api/qt/stock/kline/get", params)
            rows = (data.get("data") or {}).get("klines", [])
        native = []
        for raw in rows:
            row = raw.split(",")
            if len(row) < 7 or not first <= row[0][:10] <= last:
                continue
            mapped = dict(timestamp=row[0], open=row[1], close=row[2], high=row[3], low=row[4], volume=row[5], turnover=row[6])
            if valid_ohlc(mapped):
                native.append(mapped)
        bars = normalize_minute_rows(native, symbol, interval, now)
        for bar in bars:
            bar["source"] = "eastmoney"
    if not bars:
        raise ValueError("no valid unadjusted native bars from public source")
    logging.info("public history fetched symbol=%s interval=%s source=%s count=%d", symbol, interval, bars[0]["source"], len(bars))
    return bars
