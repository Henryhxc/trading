# Workbench Layout Independent Review

Review date: 2026-10-24 (Asia/Hong_Kong)

## Decision

Implementation source review: `PASS_WITH_NOTES` after the no-quote fallback correction. Visual review: `BLOCKED`; the required in-app-browser three-viewport evidence is not available in this review session. Do not treat this as acceptance of the broader Phase 2/3 or final project gates.

## No-quote fallback finding and recheck

The earlier provider flow treated any successful `unavailable` quote as a reason to continue to the mock provider. MySQL directory rows with `last_price = NULL` are represented as `DataStatusUnavailable`, with no price and source `none`; for the three demo symbols this could therefore have exposed a fabricated mock price and mock candles instead of the directory's unavailable state.

The current source fixes the quote path: `apps/backend/internal/integration/marketdata/provider.go:52-62` only continues fallback for an unavailable quote that still has a price. An unavailable quote with `LastPrice == nil` is returned unchanged. `TestProviderKeepsKnownSymbolWithoutCachedQuoteUnavailable` at `apps/backend/internal/integration/marketdata/provider_test.go:134-151` asserts no price, `unavailable`, and source `none`. The test and full marketdata package passed during this review:

```text
go test ./internal/integration/marketdata -run '^TestProviderKeepsKnownSymbolWithoutCachedQuoteUnavailable$' -count=1
PASS
go test ./internal/integration/marketdata -count=1
PASS
```

The directory import evidence also shows a no-price imported row (`00001 … unavailable NULL`) in `results/workbench-layout/directory-import-isolated.txt`; the larger MySQL directory test passes in `directory-green.txt`. The historical failing stdout for this exact provider regression was not present among the saved artifacts, so this report does not claim a preserved RED run. The unrelated `mock-directory-red.txt` concerns mock directory enumeration, not this regression. Empty candles from an existing MySQL directory row return successfully as an empty series, so the provider does not reach mock historical candles on that path.

## Source review

- `WorkbenchPage` loads the full directory using an empty query, defaults to Market, refreshes directory/watchlist every 60 seconds, persists the active symbol, restores the favorites scroll anchor, supports reorder, and refreshes account/positions/watchlist/directory after an order.
- `MarketSidebar` filters the loaded directory, displays distinct Favorite/Position markers, renders missing prices as an em dash, and discloses the imported directory's future-declared source date. `QuotePanel` shows provider/status and empty quote values.
- The workbench source and styles implement the left directory / center quote-and-chart / right paper-order arrangement, the logo on a light plaque, dark blue-grey surfaces, blue interaction accents, green/red movement values with percentages, and mobile list → chart → order stacking. Full account, positions, and order history are rendered in the My positions view; the ticket retains only compact available cash.
- The chart UI has period controls, half-year bounds, refresh/error/empty states, source/status labeling, and OHLC/volume rendering. These are source-level observations only, not browser-confirmed layout or interaction results.

## Checks and evidence

- Saved `results/workbench-layout/frontend-regression.txt`: frontend source guard plus 4 test files / 13 tests passed.
- Saved `frontend-build.txt`: TypeScript and Vite build passed.
- Saved `go-test.txt` and `go-vet.txt`: full Go suite and vet passed before this review's focused provider rerun.
- Saved `directory-parser-green.txt`, `directory-green.txt`, and `directory-import-isolated.txt`: directory parser and isolated MySQL directory checks passed; import metadata is limited to HKD equities and is marked metadata-only. `directory-source.json` explicitly records that the source declares 2026-10-26, later than the local import date 2026-10-24.
- Docker frontend/backend image build logs exist; this review did not claim a fresh full-stack cold-start or runtime verification.

## Browser gate

The required Phase 2 workbench screenshots directory contains only `.gitkeep`; the Phase 3 directory has login/register screenshots but no workbench/trading-flow screenshots. In this review session, the selected in-app browser was unavailable (`Browser is not available: iab`; browser discovery returned no available browsers), so I did not navigate to localhost or create screenshots. The caller-provided image shows a blank new-tab surface and is not evidence of the workbench. Three-viewport visual checks, interaction states, and screenshot comparison remain pending; no visual pass is claimed.

## Remaining gate state

Keep browser-dependent tasks and the aggregate gate open until the project captures the required in-app-browser evidence. Real AkShare upstream success remains unproven by the recorded `RemoteDisconnected` result, and existing unchecked Phase 2/3 tasks are not changed by this review. No production source, task checkbox, or orchestration state was modified.
