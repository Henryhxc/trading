# Phase 2 Provider Config Rework Handoff

Status: `DONE` (bounded provider-config packet only)

Date: 2026-10-24 (Asia/Hong_Kong)

## Scope and outcome

This rework removes the runtime's hard-coded use of the MySQL market repository in the authorized provider slot.

- Unset `MARKET_DATA_MODE` defaults explicitly to `mock`.
- `mock` assembles only `MockMarketDataProvider`; quote API responses are `source=mock`, `data_status=simulated`.
- Explicit `repository` mode places the MySQL market repository in the non-authorized slot and retains mock fallback.
- Non-authorized quote statuses are allow-listed to `unofficial`, `delayed`, `stale`, and `unavailable`. `live`, `simulated`, and unknown status values from that slot are downgraded to `stale`.
- `authorized`, `public`, and unknown modes fail during configuration loading. No nonexistent adapter is represented as connected.
- `repository` mode without a repository dependency also fails closed.
- Docker Compose and the market-data boundary document use the same `mock` default and describe the implemented modes accurately.

No external quote feed or broker was contacted. No shared database, migration, frontend, business handler/service, task checkbox, phase controller, or orchestration state was changed.

## TDD evidence

RED was captured before production implementation in [`results/phase2-provider-config-red.md`](../../results/phase2-provider-config-red.md).

Initial RED command from `apps/backend`:

```text
gofmt -w internal/config/config_test.go cmd/api/main_test.go && go test ./internal/config ./cmd/api -count=1
```

Exit `1`: provider mode constants, error-returning `Load`, and runtime builder were absent.

A second RED cycle covered the incomplete non-authorized status guard:

```text
gofmt -w internal/integration/marketdata/provider_test.go && go test ./internal/integration/marketdata -run 'TestProviderDowngradesStatusesNotAllowedForUnofficialSources|TestProviderFallsBackToMockAndPreservesSimulatedStatus' -count=1
```

Exit `1`: the non-authorized slot leaked `simulated` and an arbitrary `unexpected` status. The `live` downgrade and separate mock `simulated` behavior already passed.

Both focused suites passed after the minimum implementations.

## Behavioral coverage

- `internal/config/config_test.go`
  - unset mode -> `mock`;
  - explicit `repository` accepted;
  - `authorized`, `public`, and unknown modes rejected.
- `cmd/api/main_test.go`
  - mock runtime composition returns mock/simulated;
  - repository is composed as unofficial and a claimed live quote is downgraded;
  - missing repository dependency fails closed;
  - an authenticated request through the real HTTP router, market service, and configured mock provider returns HTTP 200 with mock/simulated.
- `internal/integration/marketdata/provider_test.go`
  - unauthorized live/simulated/unknown statuses become stale;
  - the separate mock fallback retains simulated.

## Verification

All listed commands exited `0`.

```text
cd apps/backend && go test ./internal/integration/marketdata ./internal/config ./cmd/api -count=1
PASS

cd apps/backend && go test ./cmd/api -run 'TestMockModeQuoteAPISmoke|TestBuildMarketDataProvider' -count=1 -v
PASS (GET /api/v1/markets/quotes/00700 -> 200; decoded source/status = mock/simulated)

cd apps/backend && go test ./... -count=1
PASS

cd apps/backend && go vet ./...
PASS

env -u MARKET_DATA_MODE docker compose -f deploy/docker-compose.yml config --format json | jq -r '.services.backend.environment.MARKET_DATA_MODE'
mock

docker compose -f deploy/docker-compose.yml config -q
PASS

npx openspec validate ux-securities-hkex-trading-terminal --strict
Change 'ux-securities-hkex-trading-terminal' is valid
```

Production file line counts remain below 500:

```text
95  apps/backend/internal/config/config.go
98  apps/backend/cmd/api/main.go
173 apps/backend/internal/integration/marketdata/provider.go
```

## Files written

- `apps/backend/internal/config/config.go`
- `apps/backend/internal/config/config_test.go`
- `apps/backend/cmd/api/main.go`
- `apps/backend/cmd/api/main_test.go`
- `apps/backend/internal/integration/marketdata/provider.go`
- `apps/backend/internal/integration/marketdata/provider_test.go`
- `deploy/docker-compose.yml`
- `docs/api/hkex-market-data.md`
- `openspec/changes/ux-securities-hkex-trading-terminal/results/phase2-provider-config-red.md`
- `openspec/changes/ux-securities-hkex-trading-terminal/agent-dispatch/handoffs/phase2-provider-config-rework-handoff.md`

## Runtime handoff note

`docker compose ... ps` returned no running services during this packet. Per the assignment, this agent did not build/restart the shared Compose stack. The API smoke above is an in-process request through the production router/service/provider stack and does not touch MySQL. The main agent still owns the rebuilt-container localhost GET smoke during aggregate runtime acceptance.

The workspace root has no `.git` directory, so `git diff --check` is unavailable here; this is reported rather than claimed as PASS.
