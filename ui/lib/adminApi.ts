import { ADMIN_TOKEN_KEY } from "@/constants/storage";

export const getAdminToken = () => (typeof window === "undefined" ? "" : sessionStorage.getItem(ADMIN_TOKEN_KEY) ?? "");
export const setAdminToken = (t: string) => sessionStorage.setItem(ADMIN_TOKEN_KEY, t);
export const clearAdminToken = () => sessionStorage.removeItem(ADMIN_TOKEN_KEY);

export async function adminApi<T = any>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Authorization", getAdminToken());
  if (init.body) headers.set("Content-Type", "application/json");
  let res: Response;
  try { res = await fetch(path, { ...init, headers }); }
  catch { throw new Error("Can't reach the cinema server."); }
  const body = await res.json().catch(() => null);
  if (!res.ok) {
    if (res.status === 401) clearAdminToken();
    throw new Error(body?.message ?? "Request failed.");
  }
  return body as T;
}

/** Poster or backdrop for the admin page. `bust` forces a reload after regeneration. */
export function adminImg(kind: "posters" | "backdrops", id: string, v: number, bust = 0): string {
  return `/api/admin-img/${kind}/${id}?v=${v}&b=${bust}&token=${encodeURIComponent(getAdminToken())}`;
}

/** Sends an image as the raw request body (used for poster and backdrop uploads). */
export async function adminUpload<T = any>(path: string, file: File, method: "PUT" | "POST" = "PUT"): Promise<T> {
  let res: Response;
  try { res = await fetch(path, { method, headers: { Authorization: getAdminToken(), "Content-Type": file.type || "application/octet-stream" }, body: file }); }
  catch { throw new Error("Can't reach the cinema server."); }
  const body = await res.json().catch(() => null);
  if (!res.ok) throw new Error(body?.message ?? "Upload failed.");
  return body as T;
}
