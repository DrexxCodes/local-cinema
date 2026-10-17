import { SESSION_KEY } from "@/constants/storage";
import type { Session } from "@/types";

export function getSession(): Session | null {
  try { const raw = localStorage.getItem(SESSION_KEY); return raw ? JSON.parse(raw) : null; }
  catch { return null; }
}
export function setSession(s: Session) { localStorage.setItem(SESSION_KEY, JSON.stringify(s)); }
export function clearSession() { localStorage.removeItem(SESSION_KEY); }
