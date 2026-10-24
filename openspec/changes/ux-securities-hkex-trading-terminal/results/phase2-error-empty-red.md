# Phase 2 Error and Search Empty-State RED

Date: 2026-10-24 (Asia/Hong_Kong)

## Scope

This bounded RED check covers the latest Phase 2 review findings addressed by
this rework:

- quote and daily-candle provider outages must map to a typed market-unavailable
  error and HTTP 503, while other lookup failures remain HTTP 404;
- a successful zero-match search must render an explicit empty-result state,
  distinct from a search transport error.

## Expected failure

Command:

```text
python3 tests/phase2_error_empty_rework_red.py
```

Result: **RED (expected)**. Both focused checks failed before the production
changes:

```text
Ran 2 tests ... FAILED
test_market_domain_exposes_shared_unavailable_error_and_handlers_map_it ... FAIL
  AssertionError: 'ErrUnavailable' not found in domain/market/quote.go
test_search_success_empty_state_is_distinct_from_transport_error ... FAIL
  AssertionError: searchState/searchStatus was not present in WorkbenchPage.tsx
```

The failures are caused by the missing shared typed error mapping and missing
frontend search outcome state, not by test collection or environment errors.
