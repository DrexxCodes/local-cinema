import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MY_LIST_PREFIX } from "@/constants/storage";
import { useMyList } from "./useMyList";

describe("useMyList", () => {
  it("starts empty", () => {
    const { result } = renderHook(() => useMyList("u1"));
    expect(result.current.ids).toEqual([]);
    expect(result.current.has("a")).toBe(false);
  });

  it("adds new movies at the front and removes them again", () => {
    const { result } = renderHook(() => useMyList("u1"));
    act(() => result.current.toggle("a"));
    act(() => result.current.toggle("b"));
    expect(result.current.ids).toEqual(["b", "a"]);
    expect(result.current.has("a")).toBe(true);
    act(() => result.current.toggle("a"));
    expect(result.current.ids).toEqual(["b"]);
  });

  it("keeps the list on this device, per room", () => {
    const { result } = renderHook(() => useMyList("u1"));
    act(() => result.current.toggle("a"));
    expect(JSON.parse(localStorage.getItem(MY_LIST_PREFIX + "u1")!)).toEqual(["a"]);
    expect(localStorage.getItem(MY_LIST_PREFIX + "u2")).toBeNull();
  });

  it("loads what was saved before", () => {
    localStorage.setItem(MY_LIST_PREFIX + "u1", JSON.stringify(["x", "y"]));
    const { result } = renderHook(() => useMyList("u1"));
    expect(result.current.ids).toEqual(["x", "y"]);
  });

  it("ignores damaged storage", () => {
    localStorage.setItem(MY_LIST_PREFIX + "u1", "{oops");
    const { result } = renderHook(() => useMyList("u1"));
    expect(result.current.ids).toEqual([]);
  });

  it("switches list when the room changes", () => {
    localStorage.setItem(MY_LIST_PREFIX + "u1", JSON.stringify(["one"]));
    localStorage.setItem(MY_LIST_PREFIX + "u2", JSON.stringify(["two"]));
    const { result, rerender } = renderHook(({ id }) => useMyList(id), { initialProps: { id: "u1" } });
    expect(result.current.ids).toEqual(["one"]);
    rerender({ id: "u2" });
    expect(result.current.ids).toEqual(["two"]);
  });

  it("does not save before a room is known", () => {
    const { result } = renderHook(() => useMyList(undefined));
    act(() => result.current.toggle("a"));
    expect(result.current.ids).toEqual(["a"]);
    expect(localStorage.length).toBe(0);
  });
});
