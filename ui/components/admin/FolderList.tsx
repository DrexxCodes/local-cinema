"use client";
import { useState } from "react";
import { adminApi } from "@/lib/adminApi";
import type { FolderCheck } from "@/types";
import { fieldClass } from "./ui";

type Props = { folders: string[]; onChange: (f: string[]) => void };

/** Editable list of movie folders. New folders are checked on the host before they are added. */
export default function FolderList({ folders, onChange }: Props) {
  const [draft, setDraft] = useState("");
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);

  async function add() {
    const path = draft.trim();
    if (!path || folders.includes(path)) return;
    setBusy(true); setNote("");
    try {
      const v = await adminApi<FolderCheck>("/api/admin/validate-folder", { method: "POST", body: JSON.stringify({ path }) });
      if (!v.valid) { setNote(v.warnings.join(" ")); return; }
      setNote(`${v.video_count} videos, ${(v.total_bytes / 1e9).toFixed(1)} GB. ${v.warnings.join(" ")}`);
      onChange([...folders, path]);
      setDraft("");
    } catch (e) { setNote((e as Error).message); } finally { setBusy(false); }
  }

  return (
    <div>
      <ul className="space-y-2">
        {folders.map((f) => (
          <li key={f} className="flex items-center justify-between gap-3 rounded-md border border-edge bg-panel px-3 py-2 text-sm">
            <span className="break-all">{f}</span>
            <button type="button" className="shrink-0 text-dim hover:text-red-400" onClick={() => onChange(folders.filter((x) => x !== f))}>Remove</button>
          </li>
        ))}
      </ul>
      <div className="mt-3 flex gap-2">
        <input className={`${fieldClass} mt-0`} value={draft} placeholder="D:\Movies or /mnt/movies" aria-label="Folder path"
          onChange={(e) => setDraft(e.target.value)} onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); add(); } }} />
        <button type="button" onClick={add} disabled={busy} className="rounded-md border border-edge px-4 disabled:opacity-60">Add</button>
      </div>
      {note && <p className="mt-2 text-sm text-dim">{note}</p>}
    </div>
  );
}
