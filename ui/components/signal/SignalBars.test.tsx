import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { SignalLevel } from "@/types";
import SignalBars from "./SignalBars";

const bars = (c: HTMLElement) => Array.from(c.querySelectorAll("span > span")) as HTMLElement[];
const lit = (c: HTMLElement) => bars(c).filter((b) => !b.className.includes("bg-white/25"));

describe("SignalBars", () => {
  it("always draws four bars of rising height", () => {
    const { container } = render(<SignalBars level={2} />);
    expect(bars(container).map((b) => b.style.height)).toEqual(["25%", "50%", "75%", "100%"]);
  });

  it.each([[0, 0], [1, 1], [2, 2], [3, 3], [4, 4]] as [SignalLevel, number][])("level %i lights %i bars", (level, count) => {
    const { container } = render(<SignalBars level={level} />);
    expect(lit(container)).toHaveLength(count);
  });

  it("colours by quality", () => {
    expect(lit(render(<SignalBars level={4} />).container)[0].className).toContain("emerald");
    expect(lit(render(<SignalBars level={2} />).container)[0].className).toContain("amber");
    expect(lit(render(<SignalBars level={1} />).container)[0].className).toContain("red");
  });

  it("lights nothing before the first reading", () => {
    expect(lit(render(<SignalBars level={4} ready={false} />).container)).toHaveLength(0);
  });
});
