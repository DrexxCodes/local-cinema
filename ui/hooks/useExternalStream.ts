"use client";
import { useCallback, useState } from "react";
import { api, ApiError } from "@/lib/api";
import type { LimitInfo, StreamLink, StreamState } from "@/types";

/** Creates a stream URL for another player and handles the "profile is full" case. */
export function useExternalStream(movieId: string) {
  const [state, setState] = useState<StreamState>({ kind: "idle" });

  const request = useCallback(async () => {
    setState({ kind: "loading" });
    try {
      const link = await api<StreamLink>(`/api/movies/${movieId}/stream-url`, { method: "POST" });
      setState({ kind: "ready", link });
    } catch (e) {
      const err = e as ApiError;
      if (err.code === "max_connections_reached") setState({ kind: "limit", limit: err.body as LimitInfo });
      else if (err.code === "external_not_allowed") setState({ kind: "denied", message: err.message });
      else setState({ kind: "error", message: err.message });
    }
  }, [movieId]);

  /** Disconnects another device on this profile, then tries again. */
  const disconnect = useCallback(async (id: string) => {
    await api(`/api/sessions/${id}`, { method: "DELETE" }).catch(() => {});
    await request();
  }, [request]);

  /** Sends an upgrade request and shows it as pending. */
  const sendUpgrade = useCallback(async (slots: number) => {
    await api("/api/upgrade-requests", { method: "POST", body: JSON.stringify({ slots }) });
    setState((s) => (s.kind === "limit" ? { kind: "limit", limit: { ...s.limit, request: { status: "pending", slots, granted: 0 } } } : s));
  }, []);

  return { state, request, disconnect, sendUpgrade };
}
