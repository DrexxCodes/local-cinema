export function formatDuration(sec: number): string {
  if (!sec) return "";
  const h = Math.floor(sec / 3600), m = Math.floor((sec % 3600) / 60);
  return h ? `${h}h ${m}m` : `${m}m`;
}

export function formatClock(sec: number): string {
  const s = Math.floor(sec), h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60), r = s % 60;
  return (h ? `${h}:${String(m).padStart(2, "0")}` : `${m}`) + `:${String(r).padStart(2, "0")}`;
}

export function formatSize(b: number): string {
  return b > 1e9 ? `${(b / 1e9).toFixed(1)} GB` : `${Math.round(b / 1e6)} MB`;
}

/** Parses "90", "1:30" or "1:02:03" into seconds. Returns 0 when empty or invalid. */
export function parseClock(text: string): number {
  const parts = text.trim().split(":").map(Number);
  if (!parts.length || parts.some((n) => !Number.isFinite(n) || n < 0)) return 0;
  return parts.reduce((total, n) => total * 60 + n, 0);
}

export function parseDate(s: string): number {
  return Date.parse(s.replace(" ", "T"));
}
