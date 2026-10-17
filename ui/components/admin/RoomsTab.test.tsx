import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { RoomsResult } from "@/types";

vi.mock("qrcode", () => ({ default: { toDataURL: vi.fn().mockResolvedValue("data:image/png;base64,AAAA") } }));
import RoomsTab from "./RoomsTab";

let finish!: (r: RoomsResult | { status: number; message: string }) => void;

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn((path: string) => {
    if (path === "/api/setup/interfaces") return Promise.resolve({ ok: true, json: async () => [{ ip: "192.168.1.9" }] });
    return new Promise((resolve) => {
      finish = (r) => resolve("status" in r
        ? { ok: false, status: r.status, json: async () => ({ message: r.message }) }
        : { ok: true, status: 200, json: async () => r });
    });
  }));
});
afterEach(() => vi.unstubAllGlobals());

const submit = () => fireEvent.click(screen.getByRole("button", { name: "Create profiles" }));

describe("RoomsTab (Profiles)", () => {
  it("uses the word profile, not room", () => {
    render(<RoomsTab />);
    expect(screen.getByText("First profile number")).toBeInTheDocument();
    expect(screen.getByText("Devices per profile")).toBeInTheDocument();
    expect(screen.queryByText(/room/i)).toBeNull();
    expect((screen.getByLabelText("Prefix") as HTMLInputElement).value).toBe("profile");
  });

  it("tells the admin profiles are being created and blocks a second click", async () => {
    render(<RoomsTab />);
    fireEvent.change(screen.getByLabelText("First profile number"), { target: { value: "125" } });
    fireEvent.change(screen.getByLabelText("Last profile number"), { target: { value: "200" } });
    submit();
    expect(await screen.findByRole("status")).toHaveTextContent("Creating 76 profiles");
    const button = screen.getByRole("button", { name: "Creating profiles..." });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute("aria-busy", "true");
    fireEvent.click(button);
    expect(vi.mocked(fetch).mock.calls.filter((c) => c[0] === "/api/admin/rooms/bulk")).toHaveLength(1);
    finish({ created: [], skipped: [] });
    await waitFor(() => expect(screen.queryByText(/Creating 76 profiles/)).toBeNull());
  });

  it("confirms how many were created", async () => {
    render(<RoomsTab />);
    submit();
    finish({ created: [{ room: "101", email: "profile101@cinema.local", password: "pw" }], skipped: [] });
    expect(await screen.findByText("1 profile created.")).toBeInTheDocument();
    expect(screen.getByText("Profile 101")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create profiles" })).toBeEnabled();
  });

  it("says so when nothing new was created", async () => {
    render(<RoomsTab />);
    submit();
    finish({ created: [], skipped: ["101", "102"] });
    expect(await screen.findByText(/every profile in that range already exists/)).toBeInTheDocument();
  });

  it("shows the server's error and enables the button again", async () => {
    render(<RoomsTab />);
    submit();
    finish({ status: 400, message: "Provide a valid from/to range (max 1000 profiles)." });
    expect(await screen.findByRole("alert")).toHaveTextContent("max 1000 profiles");
    expect(screen.queryByText(/Creating \d+ profile/)).toBeNull();
    expect(screen.getByRole("button", { name: "Create profiles" })).toBeEnabled();
  });
});
