import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import SortBar from "./SortBar";

const base = { sort: "latest" as const, reversed: false, onSort: () => {}, minRating: 0, onRating: () => {}, showRating: true };

describe("SortBar", () => {
  it("marks the active sort and reports taps", () => {
    const onSort = vi.fn();
    render(<SortBar {...base} onSort={onSort} />);
    expect(screen.getByRole("button", { name: "Latest" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "A-Z" })).toHaveAttribute("aria-pressed", "false");
    fireEvent.click(screen.getByRole("button", { name: "A-Z" }));
    expect(onSort).toHaveBeenCalledWith("az");
    fireEvent.click(screen.getByRole("button", { name: "Latest" }));
    expect(onSort).toHaveBeenLastCalledWith("latest"); // tapping the active one lets the page flip the order
  });

  it("shows the order arrow only on the active sort, turned when reversed", () => {
    const { container, rerender } = render(<SortBar {...base} />);
    expect(container.querySelectorAll("button svg")).toHaveLength(1);
    expect(container.querySelector("button svg")!.getAttribute("class")).not.toContain("rotate-180");
    rerender(<SortBar {...base} reversed />);
    expect(container.querySelector("button svg")!.getAttribute("class")).toContain("rotate-180");
  });

  it("has a rating slider when asked, showing one decimal", () => {
    const onRating = vi.fn();
    render(<SortBar {...base} minRating={7} onRating={onRating} />);
    expect(screen.getByText("7.0")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Minimum rating"), { target: { value: "8.5" } });
    expect(onRating).toHaveBeenCalledWith(8.5);
  });

  it("hides the slider when not wanted", () => {
    render(<SortBar {...base} showRating={false} />);
    expect(screen.queryByLabelText("Minimum rating")).toBeNull();
  });
});
