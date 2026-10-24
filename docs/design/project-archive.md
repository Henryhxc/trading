# UX Securities Project Archive

## Phase 1 Foundation

- Business semantics source: `openspec/changes/ux-securities-hkex-trading-terminal/design.md`.
- Backend technical contract: `docs/design/develop-backend-design.md`.
- Frontend technical contract: `docs/design/develop-frontend-design.md`.
- Task traceability: `docs/design/develop-design-task-linkage.md`.
- Brand logo: `docs/assets/ux-securities-logo.png`.
- Local Compose entry: `deploy/docker-compose.yml`.

Phase 1 code establishes registration, login, logout, current-user lookup, session storage/revocation/expiry, initial account creation, health, and local deployment topology. It does not provide stocks, quotes, watchlists, orders, positions, or execution. Those remain assigned to later phases.

Default host ports are frontend 3000, backend 8080, MySQL 3306, and phpMyAdmin 8081. Compose supports `FRONTEND_PORT`, `BACKEND_PORT`, `MYSQL_PORT`, and `PHPMYADMIN_PORT` overrides. phpMyAdmin is opt-in via `--profile admin` and should remain local-only. The defaults `DB_PASSWORD=local-dev-only` and `MYSQL_ROOT_PASSWORD=local-root-dev-only` are for isolated development only and must not be deployed.

Current delivery evidence is recorded in phase handoffs, `results/phase1-main-agent-qa-2026-10-24.md`, database change reports, screenshots, and the main-agent orchestration record. Phase 1 remains unaccepted until its required browser and phpMyAdmin UI gates and independent re-review pass. Source checks and builds alone are not evidence of runtime acceptance.
