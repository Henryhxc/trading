"""Independent, bounded AkShare history collection for native candle periods."""
from __future__ import annotations

import logging
from datetime import datetime
from typing import Any

from .normalization import normalize_history_rows, normalize_minute_rows


def collect_history(api: Any, config: Any, now: datetime, fallback=None) -> list[dict]:
    start, end = config.history_bounds(now)
    start_time = datetime.strptime(start, "%Y%m%d").strftime("%Y-%m-%d 00:00:00")
    end_time = datetime.strptime(end, "%Y%m%d").strftime("%Y-%m-%d 23:59:59")
    collected = []
    attempts = 0
    for symbol in config.symbols:
        for interval in ("1d", *config.minute_periods):
            attempts += 1
            try:
                if config.history_provider == "public":
                    if fallback is None:
                        raise ValueError("public history adapter required")
                    collected.extend(fallback(symbol, interval, start, end, now))
                    continue
                if interval == "1d":
                    frame = api.stock_hk_hist(symbol=symbol, period="daily", start_date=start, end_date=end, adjust="")
                    rows = normalize_history_rows(frame.to_dict("records"), symbol, now)
                else:
                    frame = api.stock_hk_hist_min_em(symbol=symbol, period=interval[:-1], adjust="", start_date=start_time, end_date=end_time)
                    rows = normalize_minute_rows(frame.to_dict("records"), symbol, interval, now)
                collected.extend(rows)
                logging.info("history fetched symbol=%s interval=%s count=%d", symbol, interval, len(rows))
            except Exception as exc:
                # Do not log exception payloads, vendor URLs, or credentials.
                logging.warning("history unavailable symbol=%s interval=%s error=%s", symbol, interval, type(exc).__name__)
                if fallback is not None and config.history_provider != "public":
                    try:
                        collected.extend(fallback(symbol, interval, start, end, now))
                    except Exception as public_error:
                        logging.warning("public history unavailable symbol=%s interval=%s error=%s", symbol, interval, type(public_error).__name__)
    if attempts and not collected:
        raise RuntimeError("all configured native history sources unavailable")
    return collected
