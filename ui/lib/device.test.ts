import { afterEach, describe, expect, it, vi } from "vitest";
import { DEVICE_KEY } from "@/constants/storage";
import { deviceId, deviceName, isIOS } from "./device";

function fakeNavigator(ua: string, platform = "", maxTouchPoints = 0) {
  vi.stubGlobal("navigator", { userAgent: ua, platform, maxTouchPoints });
}
afterEach(() => vi.unstubAllGlobals());

describe("deviceId", () => {
  it("makes a version 4 UUID and keeps it", () => {
    const id = deviceId();
    expect(id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
    expect(deviceId()).toBe(id);
    expect(localStorage.getItem(DEVICE_KEY)).toBe(id);
  });
  it("reuses an id that is already stored", () => {
    localStorage.setItem(DEVICE_KEY, "existing");
    expect(deviceId()).toBe("existing");
  });
  it("makes a different id for each browser", () => {
    const a = deviceId();
    localStorage.clear();
    expect(deviceId()).not.toBe(a);
  });
});

describe("deviceName", () => {
  const cases: [string, string][] = [
    ["Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)", "iPhone"],
    ["Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X)", "iPad"],
    ["Mozilla/5.0 (Linux; Android 14; Pixel 8)", "Android"],
    ["Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", "Mac"],
    ["Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Windows PC"],
    ["Mozilla/5.0 (X11; Linux x86_64)", "Linux PC"],
    ["SomethingElse/1.0", "Browser"],
  ];
  it.each(cases)("%s -> %s", (ua, name) => {
    fakeNavigator(ua);
    expect(deviceName()).toBe(name);
  });
});

describe("isIOS", () => {
  it("is true for iPhone and iPad user agents", () => {
    fakeNavigator("Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)");
    expect(isIOS()).toBe(true);
  });
  it("is true for an iPad that reports itself as a Mac", () => {
    fakeNavigator("Mozilla/5.0 (Macintosh)", "MacIntel", 5);
    expect(isIOS()).toBe(true);
  });
  it("is false for a real Mac and for Android", () => {
    fakeNavigator("Mozilla/5.0 (Macintosh)", "MacIntel", 0);
    expect(isIOS()).toBe(false);
    fakeNavigator("Mozilla/5.0 (Linux; Android 14)", "Linux armv8l", 5);
    expect(isIOS()).toBe(false);
  });
});
