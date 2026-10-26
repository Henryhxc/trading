import { fireEvent, render, screen, within } from "../../test-utils";
import { describe, expect, it } from "vitest";
import { QuantAssistant } from "./QuantAssistant";
import { classifyIntents, loadLongTermMemory, saveLongTermMemory } from "./assistantMemory";

function createMemoryStorage(): Storage {
  const values = new Map<string, string>();
  return {
    get length() { return values.size; },
    clear: () => values.clear(),
    getItem: key => values.get(key) ?? null,
    key: index => [...values.keys()][index] ?? null,
    removeItem: key => values.delete(key),
    setItem: (key, value) => values.set(key, value),
  };
}

describe("QuantAssistant", () => {
  it("recognizes several financial intents in one prompt and carries them into a contextual follow-up", () => {
    const intents = classifyIntents("宏观利率变化会怎样影响我的持仓？也比较量化回撤风险");

    expect(intents).toEqual(expect.arrayContaining(["macro", "portfolio", "quant", "risk"]));
    expect(classifyIntents("那第二种呢？", intents)).toEqual(intents);
  });

  it("persists only user-confirmed long-term memory preferences", () => {
    const storage = createMemoryStorage();
    saveLongTermMemory("Prefer downside-first analysis", storage);

    expect(loadLongTermMemory(storage)).toEqual(["Prefer downside-first analysis"]);
    saveLongTermMemory("Prefer downside-first analysis", storage);
    expect(loadLongTermMemory(storage)).toHaveLength(1);
  });

  it("labels multiple detected intents and applies saved memory to a follow-up round", () => {
    const storage = createMemoryStorage();
    saveLongTermMemory("Show downside scenarios first", storage);
    render(<QuantAssistant memoryStorage={storage} />);

    fireEvent.click(screen.getByRole("button", { name: "Open quant assistants" }));
    const panel = screen.getByRole("dialog", { name: "Quant assistants" });
    fireEvent.change(within(panel).getByRole("textbox", { name: "Ask a question" }), {
      target: { value: "How do interest rates affect my portfolio drawdown risk?" },
    });
    fireEvent.click(within(panel).getByRole("button", { name: "Send message" }));

    expect(within(panel).getByText("Macro", { exact: true })).toBeTruthy();
    expect(within(panel).getAllByText("Portfolio", { exact: true }).length).toBeGreaterThan(0);
    expect(within(panel).getAllByText("Risk", { exact: true }).length).toBeGreaterThan(0);
    expect(within(panel).getAllByText(/Show downside scenarios first/)).toHaveLength(2);
  });

  it("lets the user save and remove a long-term preference from the assistant", () => {
    const storage = createMemoryStorage();
    render(<QuantAssistant memoryStorage={storage} />);

    fireEvent.click(screen.getByRole("button", { name: "Open quant assistants" }));
    const panel = screen.getByRole("dialog", { name: "Quant assistants" });
    fireEvent.click(within(panel).getByText("Long-term memory"));
    fireEvent.change(within(panel).getByRole("textbox", { name: "Memory preference" }), {
      target: { value: "Explain assumptions before strategies" },
    });
    fireEvent.click(within(panel).getByRole("button", { name: "Remember" }));

    expect(loadLongTermMemory(storage)).toEqual(["Explain assumptions before strategies"]);
    fireEvent.change(within(panel).getByRole("textbox", { name: "Ask a question" }), {
      target: { value: "What about that?" },
    });
    fireEvent.click(within(panel).getByRole("button", { name: "Send message" }));
    expect(within(panel).getAllByText(/Explain assumptions before strategies/)).toHaveLength(2);

    fireEvent.click(within(panel).getByRole("button", { name: "Forget Explain assumptions before strategies" }));
    expect(loadLongTermMemory(storage)).toEqual([]);
  });

  it("opens with a visible sample user question and mock quant response", () => {
    render(<QuantAssistant />);

    fireEvent.click(screen.getByRole("button", { name: "Open quant assistants" }));
    const panel = screen.getByRole("dialog", { name: "Quant assistants" });

    expect(within(panel).getByText("How should I evaluate a hypothetical SpaceX position and the Starlink trend?", { exact: true })).toBeTruthy();
    expect(within(panel).getByText("Parallel analysis · 3 agents", { exact: true })).toBeTruthy();
    expect(within(panel).getByText("If the theme suffers a hypothetical 40% selloff, compare a left-side entry with a quant pullback plan.", { exact: true })).toBeTruthy();
    expect(within(panel).getByText("Parallel analysis · 4 agents", { exact: true })).toBeTruthy();
    expect(within(panel).getByText(/SpaceX is privately held/)).toBeTruthy();
    expect(within(panel).getByText(/Quant pullback plan/)).toBeTruthy();
    expect(within(panel).getAllByText("Harness synthesis")).toHaveLength(2);
  });

  it("switches between quant agents and answers with a clearly marked mock analysis", () => {
    render(<QuantAssistant />);

    fireEvent.click(screen.getByRole("button", { name: "Open quant assistants" }));
    const panel = screen.getByRole("dialog", { name: "Quant assistants" });
    fireEvent.click(within(panel).getByRole("button", { name: /Risk monitor/ }));
    fireEvent.change(within(panel).getByRole("textbox", { name: "Ask a question" }), {
      target: { value: "How should I think about volatility?" },
    });
    fireEvent.click(within(panel).getByRole("button", { name: "Send message" }));

    expect(within(panel).getByText("How should I think about volatility?")).toBeTruthy();
    expect(within(panel).getAllByText(/Risk lens:/).length).toBeGreaterThan(0);
    expect(within(panel).getAllByText(/Portfolio lens:/).length).toBeGreaterThan(0);
    expect(within(panel).getAllByText(/Macro lens:/).length).toBeGreaterThan(0);
    expect(within(panel).getAllByText(/Quant pullback plan:/).length).toBeGreaterThan(0);
    expect(within(panel).getAllByText("Harness synthesis", { selector: "strong" })).toHaveLength(3);
    expect(within(panel).getByText(/SpaceX is private; Starlink is not a standalone listed stock/)).toBeTruthy();
  });

  it("responds to a drawdown question with a calculated hypothetical stress case", () => {
    render(<QuantAssistant />);

    fireEvent.click(screen.getByRole("button", { name: "Open quant assistants" }));
    const panel = screen.getByRole("dialog", { name: "Quant assistants" });
    fireEvent.change(within(panel).getByRole("textbox", { name: "Ask a question" }), {
      target: { value: "What if it falls another 40% after my position is already down 40%?" },
    });
    fireEvent.click(within(panel).getByRole("button", { name: "Send message" }));

    expect(within(panel).getByText(/A hypothetical 100-unit investment falls to 60, then 36/)).toBeTruthy();
    expect(within(panel).getByText(/cumulative loss of 64%/)).toBeTruthy();
    expect(within(panel).getAllByText(/not actual SpaceX market data/).length).toBeGreaterThan(0);
    expect(within(panel).getByText(/hypothetical \$10,000 portfolio/)).toBeTruthy();
    expect(within(panel).getByText(/position value is capped at \$1,000/)).toBeTruthy();
  });
});
