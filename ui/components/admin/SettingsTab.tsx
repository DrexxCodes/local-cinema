"use client";
import { useEffect, useState } from "react";
import { useCinema } from "@/hooks/useCinema";
import { adminApi, clearAdminToken } from "@/lib/adminApi";
import type { AdminSettings } from "@/types";
import PasswordInput from "../ui/PasswordInput";
import AboutCard from "./AboutCard";
import FolderList from "./FolderList";
import ToolsCard from "./ToolsCard";
import { Field, fieldClass, Notice, Toggle } from "./ui";

const Card = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <section className="rounded-2xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur">
    <h2 className="title-face text-lg">{title}</h2>
    <div className="mt-4 space-y-4">{children}</div>
  </section>
);

export default function SettingsTab({ onAuthLost }: { onAuthLost: () => void }) {
  const { refresh } = useCinema();
  const [s, setS] = useState<AdminSettings | null>(null);
  const [tmdb, setTmdb] = useState("");
  const [clearKey, setClearKey] = useState(false);
  const [msg, setMsg] = useState<{ kind: "ok" | "error"; text: string } | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => { adminApi<AdminSettings>("/api/admin/settings").then(setS).catch((e) => setMsg({ kind: "error", text: e.message })); }, []);
  if (!s) return <p className="mt-6 text-dim">{msg?.text ?? "Loading"}</p>;

  const set = <K extends keyof AdminSettings>(k: K, v: AdminSettings[K]) => setS((p) => (p ? { ...p, [k]: v } : p));
  const port = typeof window !== "undefined" ? window.location.port : "";

  async function save(e: React.FormEvent) {
    e.preventDefault();
    if (!s) return;
    setBusy(true); setMsg(null);
    try {
      const next = await adminApi<AdminSettings>("/api/admin/settings", {
        method: "PUT",
        body: JSON.stringify({
          cinema_name: s.cinema_name, tagline: s.tagline, accent_color: s.accent_color,
          movie_folders: s.movie_folders, watch_folders: s.watch_folders,
          global_max_streams: s.global_max_streams, mdns_name: s.mdns_name,
          ...(clearKey ? { clear_tmdb_key: true } : tmdb.trim() ? { tmdb_api_key: tmdb.trim() } : {}),
        }),
      });
      setS(next); setTmdb(""); setClearKey(false); refresh();
      setMsg({ kind: "ok", text: "Saved. Folder changes start a rescan automatically." });
    } catch (e) { setMsg({ kind: "error", text: (e as Error).message }); } finally { setBusy(false); }
  }

  return (
    <div className="mt-6 grid gap-6 lg:grid-cols-2">
      <form onSubmit={save} className="space-y-6 lg:col-span-2 lg:grid lg:grid-cols-2 lg:gap-6 lg:space-y-0">
        <Card title="Name and look">
          <Field label="Cinema name"><input className={fieldClass} value={s.cinema_name} maxLength={50} onChange={(e) => set("cinema_name", e.target.value)} /></Field>
          <Field label="Tagline"><input className={fieldClass} value={s.tagline} maxLength={120} onChange={(e) => set("tagline", e.target.value)} /></Field>
          <Field label="Accent color">
            <input type="color" value={s.accent_color || "#7c3aed"} onChange={(e) => set("accent_color", e.target.value)} className="mt-1 block h-10 w-20 rounded border border-edge bg-panel" />
          </Field>
        </Card>

        <Card title="Limits and discovery">
          <Field label="Maximum devices watching at once">
            <input type="number" min={1} className={fieldClass} value={s.global_max_streams} onChange={(e) => set("global_max_streams", +e.target.value)} />
          </Field>
          {s.global_max_streams > 20 && <p className="text-sm text-amber-400">Above 20 streams, your WiFi is usually the bottleneck.</p>}
          <Field label="Network name" hint={`Guests open http://${s.mdns_name}.local${port ? ":" + port : ""}`}>
            <input className={fieldClass} value={s.mdns_name} onChange={(e) => set("mdns_name", e.target.value.replace(/[^a-z0-9-]/gi, "").toLowerCase())} />
          </Field>
          <Field label={s.has_tmdb_key && !clearKey ? "TMDB API key (one is saved, type to replace)" : "TMDB API key (optional, adds posters and synopses)"}>
            <input className={fieldClass} value={tmdb} onChange={(e) => setTmdb(e.target.value)} placeholder={s.has_tmdb_key && !clearKey ? "Saved key hidden" : ""} />
          </Field>
          {s.has_tmdb_key && <button type="button" onClick={() => setClearKey((c) => !c)} className="text-sm text-dim underline">{clearKey ? "Keep the saved key" : "Remove the saved key"}</button>}
        </Card>

        <div className="lg:col-span-2">
          <Card title="Movie folders">
            <FolderList folders={s.movie_folders} onChange={(f) => set("movie_folders", f)} />
            <Toggle checked={s.watch_folders} onChange={(v) => set("watch_folders", v)} label="Watch folders and rescan when files change" />
          </Card>
        </div>

        <div className="flex flex-wrap items-center gap-4 lg:col-span-2">
          <button disabled={busy} className="rounded-md bg-accent px-6 py-2.5 font-medium disabled:opacity-60">{busy ? "Saving" : "Save changes"}</button>
          {msg && <Notice kind={msg.kind}>{msg.text}</Notice>}
        </div>
      </form>

      <div className="lg:col-span-2"><AccountForm current={s.admin_email} onAuthLost={onAuthLost} /></div>
      <ToolsCard />
      <AboutCard version={s.version} />
    </div>
  );
}

function AccountForm({ current, onAuthLost }: { current: string; onAuthLost: () => void }) {
  const [email, setEmail] = useState(current);
  const [pw, setPw] = useState("");
  const [next, setNext] = useState("");
  const [again, setAgain] = useState("");
  const [msg, setMsg] = useState<{ kind: "ok" | "error"; text: string } | null>(null);

  async function go(e: React.FormEvent) {
    e.preventDefault();
    if (next && next !== again) return setMsg({ kind: "error", text: "The new passwords do not match." });
    if (!pw) return setMsg({ kind: "error", text: "Enter your current password to confirm." });
    try {
      const r = await adminApi<{ relogin: boolean }>("/api/admin/account", {
        method: "PUT", body: JSON.stringify({ current_password: pw, email, new_password: next }),
      });
      setPw(""); setNext(""); setAgain("");
      if (r.relogin) { clearAdminToken(); onAuthLost(); return; }
      setMsg({ kind: "ok", text: "Admin account updated." });
    } catch (e) { setMsg({ kind: "error", text: (e as Error).message }); }
  }

  return (
    <form onSubmit={go} className="rounded-2xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur">
      <h2 className="title-face text-lg">Admin account</h2>
      <div className="mt-4 grid gap-4 sm:grid-cols-2">
        <Field label="Email"><input type="email" className={fieldClass} value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="username" /></Field>
        <Field label="Current password (required)"><PasswordInput className={fieldClass} value={pw} onChange={(e) => setPw(e.target.value)} autoComplete="current-password" /></Field>
        <Field label="New password (8+ characters, optional)"><PasswordInput className={fieldClass} value={next} onChange={(e) => setNext(e.target.value)} autoComplete="new-password" /></Field>
        <Field label="Confirm new password"><PasswordInput className={fieldClass} value={again} onChange={(e) => setAgain(e.target.value)} autoComplete="new-password" /></Field>
      </div>
      <div className="mt-5 flex flex-wrap items-center gap-4">
        <button className="rounded-md border border-edge px-5 py-2.5 hover:border-accent">Update account</button>
        {msg && <Notice kind={msg.kind}>{msg.text}</Notice>}
      </div>
    </form>
  );
}
