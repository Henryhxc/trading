# Design and Task Traceability

Task identifiers below mirror `openspec/changes/ux-securities-hkex-trading-terminal/tasks.md`; this file does not replace that checklist or its phase gates.

| Tasks | Contract / responsibility | Implementation or governing artifact | Verification evidence |
|---|---|---|---|
| 0.1 | Record project facts and requirements before implementation | `AGENTS.md`, `agent.md`, `docs/PRD.md`, research, change specs | Phase handoff cites source priority and unresolved conflicts |
| 0.2 | Preserve the supplied UX Securities logo and brand tokens | `docs/assets/ux-securities-logo.png`, `docs/design/design.md` | `docs/evidence/reports/brand-token-source.md` and rendered UI evidence |
| 0.3 | Maintain the project archive and keep production files under 500 lines | `docs/design/*`, `docs/api/*`, `docs/evidence/*`, `apps/*`, `deploy/*`, `db/migrations/*` | Required files exist; phase review includes a production-source line-count scan |
| 0.4 | Keep unresolved product choices explicit and configurable | `docs/PRD.md`, change `design.md`, runtime configuration | Handoff records open decisions and verifies defaults are not represented as user-confirmed |
| 0.5 | Enforce ordered RED → GREEN → targeted verification → UI → review → acceptance | `agent-dispatch/phase-controller.md`, per-phase packets and orchestration state | Gate evidence and state transitions recorded by the main agent |
| 1.1–1.3 | React routing, session state, API transport, error boundary, auth UI and frontend archive | `apps/frontend/src/app/*`, `src/pages/AuthPage.tsx`, `src/shared/api/*`, `src/styles.css`, `docs/design/develop-frontend-design.md` | Frontend build and login/register state screenshots at required viewports |
| 1.4 | Go MVC foundation, request IDs, health and common errors | `apps/backend/internal/{httpapi,service,domain,repository}`, `cmd/api` | `go test ./...`, `go vet ./...`, health/API smoke |
| 1.5 | Create users, sessions and initial accounts with constraints and rollback | `db/migrations/001_users_sessions_accounts.sql`, matching down migration | Isolated apply/down/reapply test and Phase 1 DB report |
| 1.6 | Registration, login, logout, session expiry/revocation and actor isolation | Auth service, repository, HTTP API and session provider | RED/GREEN evidence, auth lifecycle API smoke, user/session assertions |
| 1.7 | Local frontend/backend/MySQL Compose topology and optional phpMyAdmin | `deploy/docker-compose.yml`, app Dockerfiles and frontend proxy | Cold start, healthchecks, host URLs, internal API proxy and phpMyAdmin evidence |
| 1.8–1.9 | Auth behavior RED/GREEN and visual state coverage | Phase 1 TDD handoff and `results/screenshots/phase1-foundation/` | Auth/session tests; logo, focus, loading/error, route and responsive screenshots |
| 1.10–1.12 | Independent review, DB report and Phase 1 acceptance | `agent-dispatch/reviews/phase1-review.md`, QA/DB reports, orchestration state | Review conclusion and all required runtime/browser evidence; main agent alone updates checkboxes |
| 2.1–2.3 | Responsive market workbench, quote/candle presentation and watchlist interactions | `apps/frontend/src/features/market/*`, watchlist UI/model/API client, frontend design archive | Build and market-workbench screenshots for loading/empty/error/success and three viewports |
| 2.4–2.7 | Provider contract, mock/public adapters, market/watchlist APIs and source-state documentation | Go market/provider/watchlist modules; migration `002_stocks_daily_quotes_watchlists.sql`; `docs/api/hkex-market-data.md` | Provider/fallback/contract tests, API smoke, DB report; public feeds never report `live` |
| 2.8–2.12 | Market RED/GREEN, visual QA, independent review and Phase 2 acceptance | Phase 2 handoffs/review, `results/screenshots/phase2-market-workbench/`, DB report | Provider fixtures, user-isolation/idempotency tests, screenshot and review gates |
| 3.1–3.3 | Order form and linked account, positions and order-history UI | Trading/account frontend features and API DTO mapping | Build and full paper-trading UI flow at three viewports |
| 3.4–3.7 | Reuse Phase 1 accounts; add orders/positions, server-side validation, idempotency, atomic fills and APIs | Migration `003_orders_positions.sql`; Go paper-trading service/repository/handlers | DB report plus validation, concurrency, rollback, state-machine and user-isolation tests |
| 3.8–3.12 | Trading RED/GREEN, end-to-end flow, independent review and final acceptance | Phase 3 handoffs/review, `results/screenshots/phase3-paper-trading/`, final report | Complete simulated buy/sell lifecycle, Docker final smoke, strict OpenSpec validation |
| 4.1–4.2 | Final task/evidence audit; report remaining risks honestly | `tasks.md`, orchestration state, evidence reports | All required checkboxes and gates verified before overall completion |
