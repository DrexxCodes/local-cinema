import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { PING_INTERVAL_MS, SAMPLE_WINDOW } from "@/constants/signal";

type Monitor = typeof import("./monitor");

// The monitor keeps its state in the module, so every test loads a fresh copy.
async function load(): Promise<Monitor> {
  vi.resetModules();
  return import("./monitor");
}

const flush = () => vi.advanceTimersByTimeAsync(0);

beforeEach(() => {
  vi.useFakeTimers();
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true }));
});
afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("signal monitor", () => {
  it("starts in the checking state and does not probe without a listener", async () => {
    const m = await load();
    expect(m.getSignal().ready).toBe(false);
    await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS * 3);
    expect(fetch).not.toHaveBeenCalled();
  });

  it("probes at once when the first listener arrives, then on a timer", async () => {
    const m = await load();
    const stop = m.subscribeSignal(() => {});
    await flush();
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(vi.mocked(fetch).mock.calls[0][0]).toMatch(/^\/api\/ping\?t=\d+$/);
    await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS);
    expect(fetch).toHaveBeenCalledTimes(2);
    stop();
  });

  it("tells listeners after each probe and reports a reading", async () => {
    const m = await load();
    const listener = vi.fn();
    const stop = m.subscribeSignal(listener);
    await flush();
    expect(listener).toHaveBeenCalled();
    expect(m.getSignal().ready).toBe(true);
    expect(m.getSignal().history).toHaveLength(1);
    stop();
  });

  it("shares one timer between listeners and stops when the last one leaves", async () => {
    const m = await load();
    const stopA = m.subscribeSignal(() => {});
    const stopB = m.subscribeSignal(() => {});
    await flush();
    expect(fetch).toHaveBeenCalledTimes(1); // the second listener did not start another probe
    stopA();
    await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS);
    expect(fetch).toHaveBeenCalledTimes(2); // still running for B
    stopB();
    await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS * 3);
    expect(fetch).toHaveBeenCalledTimes(2); // stopped
  });

  it("counts a failed request as a lost probe", async () => {
    vi.mocked(fetch).mockRejectedValue(new TypeError("offline"));
    const m = await load();
    const stop = m.subscribeSignal(() => {});
    await flush();
    expect(m.getSignal()).toMatchObject({ ready: true, level: 0, loss: 1 });
    stop();
  });

  it("counts a bad status as a lost probe", async () => {
    vi.mocked(fetch).mockResolvedValue({ ok: false } as Response);
    const m = await load();
    const stop = m.subscribeSignal(() => {});
    await flush();
    expect(m.getSignal().loss).toBe(1);
    stop();
  });

  it("does not probe while the tab is hidden", async () => {
    const m = await load();
    Object.defineProperty(document, "hidden", { value: true, configurable: true });
    const stop = m.subscribeSignal(() => {});
    await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS * 2);
    expect(fetch).not.toHaveBeenCalled();
    stop();
    Object.defineProperty(document, "hidden", { value: false, configurable: true });
  });

  it("records a lost probe without calling the server when the browser is offline", async () => {
    const m = await load();
    vi.spyOn(navigator, "onLine", "get").mockReturnValue(false);
    const stop = m.subscribeSignal(() => {});
    await flush();
    expect(fetch).not.toHaveBeenCalled();
    expect(m.getSignal().loss).toBe(1);
    stop();
  });

  it("keeps only the latest samples", async () => {
    const m = await load();
    const stop = m.subscribeSignal(() => {});
    await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS * (SAMPLE_WINDOW + 5));
    expect(m.getSignal().history).toHaveLength(SAMPLE_WINDOW);
    stop();
  });

  it("gives the server render the checking state", async () => {
    const m = await load();
    expect(m.getServerSignal().ready).toBe(false);
  });
});
