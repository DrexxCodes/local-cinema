import { describe, expect, it } from "vitest";
import { SESSION_KEY } from "@/constants/storage";
import { clearSession, getSession, setSession } from "./session";

const session = { token: "t", session_id: "s", user: { id: "u", room_number: "101", display_name: "Room 101" } };

describe("session storage", () => {
  it("is null when nobody is signed in", () => {
    expect(getSession()).toBeNull();
  });
  it("stores and returns the session", () => {
    setSession(session);
    expect(getSession()).toEqual(session);
  });
  it("clears it", () => {
    setSession(session);
    clearSession();
    expect(getSession()).toBeNull();
  });
  it("treats damaged storage as signed out instead of throwing", () => {
    localStorage.setItem(SESSION_KEY, "{not json");
    expect(getSession()).toBeNull();
  });
});
