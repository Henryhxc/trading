# Demo Portfolio Seed Evidence

Date: 2026-10-25

The local MySQL demo account (`021021hxc@gmail.com`) was seeded with paper-trading records using `tools/seed_demo_portfolio.sql`.

| Check | Result |
| --- | --- |
| Open positions | 4 |
| Seeded filled orders | 12 |
| Order date range | 2026-06-03 to 2026-10-10 |
| Cash balance | HKD 815,300.00 |
| Available cash | HKD 815,300.00 |
| Seed repeat | Idempotent by fixed client order IDs |

Open quantities are `00700` 200, `09988` 300, `01810` 500, and `00005` 1,000. The seed is development-only and does not connect to a broker or send real orders.
