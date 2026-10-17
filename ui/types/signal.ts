export type SignalLevel = 0 | 1 | 2 | 3 | 4;

export type SignalReading = {
  /** False until the first probe finishes. */
  ready: boolean;
  /** 0 (offline) to 4 (excellent). */
  level: SignalLevel;
  label: string;
  /** Median round trip in ms over the sample window, null when every probe failed. */
  rtt: number | null;
  /** Mean change between consecutive round trips, in ms. */
  jitter: number;
  /** Share of failed probes, 0 to 1. */
  loss: number;
  /** Recent round trips in ms, null marks a failed probe. */
  history: (number | null)[];
};
