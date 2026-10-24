# IAB Browser Check: 2026-10-24

## Environment

- Browser: Codex in-app browser (IAB), connected successfully in this run.
- Frontend: `http://localhost:3000`; the page resolved to `/login` and `/register`.
- Synthetic-account registration was not submitted. That database write is pending explicit action-time confirmation.

## Captures and observations

The login and registration pages were captured at `1440x900`, `1280x800`, and `390x844` under `results/screenshots/phase3-paper-trading/`. The measured login document scroll dimensions matched each viewport exactly: `1440x900`, `1280x800`, and `390x844`; no horizontal overflow was observed. The registration form and UX Securities logo were fully visible in the inspected desktop and mobile captures.

These captures prove only the unauthenticated entry surface. The following required Phase 2/3 browser states were not exercised: authenticated search, quote/candles, watchlist, stream update/reconnect, stale/unavailable, order confirmation/fill, positions/history, and their success/error/empty states. Accordingly, tasks `2.9`, `3.9`, and the associated browser gates remain open.

## Current verification

- `GOCACHE=/tmp/ux-go-build go test ./... -count=1`: PASS.
- `GOCACHE=/tmp/ux-go-build go vet ./...`: PASS.
- `npm run test:frontend`: PASS (6 source assertions).
- `npm run build`: PASS.
- `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s services/akshare/tests -p 'test_*.py' -v`: PASS (8 tests).
- `openspec validate ux-securities-hkex-trading-terminal --strict`: PASS.
- `docker ps` showed the target frontend/backend/MySQL/AkShare/phpMyAdmin containers running; the core containers were healthy and phpMyAdmin returned HTTP 200. A `docker compose ps` invocation using the default `deploy` project showed only `deploy-mysql`; this is not evidence of a fresh cold-start of the complete target project.
- A fresh real AkShare `stock_hk_spot_em()` pull failed with `RemoteDisconnected`. No successful upstream-to-webhook-to-MySQL-to-REST-to-WebSocket chain is claimed.

## Decision

Partial evidence only; Phase 2/3 remain `rework`. Do not check browser, upstream-success, independent-review, or final-gate tasks on this basis.
