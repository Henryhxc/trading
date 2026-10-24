import { readFile } from "node:fs/promises";

const root = new URL("../src/", import.meta.url);
const files = {
  api: new URL("shared/api/trading.ts", root),
  page: new URL("pages/TradingPanel.tsx", root),
  portfolio: new URL("features/trading/PortfolioPanel.tsx", root),
  orders: new URL("features/trading/OrderHistoryPanel.tsx", root),
};

const sources = Object.fromEntries(
  await Promise.all(Object.entries(files).map(async ([name, url]) => {
    try {
      return [name, await readFile(url, "utf8")];
    } catch {
      return [name, ""];
    }
  })),
);

const assertions = [
  ["trading API stays on internal account endpoint", sources.api.includes('"/account"')],
  ["trading API exposes idempotent order submission", sources.api.includes("Idempotency-Key")],
  ["trading API sends the same idempotency key in the JSON body", sources.api.includes("client_order_id")],
  ["trading panel renders order form", sources.page.includes("Place paper order")],
  ["portfolio and order history pages stay separate", sources.portfolio.includes("Open positions") && sources.orders.includes("Order history")],
  ["trading panel refreshes after submit", sources.page.includes("onOrderCreated")],
];

const failures = assertions.filter(([, passed]) => !passed);
if (failures.length > 0) {
  console.error(`Frontend source guard failed: ${failures.length}/${assertions.length}`);
  for (const [name] of failures) console.error(`- ${name}`);
  process.exit(1);
}
console.log(`Frontend source guard passed: ${assertions.length} assertions`);
