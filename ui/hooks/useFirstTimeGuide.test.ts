import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { FIRST_TIME_KEY } from "@/constants/tour";
import { useFirstTimeGuide } from "./useFirstTimeGuide";

describe("useFirstTimeGuide", () => {
  it("shows nothing until a movie starts", () => {
    const { result } = renderHook(() => useFirstTimeGuide(false));
    expect(result.current.show).toBe(false);
    expect(localStorage.getItem(FIRST_TIME_KEY)).toBeNull();
  });

  it("shows the guide on the first movie and remembers it was created", () => {
    const { result } = renderHook(() => useFirstTimeGuide(true));
    expect(result.current.show).toBe(true);
    expect(localStorage.getItem(FIRST_TIME_KEY)).toBe("true");
  });

  it("finish hides it for good", () => {
    const { result, unmount } = renderHook(() => useFirstTimeGuide(true));
    act(() => result.current.finish());
    expect(result.current.show).toBe(false);
    expect(localStorage.getItem(FIRST_TIME_KEY)).toBe("false");
    unmount();
    expect(renderHook(() => useFirstTimeGuide(true)).result.current.show).toBe(false);
  });

  it("starts when the movie starts, not before", () => {
    const { result, rerender } = renderHook(({ started }) => useFirstTimeGuide(started), { initialProps: { started: false } });
    rerender({ started: true });
    expect(result.current.show).toBe(true);
  });

  it("checks only once per page", () => {
    const { result, rerender } = renderHook(({ started }) => useFirstTimeGuide(started), { initialProps: { started: true } });
    act(() => result.current.finish());
    rerender({ started: false });
    rerender({ started: true });
    expect(result.current.show).toBe(false);
  });
});
