import { afterEach, describe, expect, it, vi } from "vitest";
import { copyText } from "./clipboard";

afterEach(() => vi.unstubAllGlobals());

describe("copyText", () => {
  it("uses the clipboard API on a secure page", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("navigator", { clipboard: { writeText } });
    vi.stubGlobal("isSecureContext", true);
    expect(await copyText("hello")).toBe(true);
    expect(writeText).toHaveBeenCalledWith("hello");
  });

  it("falls back to the selection method on plain http", async () => {
    vi.stubGlobal("navigator", {});
    vi.stubGlobal("isSecureContext", false);
    const exec = vi.fn().mockReturnValue(true);
    document.execCommand = exec;
    expect(await copyText("http://192.168.1.5:8765/stream/x")).toBe(true);
    expect(exec).toHaveBeenCalledWith("copy");
    expect(document.querySelector("textarea")).toBeNull(); // the helper field is removed again
  });

  it("falls back when the clipboard API refuses", async () => {
    vi.stubGlobal("navigator", { clipboard: { writeText: vi.fn().mockRejectedValue(new Error("denied")) } });
    vi.stubGlobal("isSecureContext", true);
    document.execCommand = vi.fn().mockReturnValue(true);
    expect(await copyText("x")).toBe(true);
  });

  it("reports failure when nothing works", async () => {
    vi.stubGlobal("navigator", {});
    vi.stubGlobal("isSecureContext", false);
    document.execCommand = vi.fn().mockReturnValue(false);
    expect(await copyText("x")).toBe(false);
    document.execCommand = vi.fn().mockImplementation(() => { throw new Error("blocked"); });
    expect(await copyText("x")).toBe(false);
  });
});
