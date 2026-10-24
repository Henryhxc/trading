# Phase 2 Contract Rework RED

Date: 2026-10-24

Scope: OpenAPI response-status parity for the market/watchlist handlers and
strict daily-candle query validation.

## Expected failures

```text
python3 tests/phase2_contract_rework_red.py
FAILED (failures=7)

Omitted OpenAPI statuses:
- GET /api/v1/markets/symbols: 400, 503
- GET /api/v1/markets/candles/{symbol}: 400, 404
- GET /api/v1/watchlist: 500
- POST /api/v1/watchlist: 400
- PUT /api/v1/watchlist/order: 400, 500
- DELETE /api/v1/watchlist/{symbol}: 500
```

```text
cd apps/backend && go test ./internal/httpapi -run 'TestGetDailyCandles' -count=1
FAIL

TestGetDailyCandlesRejectsInvalidQueryBeforeProviderCall:
- malformed from date returned 200 (want 400)
- malformed/invalid to/range and limit 0, 366, non-numeric returned 200 (want 400)
```

The valid date-range/default-limit test was present in the same run and was
the control case; implementation changes must preserve its expected 200 and
provider limit 90 behavior.
