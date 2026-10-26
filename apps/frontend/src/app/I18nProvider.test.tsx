import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { I18nProvider, LanguageToggle, useI18n } from "./I18nProvider";

function Probe() {
  const { language, setLanguage, t } = useI18n();
  return <div><span>{t("Market")}</span><span>{language}</span><span>{t("Candlestick chart")}</span><span>{t("Exchange currencies")}</span><button onClick={() => setLanguage(language === "en" ? "zh" : "en")}>toggle</button><LanguageToggle /></div>;
}

describe("I18nProvider", () => {
  afterEach(() => sessionStorage.clear());

  it("switches language and remembers the choice across renders", () => {
    render(<I18nProvider><Probe /></I18nProvider>);
    expect(screen.getByText("Market")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "toggle" }));
    expect(screen.getByText("行情")).toBeTruthy();
    expect(sessionStorage.getItem("ux_language")).toBe("zh");
    expect(document.documentElement.lang).toBe("zh-CN");
    expect(screen.getByText("K 线行情")).toBeTruthy();
    expect(screen.getByText("货币兑换")).toBeTruthy();
  });

  it("restores the saved language on the next app mount", () => {
    sessionStorage.setItem("ux_language", "zh");
    render(<I18nProvider><Probe /></I18nProvider>);
    expect(screen.getByText("行情")).toBeTruthy();
  });

  it("switches back to English and keeps chart copy in the selected language", () => {
    sessionStorage.setItem("ux_language", "zh");
    render(<I18nProvider><Probe /></I18nProvider>);
    fireEvent.click(screen.getByRole("button", { name: "EN" }));
    expect(screen.getByText("Market")).toBeTruthy();
    expect(screen.getByText("Candlestick chart")).toBeTruthy();
    expect(document.documentElement.lang).toBe("en");
    expect(sessionStorage.getItem("ux_language")).toBe("en");
  });
});
