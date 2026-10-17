import { PING_INTERVAL_MS, PING_TIMEOUT_MS, SAMPLE_WINDOW } from "@/constants/signal";
import type { SignalReading } from "@/types";
import { classify, INITIAL_READING } from "./classify";

// One shared monitor per tab. It only probes while something is subscribed.
let reading: SignalReading = INITIAL_READING;
let history: (number | null)[] = [];
let timer: ReturnType<typeof setInterval> | null = null;
const listeners = new Set<() => void>();

function push(sample: number | null) {
  history = [...history, sample].slice(-SAMPLE_WINDOW);
  reading = classify(history);
  listeners.forEach((l) => l());
}

async function probe() {
  if (typeof document !== "undefined" && document.hidden) return; // save battery in background tabs
  if (typeof navigator !== "undefined" && navigator.onLine === false) return push(null);
  const ctrl = new AbortController();
  const timeout = setTimeout(() => ctrl.abort(), PING_TIMEOUT_MS);
  const t0 = performance.now();
  try {
    const res = await fetch(`/api/ping?t=${Date.now()}`, { cache: "no-store", signal: ctrl.signal });
    if (!res.ok) throw new Error("bad status");
    push(performance.now() - t0);
  } catch {
    push(null);
  } finally {
    clearTimeout(timeout);
  }
}

export function subscribeSignal(listener: () => void): () => void {
  listeners.add(listener);
  if (listeners.size === 1) {
    probe();
    timer = setInterval(probe, PING_INTERVAL_MS);
  }
  return () => {
    listeners.delete(listener);
    if (listeners.size === 0 && timer) { clearInterval(timer); timer = null; }
  };
}

export const getSignal = (): SignalReading => reading;
export const getServerSignal = (): SignalReading => INITIAL_READING;
