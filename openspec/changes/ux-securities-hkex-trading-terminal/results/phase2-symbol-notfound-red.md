# Phase 2 Unknown-Symbol Rework RED

Date: 2026-10-24

The focused guards were added before implementation:

```text
$ python3 tests/phase2_symbol_notfound_red.py
FFF
FAIL: test_typed_not_found_and_repository_translation_exist
  AssertionError: 'ErrNotFound' not found in apps/backend/internal/domain/market/quote.go
FAIL: test_provider_and_mock_preserve_not_found
  AssertionError: provider composition did not inspect market.ErrNotFound
FAIL: test_candles_document_unavailable_response
  AssertionError: candles OpenAPI path did not declare the shared 503 Unavailable response
```

The focused Go tests also failed to compile as expected because the new sentinel
was not yet defined:

```text
$ cd apps/backend && go test ./internal/integration/marketdata ./internal/httpapi
FAIL
undefined: market.ErrNotFound
```

This is the valid RED baseline for the bounded rework. No task checkbox or
browser evidence was changed.
