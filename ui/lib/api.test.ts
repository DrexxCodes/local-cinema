import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SESSION_KEY } from "@/constants/storage";
import { ApiError, api, withAuth } from "./api";

const session = { token: "tok", session_id: "sid", user: { id: "u", room_number: "1", display_name: "R" } };
const signIn = () => localStorage.setItem(SESSION_KEY, JSON.stringify(session));

function reply(status: number, body: unknown) {
  const fetchMock = vi.fn().mockResolvedValue({ ok: status >= 200 && status < 300, status, json: async () => body });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}
const sentHeaders = (f: ReturnType<typeof vi.fn>) => f.mock.calls[0][1].headers as Headers;

afterEach(() => vi.unstubAllGlobals());

describe("api", () => {
  it("returns the JSON body", async () => {
    reply(200, [{ id: "a" }]);
    expect(await api("/api/movies")).toEqual([{ id: "a" }]);
  });

  it("sends the sign in headers when signed in", async () => {
    signIn();
    const f = reply(200, {});
    await api("/api/movies");
    expect(sentHeaders(f).get("Authorization")).toBe("Bearer tok");
    expect(sentHeaders(f).get("X-Session-Id")).toBe("sid");
  });

  it("sends none when signed out", async () => {
    const f = reply(200, {});
    await api("/api/movies");
    expect(sentHeaders(f).has("Authorization")).toBe(false);
  });

  it("sets JSON content type for a body, and keeps one that was given", async () => {
    let f = reply(200, {});
    await api("/x", { method: "POST", body: "{}" });
    expect(sentHeaders(f).get("Content-Type")).toBe("application/json");
    f = reply(200, {});
    await api("/x", { method: "POST", body: "raw", headers: { "Content-Type": "text/plain" } });
    expect(sentHeaders(f).get("Content-Type")).toBe("text/plain");
  });

  it("throws an ApiError with the server's code and message", async () => {
    reply(429, { error: "too_many", message: "Slow down." });
    await expect(api("/api/auth/login")).rejects.toMatchObject({ status: 429, code: "too_many", message: "Slow down." });
    await expect(api("/api/auth/login")).rejects.toBeInstanceOf(ApiError);
  });

  it("uses a generic message when the error has no body", async () => {
    const f = vi.fn().mockResolvedValue({ ok: false, status: 500, json: async () => { throw new Error("no json"); } });
    vi.stubGlobal("fetch", f);
    await expect(api("/x")).rejects.toMatchObject({ status: 500, code: "error", message: "Something went wrong." });
  });

  it("explains a network failure", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("failed")));
    await expect(api("/x")).rejects.toMatchObject({ status: 0, code: "network" });
  });

  it("ends the session on 401 and sends the guest to the login page", async () => {
    signIn();
    const hrefSet = vi.fn();
    vi.stubGlobal("location", { set href(v: string) { hrefSet(v); } });
    reply(401, { error: "unauthorized", message: "Signed out." });
    await expect(api("/api/movies")).rejects.toBeInstanceOf(ApiError);
    expect(localStorage.getItem(SESSION_KEY)).toBeNull();
    expect(hrefSet).toHaveBeenCalledWith("/login/?ended=1");
  });

  it("keeps the session when the login itself answers 401", async () => {
    signIn();
    reply(401, { error: "bad_login", message: "Wrong password." });
    await expect(api("/api/auth/login", { method: "POST" })).rejects.toMatchObject({ code: "bad_login" });
    expect(localStorage.getItem(SESSION_KEY)).not.toBeNull();
  });
});

describe("withAuth", () => {
  beforeEach(() => localStorage.clear());
  it("leaves the path alone when signed out", () => {
    expect(withAuth("/stream/a")).toBe("/stream/a");
  });
  it("adds the credentials with ? or & as needed, and encodes them", () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify({ ...session, token: "a b&c" }));
    expect(withAuth("/stream/a")).toBe("/stream/a?token=a%20b%26c&sid=sid");
    expect(withAuth("/img?v=2")).toBe("/img?v=2&token=a%20b%26c&sid=sid");
  });
});
