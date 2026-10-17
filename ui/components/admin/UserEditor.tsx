"use client";
import { useState } from "react";
import { adminApi } from "@/lib/adminApi";
import type { AdminUser } from "@/types";
import { Field, fieldClass, Modal, Notice, Toggle } from "./ui";

type Props = { user: AdminUser; onClose: () => void; onSaved: () => void };

export default function UserEditor({ user, onClose, onSaved }: Props) {
  const [name, setName] = useState(user.display_name);
  const [max, setMax] = useState(user.max_connections);
  const [active, setActive] = useState(user.active);
  const [external, setExternal] = useState(user.allow_external);
  const [notes, setNotes] = useState(user.notes);
  const [pw, setPw] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");
  const [confirmDelete, setConfirmDelete] = useState(false);

  async function save(e: React.FormEvent) {
    e.preventDefault(); setBusy(true); setErr("");
    try {
      await adminApi(`/api/admin/users/${user.id}`, {
        method: "PATCH",
        body: JSON.stringify({ display_name: name, max_connections: max, active, notes, allow_external: external, ...(pw ? { password: pw } : {}) }),
      });
      onSaved(); onClose();
    } catch (e) { setErr((e as Error).message); } finally { setBusy(false); }
  }

  async function remove() {
    setBusy(true); setErr("");
    try { await adminApi(`/api/admin/users/${user.id}`, { method: "DELETE" }); onSaved(); onClose(); }
    catch (e) { setErr((e as Error).message); setBusy(false); }
  }

  return (
    <Modal title={`Profile ${user.room_number || user.username}`} onClose={onClose}>
      <form onSubmit={save} className="mt-4 space-y-4">
        <p className="text-xs text-dim">Username: {user.username}</p>
        <Field label="Display name"><input className={fieldClass} value={name} onChange={(e) => setName(e.target.value)} /></Field>
        <Field label="Devices at once"><input type="number" min={1} max={50} className={fieldClass} value={max} onChange={(e) => setMax(+e.target.value)} /></Field>
        <Field label="New password (leave blank to keep)" hint="Changing it signs the profile out everywhere.">
          <input className={fieldClass} value={pw} onChange={(e) => setPw(e.target.value)} autoComplete="new-password" />
        </Field>
        <Field label="Notes"><input className={fieldClass} value={notes} onChange={(e) => setNotes(e.target.value)} /></Field>
        <Toggle checked={active} onChange={setActive} label={active ? "Profile is active" : "Profile is disabled"} />
        <div>
          <Toggle checked={external} onChange={setExternal} label={external ? "External streaming is allowed" : "External streaming is off"} />
          <p className="mt-1 text-xs text-dim">Lets this profile copy a stream URL for VLC and other players. Each one uses a device connection.</p>
        </div>
        {err && <Notice kind="error">{err}</Notice>}
        <div className="flex flex-wrap items-center gap-3 pt-2">
          <button disabled={busy} className="rounded-md bg-accent px-5 py-2 font-medium disabled:opacity-60">Save</button>
          <button type="button" onClick={onClose} className="rounded-md border border-edge px-5 py-2">Cancel</button>
          {confirmDelete ? (
            <button type="button" onClick={remove} disabled={busy} className="ml-auto rounded-md bg-red-500 px-4 py-2 text-sm font-medium text-white">Confirm delete</button>
          ) : (
            <button type="button" onClick={() => setConfirmDelete(true)} className="ml-auto text-sm text-red-400">Delete profile</button>
          )}
        </div>
      </form>
    </Modal>
  );
}
