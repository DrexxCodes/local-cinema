"use client";
import { useCallback, useEffect, useState } from "react";
import { adminApi } from "@/lib/adminApi";
import type { UpgradeRequest } from "@/types";
import { Notice } from "./ui";

function Row({ r, onDone }: { r: UpgradeRequest; onDone: () => void }) {
  const [slots, setSlots] = useState(r.slots);
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  async function resolve(action: "approve" | "decline") {
    setBusy(true); setErr("");
    try { await adminApi(`/api/admin/upgrade-requests/${r.id}`, { method: "PATCH", body: JSON.stringify({ action, slots }) }); onDone(); }
    catch (e) { setErr((e as Error).message); setBusy(false); }
  }

  return (
    <li className="rounded-xl border border-edge bg-panel/80 p-4">
      <div className="flex flex-wrap items-center gap-x-6 gap-y-2">
        <div>
          <p className="font-semibold">Profile {r.room_number} <span className="font-normal text-dim">{r.display_name}</span></p>
          <p className="text-sm text-dim">Using {r.online} of {r.max_connections} connections. Asked for {r.slots} more.</p>
        </div>
        <label className="ml-auto flex items-center gap-2 text-sm text-dim">Give
          <input type="number" min={1} max={20} value={slots} onChange={(e) => setSlots(Math.max(1, Math.min(20, +e.target.value || 1)))}
            className="w-20 rounded-md border border-edge bg-panel px-2 py-1.5 text-ink" aria-label="Extra connections to give" />
          more, for {r.max_connections + slots} in total
        </label>
        <button disabled={busy} onClick={() => resolve("approve")} className="rounded-md bg-accent px-4 py-2 text-sm font-medium disabled:opacity-60">Approve</button>
        <button disabled={busy} onClick={() => resolve("decline")} className="rounded-md border border-edge px-4 py-2 text-sm hover:border-red-400 disabled:opacity-60">Decline</button>
      </div>
      {err && <Notice kind="error">{err}</Notice>}
    </li>
  );
}

export default function RequestsTab() {
  const [pending, setPending] = useState<UpgradeRequest[] | null>(null);
  const [done, setDone] = useState<UpgradeRequest[]>([]);
  const [err, setErr] = useState("");

  const load = useCallback(async () => {
    try {
      setPending(await adminApi<UpgradeRequest[]>("/api/admin/upgrade-requests"));
      setDone(await adminApi<UpgradeRequest[]>("/api/admin/upgrade-requests?status=resolved"));
      setErr("");
    } catch (e) { setErr((e as Error).message); }
  }, []);

  useEffect(() => { load(); const t = setInterval(load, 10000); return () => clearInterval(t); }, [load]);

  return (
    <section className="mt-6">
      <p className="max-w-xl text-sm text-dim">Guests who hit their device limit can ask for more connections. Change the number if you want to give a different amount. You can also edit a profile&apos;s limit any time in the Users tab.</p>
      {err && <Notice kind="error">{err}</Notice>}
      <ul className="mt-5 space-y-3">{pending?.map((r) => <Row key={r.id} r={r} onDone={load} />)}</ul>
      {pending && pending.length === 0 && <p className="mt-6 text-dim">No requests waiting.</p>}

      {done.length > 0 && (
        <details className="mt-8">
          <summary className="cursor-pointer text-sm text-dim">Earlier requests ({done.length})</summary>
          <ul className="mt-3 space-y-1 text-sm">
            {done.map((r) => (
              <li key={r.id} className="flex flex-wrap gap-x-4 text-dim">
                <span className="text-ink">Profile {r.room_number}</span>
                <span>{r.status === "approved" ? `approved +${r.granted}` : "declined"}</span>
                <span>asked for {r.slots}</span>
                <span>{r.created.slice(0, 10)}</span>
              </li>
            ))}
          </ul>
        </details>
      )}
    </section>
  );
}
