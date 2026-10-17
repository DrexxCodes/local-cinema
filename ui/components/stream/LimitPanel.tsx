"use client";
import { useState } from "react";
import type { LimitInfo } from "@/types";

type Props = { limit: LimitInfo; onDisconnect: (id: string) => Promise<void>; onUpgrade: (slots: number) => Promise<void> };

/** Shown when the profile has used all its connections. */
export default function LimitPanel({ limit, onDisconnect, onUpgrade }: Props) {
  const [asking, setAsking] = useState(false);
  const [slots, setSlots] = useState(1);
  const [busy, setBusy] = useState("");
  const [err, setErr] = useState("");
  const req = limit.request;

  async function send(e: React.FormEvent) {
    e.preventDefault(); setBusy("upgrade"); setErr("");
    try { await onUpgrade(slots); setAsking(false); } catch (x) { setErr((x as Error).message); } finally { setBusy(""); }
  }

  return (
    <div>
      <p className="font-semibold">You have reached the maximum connection for this profile</p>
      <p className="mt-1 text-sm text-dim">This profile can use {limit.max} device{limit.max === 1 ? "" : "s"} at once. These are connected now:</p>

      <ul className="mt-3 space-y-2">
        {limit.devices.map((d) => (
          <li key={d.id} className="flex items-center gap-3 rounded-lg border border-edge bg-white/[0.04] px-3 py-2 text-sm">
            <div className="min-w-0 flex-1">
              <p className="truncate">{d.device_name || "Device"}{d.current && <span className="ml-2 rounded bg-accent px-1.5 py-0.5 text-[10px] font-bold">This device</span>}</p>
              <p className="truncate text-xs text-dim">{[d.ip, `since ${d.started_at.slice(11, 16)}`].filter(Boolean).join(", ")}</p>
            </div>
            {!d.current && (
              <button disabled={!!busy} onClick={async () => { setBusy(d.id); await onDisconnect(d.id); setBusy(""); }}
                className="shrink-0 rounded border border-edge px-2.5 py-1 text-xs hover:border-red-400 disabled:opacity-50">Disconnect</button>
            )}
          </li>
        ))}
      </ul>

      <div className="mt-4 rounded-lg border border-edge p-3 text-sm">
        {req?.status === "pending" && !asking && (
          <p>Your request for {req.slots} more connection{req.slots === 1 ? "" : "s"} is waiting for approval. <button onClick={() => { setSlots(req.slots); setAsking(true); }} className="underline">Change it</button></p>
        )}
        {req?.status === "declined" && !asking && <p className="text-dim">Your last request was declined.</p>}
        {req?.status === "approved" && !asking && <p className="text-dim">Your last request was approved. If you are still at the limit, ask again.</p>}
        {asking ? (
          <form onSubmit={send} className="flex flex-wrap items-end gap-3">
            <label className="text-dim">How many extra connections do you need?
              <input type="number" min={1} max={20} value={slots} onChange={(e) => setSlots(Math.max(1, Math.min(20, +e.target.value || 1)))}
                className="mt-1 block w-24 rounded-md border border-edge bg-panel px-3 py-2 text-ink" />
            </label>
            <button disabled={busy === "upgrade"} className="rounded-md bg-accent px-4 py-2 font-medium disabled:opacity-60">Send request</button>
            <button type="button" onClick={() => setAsking(false)} className="px-2 py-2 text-dim">Cancel</button>
          </form>
        ) : (!req || req.status !== "pending") && (
          <button onClick={() => { setSlots(1); setAsking(true); }} className="rounded-md bg-accent px-4 py-2 font-medium">Request Upgrade</button>
        )}
        {err && <p role="alert" className="mt-2 text-red-400">{err}</p>}
      </div>
    </div>
  );
}
