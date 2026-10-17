import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { setAdminToken } from "@/lib/adminApi";
import FolderList from "./FolderList";

function validate(body: unknown, status = 200) {
  const f = vi.fn().mockResolvedValue({ ok: status === 200, status, json: async () => body });
  vi.stubGlobal("fetch", f);
  return f;
}
afterEach(() => vi.unstubAllGlobals());

const field = () => screen.getByLabelText("Folder path") as HTMLInputElement;
const type = (v: string) => fireEvent.change(field(), { target: { value: v } });

describe("FolderList", () => {
  it("lists folders and removes one", () => {
    const onChange = vi.fn();
    render(<FolderList folders={["/a", "/b"]} onChange={onChange} />);
    fireEvent.click(screen.getAllByRole("button", { name: "Remove" })[0]);
    expect(onChange).toHaveBeenCalledWith(["/b"]);
  });

  it("checks a folder on the host before adding it", async () => {
    setAdminToken("T");
    const f = validate({ valid: true, video_count: 12, total_bytes: 3.5e9, warnings: [] });
    const onChange = vi.fn();
    render(<FolderList folders={["/a"]} onChange={onChange} />);
    type("  /movies  ");
    fireEvent.click(screen.getByRole("button", { name: "Add" }));
    await waitFor(() => expect(onChange).toHaveBeenCalledWith(["/a", "/movies"]));
    expect(f.mock.calls[0][0]).toBe("/api/admin/validate-folder");
    expect(JSON.parse(f.mock.calls[0][1].body)).toEqual({ path: "/movies" });
    expect(screen.getByText(/12 videos, 3.5 GB/)).toBeInTheDocument();
    expect(field().value).toBe("");
  });

  it("does not add a folder the host rejects, and shows why", async () => {
    validate({ valid: false, video_count: 0, total_bytes: 0, warnings: ["That folder does not exist."] });
    const onChange = vi.fn();
    render(<FolderList folders={[]} onChange={onChange} />);
    type("/nope");
    fireEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(await screen.findByText("That folder does not exist.")).toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
    expect(field().value).toBe("/nope");
  });

  it("shows the error when the check fails", async () => {
    validate({ message: "Sign in again." }, 401);
    render(<FolderList folders={[]} onChange={vi.fn()} />);
    type("/x");
    fireEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(await screen.findByText("Sign in again.")).toBeInTheDocument();
  });

  it("ignores empty input and folders already in the list", () => {
    const f = validate({ valid: true, video_count: 1, total_bytes: 1, warnings: [] });
    render(<FolderList folders={["/a"]} onChange={vi.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: "Add" }));
    type("/a");
    fireEvent.click(screen.getByRole("button", { name: "Add" }));
    expect(f).not.toHaveBeenCalled();
  });

  it("adds on Enter without submitting a surrounding form", async () => {
    validate({ valid: true, video_count: 1, total_bytes: 1e6, warnings: [] });
    const onSubmit = vi.fn((e) => e.preventDefault());
    const onChange = vi.fn();
    render(<form onSubmit={onSubmit}><FolderList folders={[]} onChange={onChange} /></form>);
    type("/m");
    fireEvent.keyDown(field(), { key: "Enter" });
    await waitFor(() => expect(onChange).toHaveBeenCalledWith(["/m"]));
    expect(onSubmit).not.toHaveBeenCalled();
  });
});
