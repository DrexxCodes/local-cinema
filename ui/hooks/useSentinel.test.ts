import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useSentinel } from "./useSentinel";

type Callback = (entries: { isIntersecting: boolean }[]) => void;
let observers: { cb: Callback; observe: ReturnType<typeof vi.fn>; disconnect: ReturnType<typeof vi.fn>; options?: IntersectionObserverInit }[];

beforeEach(() => {
  observers = [];
  vi.stubGlobal("IntersectionObserver", class {
    observe = vi.fn();
    disconnect = vi.fn();
    constructor(public cb: Callback, public options?: IntersectionObserverInit) { observers.push(this); }
  });
});
afterEach(() => vi.unstubAllGlobals());

const el = () => document.createElement("div");

describe("useSentinel", () => {
  it("does nothing until it has an element", () => {
    renderHook(() => useSentinel(() => {}, true, 0));
    expect(observers).toHaveLength(0);
  });

  it("does nothing while disabled", () => {
    const { result } = renderHook(() => useSentinel(() => {}, false, 0));
    act(() => result.current(el()));
    expect(observers).toHaveLength(0);
  });

  it("calls back when the element comes near the screen, with the margin it was given", () => {
    const onVisible = vi.fn();
    const { result } = renderHook(() => useSentinel(onVisible, true, 0, "100px"));
    const node = el();
    act(() => result.current(node));
    expect(observers[0].observe).toHaveBeenCalledWith(node);
    expect(observers[0].options).toEqual({ rootMargin: "100px" });
    observers[0].cb([{ isIntersecting: false }]);
    expect(onVisible).not.toHaveBeenCalled();
    observers[0].cb([{ isIntersecting: true }]);
    expect(onVisible).toHaveBeenCalledTimes(1);
  });

  it("uses the latest callback without watching again", () => {
    const first = vi.fn(), second = vi.fn();
    const { result, rerender } = renderHook(({ cb }) => useSentinel(cb, true, 0), { initialProps: { cb: first } });
    act(() => result.current(el()));
    rerender({ cb: second });
    expect(observers).toHaveLength(1);
    observers[0].cb([{ isIntersecting: true }]);
    expect(second).toHaveBeenCalled();
    expect(first).not.toHaveBeenCalled();
  });

  it("looks again when the watched value changes, so a still visible element loads more", () => {
    const { result, rerender } = renderHook(({ w }) => useSentinel(() => {}, true, w), { initialProps: { w: 1 } });
    act(() => result.current(el()));
    rerender({ w: 2 });
    expect(observers).toHaveLength(2);
    expect(observers[0].disconnect).toHaveBeenCalled();
  });

  it("stops watching when it unmounts", () => {
    const { result, unmount } = renderHook(() => useSentinel(() => {}, true, 0));
    act(() => result.current(el()));
    unmount();
    expect(observers[0].disconnect).toHaveBeenCalled();
  });
});
