# Phase 2 Unknown-Symbol Rework Handoff

Date: 2026-10-24
Status: GREEN_SYMBOL_NOTFOUND; BROWSER_PENDING
Scope: unknown-symbol semantics and candle 503 contract only

## RED

Added `tests/phase2_symbol_notfound_red.py` and focused Go tests before the
implementation. The expected failures are recorded in
`results/phase2-symbol-notfound-red.md`: the market sentinel and provider
guards were absent, the candle OpenAPI path lacked 503, and the Go tests could
not compile against `market.ErrNotFound`.

## GREEN changes

- Added canonical `market.ErrNotFound`, distinct from `market.ErrUnavailable`.
- `MarketRepository.Quote` translates `sql.ErrNoRows` to `market.ErrNotFound`.
- `MarketRepository.Candles` checks stock existence first. Missing stocks return
  `market.ErrNotFound`; existing stocks with no daily rows return a successful,
  non-nil empty slice.
- `marketdata.Provider` continues fallback after a typed not-found result and
  returns `ErrNotFound` only when all candidates report an unknown symbol.
  Generic provider errors and unavailable statuses remain `ErrUnavailable`.
- `MockProvider` returns `market.ErrNotFound` for unknown quotes and candles.
- Quote and candle handlers explicitly map `ErrNotFound` to HTTP 404 and
  `ErrUnavailable` to HTTP 503.
- Added the candles `503 Unavailable` response to `docs/api/openapi.yaml`.

## Focused evidence

```text
$ python3 tests/phase2_symbol_notfound_red.py
Ran 3 tests in 0.001s
OK

$ cd apps/backend && go test ./internal/integration/marketdata ./internal/httpapi ./internal/repository/mysqlrepo
ok
```

The provider tests cover repository-shaped not-found fallback to a known mock,
all-candidate unknown quote/candle results, known empty candles, mock unknowns,
and provider outage status. HTTP tests cover 404 mapping for unknown quote and
candle symbols and 503 mapping for outages.

## Broader verification

- `cd apps/backend && go test ./... -count=1` PASS
- `cd apps/backend && go vet ./...` PASS
- `python3 tests/phase2_symbol_notfound_red.py` PASS
- `python3 tests/phase2_contract_rework_red.py` PASS
- `python3 tests/phase2_rework_contract_test.py` PASS
- OpenAPI PyYAML parse PASS
- `npm run build` in `apps/frontend` PASS
- `npx openspec validate ux-securities-hkex-trading-terminal --strict` PASS
- Git status/diff checks remain unavailable because this workspace has no Git
  metadata; this is an environment limitation, not a code failure.

## Remaining gate

No browser evidence or task checkbox was changed. Phase 2 still requires the
user-selected in-app browser (IAB) workbench screenshots and interaction/state
evidence at the required viewports before independent review can pass.
