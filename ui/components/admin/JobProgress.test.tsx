import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { AdminStats, ScanState } from "@/types";
import JobProgress from "./JobProgress";

const idle: ScanState = { running: false, total: 0, done: 0, current: "", percent: 0 };
function stats(over: Partial<AdminStats> = {}): AdminStats {
  return {
    active_sessions: 0, movies: 0, uptime_seconds: 0, scan: idle, thumbs: idle, convert: idle, convert_error: "",
    scan_issues: [], subs: idle, subs_msg: "", pending_upgrades: 0, update_available: "", ffmpeg_install: idle, ffmpeg: true, ...over,
  };
}
const ui = (s: AdminStats | null, cancel = () => {}) => render(<JobProgress stats={s} onCancelConvert={cancel} />);

describe("JobProgress", () => {
  it("renders nothing before the first stats arrive", () => {
    expect(ui(null).container).toBeEmptyDOMElement();
  });

  it("shows nothing while every job is idle", () => {
    expect(ui(stats()).container.textContent).toBe("");
  });

  it("shows the scan and thumbnail progress", () => {
    ui(stats({
      scan: { ...idle, running: true, total: 10, done: 4, current: "Heat.mkv" },
      thumbs: { ...idle, running: true, total: 3, done: 1, current: "Alien.mp4" },
    }));
    expect(screen.getByText(/Scanning 4 of 10/)).toBeInTheDocument();
    expect(screen.getByText("Heat.mkv")).toBeInTheDocument();
    expect(screen.getByText(/Thumbnails 1 of 3/)).toBeInTheDocument();
  });

  it("shows conversion with its percent, a bar, and a working cancel button", () => {
    const cancel = vi.fn();
    const { container } = ui(stats({ convert: { running: true, total: 5, done: 1, current: "Big.mkv", percent: 40 } }), cancel);
    expect(screen.getByText(/Converting 2 of 5/)).toBeInTheDocument();
    expect(screen.getByText(/40%/)).toBeInTheDocument();
    expect((container.querySelector("[style*='width']") as HTMLElement).style.width).toBe("40%");
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(cancel).toHaveBeenCalledTimes(1);
  });

  it("never counts past the last movie", () => {
    ui(stats({ convert: { running: true, total: 2, done: 2, current: "x", percent: 0 } }));
    expect(screen.getByText(/Converting 2 of 2/)).toBeInTheDocument();
  });

  it("lists files the last scan skipped, with the right plural", () => {
    ui(stats({ scan_issues: ["a.mkv: unreadable"] }));
    expect(screen.getByText("The last scan skipped 1 file")).toBeInTheDocument();
    expect(screen.getByText("a.mkv: unreadable")).toBeInTheDocument();
  });
  it("pluralises for several files", () => {
    ui(stats({ scan_issues: ["a", "b"] }));
    expect(screen.getByText("The last scan skipped 2 files")).toBeInTheDocument();
  });
  it("hides the skipped files while a scan is running", () => {
    ui(stats({ scan_issues: ["a"], scan: { ...idle, running: true, total: 1, done: 0, current: "" } }));
    expect(screen.queryByText(/skipped/)).toBeNull();
  });

  it("reports a conversion problem, but not while converting", () => {
    ui(stats({ convert_error: "ffmpeg exited 1" }));
    expect(screen.getByRole("alert")).toHaveTextContent("Conversion problem: ffmpeg exited 1");
  });
  it("does not show an old conversion problem during a new conversion", () => {
    ui(stats({ convert_error: "old", convert: { running: true, total: 1, done: 0, current: "x", percent: 0 } }));
    expect(screen.queryByRole("alert")).toBeNull();
  });
});
