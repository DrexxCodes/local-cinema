import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { Session } from "@/types";
import { useCatalog } from "./useCatalog";

const session: Session = { token: "t", session_id: "s", user: { id: "u", room_number: "1", display_name: "R" } };

function routes(map: Record<string, unknown | Error>) {
  vi.stubGlobal("fetch", vi.fn(async (path: string) => {
    const r = map[path.split("?")[0]];
    if (r instanceof Error) throw r;
    return { ok: true, status: 200, json: async () => r };
  }));
}
afterEach(() => vi.unstubAllGlobals());

describe("useCatalog", () => {
  it("does not load without a session", () => {
    routes({});
    const { result } = renderHook(() => useCatalog(null));
    expect(result.current.movies).toBeNull();
    expect(fetch).not.toHaveBeenCalled();
  });

  it("loads movies and progress, and indexes progress by movie", async () => {
    routes({
      "/api/movies": [{ id: "m1" }, { id: "m2" }],
      "/api/playback": [{ movie: "m2", position_seconds: 30, completed: false }],
    });
    const { result } = renderHook(() => useCatalog(session));
    expect(result.current.movies).toBeNull(); // still loading
    await waitFor(() => expect(result.current.movies).toHaveLength(2));
    await waitFor(() => expect(result.current.progress).toHaveLength(1));
    expect(result.current.progressById.get("m2")?.position_seconds).toBe(30);
    expect(result.current.progressById.has("m1")).toBe(false);
  });

  it("shows an empty catalog when the list cannot be loaded", async () => {
    routes({ "/api/movies": new TypeError("offline"), "/api/playback": [] });
    const { result } = renderHook(() => useCatalog(session));
    await waitFor(() => expect(result.current.movies).toEqual([]));
  });

  it("still shows movies when progress cannot be loaded", async () => {
    routes({ "/api/movies": [{ id: "m1" }], "/api/playback": new TypeError("offline") });
    const { result } = renderHook(() => useCatalog(session));
    await waitFor(() => expect(result.current.movies).toHaveLength(1));
    expect(result.current.progress).toEqual([]);
  });
});
