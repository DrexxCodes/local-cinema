"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { LANGUAGES } from "@/constants/languages";
import { adminApi, adminUpload } from "@/lib/adminApi";
import type { AdminStats, SubTrack } from "@/types";
import { fieldClass, Notice } from "./ui";

const SOURCE: Record<string, string> = { file: "Next to the movie", upload: "Uploaded", embedded: "From the video" };
const small = "rounded border border-edge px-2 py-1 text-xs hover:border-accent disabled:opacity-50";

/** Subtitle tracks of one movie: upload several, rename, choose the default, fix timing, extract from the video. */
export default function SubtitleManager({ movieId }: { movieId: string }) {
  const [tracks, setTracks] = useState<SubTrack[]>([]);
  const [msg, setMsg] = useState<{ kind: "ok" | "error"; text: string } | null>(null);
  const [lang, setLang] = useState("en");
  const [label, setLabel] = useState("English");
  const [extracting, setExtracting] = useState(false);
  const [busy, setBusy] = useState(false);
  const file = useRef<HTMLInputElement>(null);
  const base = `/api/admin/movies/${movieId}/subtitles`;

  const load = useCallback(() => adminApi<SubTrack[]>(base).then(setTracks).catch((e) => setMsg({ kind: "error", text: e.message })), [base]);
  useEffect(() => { load(); }, [load]);

  // While the video is being read, poll until the job ends.
  useEffect(() => {
    if (!extracting) return;
    const t = setInterval(async () => {
      const s = await adminApi<AdminStats>("/api/admin/stats").catch(() => null);
      if (s && !s.subs.running) { setExtracting(false); setMsg({ kind: "ok", text: s.subs_msg || "Done." }); load(); }
    }, 1500);
    return () => clearInterval(t);
  }, [extracting, load]);

  async function act(fn: () => Promise<unknown>, ok?: string) {
    setBusy(true); setMsg(null);
    try { await fn(); if (ok) setMsg({ kind: "ok", text: ok }); await load(); }
    catch (e) { setMsg({ kind: "error", text: (e as Error).message }); }
    finally { setBusy(false); }
  }

  const patch = (key: string, body: object, ok?: string) =>
    act(() => adminApi(`${base}/${key}`, { method: "PATCH", body: JSON.stringify(body) }), ok);

  const upload = (f?: File) => f && act(
    () => adminUpload(`${base}?lang=${encodeURIComponent(lang)}&label=${encodeURIComponent(label)}&name=${encodeURIComponent(f.name)}`, f, "POST"),
    "Subtitles added.");

  const extract = () => act(async () => {
    await adminApi(`${base}/extract`, { method: "POST" });
    setExtracting(true);
  });

  return (
    <section className="mt-8 border-t border-edge pt-5" aria-label="Subtitles">
      <h3 className="title-face text-lg">Subtitles</h3>
      <p className="mt-1 text-xs text-dim">Guests pick a track in the player. Files named like the movie (Movie.en.srt) in the movie folder show up here too.</p>

      <ul className="mt-3 space-y-2">
        {tracks.map((t) => (
          <li key={t.key} className="flex flex-wrap items-center gap-2 rounded-lg border border-edge px-3 py-2 text-sm">
            {t.editable ? (
              <input defaultValue={t.label} maxLength={40} aria-label="Track name" onBlur={(e) => { const v = e.target.value.trim(); if (v && v !== t.label) patch(t.key, { label: v }); }}
                className="w-40 rounded border border-edge bg-panel px-2 py-1" />
            ) : <span className="w-40 truncate">{t.label}</span>}
            <span className="text-xs text-dim">{t.lang} · {SOURCE[t.source]}</span>
            <label className="ml-auto flex items-center gap-1.5 text-xs text-dim">
              <input type="radio" name={`default-${movieId}`} checked={t.default} onChange={() => patch(t.key, { default: true })} className="accent-[var(--accent)]" /> Starts on
            </label>
            {t.editable && (
              <>
                <button disabled={busy} onClick={() => patch(t.key, { shift_ms: -500 })} className={small} title="Show subtitles half a second earlier">-0.5s</button>
                <button disabled={busy} onClick={() => patch(t.key, { shift_ms: 500 })} className={small} title="Show subtitles half a second later">+0.5s</button>
                <button disabled={busy} onClick={() => act(() => adminApi(`${base}/${t.key}`, { method: "DELETE" }))} className={`${small} hover:border-red-400`}>Delete</button>
              </>
            )}
          </li>
        ))}
        {tracks.length === 0 && <li className="text-sm text-dim">No subtitles yet.</li>}
      </ul>
      {tracks.some((t) => t.default) && <button onClick={() => { const d = tracks.find((t) => t.default); if (d) patch(d.key, { default: false }); }} className="mt-2 text-xs text-dim underline">Start with subtitles off</button>}

      <div className="mt-4 flex flex-wrap items-end gap-3">
        <label className="text-xs text-dim">Language
          <select value={lang} onChange={(e) => { setLang(e.target.value); setLabel(LANGUAGES.find((l) => l.code === e.target.value)?.name ?? e.target.value); }} className={`${fieldClass} mt-1 w-40`}>
            {LANGUAGES.map((l) => <option key={l.code} value={l.code}>{l.name}</option>)}
          </select>
        </label>
        <label className="text-xs text-dim">Name shown in the player
          <input value={label} maxLength={40} onChange={(e) => setLabel(e.target.value)} className={`${fieldClass} mt-1 w-44`} />
        </label>
        <input ref={file} type="file" accept=".srt,.vtt,.ass,.ssa,.sub" hidden onChange={(e) => { upload(e.target.files?.[0]); e.target.value = ""; }} />
        <button type="button" disabled={busy || !label.trim()} onClick={() => file.current?.click()} className="rounded-md bg-accent px-4 py-2 text-sm font-medium disabled:opacity-60">Upload subtitle file</button>
        <button type="button" disabled={busy || extracting} onClick={extract} className="rounded-md border border-edge px-4 py-2 text-sm disabled:opacity-60">{extracting ? "Reading the video" : "Extract from video"}</button>
      </div>
      <p className="mt-2 text-xs text-dim">Accepts .srt, .vtt, .ass and .ssa up to 2 MB. Extract pulls text subtitle tracks out of MKV and MP4 files.</p>
      {msg && <Notice kind={msg.kind}>{msg.text}</Notice>}
    </section>
  );
}
