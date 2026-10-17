import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import SeekBar from "./SeekBar";

const vars = (el: HTMLElement) => ({ p: el.style.getPropertyValue("--p"), b: el.style.getPropertyValue("--b") });

describe("SeekBar", () => {
  it("shows the position as a share of the length", () => {
    render(<SeekBar value={25} max={100} label="Seek" onChange={() => {}} />);
    expect(vars(screen.getByLabelText("Seek")).p).toBe("25%");
  });

  it("shows the loaded part, never less than the position", () => {
    render(<SeekBar value={25} max={100} buffered={60} label="Seek" onChange={() => {}} />);
    expect(vars(screen.getByLabelText("Seek")).b).toBe("60%");
    render(<SeekBar value={50} max={100} buffered={10} label="Seek2" onChange={() => {}} />);
    expect(vars(screen.getByLabelText("Seek2")).b).toBe("50%");
  });

  it("copes with an unknown length and a position past the end", () => {
    render(<SeekBar value={5} max={0} label="A" onChange={() => {}} />);
    expect(screen.getByLabelText("A")).toHaveAttribute("max", "1");
    render(<SeekBar value={500} max={100} label="B" onChange={() => {}} />);
    expect(vars(screen.getByLabelText("B")).p).toBe("100%");
  });

  it("reports a number when dragged", () => {
    const onChange = vi.fn();
    render(<SeekBar value={0} max={100} label="Seek" onChange={onChange} />);
    fireEvent.change(screen.getByLabelText("Seek"), { target: { value: "42" } });
    expect(onChange).toHaveBeenCalledWith(42);
  });

  it("passes the spoken value and the tour marker through", () => {
    render(<SeekBar value={1} max={10} label="Seek" valueText="1 minute" tour="seek" onChange={() => {}} />);
    const el = screen.getByLabelText("Seek");
    expect(el).toHaveAttribute("aria-valuetext", "1 minute");
    expect(el).toHaveAttribute("data-tour", "seek");
  });
});
