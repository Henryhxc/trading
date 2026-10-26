import { createElement, type ReactElement } from "react";
import { render as rtlRender } from "@testing-library/react";
import { I18nProvider } from "./app/I18nProvider";

export * from "@testing-library/react";

export function render(ui: ReactElement, options?: Parameters<typeof rtlRender>[1]) {
  return rtlRender(createElement(I18nProvider, null, ui), options);
}
