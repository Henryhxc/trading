# Personal Trading Terminal — Engineering Notes

## Architecture

```text
React frontend
    │ /api/v1
    ▼
Go HTTP handlers and middleware
    ▼
Services / use cases
    ├── repositories ──► MySQL
    └── market adapters ──► unofficial development data
```

- The React frontend owns pages, feature components, state, and presentation.
- The Go backend separates handlers, services, domain models, repositories, and integrations.
- MySQL stores users, session digests, favorites, simulated orders, positions, and balances.
- Docker Compose runs the frontend, backend, database, and optional data collector locally.

## Market-data boundary

The MVP uses an adapter abstraction so the interface does not depend directly on a public source. The frontend calls only the internal API. Data returned by an unofficial provider is labeled accordingly and must not be described as an official HKEX real-time feed.

Expected quote fields include symbol, display name, last price, change, change percentage, volume, turnover, timestamp, source, and data status. The system should fall back explicitly when a provider times out, changes format, becomes stale, or is unavailable.

## Authentication and isolation

- Passwords are stored as one-way bcrypt hashes; plaintext is not persisted.
- The browser receives a random `HttpOnly` session cookie.
- MySQL stores only the SHA-256 digest of the session token.
- Protected requests validate expiry, revocation, and active-user status.
- Service and repository queries are scoped by the authenticated actor so changing a client-side identifier cannot expose another user's favorites, orders, or positions.
- Production deployment should add HTTPS, secure cookies, explicit origin controls, CSRF review, rate limiting, key rotation, and external security review.

## Service authentication

Internal callbacks can use HMAC-style signatures with a shared secret, timestamp, and maximum request age. The secret belongs in environment variables or a secret manager. HMAC protects integrity and replay resistance but does not replace TLS, user authentication, or broker authorization.

## Database safety

Repositories use `database/sql` placeholders and bound parameters for user-controlled values. Dynamic identifiers must come from fixed allow-lists. SQL text, passwords, tokens, database credentials, and raw provider payloads must not be returned to the browser or written to logs.

## Delivery approach

1. Define business states, authorization, failure branches, and acceptance criteria in a specification.
2. Keep a shared OpenAPI contract between the frontend and backend.
3. Implement behavior through RED → minimal GREEN → targeted tests → integration/E2E.
4. Review module boundaries and run visual checks against the intended workspace states.
5. Store screenshots and reports as evidence rather than treating an implementation claim as proof.

## Local verification

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

## Scope disclaimer

This is a local simulated-trading prototype. It does not grant access to a licensed market-data feed, a brokerage account, real funds, or real-order execution. Those capabilities require separate commercial, technical, security, and regulatory work.
