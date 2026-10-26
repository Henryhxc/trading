import React from "react";
import { createRoot } from "react-dom/client";
import { AppRouter } from "./app/AppRouter";
import { ErrorBoundary } from "./app/ErrorBoundary";
import { SessionProvider } from "./app/SessionProvider";
import { I18nProvider } from "./app/I18nProvider";
import "./styles.css";

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <I18nProvider>
        <SessionProvider>
          <AppRouter />
        </SessionProvider>
      </I18nProvider>
    </ErrorBoundary>
  </React.StrictMode>,
);
