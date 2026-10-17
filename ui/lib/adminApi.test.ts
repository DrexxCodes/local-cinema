import { afterEach, describe, expect, it, vi } from "vitest";
import { ADMIN_TOKEN_KEY } from "@/constants/storage";
import { adminApi, adminImg, adminUpload, clearAdminToken, getAdminToken, setAdminToken } from "./adminApi";

function reply(status: number, body: unknown) {
  const f = vi.fn().mockResolvedValue({ ok: status >= 200 && status < 300, status, json: async () => body });
  vi.stubGlobal("fetch", f);
  return f;
}
afterEach(() => vi.unstubAllGlobals());

describe("admin token", () => {
  it("is kept for the tab only (sessionStorage)", () => {
    expect(getAdminToken()).toBe("");
    setAdminToken("abc");
    expect(getAdminToken()).toBe("abc");
    expect(sessionStorage.getItem(ADMIN_TOKEN_KEY)).toBe("abc");
    expect(localStorage.getItem(ADMIN_TOKEN_KEY)).toBeNull();
    clearAdminToken();
    expect(getAdminToken()).toBe("");
  });
});

describe("adminApi", () => {
  it("sends the token and returns the body", async () => {
    setAdminToken("T");
    const f = reply(200, { ok: true });
    expect(await adminApi("/api/admin/stats")).toEqual({ ok: true });
    expect((f.mock.calls[0][1].headers as Headers).get("Authorization")).toBe("T");
  });
  it("sets JSON content type only when there is a body", async () => {
    let f = reply(200, {});
    await adminApi("/x");
    expect((f.mock.calls[0][1].headers as Headers).has("Content-Type")).toBe(false);
    f = reply(200, {});
    await adminApi("/x", { method: "PUT", body: "{}" });
    expect((f.mock.calls[0][1].headers as Headers).get("Content-Type")).toBe("application/json");
  });
  it("throws the server's message and forgets the token on 401", async () => {
    setAdminToken("T");
    reply(401, { message: "Sign in again." });
    await expect(adminApi("/x")).rejects.toThrow("Sign in again.");
    expect(getAdminToken()).toBe("");
  });
  it("keeps the token on other errors", async () => {
    setAdminToken("T");
    reply(500, null);
    await expect(adminApi("/x")).rejects.toThrow("Request failed.");
    expect(getAdminToken()).toBe("T");
  });
  it("explains a network failure", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new Error("down")));
    await expect(adminApi("/x")).rejects.toThrow("Can't reach the cinema server.");
  });
});

describe("adminImg", () => {
  it("builds the URL with version, cache buster and encoded token", () => {
    setAdminToken("a b");
    expect(adminImg("posters", "m1", 3, 9)).toBe("/api/admin-img/posters/m1?v=3&b=9&token=a%20b");
  });
});

describe("adminUpload", () => {
  const file = new File(["x"], "p.png", { type: "image/png" });
  it("sends the file as the raw body with its type", async () => {
    setAdminToken("T");
    const f = reply(200, { id: "m1" });
    expect(await adminUpload("/api/admin/movies/m1/art/poster", file)).toEqual({ id: "m1" });
    const init = f.mock.calls[0][1];
    expect(init.method).toBe("PUT");
    expect(init.body).toBe(file);
    expect(init.headers).toEqual({ Authorization: "T", "Content-Type": "image/png" });
  });
  it("shows the server's message when the upload is refused", async () => {
    reply(400, { message: "Use a JPEG, PNG or WebP image." });
    await expect(adminUpload("/x", file)).rejects.toThrow("Use a JPEG, PNG or WebP image.");
  });
});
