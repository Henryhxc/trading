# Personal Trading Terminal — Portfolio Overview

This project explores how a personal investor can monitor Hong Kong equities, organize favorites, inspect price history, simulate orders, track a portfolio, and use an AI assistant without switching between several disconnected tools.

## 1. Workspace first

The main experience is a dense dark workspace. Navigation stays on the left, the active task occupies the center, and contextual actions remain on the right where appropriate.

![Market workspace](../evidence/screenshots/personal-terminal/dark/market.png)

The market directory and favorites are intentionally different views: the market page supports discovery across the directory, while favorites keeps a short personal list that can be reordered.

![Favorites](../evidence/screenshots/personal-terminal/dark/favorites.png)

## 2. Portfolio and performance

Positions show current holdings and support distinct day, month, and year performance views. Each screenshot represents a separate state rather than repeating the same page.

![Positions](../evidence/screenshots/personal-terminal/dark/positions.png)

| Period | Screenshot |
| --- | --- |
| Day | [View](../evidence/screenshots/personal-terminal/dark/positions-profit-day.png) |
| Month | [View](../evidence/screenshots/personal-terminal/dark/positions-profit-month.png) |
| Year | [View](../evidence/screenshots/personal-terminal/dark/positions-profit-year.png) |

## 3. Orders and cash tools

The prototype includes paper-order history and account tools for cash overview, deposits, withdrawals, exchange, financing, and transaction history. These pages demonstrate the information architecture; they do not perform real money movement.

![Orders](../evidence/screenshots/personal-terminal/dark/orders.png)

![Cash overview](../evidence/screenshots/personal-terminal/dark/cash-overview.png)

## 4. Notes and AI assistant

Notes preserve research beside the workspace. The assistant is positioned after the core workbench because it supports the workflow rather than replacing it. It can help interpret information, summarize context, and guide the next action, while order execution remains bounded by explicit user confirmation and backend authorization.

![Notes](../evidence/screenshots/personal-terminal/dark/notes-preview.png)

![AI assistant](../evidence/screenshots/personal-terminal/dark/assistant.png)

## 5. What I focused on

- Translating prototypes into a coherent desktop information hierarchy.
- Keeping discovery, favorites, chart analysis, orders, and portfolio states visually distinct.
- Designing a dark financial interface with readable density and consistent emphasis.
- Defining clear boundaries between unofficial market data, simulated trading, and real brokerage execution.
- Using specification-led development, tests, review gates, and visual evidence to keep implementation traceable.

For a deeper product walkthrough, see [README-product.md](README-product.md). For architecture and security details, see [README-engineering.md](README-engineering.md).
