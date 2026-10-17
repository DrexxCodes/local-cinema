import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import CompatBadge from "./CompatBadge";

describe("CompatBadge", () => {
  it("shows nothing for movies that play everywhere", () => {
    expect(render(<CompatBadge status="ok" />).container).toBeEmptyDOMElement();
    expect(render(<CompatBadge status="unknown" />).container).toBeEmptyDOMElement();
  });
  it("warns about movies that may not play", () => {
    render(<CompatBadge status="convert" />);
    expect(screen.getByText("May not play")).toBeInTheDocument();
  });
  it("says when a converted copy is used", () => {
    render(<CompatBadge status="converted" />);
    expect(screen.getByText("Converted")).toBeInTheDocument();
  });
  it("hidden beats the compatibility status", () => {
    render(<CompatBadge status="convert" hidden />);
    expect(screen.getByText("Hidden")).toBeInTheDocument();
    expect(screen.queryByText("May not play")).toBeNull();
  });
  it("a missing file beats everything", () => {
    render(<CompatBadge status="convert" hidden missing />);
    expect(screen.getByText("File missing")).toBeInTheDocument();
    expect(screen.queryByText("Hidden")).toBeNull();
  });
  it("shows a label even for a movie that plays everywhere when it is hidden", () => {
    render(<CompatBadge status="ok" hidden />);
    expect(screen.getByText("Hidden")).toBeInTheDocument();
  });
});
