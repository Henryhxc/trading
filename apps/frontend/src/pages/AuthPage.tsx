import { FormEvent, useState } from "react";
import { useSession } from "../app/SessionProvider";
import { navigate } from "../app/navigation";
import { authApi } from "../shared/api/auth";
import { LanguageToggle, useI18n } from "../app/I18nProvider";

export type AuthMode = "login" | "register";

export function AuthPage({ mode }: { mode: AuthMode }) {
  const session = useSession();
  const { t } = useI18n();
  const [identity, setIdentity] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setMessage("");
    try {
      const result = mode === "login"
        ? await authApi.login(identity, password)
        : await authApi.register(username, email, password);
      if (result?.user) {
        session.acceptUser(result.user);
        navigate("/workspace");
      }
    } catch (error) {
      setMessage(t(error instanceof Error ? error.message : "Request could not be completed"));
    } finally {
      setBusy(false);
    }
  }

  async function signOut() {
    try {
      await authApi.logout();
      session.clearUser();
      setPassword("");
      navigate("/login");
    } catch {
      setMessage(t("Sign out could not be completed. Please retry."));
    }
  }

  if (session.status === "authenticated" && session.user) {
    return (
      <main className="auth-layout terminal-auth">
        <LanguageToggle />
        <img className="brand-logo terminal-auth-logo" src="/usmart-logo.png" alt="USMART 盈立證券" />
        <section className="account-panel" aria-live="polite">
          <p className="eyebrow">{t("PERSONAL ACCOUNT")}</p>
          <h1>{t("Welcome, {name}", { name: session.user.username })}</h1>
          <p className="account-email">{session.user.email}</p>
          <p className="account-note">{t("Your secure session is active.")}</p>
          {message && <p className="error-message" role="alert">{message}</p>}
          <button className="primary-button" type="button" onClick={signOut}>{t("Sign out")}</button>
        </section>
      </main>
    );
  }

  return (
    <main className="auth-layout terminal-auth">
      <LanguageToggle />
      <img className="brand-logo terminal-auth-logo" src="/usmart-logo.png" alt="USMART 盈立證券" />
      <section className="auth-panel">
        <div className="auth-heading">
          <p className="eyebrow">{t("USMART 盈立證券")}</p>
          <h1>{t(mode === "login" ? "Sign in" : "Create your account")}</h1>
          <p className="subheading">{t("Hong Kong market workspace")}</p>
        </div>
        <nav className="mode-switch" aria-label={t("Account access")}>
          <button className={mode === "login" ? "mode active" : "mode"} aria-current={mode === "login" ? "page" : undefined} type="button" onClick={() => navigate("/login")}>{t("Sign in")}</button>
          <button className={mode === "register" ? "mode active" : "mode"} aria-current={mode === "register" ? "page" : undefined} type="button" onClick={() => navigate("/register")}>{t("Register")}</button>
        </nav>
        <form onSubmit={submit} noValidate>
          {mode === "register" ? (
            <>
              <label htmlFor="username">{t("Username")}</label>
              <input id="username" name="username" autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} required minLength={3} />
              <label htmlFor="email">{t("Email")}</label>
              <input id="email" name="email" type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} required maxLength={128} />
            </>
          ) : (
            <>
              <label htmlFor="identity">{t("Username or email")}</label>
              <input id="identity" name="identity" autoComplete="username" value={identity} onChange={(event) => setIdentity(event.target.value)} required />
            </>
          )}
          <label htmlFor="password">{t("Password")}</label>
          <input id="password" name="password" type="password" autoComplete={mode === "login" ? "current-password" : "new-password"} value={password} onChange={(event) => setPassword(event.target.value)} required minLength={12} />
          {mode === "register" && <p className="field-hint">{t("Use at least 12 characters.")}</p>}
          {message && <p className="error-message" role="alert">{message}</p>}
          <button className="primary-button" type="submit" disabled={busy}>{t(busy ? "Please wait" : mode === "login" ? "Sign in" : "Create account")}</button>
        </form>
        <p className="security-note">{t("Paper trading only · No real orders are submitted")}</p>
      </section>
      <footer>{t("Market data and simulated account services are provided for development use.")}</footer>
    </main>
  );
}
