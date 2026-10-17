import { describe, expect, it } from "vitest";
import { formatClock, formatDuration, formatSize, parseClock, parseDate } from "./format";

describe("formatDuration", () => {
  it("shows hours and minutes, or minutes alone", () => {
    expect(formatDuration(5400)).toBe("1h 30m");
    expect(formatDuration(3600)).toBe("1h 0m");
    expect(formatDuration(45 * 60)).toBe("45m");
  });
  it("is empty when the length is unknown", () => {
    expect(formatDuration(0)).toBe("");
  });
});

describe("formatClock", () => {
  it("pads seconds, and minutes once there are hours", () => {
    expect(formatClock(0)).toBe("0:00");
    expect(formatClock(65)).toBe("1:05");
    expect(formatClock(3725)).toBe("1:02:05");
  });
  it("ignores fractions of a second", () => {
    expect(formatClock(59.9)).toBe("0:59");
  });
});

describe("formatSize", () => {
  it("uses MB below 1 GB and GB above", () => {
    expect(formatSize(750e6)).toBe("750 MB");
    expect(formatSize(2.5e9)).toBe("2.5 GB");
  });
});

describe("parseClock", () => {
  it("reads seconds, m:ss and h:mm:ss", () => {
    expect(parseClock("90")).toBe(90);
    expect(parseClock("1:30")).toBe(90);
    expect(parseClock("1:02:03")).toBe(3723);
    expect(parseClock("  2:00 ")).toBe(120);
  });
  it("gives 0 for text that is not a time", () => {
    expect(parseClock("")).toBe(0);
    expect(parseClock("abc")).toBe(0);
    expect(parseClock("1:xx")).toBe(0);
    expect(parseClock("-5")).toBe(0);
  });
  it("round trips with formatClock", () => {
    expect(parseClock(formatClock(3725))).toBe(3725);
  });
});

describe("parseDate", () => {
  it("reads the database date format", () => {
    expect(parseDate("2026-01-02 03:04:05.000Z")).toBe(Date.parse("2026-01-02T03:04:05.000Z"));
  });
});
