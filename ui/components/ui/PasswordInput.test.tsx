import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import PasswordInput from "./PasswordInput";

describe("PasswordInput", () => {
  it("hides the text until asked", () => {
    const { container } = render(<PasswordInput defaultValue="secret" aria-label="Password" />);
    const input = container.querySelector("input")!;
    expect(input.type).toBe("password");
    fireEvent.click(screen.getByRole("button", { name: "Show password" }));
    expect(input.type).toBe("text");
    expect(screen.getByRole("button", { name: "Hide password" })).toHaveAttribute("aria-pressed", "true");
    fireEvent.click(screen.getByRole("button", { name: "Hide password" }));
    expect(input.type).toBe("password");
  });

  it("keeps the typed text when toggling", () => {
    const onChange = vi.fn();
    const { container } = render(<PasswordInput onChange={onChange} />);
    const input = container.querySelector("input")!;
    fireEvent.change(input, { target: { value: "abc" } });
    fireEvent.click(screen.getByRole("button"));
    expect(input.value).toBe("abc");
    expect(onChange).toHaveBeenCalled();
  });

  it("does not submit a form when the eye button is pressed", () => {
    const onSubmit = vi.fn((e) => e.preventDefault());
    render(<form onSubmit={onSubmit}><PasswordInput /></form>);
    fireEvent.click(screen.getByRole("button"));
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("moves top margins to the wrapper and leaves the rest on the field", () => {
    const { container } = render(<PasswordInput className="mt-2 w-full rounded" />);
    expect(container.firstElementChild!.className).toContain("mt-2");
    const input = container.querySelector("input")!;
    expect(input.className).toContain("w-full rounded");
    expect(input.className).not.toContain("mt-2");
    expect(input.className).toContain("pr-11");
  });
});
