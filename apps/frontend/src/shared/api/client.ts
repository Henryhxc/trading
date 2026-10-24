export class ApiError extends Error {
  constructor(message: string, readonly status: number) {
    super(message);
    this.name = "ApiError";
  }
}

export async function apiRequest<T>(path: string, body?: unknown, method?: string, extraHeaders?: Record<string, string>): Promise<T | null> {
  const response = await fetch(`/api/v1${path}`, {
    method: method || (body === undefined ? "GET" : "POST"),
    credentials: "include",
    headers: { ...(body === undefined ? {} : { "Content-Type": "application/json" }), ...extraHeaders },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (response.status === 204) return null;
  const payload = await response.json() as { message?: string };
  if (!response.ok) throw new ApiError(payload.message || "Request could not be completed", response.status);
  return payload as T;
}
