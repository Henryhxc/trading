import { useEffect, useState } from "react";
import { AuthPage } from "../pages/AuthPage";
import { WorkbenchPage } from "../pages/WorkbenchPage";
import { useSession } from "./SessionProvider";
import { navigate } from "./navigation";
import { useI18n } from "./I18nProvider";

export function AppRouter() {
  const [path, setPath] = useState(window.location.pathname);
  const session = useSession();
  const { t } = useI18n();

  useEffect(() => {
    const updatePath = () => setPath(window.location.pathname);
    window.addEventListener("popstate", updatePath);
    return () => window.removeEventListener("popstate", updatePath);
  }, []);

  useEffect(() => {
    if (path === "/") navigate("/login");
    else if (session.status === "authenticated" && (path === "/login" || path === "/register")) navigate("/workspace");
    else if (session.status === "anonymous" && path === "/workspace") navigate("/login");
    else if (!["/login", "/register", "/workspace"].includes(path)) navigate("/login");
  }, [path, session.status]);

  if (session.status === "loading") {
    return <main className="runtime-state" role="status">{t("Checking your session…")}</main>;
  }
  if (session.status === "error") {
    return <main className="runtime-state" role="alert"><h1>{t("Session unavailable")}</h1><p>{t("Your sign-in status could not be verified.")}</p><button className="primary-button" onClick={session.retry}>{t("Retry")}</button></main>;
  }
  if (path === "/workspace" && session.status === "anonymous") {
    return <main className="runtime-state" role="status">{t("Opening sign in…")}</main>;
  }
  if (path === "/workspace" && session.status === "authenticated") return <WorkbenchPage />;
  return <AuthPage mode={path === "/register" ? "register" : "login"} />;
}
