import React from "react";
import { createRoot } from "react-dom/client";
import { AppRouter } from "./app/AppRouter";
import { ErrorBoundary } from "./app/ErrorBoundary";
import { SessionProvider } from "./app/SessionProvider";
import "./styles.css";

createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <SessionProvider>
        <AppRouter />
      </SessionProvider>
    </ErrorBoundary>
  </React.StrictMode>,
);
