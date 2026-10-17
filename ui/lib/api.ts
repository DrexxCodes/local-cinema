import { clearSession, getSession } from "./session";

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string, public body?: any) { super(message); }
}

export async function api<T = any>(path: string, init: RequestInit = {}): Promise<T> {
  const s = getSession();
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
  if (s) { headers.set("Authorization", `Bearer ${s.token}`); headers.set("X-Session-Id", s.session_id); }
  let res: Response;
  try { res = await fetch(path, { ...init, headers }); }
  catch { throw new ApiError(0, "network", "Can't reach the cinema server. Check your WiFi."); }
  const body = await res.json().catch(() => null);
  if (!res.ok) {
    if (res.status === 401 && s && path !== "/api/auth/login") {
      clearSession();
      if (typeof window !== "undefined") window.location.href = "/login/?ended=1";
    }
    throw new ApiError(res.status, body?.error ?? "error", body?.message ?? "Something went wrong.", body);
  }
  return body as T;
}

// <img> and <video> cannot send headers, so pass credentials in the query string.
export function withAuth(path: string): string {
  const s = getSession();
  if (!s) return path;
  const sep = path.includes("?") ? "&" : "?";
  return `${path}${sep}token=${encodeURIComponent(s.token)}&sid=${encodeURIComponent(s.session_id)}`;
}
