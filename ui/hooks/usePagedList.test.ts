import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { Page } from "@/types";
import { usePagedList } from "./usePagedList";

type Item = { id: string };
const items = (...ids: string[]): Item[] => ids.map((id) => ({ id }));
const page = (ids: string[], total: number, has_more: boolean): Page<Item> => ({ items: items(...ids), total, has_more });

describe("usePagedList", () => {
  it("loads the first page on mount", async () => {
    const fetchPage = vi.fn().mockResolvedValue(page(["a", "b"], 4, true));
    const { result } = renderHook(() => usePagedList(fetchPage, ""));
    await waitFor(() => expect(result.current.items).toHaveLength(2));
    expect(fetchPage).toHaveBeenCalledWith(0);
    expect(result.current).toMatchObject({ total: 4, hasMore: true, loading: false, error: "" });
  });

  it("asks for the next page from the number of items it already has", async () => {
    const fetchPage = vi.fn()
      .mockResolvedValueOnce(page(["a", "b"], 3, true))
      .mockResolvedValueOnce(page(["c"], 3, false));
    const { result } = renderHook(() => usePagedList(fetchPage, ""));
    await waitFor(() => expect(result.current.items).toHaveLength(2));
    await act(() => result.current.loadMore());
    expect(fetchPage).toHaveBeenLastCalledWith(2);
    expect(result.current.items.map((i) => i.id)).toEqual(["a", "b", "c"]);
    expect(result.current.hasMore).toBe(false);
  });

  it("skips items it already has", async () => {
    const fetchPage = vi.fn()
      .mockResolvedValueOnce(page(["a", "b"], 3, true))
      .mockResolvedValueOnce(page(["b", "c"], 3, false));
    const { result } = renderHook(() => usePagedList(fetchPage, ""));
    await waitFor(() => expect(result.current.items).toHaveLength(2));
    await act(() => result.current.loadMore());
    expect(result.current.items.map((i) => i.id)).toEqual(["a", "b", "c"]);
  });

  it("stops when a page comes back empty, even if the server says there is more", async () => {
    const fetchPage = vi.fn().mockResolvedValue(page([], 5, true));
    const { result } = renderHook(() => usePagedList(fetchPage, ""));
    await waitFor(() => expect(fetchPage).toHaveBeenCalled());
    await waitFor(() => expect(result.current.hasMore).toBe(false));
  });

  it("does not run two loads at once", async () => {
    let release!: (p: Page<Item>) => void;
    const fetchPage = vi.fn().mockImplementation(() => new Promise<Page<Item>>((r) => { release = r; }));
    const { result } = renderHook(() => usePagedList(fetchPage, ""));
    await waitFor(() => expect(fetchPage).toHaveBeenCalledTimes(1));
    act(() => { result.current.loadMore(); result.current.loadMore(); });
    expect(fetchPage).toHaveBeenCalledTimes(1);
    await act(async () => release(page(["a"], 1, false)));
    await waitFor(() => expect(result.current.loading).toBe(false));
  });

  it("shows the error and stops loading", async () => {
    const fetchPage = vi.fn().mockRejectedValue(new Error("Could not list movies."));
    const { result } = renderHook(() => usePagedList(fetchPage, ""));
    await waitFor(() => expect(result.current.error).toBe("Could not list movies."));
    expect(result.current.hasMore).toBe(false);
    expect(result.current.loading).toBe(false);
  });

  it("starts again from the first page when the key changes", async () => {
    const fetchPage = vi.fn().mockImplementation(async (offset: number) => page([`r${offset}-${fetchPage.mock.calls.length}`], 1, false));
    const { result, rerender } = renderHook(({ k }) => usePagedList(fetchPage, k), { initialProps: { k: "a" } });
    await waitFor(() => expect(result.current.items).toHaveLength(1));
    const first = result.current.items[0].id;
    rerender({ k: "b" });
    await waitFor(() => expect(result.current.items[0]?.id).not.toBe(first));
    expect(result.current.items).toHaveLength(1);
    expect(fetchPage).toHaveBeenLastCalledWith(0);
  });

  it("ignores an answer that arrives after the key changed", async () => {
    const resolvers: ((p: Page<Item>) => void)[] = [];
    const fetchPage = vi.fn().mockImplementation(() => new Promise<Page<Item>>((r) => resolvers.push(r)));
    const { result, rerender } = renderHook(({ k }) => usePagedList(fetchPage, k), { initialProps: { k: "old" } });
    await waitFor(() => expect(resolvers).toHaveLength(1));
    rerender({ k: "new" });
    await waitFor(() => expect(resolvers).toHaveLength(2));
    await act(async () => resolvers[1](page(["new1"], 1, false)));
    await act(async () => resolvers[0](page(["old1"], 1, false)));
    expect(result.current.items.map((i) => i.id)).toEqual(["new1"]);
  });

  it("reload refreshes from the first page", async () => {
    const fetchPage = vi.fn().mockResolvedValue(page(["a"], 1, false));
    const { result } = renderHook(() => usePagedList(fetchPage, ""));
    await waitFor(() => expect(result.current.items).toHaveLength(1));
    await act(async () => result.current.reload());
    await waitFor(() => expect(fetchPage).toHaveBeenCalledTimes(2));
    expect(fetchPage).toHaveBeenLastCalledWith(0);
    expect(result.current.items).toHaveLength(1);
  });
});
