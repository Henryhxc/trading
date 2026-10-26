# UX Securities

Personal Hong Kong equities terminal branded for UX Securities Limited. React frontend, Go backend, MySQL persistence, Docker Compose deployment, and an optional AkShare data collector.

This is a development application for **simulated trading only**. It does not connect to a real broker or send real orders. Public market data is unofficial and may be delayed, stale, or unavailable.

## Demo account

- Account email: `021021hxc@gmail.com`
- Username: `021021hxc`
- Password： 'qweQWE123...'
- Password: provisioned locally and intentionally not committed to Git
- Current favorites: Tencent (`00700`), Alibaba (`09988`), Xiaomi (`01810`)

The demo password was reset in the local database. Use the password supplied separately by the project owner; do not copy it into this README, screenshots, logs or frontend bundles.

## Run locally

Install Docker with Compose, then run from the project root:

```sh
cd "/Users/hongxincong/project/UX Security"
docker compose up
```

This root-level command starts MySQL, the backend, the frontend and the AkShare collector. Use `Ctrl+C` to stop the foreground stack, or add `-d` to run it in the background:

```sh
cd "/Users/hongxincong/project/UX Security"
docker compose up -d
```

- Frontend: http://localhost:3000
- Backend health: http://localhost:8080/health
- phpMyAdmin: http://localhost:8081
- MySQL: localhost:3306

The backend applies database migrations at startup. Register a local account through the frontend; existing local users and database volumes are not included in Git.

Compose contains development-only fallback credentials. Set your own `DB_PASSWORD`, `MYSQL_ROOT_PASSWORD`, and `AKSHARE_WEBHOOK_SECRET` through environment variables or a local ignored environment file before exposing any service. Do not use the development defaults in production. phpMyAdmin is intentionally opt-in because port 8081 is commonly occupied by other local projects:

```sh
docker compose --profile admin up -d
```

## Security notes

- **Password algorithms:** passwords are stored as one-way bcrypt hashes with a work factor managed by the Go service; plaintext passwords are never persisted or logged. Argon2id is the preferred migration option if the deployment standard requires memory-hard hashing. The system should select one password-hashing scheme per stored record and identify it by the algorithm prefix, rather than stacking multiple hashes or treating encryption as password storage.
- **Session authentication:** the browser receives an opaque, randomly generated `HttpOnly` session cookie. The database stores only its SHA-256 digest, and each protected request checks expiry, revocation, and active-user status. In production, serve the frontend and API over HTTPS with `COOKIE_SECURE=true` and an explicit allowed origin.
- **Symmetric authentication:** service-to-service callbacks such as the AkShare webhook use a shared secret and HMAC-style request authentication with timestamp/age checks. This is for integrity and replay resistance; it is not a replacement for TLS, user login, or broker authorization. Shared secrets belong in environment variables or a secret manager, never in source, screenshots, or logs.
- **Frontend/backend access:** the React app calls only the internal `/api/v1` API. Authorization is enforced in Go middleware and repositories using the session actor, so the client cannot choose another user's account, orders, positions, or watchlist by changing an ID in the request.
- **SQL injection prevention:** database code uses `database/sql` placeholders (`?`) and bound arguments for user-controlled values. Dynamic SQL is limited to allow-listed table/column fragments selected by code; user input is never concatenated into SQL. Error responses and logs omit SQL text, credentials, tokens, passwords, and raw provider payloads.
- **Scope:** these controls protect the local simulated-trading prototype. They do not grant real broker access, real-order permission, or HKEX feed authorization. Before production, add key rotation, rate limiting, CSRF/origin review, dependency scanning, audit retention, and an external security review.

The Compose entrypoint defaults to the AkShare ingestion mode. To use the mock provider instead:

```sh
MARKET_DATA_MODE=mock docker compose up -d
```

Upstream availability is not guaranteed. The application identifies unavailable data rather than claiming official HKEX real-time access.

## Demo data

The active development database has 125 public Tencent daily bars for each of Alibaba and Xiaomi covering `2026-04-24` through `2026-10-23`; the latest quote and bars are labeled `tencent` / `unofficial` and are not official HKEX real-time data. The collector defaults now include both `09988` and `01810` for subsequent updates.

The demo account also has a repeatable paper-trading portfolio seed: four open positions (`00700`, `09988`, `01810`, `00005`) and twelve filled orders dated from June through October 2026. To apply it to the local demo account after the database is running, use:

```sh
MYSQL_CONTAINER="$(docker compose ps -q mysql)"
docker exec -i "$MYSQL_CONTAINER" sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" --database="$MYSQL_DATABASE"' < tools/seed_demo_portfolio.sql
```

The seed is idempotent by `client_order_id`, recalculates the four demo positions, and keeps the account cash balance consistent with the seeded fills. It contains no credentials.

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
