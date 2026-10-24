# Phase 2 Rework Handoff

Status: `GREEN_CONTRACTS; BROWSER_PENDING`

## Scope

This bounded rework addresses the independent Phase 2 review findings only:

- reject empty or overlong market-search queries before invoking a provider;
- return `201 Created` only for a newly inserted favorite and `200 OK` for an existing idempotent relationship;
- expose favorite reorder controls and call `PUT /api/v1/watchlist/order`;
- persist and restore the workbench favorite-list scroll anchor alongside the active symbol.

No Phase 3 work, global orchestration edits, task checkbox edits, or browser evidence claims were made.

## RED

Added [`tests/phase2_rework_contract_test.py`](../../../../../tests/phase2_rework_contract_test.py) before production edits. The initial run failed all three focused checks (`Ran 3 tests ... FAILED`):

- empty search validation was absent;
- duplicate-add creation signaling and `200` handler branch were absent;
- frontend reorder and scroll-anchor contracts were absent.

The main-agent RED capture is recorded at [`results/phase2-rework-red.md`](../results/phase2-rework-red.md).

## GREEN changes

- `MarketRepository.AddWatchlist` now returns `(item, created, error)` based on `RowsAffected`; service and HTTP contracts carry the flag, and the handler selects 201/200.
- `searchSymbols` trims and validates the query (`1..80` characters) and returns the shared 400 `invalid_input` error without calling the provider for invalid input.
- `marketApi.reorderWatchlist` calls the existing authenticated order route. `WorkbenchPage` adds accessible up/down controls, optimistic reorder with reload on failure, and `ux_watchlist_scroll_anchor` persistence/restoration on the scrollable favorites list.
- Added behavioral HTTP tests for empty-query rejection and 201/200 add semantics in `internal/httpapi/market_handler_test.go`.

## Verification

Commands and results:

```text
cd apps/backend && go test ./... -count=1
PASS

cd apps/backend && go vet ./...
PASS

cd apps/frontend && npm run build
PASS (TypeScript and Vite production build)

python3 tests/phase2_rework_contract_test.py
Ran 3 tests ... OK

python3 tests/phase2_market_red_test.py
Ran 5 tests ... OK

openspec validate ux-securities-hkex-trading-terminal --strict
Change 'ux-securities-hkex-trading-terminal' is valid
```

`git diff --check` could not run because this workspace has no Git metadata; no whitespace diagnostics were observed from the formatter/build/test commands.

## Gate state

The code and contract checks are green for this rework. Phase 2 visual/IAB evidence remains pending user confirmation and must be captured separately at the required viewports before phase acceptance. Phase 3 remains blocked by the parent gate.
