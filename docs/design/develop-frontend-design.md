# Phase 1 Frontend Design

## Scope

Phase 1 delivers the branded registration/login shell and a minimal authenticated session confirmation/logout state. It does not render a market workbench or local account balance, and it does not submit real or simulated orders.

## Directory and file contracts

```text
apps/frontend/
  index.html                    app mount and document metadata
  src/main.tsx                  React entry point and global error boundary
  src/app/AppRouter.tsx         browser routes and protected session route
  src/app/SessionProvider.tsx    shared session bootstrap and state
  src/app/ErrorBoundary.tsx     render-failure recovery boundary
  src/app/navigation.ts         same-origin history navigation helper
  src/pages/AuthPage.tsx        login/register form and authenticated shell
  src/shared/api/client.ts      same-origin API transport and status errors
  src/shared/api/auth.ts        authentication endpoint DTOs and methods
  src/styles.css                brand tokens, responsive form, focus/error states
  vite.config.ts                React development/build configuration
  nginx.conf                    static SPA and internal API reverse proxy
  Dockerfile                    frontend build and nginx runtime
```

The browser calls only same-origin `/api/v1/*` endpoints. The API client sets `credentials: "include"`; no token is read from JavaScript storage. Nginx forwards API requests to the Compose `backend` service and forwards `/health` for a deployment smoke check.

## Interaction and state

- On load, the session provider calls `GET /api/v1/me`. While pending, the router renders a session-check state rather than a signed-out form. A 401 resolves to anonymous; network or server failures render a distinct retryable verification error.
- `/login` and `/register` are public authentication routes. `/workspace` is protected: anonymous users return to sign-in, while authenticated users see only the Phase 1 account confirmation state. No market workbench is added in this phase.
- Sign-in posts `{identity,password}` to `/api/v1/auth/login`.
- Registration posts `{username,email,password}` to `/api/v1/auth/register`; the password hint matches the backend 12-character minimum.
- Successful authentication displays the returned username/email and a sign-out command; it does not pretend a workbench exists before later phases add routes.
- Form/network failures show an inline alert while preserving entered identity data. Busy state disables submit to avoid accidental duplicate form requests.
- Sign-out posts to `/api/v1/auth/logout`; success clears local identity and returns to sign-in. A failed request preserves the session view and shows retry guidance.

## Visual and accessibility contract

The logo is included at its native aspect ratio from `docs/assets/ux-securities-logo.png`. Brand variables are `#1D41A9`, `#2857E2`, `#0B0B0D`, `#FFFFFF`, and `#6B7280`. Inputs have visible keyboard focus, associated labels, semantic autocomplete values, and errors use `role=alert`. Login/register choices navigate routes and expose the current page. Layout is constrained to 440px and switches to reduced margins/padding under 520px without horizontal scrolling at 390px.

## Failure branches and validation

- 400/409/401 API errors display the server's safe user-facing message.
- Non-JSON/network/upstream failure gets a generic client message.
- Protected session lookup failure does not reveal server details and can be retried separately from a signed-out session.
- `npm install && npm run build` validates the production bundle.
- Browser checks at 1440x900, 1280x800, and 390x844, including focus, registration/login errors and success, are required from the validation lane; not completed by a build check.

Logo path: `docs/assets/ux-securities-logo.png` is copied into the frontend image at build time. No provider calls, provider secrets, account records, or order state are held in frontend storage.

## Phase 2 workbench additions

`src/pages/WorkbenchPage.tsx` renders the protected market workspace: internal search, favorites, quote summary, source/status badge, daily close bars, loading, empty, stale/unavailable and error states. `src/shared/api/market.ts` is the only market data client and calls same-origin `/api/v1/markets/*` and `/api/v1/watchlist` routes; no provider URL or credential is present in the bundle. The active symbol is retained in `sessionStorage` so returning from a filtered list preserves the selected security.

The two-column desktop layout collapses to a single-column mobile layout below 760px. Brand tokens and transaction semantic colors remain in `styles.css`; status badges include text as well as color. Browser validation owns responsive screenshots and overflow/focus checks.
