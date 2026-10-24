# Workbench deployment and IAB verification

Date: 2026-10-24, Asia/Hong_Kong. Scope: additional workbench layout/background requirements, not aggregate Phase 2/3 acceptance.

## Deployed result

- Correct Compose project: `ux-securities-hkex-trading-terminal`, config `deploy/docker-compose.yml`.
- Backend and frontend images rebuilt and recreated with `MARKET_DATA_MODE=akshare`. Existing MySQL, account data and AkShare configuration retained; no volume reset.
- Observed frontend/backend/MySQL/AkShare healthy; phpMyAdmin running. Backend `/health` returned `{"status":"ok"}`.
- Frontend build assets match deployed page: `index-DsigX0q6.js`, `index-DYlCPnbW.css`.
- Entire workbench uses dark navy background, blue-grey surfaces, restrained logo-blue controls, green/red signed movements, gold favorites and cyan position markers. The original logo retains its proportions on a small white plaque. This follows the frontend UI skill's semantic color and component-state guidance.

## Directory import

Imported 2,791 HKD Equity metadata rows from the official securities-list XLSX, not an official quote feed. See `directory-source.json` for URL, hash and the declared **2026-10-26** source date, which is later than local **2026-10-24**. This limitation is disclosed in the UI; do not call it today's trading-status snapshot.

The importer upserts metadata in a transaction without overwriting existing prices/history or user/trading data. Post-import, the active development database returned:

```text
total  unique_symbols  no_quote_rows
2791   2791            2788
```

Only three existing rows have cached prices; listing all metadata does not imply complete quote coverage. Isolated idempotent double-import evidence is in `directory-import-isolated.txt`.

## Regression and fresh checks

The provider regression `TestProviderKeepsKnownSymbolWithoutCachedQuoteUnavailable` was first observed failing because a successful unavailable NULL-price row fell through to mock (`Source:mock`, `DataStatus:simulated`). This is an observation summary, not an archived raw RED transcript. The fix preserves NULL-price unavailable rows while retaining fallback for repository misses. Focused GREEN was independently rerun in `independent-review.md`; the main agent also reran the full suite after the fix.

Fresh final-source commands and results:

| Command | Result |
| --- | --- |
| `go test ./... -count=1` | PASS; all Go packages with tests passed |
| `go vet ./...` | PASS; exit 0, no diagnostics |
| `npm run test:frontend` | PASS; 6 source guards, 4 test files, 13 tests |
| `npm run build` | PASS; TypeScript and Vite, 54 modules |
| `openspec validate ux-securities-hkex-trading-terminal --strict` | PASS |

These checks are not real-upstream or successful browser-trade evidence.

## Actual browser evidence

User-selected in-app browser, authenticated existing account. Actual viewport **1311×1076**; DOM width **1311** (no horizontal overflow), document height **1158** on the market page (vertical scroll required). No viewport emulation or CSS zoom was used.

- `market-weekly-iab-1311x1076.png`: default full directory (2,791), left market / middle quote and K-line / right order ticket, nine period controls, half-year bounds, signed percentage, stale source warning. Weekly chart rendered 27 aggregated bars from the cached daily series; this does not prove AkShare upstream history success.
- `favorites-iab-1311x1076.png`: one saved security with gold star; same three-column layout and background. Search `00700` produced one item; clearing search restored the full list.
- `positions-empty-iab-1311x1076.png`: dedicated account/positions/history page, empty positions and history. These full sections do not repeat on Market. Initial loading indicators were also observed before the empty states.
- `no-quote-iab-1311x1076.png`: selected `00001`, no cached spot price, unavailable status, no fabricated history, disabled order submission.
- `minute-empty-iab-1311x1076.png`: selected 1-minute interval; native-minute unavailable message and explicit upstream/half-year limitation, not daily bars relabeled as minute data.

Main agent opened market, favorites, positions and no-quote captures and inspected the live minute state. Background continuity, logo proportions, panel separation, chart visibility and order-ticket placement are consistent at this viewport. The current account has no holdings; cyan held-security markers are covered by automated tests, not a nonempty browser portfolio. No paper order was submitted in this layout verification and no successful fill is claimed. Browser left on Market with the full directory and Tencent weekly chart.

## Still open

- Required Phase 2/3 viewports `1440×900`, `1280×800`, `390×844` and complete order/fill/WebSocket scenarios are not replaced by this actual-viewport evidence.
- Independent review source decision is PASS_WITH_NOTES; its browser visual review remains unavailable. Main-agent screenshots do not constitute independent visual acceptance.
- Real AkShare upstream previously returned `RemoteDisconnected`; a successful scheduler → webhook → DB → REST → authenticated WebSocket chain is still unproven.
- There is no Git repository, so `git diff --check` is unavailable. Overall status stays `rework`.
