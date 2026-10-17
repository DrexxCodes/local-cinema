"use client";
import { useEffect, useState } from "react";
import { adminApi } from "@/lib/adminApi";
import type { UpdateInfo } from "@/types";
import { Notice } from "./ui";

const when = (iso: string) => (iso ? new Date(iso).toLocaleString() : "never");

/** Current version, and a button that asks GitHub whether a newer one exists. */
export default function AboutCard({ version }: { version: string }) {
  const [u, setU] = useState<UpdateInfo | null>(null);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");

  useEffect(() => { adminApi<UpdateInfo>("/api/admin/update").then(setU).catch((e) => setErr(e.message)); }, []);

  async function check() {
    setBusy(true); setErr("");
    try { setU(await adminApi<UpdateInfo>("/api/admin/update/check", { method: "POST" })); }
    catch (e) { setErr((e as Error).message); } finally { setBusy(false); }
  }

  return (
    <section className="rounded-2xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur">
      <h2 className="title-face text-lg">About and updates</h2>
      <p className="mt-3 text-sm">Version <span className="font-semibold tabular-nums">{version === "dev" ? "development build" : version}</span></p>

      {u && !u.enabled && <p className="mt-2 text-sm text-dim">Update checks are off in development builds.</p>}
      {u?.enabled && (
        <div className="mt-2 text-sm">
          {u.error ? <Notice kind="error">{u.error}</Notice>
            : u.available ? <p className="text-emerald-400">Version {u.latest} is available.</p>
            : u.latest ? <p className="text-dim">You have the latest version.</p>
            : <p className="text-dim">Not checked yet.</p>}
          <p className="mt-1 text-xs text-dim">Last checked: {when(u.checked_at)}</p>
        </div>
      )}

      <div className="mt-4 flex flex-wrap items-center gap-3">
        <button onClick={check} disabled={busy || u?.enabled === false} className="rounded-md border border-edge px-4 py-2 text-sm hover:border-accent disabled:opacity-50">
          {busy ? "Checking" : "Check for updates"}
        </button>
        {u?.available && u.asset_url && <a href={u.asset_url} className="rounded-md bg-accent px-4 py-2 text-sm font-medium">Download {u.asset_name || "update"}</a>}
        {u?.available && u.url && <a href={u.url} target="_blank" rel="noreferrer" className="text-sm text-dim underline hover:text-ink">What&apos;s new</a>}
      </div>
      {u?.available && u.notes && <pre className="mt-4 max-h-48 overflow-y-auto whitespace-pre-wrap rounded-lg border border-edge bg-panel/60 p-3 text-xs text-dim">{u.notes}</pre>}
      {u?.available && <p className="mt-3 text-xs text-dim">Run the installer to update. Your movies list, profiles and settings are kept.</p>}
      {err && <Notice kind="error">{err}</Notice>}
    </section>
  );
}
