# Phase 2 Provider Config Rework — RED Evidence

Date: 2026-10-24 (Asia/Hong_Kong)

## Intended behavior

- With `MARKET_DATA_MODE` unset, runtime selection defaults explicitly to the mock provider and quotes remain `source=mock`, `data_status=simulated`.
- The optional `repository` mode is assembled only in the non-authorized slot. A repository/cache row that claims `live` is therefore downgraded to `stale`.
- `repository` without a repository dependency, `authorized` without an implemented licensed adapter, `public` without an implemented public adapter, and unknown modes fail closed.

The tests exercise configuration parsing and the actual runtime provider composition. They do not grep production source text.

## RED command

Working directory: `apps/backend`

```text
gofmt -w internal/config/config_test.go cmd/api/main_test.go && go test ./internal/config ./cmd/api -count=1
```

Exit code: `1` (expected RED)

## Failure output

```text
# ux-securities/backend/internal/config [ux-securities/backend/internal/config.test]
internal/config/config_test.go:8:14: assignment mismatch: 2 variables but Load returns 1 value
internal/config/config_test.go:12:27: undefined: MarketDataModeMock
internal/config/config_test.go:13:64: undefined: MarketDataModeMock
internal/config/config_test.go:20:14: assignment mismatch: 2 variables but Load returns 1 value
internal/config/config_test.go:24:27: undefined: MarketDataModeRepository
internal/config/config_test.go:25:64: undefined: MarketDataModeRepository
internal/config/config_test.go:34:14: assignment mismatch: 2 variables but Load returns 1 value
FAIL ux-securities/backend/internal/config [build failed]
# ux-securities/backend/cmd/api [ux-securities/backend/cmd/api.test]
cmd/api/main_test.go:29:19: undefined: buildMarketDataProvider
cmd/api/main_test.go:29:50: undefined: config.MarketDataModeMock
cmd/api/main_test.go:52:19: undefined: buildMarketDataProvider
cmd/api/main_test.go:52:50: undefined: config.MarketDataModeRepository
cmd/api/main_test.go:66:15: undefined: buildMarketDataProvider
cmd/api/main_test.go:66:46: undefined: config.MarketDataModeRepository
FAIL ux-securities/backend/cmd/api [build failed]
FAIL
```

## Why this is the expected RED

The current configuration has no provider mode, `Load` cannot report unsupported modes, and `cmd/api/main.go` hard-codes the MySQL repository into the authorized slot. The undefined mode constants and runtime builder, plus the old single-return `Load`, are precisely the missing behavior this rework must implement. There was no unrelated syntax, fixture, network, or database failure.

## Status-guard RED cycle

After the initial configuration/runtime GREEN, a second behavior test exposed that the non-authorized slot still admitted statuses outside its allow-list.

Command (working directory `apps/backend`):

```text
gofmt -w internal/integration/marketdata/provider_test.go && go test ./internal/integration/marketdata -run 'TestProviderDowngradesStatusesNotAllowedForUnofficialSources|TestProviderFallsBackToMockAndPreservesSimulatedStatus' -count=1
```

Exit code: `1` (expected RED)

```text
--- FAIL: TestProviderDowngradesStatusesNotAllowedForUnofficialSources (0.00s)
    --- FAIL: TestProviderDowngradesStatusesNotAllowedForUnofficialSources/simulated (0.00s)
        provider_test.go:54: Quote() status = "simulated", want stale safety downgrade
    --- FAIL: TestProviderDowngradesStatusesNotAllowedForUnofficialSources/unexpected (0.00s)
        provider_test.go:54: Quote() status = "unexpected", want stale safety downgrade
FAIL
FAIL ux-securities/backend/internal/integration/marketdata 1.053s
FAIL
```

The `live` case already passed because an earlier guard downgraded it. The two failures prove the remaining bug: the non-authorized slot allowed `simulated` and arbitrary statuses. The paired mock fallback test remained green, protecting the separate rule that the real mock slot must retain `simulated`.
