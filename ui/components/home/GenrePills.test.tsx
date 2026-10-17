import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import GenrePills from "./GenrePills";

describe("GenrePills", () => {
  it("renders nothing without genres", () => {
    expect(render(<GenrePills genres={[]} chosen={[]} onToggle={() => {}} />).container).toBeEmptyDOMElement();
  });
  it("marks the chosen genres", () => {
    render(<GenrePills genres={["Drama", "Action"]} chosen={["Action"]} onToggle={() => {}} />);
    expect(screen.getByRole("button", { name: "Action" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "Drama" })).toHaveAttribute("aria-pressed", "false");
  });
  it("reports the genre that was tapped", () => {
    const onToggle = vi.fn();
    render(<GenrePills genres={["Drama"]} chosen={[]} onToggle={onToggle} />);
    fireEvent.click(screen.getByRole("button", { name: "Drama" }));
    expect(onToggle).toHaveBeenCalledWith("Drama");
  });
});
