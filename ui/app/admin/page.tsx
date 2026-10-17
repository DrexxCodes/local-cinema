"use client";
import { useCallback, useEffect, useState } from "react";
import { adminApi, clearAdminToken, getAdminToken } from "@/lib/adminApi";
import type { AdminStats } from "@/types";
import AdminLogin from "@/components/admin/AdminLogin";
import LibraryTab from "@/components/admin/LibraryTab";
import RequestsTab from "@/components/admin/RequestsTab";
import RoomsTab from "@/components/admin/RoomsTab";
import SessionsTab from "@/components/admin/SessionsTab";
import SettingsTab from "@/components/admin/SettingsTab";
import UsersTab from "@/components/admin/UsersTab";
import SignalBadge from "@/components/signal/SignalBadge";

const TABS = ["sessions", "users", "requests", "profiles", "library", "settings"] as const;
type Tab = (typeof TABS)[number];

export default function Admin() {
  const [authed, setAuthed] = useState(false);
  const [checked, setChecked] = useState(false);
  const [tab, setTab] = useState<Tab>("sessions");

  const [pending, setPending] = useState(0);
  const [update, setUpdate] = useState("");

  useEffect(() => { setAuthed(!!getAdminToken()); setChecked(true); }, []);

  // Count of upgrade requests waiting for an answer, shown on the Requests tab.
  useEffect(() => {
    if (!authed) return;
    const load = () => adminApi<AdminStats>("/api/admin/stats").then((s) => { setPending(s.pending_upgrades); setUpdate(s.update_available); }).catch(() => {});
    load();
    const t = setInterval(load, 15000);
    return () => clearInterval(t);
  }, [authed, tab]);
  const lost = useCallback(() => setAuthed(false), []);

  if (!checked) return null;
  if (!authed) return <AdminLogin onDone={() => setAuthed(true)} />;

  return (
    <main className="mx-auto max-w-6xl p-6">
      <header className="flex flex-wrap items-center gap-4 print:hidden">
        <h1 className="title-face text-3xl">Admin</h1>
        <nav className="flex flex-wrap gap-1" aria-label="Admin sections">
          {TABS.map((t) => (
            <button key={t} onClick={() => setTab(t)} aria-current={tab === t}
              className={`rounded-md px-4 py-1.5 text-sm capitalize ${tab === t ? "bg-accent" : "border border-edge"}`}>{t}{t === "requests" && pending > 0 && <span className="ml-2 rounded-full bg-red-500 px-1.5 text-xs font-bold text-white">{pending}</span>}</button>
          ))}
        </nav>
        <div className="ml-auto flex items-center gap-4">
          {update && <button onClick={() => setTab("settings")} className="rounded-full bg-emerald-500/20 px-3 py-1 text-xs text-emerald-300">Update {update} available</button>}
          <SignalBadge />
          <a href="/_/" className="text-sm text-dim">Full database dashboard</a>
          <button className="text-sm text-dim" onClick={() => { clearAdminToken(); setAuthed(false); }}>Sign out</button>
        </div>
      </header>
      {tab === "sessions" && <SessionsTab onAuthLost={lost} />}
      {tab === "users" && <UsersTab />}
      {tab === "requests" && <RequestsTab />}
      {tab === "profiles" && <RoomsTab />}
      {tab === "library" && <LibraryTab />}
      {tab === "settings" && <SettingsTab onAuthLost={lost} />}
    </main>
  );
}
