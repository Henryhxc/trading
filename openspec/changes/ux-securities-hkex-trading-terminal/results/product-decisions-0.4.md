# Product Decisions 0.4

Date: 2026-10-24

- Test identity: synthetic local account only; no real identity data.
- Starting paper cash: configurable `INITIAL_CASH_HKD`, default HKD 1,000,000.0000.
- Fees: none in the first paper-trading slice; the server remains authoritative.
- Units: whole-share quantity; lot-size validation remains server-owned and configurable.
- Matching: immediate simulated fill at the validated quote; no broker or real exchange order.
- Scope: Hong Kong equities supported by the provider/cache, with daily candles only.
- Provider: mock by default; AkShare is the selected unofficial delayed/stale development
  provider and never reports `live`; authorized HKEX provider is not configured.
- AkShare operation: continuous polling with a 15-minute minimum interval, signed internal
  webhook, MySQL commit, and post-commit authenticated WebSocket notification.

These decisions are configuration-backed where runtime behavior can vary; they do not
grant permission to connect real trading or execution venues.
