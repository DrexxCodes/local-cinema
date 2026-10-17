import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { DeviceInfo, LimitInfo } from "@/types";
import LimitPanel from "./LimitPanel";

const device = (over: Partial<DeviceInfo>): DeviceInfo => ({
  id: "d1", device_name: "iPhone", ip: "192.168.1.5", kind: "browser", current: false, started_at: "2026-01-01 20:15:00.000Z", last_seen: "", ...over,
});
const limit = (over: Partial<LimitInfo> = {}): LimitInfo => ({
  message: "", max: 2, devices: [device({ id: "a", device_name: "iPhone" }), device({ id: "b", device_name: "Laptop", current: true })], request: null, ...over,
});
const setup = (l: LimitInfo, onDisconnect = vi.fn().mockResolvedValue(undefined), onUpgrade = vi.fn().mockResolvedValue(undefined)) => {
  render(<LimitPanel limit={l} onDisconnect={onDisconnect} onUpgrade={onUpgrade} />);
  return { onDisconnect, onUpgrade };
};

describe("LimitPanel", () => {
  it("says how many devices the profile may use, with the right plural", () => {
    setup(limit({ max: 2 }));
    expect(screen.getByText(/can use 2 devices at once/)).toBeInTheDocument();
  });
  it("uses the singular for one device", () => {
    setup(limit({ max: 1 }));
    expect(screen.getByText(/can use 1 device at once/)).toBeInTheDocument();
  });

  it("lists devices, marks this one, and offers to disconnect only the others", () => {
    setup(limit());
    expect(screen.getByText("This device")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Disconnect" })).toHaveLength(1);
  });

  it("shows the start time and falls back to a generic device name", () => {
    setup(limit({ devices: [device({ device_name: "", started_at: "2026-01-01 20:15:00.000Z" })] }));
    expect(screen.getByText("Device")).toBeInTheDocument();
    expect(screen.getByText(/since 20:15/)).toBeInTheDocument();
  });

  it("disconnects the chosen device", async () => {
    const { onDisconnect } = setup(limit());
    fireEvent.click(screen.getByRole("button", { name: "Disconnect" }));
    await waitFor(() => expect(onDisconnect).toHaveBeenCalledWith("a"));
  });

  it("sends an upgrade request for the number picked", async () => {
    const { onUpgrade } = setup(limit());
    fireEvent.click(screen.getByRole("button", { name: "Request Upgrade" }));
    fireEvent.change(screen.getByLabelText(/How many extra connections/), { target: { value: "3" } });
    fireEvent.click(screen.getByRole("button", { name: "Send request" }));
    await waitFor(() => expect(onUpgrade).toHaveBeenCalledWith(3));
    await waitFor(() => expect(screen.queryByRole("button", { name: "Send request" })).toBeNull());
  });

  it("keeps the request between 1 and 20", () => {
    setup(limit());
    fireEvent.click(screen.getByRole("button", { name: "Request Upgrade" }));
    const field = screen.getByLabelText(/How many extra connections/) as HTMLInputElement;
    fireEvent.change(field, { target: { value: "99" } });
    expect(field.value).toBe("20");
    fireEvent.change(field, { target: { value: "0" } });
    expect(field.value).toBe("1");
  });

  it("shows why a request failed and keeps the form open", async () => {
    setup(limit(), undefined, vi.fn().mockRejectedValue(new Error("Too many requests.")));
    fireEvent.click(screen.getByRole("button", { name: "Request Upgrade" }));
    fireEvent.click(screen.getByRole("button", { name: "Send request" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Too many requests.");
    expect(screen.getByRole("button", { name: "Send request" })).toBeInTheDocument();
  });

  it("can cancel the form", () => {
    setup(limit());
    fireEvent.click(screen.getByRole("button", { name: "Request Upgrade" }));
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.getByRole("button", { name: "Request Upgrade" })).toBeInTheDocument();
  });

  it("shows a waiting request and lets the guest change it, but not ask again", () => {
    setup(limit({ request: { status: "pending", slots: 2, granted: 0 } }));
    expect(screen.getByText(/request for 2 more connections is waiting/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Request Upgrade" })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Change it" }));
    expect((screen.getByLabelText(/How many extra connections/) as HTMLInputElement).value).toBe("2");
  });

  it("explains a declined or approved request and lets the guest ask again", () => {
    const { unmount } = render(<LimitPanel limit={limit({ request: { status: "declined", slots: 1, granted: 0 } })} onDisconnect={vi.fn()} onUpgrade={vi.fn()} />);
    expect(screen.getByText("Your last request was declined.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Request Upgrade" })).toBeInTheDocument();
    unmount();
    render(<LimitPanel limit={limit({ request: { status: "approved", slots: 1, granted: 1 } })} onDisconnect={vi.fn()} onUpgrade={vi.fn()} />);
    expect(screen.getByText(/was approved/)).toBeInTheDocument();
  });
});
