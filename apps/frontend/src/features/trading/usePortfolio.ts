import { useCallback, useEffect, useRef, useState } from "react";
import { type Account, type Position, tradingApi } from "../../shared/api/trading";

export function usePortfolio() {
  const [account, setAccount] = useState<Account | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const request = useRef(0);
  const refresh = useCallback(async () => {
    const current = ++request.current;
    setLoading(true); setError("");
    try {
      const [balance, holdings] = await Promise.all([tradingApi.account(), tradingApi.positions()]);
      if (current === request.current) { setAccount(balance); setPositions(holdings?.items || []); }
    } catch (cause) {
      if (current === request.current) setError(cause instanceof Error ? cause.message : "Portfolio is unavailable");
    } finally { if (current === request.current) setLoading(false); }
  }, []);
  useEffect(() => { void refresh(); return () => { request.current++; }; }, [refresh]);
  return { account, positions, loading, error, refresh };
}

export const money = (value: number | null | undefined) => value == null ? "—" : `HKD ${value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
