# UX Securities

Personal Hong Kong equities terminal branded for UX Securities Limited. React frontend, Go backend, MySQL persistence, Docker Compose deployment, and an optional AkShare data collector.

This is a development application for **simulated trading only**. It does not connect to a real broker or send real orders. Public market data is unofficial and may be delayed, stale, or unavailable.

## Run locally

Install Docker with Compose, then run from the project root:

```sh
docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml --profile admin up -d --build
```

- Frontend: http://localhost:3000
- Backend health: http://localhost:8080/health
- phpMyAdmin: http://localhost:8081
- MySQL: localhost:3306

The backend applies database migrations at startup. Register a local account through the frontend; existing local users and database volumes are not included in Git.

Compose contains development-only fallback credentials. Set your own `DB_PASSWORD`, `MYSQL_ROOT_PASSWORD`, and `AKSHARE_WEBHOOK_SECRET` through environment variables or a local ignored environment file before exposing any service. Do not use the development defaults in production.

Market data defaults to mock. To enable the optional AkShare collector:

```sh
MARKET_DATA_MODE=akshare docker compose -p ux-securities-hkex-trading-terminal -f deploy/docker-compose.yml --profile admin --profile akshare up -d --build
```

Upstream availability is not guaranteed. The application identifies unavailable data rather than claiming official HKEX real-time access.

## Checks

```sh
cd apps/backend
go test ./... -count=1
go vet ./...
```

```sh
cd apps/frontend
npm ci
npm run test:frontend
npm run build
```

```sh
python3 -m unittest discover -s services/akshare/tests -p 'test_*.py'
openspec validate ux-securities-hkex-trading-terminal --strict
```

## Documentation and current limitations

- Product scope: [docs/PRD.md](docs/PRD.md)
- API contract: [docs/api/openapi.yaml](docs/api/openapi.yaml)
- Workbench design: [docs/design/workbench-layout.md](docs/design/workbench-layout.md)
- K-line history contract: [docs/design/candlestick-history.md](docs/design/candlestick-history.md)
- Latest layout verification: [verification report](openspec/changes/ux-securities-hkex-trading-terminal/results/workbench-layout/verification-2026-10-24.md)
- Outstanding acceptance tasks: [tasks.md](openspec/changes/ux-securities-hkex-trading-terminal/tasks.md)

The default directory supports metadata-only securities without quotes. The 2,791-row directory imported into the current local database is not a bundled database snapshot; see `tools/import_hk_directory.py` and the verification report for provenance and import limitations. The chart supports minute/day/week/month/year controls with a half-year default window; this does not imply that every source provides complete history or native minute data. Real upstream success and the full three-viewport trading acceptance gates remain open. Historical evidence may describe the workspace before Git initialization.

The chart uses [TradingView Lightweight Charts](https://github.com/tradingview/lightweight-charts); third-party notices are in `apps/frontend/public/third-party-notices.txt`.
