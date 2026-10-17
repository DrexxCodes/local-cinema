import { renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { HEARTBEAT_MS } from "@/constants/home";
import { SESSION_KEY } from "@/constants/storage";

const replace = vi.fn();
const router = { replace }; // Next gives back the same router object on every render, so the mock must too
vi.mock("next/navigation", () => ({ useRouter: () => router }));
vi.mock("@/lib/signal/monitor", () => ({
  subscribeSignal: vi.fn(() => vi.fn()),
  getSignal: vi.fn(() => ({ ready: true, level: 3, rtt: 42 })),
}));

import { getSignal, subscribeSignal } from "@/lib/signal/monitor";
import { useGuest } from "./useGuest";

const session = { token: "t", session_id: "s", user: { id: "u", room_number: "1", display_name: "R" } };
const heartbeats = () => vi.mocked(fetch).mock.calls.filter((c) => c[0] === "/api/auth/heartbeat");
const bodyOf = (i: number) => JSON.parse(String(heartbeats()[i][1]?.body));

beforeEach(() => {
  vi.useFakeTimers();
  replace.mockClear();
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) }));
});
afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("useGuest", () => {
  it("sends a signed out visitor to the login page", () => {
    const { result } = renderHook(() => useGuest());
    expect(result.current).toBeNull();
    expect(replace).toHaveBeenCalledWith("/login/");
    expect(fetch).not.toHaveBeenCalled();
  });

  it("returns the session and sends a heartbeat with the measured signal at once", () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    const { result } = renderHook(() => useGuest());
    expect(result.current).toEqual(session);
    expect(replace).not.toHaveBeenCalled();
    expect(heartbeats()).toHaveLength(1);
    expect(bodyOf(0)).toEqual({ rtt_ms: 42, signal: 3 });
  });

  it("repeats the heartbeat on a timer", async () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    renderHook(() => useGuest());
    await vi.advanceTimersByTimeAsync(HEARTBEAT_MS * 2);
    expect(heartbeats()).toHaveLength(3);
  });

  it("reports 0 before the first probe has finished", () => {
    vi.mocked(getSignal).mockReturnValueOnce({ ready: false, level: 0, rtt: null } as ReturnType<typeof getSignal>);
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    renderHook(() => useGuest());
    expect(bodyOf(0)).toEqual({ rtt_ms: 0, signal: 0 });
  });

  it("stops the heartbeat and the probe when the page closes", async () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    const stop = vi.fn();
    vi.mocked(subscribeSignal).mockReturnValueOnce(stop);
    const { unmount } = renderHook(() => useGuest());
    unmount();
    expect(stop).toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(HEARTBEAT_MS * 3);
    expect(heartbeats()).toHaveLength(1);
  });

  it("keeps going when a heartbeat fails", async () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    vi.mocked(fetch).mockRejectedValue(new TypeError("offline"));
    renderHook(() => useGuest());
    await vi.advanceTimersByTimeAsync(HEARTBEAT_MS);
    expect(heartbeats()).toHaveLength(2);
  });
});
