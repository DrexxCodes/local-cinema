import type { SignalLevel } from "@/types";

export const PING_INTERVAL_MS = 4000;
export const PING_TIMEOUT_MS = 3000;
/** Probes kept for the median, jitter and loss figures. */
export const SAMPLE_WINDOW = 12;

/** A link gets the first level whose limits it stays within. */
export const LEVEL_LIMITS: { level: SignalLevel; label: string; rtt: number; jitter: number; loss: number }[] = [
  { level: 4, label: "Excellent", rtt: 40, jitter: 25, loss: 0 },
  { level: 3, label: "Good", rtt: 100, jitter: 60, loss: 0.1 },
  { level: 2, label: "Fair", rtt: 250, jitter: 150, loss: 0.25 },
  { level: 1, label: "Weak", rtt: 800, jitter: 500, loss: 0.5 },
];

export const OFFLINE_LABEL = "No connection";
export const CHECKING_LABEL = "Checking";
