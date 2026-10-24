import { createContext, PropsWithChildren, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { ApiError } from "../shared/api/client";
import { authApi, User } from "../shared/api/auth";

export type SessionStatus = "loading" | "anonymous" | "authenticated" | "error";
type Session = { status: SessionStatus; user: User | null };
type SessionContextValue = Session & {
  acceptUser: (user: User) => void;
  clearUser: () => void;
  retry: () => void;
};

const SessionContext = createContext<SessionContextValue | null>(null);

export function SessionProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<Session>({ status: "loading", user: null });
  const [attempt, setAttempt] = useState(0);

  const bootstrap = useCallback(async () => {
    setSession({ status: "loading", user: null });
    try {
      const result = await authApi.current();
      setSession(result?.user
        ? { status: "authenticated", user: result.user }
        : { status: "anonymous", user: null });
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setSession({ status: "anonymous", user: null });
      } else {
        setSession({ status: "error", user: null });
      }
    }
  }, []);

  useEffect(() => { void bootstrap(); }, [bootstrap, attempt]);

  const value = useMemo<SessionContextValue>(() => ({
    ...session,
    acceptUser: (user) => setSession({ status: "authenticated", user }),
    clearUser: () => setSession({ status: "anonymous", user: null }),
    retry: () => setAttempt((current) => current + 1),
  }), [session]);

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession() {
  const value = useContext(SessionContext);
  if (!value) throw new Error("useSession must be used within SessionProvider");
  return value;
}
