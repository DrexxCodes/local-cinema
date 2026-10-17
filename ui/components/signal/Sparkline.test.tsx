import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import Sparkline from "./Sparkline";

describe("Sparkline", () => {
  it("draws one line through the samples", () => {
    const { container } = render(<Sparkline values={[10, 20, 30]} />);
    const d = container.querySelector("path")!.getAttribute("d")!;
    expect(d.match(/M/g)).toHaveLength(1);
    expect(d.match(/L/g)).toHaveLength(2);
  });

  it("marks failed probes with red ticks and breaks the line there", () => {
    const { container } = render(<Sparkline values={[10, null, 30, 40]} />);
    expect(container.querySelectorAll("rect")).toHaveLength(1);
    const d = container.querySelector("path")!.getAttribute("d")!;
    expect(d.match(/M/g)).toHaveLength(2); // the pen lifted at the gap
  });

  it("draws no line when every probe failed", () => {
    const { container } = render(<Sparkline values={[null, null]} />);
    expect(container.querySelector("path")).toBeNull();
    expect(container.querySelectorAll("rect")).toHaveLength(2);
  });

  it("copes with a single sample and with none", () => {
    expect(render(<Sparkline values={[50]} />).container.querySelector("path")).not.toBeNull();
    const empty = render(<Sparkline values={[]} />).container;
    expect(empty.querySelector("svg")).not.toBeNull();
    expect(empty.querySelector("path")).toBeNull();
  });

  it("keeps the line inside the picture, scaling to the slowest probe", () => {
    const { container } = render(<Sparkline values={[10, 1000]} />);
    const ys = [...container.querySelector("path")!.getAttribute("d")!.matchAll(/ (\d+(?:\.\d+)?) /g)].map((m) => +m[1]);
    expect(Math.min(...ys)).toBeGreaterThanOrEqual(0);
    expect(Math.max(...ys)).toBeLessThanOrEqual(32);
  });
});
