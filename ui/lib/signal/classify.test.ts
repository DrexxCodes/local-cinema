import { describe, expect, it } from "vitest";
import { CHECKING_LABEL, OFFLINE_LABEL } from "@/constants/signal";
import { INITIAL_READING, classify } from "./classify";

describe("classify", () => {
  it("is the checking state before any probe", () => {
    expect(classify([])).toBe(INITIAL_READING);
    expect(INITIAL_READING).toMatchObject({ ready: false, level: 0, label: CHECKING_LABEL });
  });

  it("is offline when every probe failed", () => {
    expect(classify([null, null])).toMatchObject({ ready: true, level: 0, label: OFFLINE_LABEL, rtt: null, loss: 1 });
  });

  it("is excellent for a fast steady link", () => {
    expect(classify([10, 12, 11, 10])).toMatchObject({ level: 4, label: "Excellent", rtt: 11, loss: 0 });
  });

  it("uses the median, so one slow probe does not drag the level down", () => {
    expect(classify([10, 12, 900, 11, 10]).rtt).toBe(11);
  });

  it("measures jitter as the mean change between probes", () => {
    // |30-10| + |10-30| = 40, over 2 steps
    expect(classify([10, 30, 10]).jitter).toBe(20);
    expect(classify([50]).jitter).toBe(0);
  });

  it("drops a level for any lost probe", () => {
    const r = classify([10, 10, 10, 10, 10, 10, 10, 10, 10, null]); // 10% loss
    expect(r.loss).toBeCloseTo(0.1);
    expect(r.level).toBe(3);
  });

  it("steps down with latency", () => {
    expect(classify([80, 80]).level).toBe(3);
    expect(classify([200, 200]).level).toBe(2);
    expect(classify([600, 600]).level).toBe(1);
  });

  it("falls back to Weak, level 1, when no limit fits but something got through", () => {
    expect(classify([2000, 2000])).toMatchObject({ level: 1, label: "Weak", rtt: 2000 });
    expect(classify([10, null, null, null])).toMatchObject({ level: 1, label: "Weak" }); // 75% loss
  });

  it("keeps the history it was given", () => {
    const h = [10, null, 12];
    expect(classify(h).history).toBe(h);
  });
});
