import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { PING_INTERVAL_MS } from "@/constants/signal";

afterEach(() => { vi.useRealTimers(); vi.unstubAllGlobals(); });

describe("useSignal", () => {
  it("follows the shared monitor", async () => {
    vi.useFakeTimers();
    vi.resetModules();
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true }));
    const { useSignal } = await import("./useSignal");
    const { result, unmount } = renderHook(() => useSignal());
    expect(result.current.ready).toBe(false);
    await act(async () => { await vi.advanceTimersByTimeAsync(0); });
    expect(result.current.ready).toBe(true);
    expect(result.current.history).toHaveLength(1);
    await act(async () => { await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS); });
    expect(result.current.history).toHaveLength(2);
    unmount();
    const calls = vi.mocked(fetch).mock.calls.length;
    await vi.advanceTimersByTimeAsync(PING_INTERVAL_MS * 3);
    expect(vi.mocked(fetch).mock.calls.length).toBe(calls); // no listener, no probes
  });
});
