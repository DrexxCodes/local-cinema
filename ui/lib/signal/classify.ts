import { CHECKING_LABEL, LEVEL_LIMITS, OFFLINE_LABEL } from "@/constants/signal";
import type { SignalReading } from "@/types";

export const INITIAL_READING: SignalReading = {
  ready: false, level: 0, label: CHECKING_LABEL, rtt: null, jitter: 0, loss: 0, history: [],
};

const median = (xs: number[]) => {
  const s = [...xs].sort((a, b) => a - b), mid = Math.floor(s.length / 2);
  return s.length % 2 ? s[mid] : (s[mid - 1] + s[mid]) / 2;
};

/** Turns the recent probe results (ms, or null for a failure) into a signal reading. */
export function classify(history: (number | null)[]): SignalReading {
  if (!history.length) return INITIAL_READING;
  const ok = history.filter((x): x is number => x !== null);
  const loss = (history.length - ok.length) / history.length;
  if (!ok.length) {
    return { ready: true, level: 0, label: OFFLINE_LABEL, rtt: null, jitter: 0, loss: 1, history };
  }
  const rtt = median(ok);
  const jitter = ok.length > 1 ? ok.slice(1).reduce((sum, v, i) => sum + Math.abs(v - ok[i]), 0) / (ok.length - 1) : 0;
  const hit = LEVEL_LIMITS.find((l) => rtt <= l.rtt && jitter <= l.jitter && loss <= l.loss);
  return {
    ready: true, level: hit?.level ?? 1, label: hit?.label ?? "Weak",
    rtt: Math.round(rtt), jitter: Math.round(jitter), loss, history,
  };
}
