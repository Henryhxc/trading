test_market_and_watchlist_tests_exist_before_green (__main__.Phase2MarketContractRed) ... FAIL
test_market_routes_cover_search_quote_and_candles (__main__.Phase2MarketContractRed) ... FAIL
test_provider_fallback_is_explicit_and_never_live_for_free_sources (__main__.Phase2MarketContractRed) ... FAIL
test_provider_interface_and_status_allowlist_exist (__main__.Phase2MarketContractRed) ... FAIL
test_watchlist_is_idempotent_and_actor_scoped (__main__.Phase2MarketContractRed) ... FAIL

======================================================================
FAIL: test_market_and_watchlist_tests_exist_before_green (__main__.Phase2MarketContractRed)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 85, in test_market_and_watchlist_tests_exist_before_green
    self.assertTrue((BACKEND / relative).exists(), f"missing RED/GREEN test: {relative}")
AssertionError: False is not true : missing RED/GREEN test: service/market/service_test.go

======================================================================
FAIL: test_market_routes_cover_search_quote_and_candles (__main__.Phase2MarketContractRed)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 55, in test_market_routes_cover_search_quote_and_candles
    self.assert_source_contains(
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 19, in assert_source_contains
    self.assertTrue(path.exists(), f"missing Phase 2 source file: {path}")
AssertionError: False is not true : missing Phase 2 source file: /Users/hongxincong/project/UX Security/apps/backend/internal/httpapi/market_handler.go

======================================================================
FAIL: test_provider_fallback_is_explicit_and_never_live_for_free_sources (__main__.Phase2MarketContractRed)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 37, in test_provider_fallback_is_explicit_and_never_live_for_free_sources
    self.assert_source_contains(
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 19, in assert_source_contains
    self.assertTrue(path.exists(), f"missing Phase 2 source file: {path}")
AssertionError: False is not true : missing Phase 2 source file: /Users/hongxincong/project/UX Security/apps/backend/internal/integration/marketdata/provider.go

======================================================================
FAIL: test_provider_interface_and_status_allowlist_exist (__main__.Phase2MarketContractRed)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 25, in test_provider_interface_and_status_allowlist_exist
    self.assert_source_contains(
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 19, in assert_source_contains
    self.assertTrue(path.exists(), f"missing Phase 2 source file: {path}")
AssertionError: False is not true : missing Phase 2 source file: /Users/hongxincong/project/UX Security/apps/backend/internal/domain/market/quote.go

======================================================================
FAIL: test_watchlist_is_idempotent_and_actor_scoped (__main__.Phase2MarketContractRed)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 64, in test_watchlist_is_idempotent_and_actor_scoped
    self.assert_source_contains(
  File "/Users/hongxincong/project/UX Security/tests/phase2_market_red_test.py", line 19, in assert_source_contains
    self.assertTrue(path.exists(), f"missing Phase 2 source file: {path}")
AssertionError: False is not true : missing Phase 2 source file: /Users/hongxincong/project/UX Security/apps/backend/internal/repository/mysqlrepo/watchlist.go

----------------------------------------------------------------------
Ran 5 tests in 0.001s

FAILED (failures=5)

exit_code=1
