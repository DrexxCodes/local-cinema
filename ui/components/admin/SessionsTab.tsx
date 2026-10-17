"use client";
import { useCallback, useEffect, useState } from "react";
import { SESSIONS_REFRESH_MS } from "@/constants/admin";
import { adminApi, getAdminToken } from "@/lib/adminApi";
import type { AdminSession, AdminStats } from "@/types";
import SignalBars from "@/components/signal/SignalBars";
import SignalPanel from "@/components/signal/SignalPanel";
import type { SignalLevel } from "@/types";

export default function SessionsTab({ onAuthLost }: { onAuthLost: () => void }) {
  const [rows, setRows] = useState<AdminSession[]>([]);
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [err, setErr] = useState("");

  const load = useCallback(async () => {
    try {
      setRows(await adminApi("/api/admin/sessions"));
      setStats(await adminApi("/api/admin/stats"));
      setErr("");
    } catch (e) {
      setErr((e as Error).message);
      if (!getAdminToken()) onAuthLost();
    }
  }, [onAuthLost]);

  useEffect(() => {
    load();
    const t = setInterval(load, SESSIONS_REFRESH_MS);
    return () => clearInterval(t);
  }, [load]);

  async function end(id: string) {
    await adminApi(`/api/admin/sessions/${id}`, { method: "DELETE" }).catch(() => {});
    load();
  }

  const up = stats ? Math.floor(stats.uptime_seconds / 60) : 0;
  const rated = rows.filter((r) => r.rtt_ms > 0);
  const avg = rated.length ? Math.round(rated.reduce((s, r) => s + r.rtt_ms, 0) / rated.length) : null;

  return (
    <section className="mt-6 grid gap-6 lg:grid-cols-[1fr_16rem]">
      <div>
        {stats && (
          <p className="text-sm text-dim">
            {stats.active_sessions} active, {stats.movies} movies, up {Math.floor(up / 60)}h {up % 60}m
            {avg !== null && `, average latency ${avg} ms`}
          </p>
        )}
        {err && <p role="alert" className="mt-2 text-sm text-red-400">{err}</p>}
        <div className="mt-4 overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="text-dim"><tr><th className="py-2">Profile</th><th>Device</th><th>Signal</th><th>IP</th><th>Started</th><th>Last seen</th><th /></tr></thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.id} className="border-t border-edge">
                  <td className="py-2">{r.room_number}</td>
                  <td>{r.device_name}{r.kind === "external" && <span className="ml-2 rounded bg-white/10 px-1.5 py-0.5 text-[10px] text-dim">Stream URL</span>}</td>
                  <td>
                    <span className="flex items-center gap-2" title={r.rtt_ms ? `${r.rtt_ms} ms` : "No reading yet"}>
                      <SignalBars level={r.signal as SignalLevel} ready={r.rtt_ms > 0} />
                      <span className="text-xs tabular-nums text-dim">{r.rtt_ms ? `${r.rtt_ms} ms` : "..."}</span>
                    </span>
                  </td>
                  <td>{r.ip}</td><td>{r.started_at.slice(11, 19)}</td><td>{r.last_seen.slice(11, 19)}</td>
                  <td><button onClick={() => end(r.id)} className="rounded border border-edge px-3 py-1 hover:border-red-400">End session</button></td>
                </tr>
              ))}
            </tbody>
          </table>
          {rows.length === 0 && <p className="mt-6 text-dim">Nobody is watching right now.</p>}
        </div>
      </div>
      <SignalPanel title="This browser to the server" className="h-fit" />
    </section>
  );
}
