# Phase 2 Error and Search Empty-State Rework Handoff

Status: `GREEN_ERROR_EMPTY; BROWSER_PENDING`

Date: 2026-10-24 (Asia/Hong_Kong)

## Scope

This bounded rework addresses only the latest Phase 2 review findings assigned
to this lane:

- provider availability failures on quote and daily-candle routes must be
  explainable HTTP 503 responses rather than 404 responses;
- successful zero-match market searches must show an explicit empty-result
  state distinct from a search transport error.

No browser evidence, task checkbox, orchestration-state, or Phase 3 changes were
made.

## RED

The focused source guard was added before the production implementation:

```text
python3 tests/phase2_error_empty_rework_red.py
Ran 2 tests ... FAILED
```

The initial failures showed that the market domain had no shared
`ErrUnavailable` identity or typed handler mapping, and `WorkbenchPage.tsx` had
no search outcome state or zero-result message. The complete capture is in
[`results/phase2-error-empty-red.md`](../../results/phase2-error-empty-red.md).

Focused HTTP tests were added before the implementation and initially failed
to compile because `market.ErrUnavailable` did not yet exist. Those tests cover
both quote and candle outage routes.

## GREEN changes

- Added the canonical `market.ErrUnavailable` sentinel in the market domain.
- Kept the existing integration package compatibility alias while making the
  domain the shared error identity.
- `getQuote` and `getDailyCandles` use `errors.Is(err, market.ErrUnavailable)`
  to return `503 market_unavailable`; non-availability errors retain the 404
  not-found mapping.
- Added `searchState` (`idle`, `loading`, `success`, `error`) and `searchError`
  to the workbench. A successful empty result renders `No securities match
  "...".` with `role="status"`; transport/input failures render a separate
  alert state.

## Verification

```text
python3 tests/phase2_error_empty_rework_red.py
Ran 2 tests ... OK

cd apps/backend && go test ./internal/httpapi ./internal/integration/marketdata ./internal/service/market -count=1
PASS

cd apps/backend && go test ./... -count=1
PASS

cd apps/backend && go vet ./...
PASS

cd apps/frontend && npm run build
PASS

npx openspec validate ux-securities-hkex-trading-terminal --strict
PASS

docker compose -f deploy/docker-compose.yml config -q
PASS
```

Browser/IAB evidence remains pending under the parent Phase 2 gate and is not
claimed by this handoff.
