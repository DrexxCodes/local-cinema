"use client";
import { useEffect, useRef, useState } from "react";
import { JOB_REFRESH_MS, MOVIES_PAGE_SIZE } from "@/constants/admin";
import { useDebounced } from "@/hooks/useDebounced";
import { usePagedList } from "@/hooks/usePagedList";
import { useSentinel } from "@/hooks/useSentinel";
import { adminApi, adminImg } from "@/lib/adminApi";
import { parseClock } from "@/lib/format";
import type { AdminMovie, AdminStats, MovieFilter, Page } from "@/types";
import { CheckIcon } from "@/components/ui/Icons";
import CompatBadge from "./CompatBadge";
import JobProgress from "./JobProgress";
import MovieEditor from "./MovieEditor";
import { fieldClass, Notice } from "./ui";

export default function LibraryTab() {
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [msg, setMsg] = useState<{ kind: "ok" | "error"; text: string } | null>(null);
  const [q, setQ] = useState("");
  const query = useDebounced(q.trim());
  const [picked, setPicked] = useState<Set<string>>(new Set());
  const [at, setAt] = useState("");
  const [bust, setBust] = useState(0);
  const [show, setShow] = useState<MovieFilter>("");
  const [editing, setEditing] = useState<AdminMovie | null>(null);
  const wasRunning = useRef(false);

  const { items, total, hasMore, loading, error, loadMore, reload } = usePagedList<AdminMovie>(
    (offset) => adminApi<Page<AdminMovie>>(`/api/admin/movies?limit=${MOVIES_PAGE_SIZE}&offset=${offset}&q=${encodeURIComponent(query)}&show=${show}`),
    `${query}|${show}`,
  );
  const sentinel = useSentinel(loadMore, hasMore && !loading, items.length);

  // Poll the jobs. When one ends, refresh the pictures and the compatibility labels.
  useEffect(() => {
    let alive = true;
    const tick = () => adminApi<AdminStats>("/api/admin/stats").then((s) => {
      if (!alive) return;
      setStats(s);
      const running = s.scan.running || s.thumbs.running || s.convert.running;
      if (wasRunning.current && !running) { setBust(Date.now()); reload(); }
      wasRunning.current = running;
    }).catch(() => {});
    tick();
    const t = setInterval(tick, JOB_REFRESH_MS);
    return () => { alive = false; clearInterval(t); };
  }, [reload]);

  const noFfmpeg = stats?.ffmpeg === false;
  const jobBusy = !!stats && (stats.scan.running || stats.thumbs.running);
  const converting = !!stats?.convert.running;
  const pickedMovies = items.filter((m) => picked.has(m.id));
  const pickedConverted = pickedMovies.filter((m) => m.compat === "converted");
  const pickedMissing = pickedMovies.filter((m) => !m.available);
  const pickedHidden = pickedMovies.filter((m) => m.hidden);
  const pickedVisible = pickedMovies.filter((m) => m.available && !m.hidden);
  const seconds = parseClock(at);

  async function run(path: string, body: object | null, ok: string) {
    setMsg(null);
    try {
      await adminApi(path, { method: "POST", ...(body ? { body: JSON.stringify(body) } : {}) });
      setMsg({ kind: "ok", text: ok });
      wasRunning.current = true;
    } catch (e) { setMsg({ kind: "error", text: (e as Error).message }); }
  }

  async function removeCopies() {
    setMsg(null);
    try {
      await Promise.all(pickedConverted.map((m) => adminApi(`/api/admin/compat/${m.id}`, { method: "DELETE" })));
      setMsg({ kind: "ok", text: "Converted copies removed. Guests get the original files again." });
      reload();
    } catch (e) { setMsg({ kind: "error", text: (e as Error).message }); }
  }

  async function setHidden(ids: string[], hidden: boolean) {
    setMsg(null);
    try {
      await adminApi("/api/admin/movies/visibility", { method: "POST", body: JSON.stringify({ ids, hidden }) });
      setMsg({ kind: "ok", text: hidden ? "Removed from the stream. Guests no longer see it." : "Back in the stream." });
      setPicked(new Set()); reload();
    } catch (e) { setMsg({ kind: "error", text: (e as Error).message }); }
  }

  async function forgetMissing() {
    setMsg(null);
    try {
      await Promise.all(pickedMissing.map((m) => adminApi(`/api/admin/movies/${m.id}`, { method: "DELETE" })));
      setMsg({ kind: "ok", text: "Removed from the list." });
      setPicked(new Set()); reload();
    } catch (e) { setMsg({ kind: "error", text: (e as Error).message }); }
  }

  const toggle = (id: string) => setPicked((p) => { const n = new Set(p); n.has(id) ? n.delete(id) : n.add(id); return n; });
  const btn = "rounded-md px-5 py-2 font-medium disabled:opacity-50";

  return (
    <section className="mt-6">
      <p className="max-w-lg text-dim">Folder changes are picked up automatically when folder watching is on. Run a manual scan after reconnecting a drive.</p>
      <div className="mt-4 flex flex-wrap gap-3">
        <button onClick={() => run("/api/admin/scan", null, "Scan started.")} disabled={jobBusy} className={`${btn} bg-accent`}>Rescan library</button>
        <button onClick={() => run("/api/admin/thumbnails", { all: true }, "Creating missing thumbnails.")} disabled={jobBusy || noFfmpeg}
          className={`${btn} border border-edge`}>Create missing thumbnails</button>
        <button onClick={() => run("/api/admin/compat", { all: true }, "Converting every movie that needs it.")} disabled={converting || noFfmpeg}
          className={`${btn} border border-edge`}>Make all compatible</button>
      </div>
      {noFfmpeg && <Notice kind="error">FFmpeg was not found on the host, so thumbnails and conversion are unavailable. Install it (winget install Gyan.FFmpeg, brew install ffmpeg or apt install ffmpeg) and restart Local Cinema.</Notice>}
      {msg && <Notice kind={msg.kind}>{msg.text}</Notice>}
      <JobProgress stats={stats} onCancelConvert={() => run("/api/admin/compat/cancel", null, "Cancelling after the current step.")} />

      <h2 className="title-face mt-10 text-xl">Movies</h2>
      <p className="mt-1 max-w-2xl text-sm text-dim">
        Pick movies, then create thumbnails (optionally from a chosen moment in the film) or make them compatible.
        A movie marked <span className="rounded bg-amber-400 px-1 text-[10px] font-bold text-black">May not play</span> is not H.264 video in an MP4 file, so some phones and browsers cannot play it.
        Making it compatible saves an MP4 copy for guests and keeps your original. Files that already hold H.264 video convert in seconds to minutes, anything else is re-encoded and can take a long time.
      </p>

      <div className="mt-4 flex flex-wrap items-center gap-3">
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Search movies" aria-label="Search movies" className={`${fieldClass} mt-0 max-w-xs`} />
        <select value={show} onChange={(e) => { setShow(e.target.value as MovieFilter); setPicked(new Set()); }} aria-label="Show"
          className={`${fieldClass} mt-0 w-auto`}>
          <option value="">All movies</option>
          <option value="hidden">Hidden from guests</option>
          <option value="missing">File missing</option>
        </select>
        <input value={at} onChange={(e) => setAt(e.target.value)} placeholder="Frame at 1:30 (optional)" aria-label="Frame position" inputMode="numeric" className={`${fieldClass} mt-0 w-52`} />
        <span className="ml-auto text-sm text-dim">{items.length} of {total}</span>
      </div>
      <div className="mt-3 flex flex-wrap items-center gap-3">
        <button disabled={!picked.size || jobBusy || noFfmpeg}
          onClick={() => run("/api/admin/thumbnails", { ids: [...picked], at: seconds, force: true }, `Generating ${picked.size} thumbnail${picked.size === 1 ? "" : "s"}.`)}
          className={`${btn} bg-accent`}>Thumbnails for selected ({picked.size})</button>
        <button disabled={!picked.size || converting || noFfmpeg}
          onClick={() => run("/api/admin/compat", { ids: [...picked] }, "Conversion started. Movies that already play everywhere are skipped.")}
          className={`${btn} border border-edge`}>Make compatible ({picked.size})</button>
        {pickedVisible.length > 0 && (
          <button onClick={() => setHidden(pickedVisible.map((m) => m.id), true)} className="text-sm text-dim underline hover:text-red-400">Hide from guests ({pickedVisible.length})</button>
        )}
        {pickedHidden.length > 0 && (
          <button onClick={() => setHidden(pickedHidden.map((m) => m.id), false)} className="text-sm text-dim underline hover:text-ink">Show to guests ({pickedHidden.length})</button>
        )}
        {pickedMissing.length > 0 && (
          <button onClick={forgetMissing} className="text-sm text-dim underline hover:text-red-400">Remove missing from list ({pickedMissing.length})</button>
        )}
        {pickedConverted.length > 0 && (
          <button onClick={removeCopies} className="text-sm text-dim underline hover:text-red-400">Remove converted copy ({pickedConverted.length})</button>
        )}
        <button onClick={() => setPicked(new Set(items.map((m) => m.id)))} className="text-sm text-dim hover:text-ink">Select loaded</button>
        <button onClick={() => setPicked(new Set(items.filter((m) => m.compat === "convert").map((m) => m.id)))} className="text-sm text-dim hover:text-ink">Select unsupported</button>
        <button onClick={() => setPicked(new Set())} className="text-sm text-dim hover:text-ink">Clear</button>
      </div>

      <ul className="mt-5 grid grid-cols-3 gap-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8">
        {items.map((m) => {
          const on = picked.has(m.id);
          return (
            <li key={m.id}>
              <button onClick={() => toggle(m.id)} aria-pressed={on} className={`relative block w-full overflow-hidden rounded-lg ring-2 ${on ? "ring-accent" : "ring-transparent"}`}>
                <div className="aspect-[2/3] bg-gradient-to-br from-accent/40 to-panel">
                  {m.has_poster && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={adminImg("posters", m.id, m.thumb_v, bust)} alt="" loading="lazy" className="h-full w-full object-cover" />
                  )}
                </div>
                <CompatBadge status={m.compat} hidden={m.hidden} missing={!m.available} />
                {on && <span className="absolute right-1.5 top-1.5 grid h-6 w-6 place-items-center rounded-full bg-accent"><CheckIcon className="h-4 w-4" /></span>}
              </button>
              <p className="mt-1 flex items-center gap-1 text-xs">
                <span className="truncate">{m.title}</span>
                <button onClick={() => setEditing(m)} className="ml-auto shrink-0 text-dim underline hover:text-ink" aria-label={`Edit ${m.title}`}>Edit</button>
              </p>
              <p className="truncate text-[11px] text-dim">{[m.container.toUpperCase(), m.video_codec].filter(Boolean).join(" ")}</p>
            </li>
          );
        })}
      </ul>
      {error && <Notice kind="error">{error}</Notice>}
      {loading && <p className="mt-4 text-sm text-dim">Loading</p>}
      {!loading && !error && items.length === 0 && <p className="mt-6 text-dim">{show === "missing" ? "No missing files." : show === "hidden" ? "No movies are hidden." : "No movies found. Check the movie folders in Settings and run a rescan."}</p>}
      {hasMore && <div ref={sentinel} className="h-8" aria-hidden />}
      {editing && <MovieEditor key={editing.id} movie={editing} onClose={() => setEditing(null)} onChanged={reload} />}
    </section>
  );
}
