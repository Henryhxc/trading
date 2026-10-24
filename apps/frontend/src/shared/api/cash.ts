import { apiRequest } from "./client";

export type CashMovement = { id: number; type: "deposit" | "withdraw"; amount: number; currency: "HKD"; method: string; status: "pending" | "completed"; created_at: string; prototype: boolean };

export const cashApi = { list: () => apiRequest<{ items: CashMovement[] }>("/cash-movements") };
