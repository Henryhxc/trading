# Phase 2 Contract Rework Handoff

Date: 2026-10-24
Owner: phase2-contract-rework-new
Scope: OpenAPI response parity and daily-candle query validation only.

## RED evidence

- Added `tests/phase2_contract_rework_red.py` before implementation.
- Added focused handler tests in `apps/backend/internal/httpapi/market_handler_test.go`.
- Baseline failures are recorded in `results/phase2-contract-rework-red.md`.
- The baseline showed seven omitted OpenAPI statuses and invalid candle
  queries reaching the provider with HTTP 200.

## GREEN changes

- `apps/backend/internal/httpapi/market_handler.go`
  - Rejects malformed or repeated `from`/`to` date parameters with 400.
  - Rejects malformed, repeated, zero, negative, or over-365 `limit` values.
  - Enforces `from <= to` when both dates are supplied.
  - Preserves omitted-limit default of 90 and does not call the provider on
    invalid input.
- `apps/backend/internal/httpapi/market_handler_test.go`
  - Covers malformed dates, reversed ranges, limit boundaries/non-numeric
    values, provider-call suppression, and valid default-limit behavior.
- `docs/api/openapi.yaml`
  - Documents handler-emitted 400/404/500/503 responses for market and
    watchlist operations, using the shared error response schemas.

## Verification

All passed:

```text
cd apps/backend && go test ./... -count=1
cd apps/backend && go vet ./...
python3 tests/phase2_contract_rework_red.py
python3 tests/phase2_rework_contract_test.py
python3 tests/phase2_market_red_test.py
cd apps/frontend && npm run build
OpenAPI YAML parse via PyYAML
npx openspec validate ux-securities-hkex-trading-terminal --strict
```

No browser evidence was changed. No task checkboxes were changed. No blocker
remains within this bounded rework; Phase 2 still requires the separate IAB
workbench evidence and independent post-rework review before acceptance.
