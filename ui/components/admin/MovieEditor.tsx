"use client";
import { useRef, useState } from "react";
import { adminApi, adminImg, adminUpload } from "@/lib/adminApi";
import type { AdminMovie } from "@/types";
import SubtitleManager from "./SubtitleManager";
import { Field, fieldClass, Modal, Notice, Toggle } from "./ui";

type Props = { movie: AdminMovie; onClose: () => void; onChanged: () => void };

/** Edit everything TMDB would fill in: title, year, synopsis, genres, rating, pictures. */
export default function MovieEditor({ movie, onClose, onChanged }: Props) {
  const [m, setM] = useState(movie);
  const [title, setTitle] = useState(movie.title);
  const [year, setYear] = useState(movie.year ? String(movie.year) : "");
  const [synopsis, setSynopsis] = useState(movie.synopsis);
  const [genres, setGenres] = useState(movie.genres.join(", "));
  const [rating, setRating] = useState(movie.rating ? String(movie.rating) : "");
  const [hidden, setHidden] = useState(movie.hidden);
  const [busy, setBusy] = useState(false);
  const [bust, setBust] = useState(0);
  const [msg, setMsg] = useState<{ kind: "ok" | "error"; text: string } | null>(null);
  const poster = useRef<HTMLInputElement>(null);
  const backdrop = useRef<HTMLInputElement>(null);

  // Copies a server response into the form (after an upload or a TMDB lookup).
  const adopt = (n: AdminMovie) => {
    setM(n); setTitle(n.title); setYear(n.year ? String(n.year) : ""); setSynopsis(n.synopsis);
    setGenres(n.genres.join(", ")); setRating(n.rating ? String(n.rating) : ""); setHidden(n.hidden);
    setBust(Date.now()); onChanged();
  };

  async function guard<T>(fn: () => Promise<T>, ok: string, then?: (r: T) => void) {
    setBusy(true); setMsg(null);
    try { const r = await fn(); then?.(r); setMsg({ kind: "ok", text: ok }); }
    catch (e) { setMsg({ kind: "error", text: (e as Error).message }); }
    finally { setBusy(false); }
  }

  const save = (e: React.FormEvent) => {
    e.preventDefault();
    return guard(() => adminApi<AdminMovie>(`/api/admin/movies/${m.id}`, {
      method: "PATCH",
      body: JSON.stringify({
        title, year: Number(year) || 0, synopsis, hidden,
        genres: genres.split(",").map((g) => g.trim()).filter(Boolean),
        rating: Number(rating) || 0,
      }),
    }), "Saved.", adopt);
  };

  const lookup = () => guard(() => adminApi<AdminMovie>(`/api/admin/movies/${m.id}/tmdb`, { method: "POST" }), "Filled in from TMDB. Review and edit anything you like.", adopt);

  const upload = (kind: "poster" | "backdrop", file?: File) => file &&
    guard(() => adminUpload<AdminMovie>(`/api/admin/movies/${m.id}/art/${kind}`, file), `${kind === "poster" ? "Poster" : "Banner"} updated.`, adopt);

  return (
    <Modal title="Edit movie" onClose={onClose} wide>
      <form onSubmit={save} className="mt-4 space-y-4">
        <p className="break-all text-xs text-dim">File: {m.file_name}</p>

        <div className="flex gap-4">
          <div className="w-28 shrink-0">
            <div className="aspect-[2/3] overflow-hidden rounded-lg bg-gradient-to-br from-accent/40 to-panel">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              {m.has_poster && <img src={adminImg("posters", m.id, m.thumb_v, bust)} alt="" className="h-full w-full object-cover" />}
            </div>
            <input ref={poster} type="file" accept="image/jpeg,image/png,image/webp" hidden onChange={(e) => { upload("poster", e.target.files?.[0]); e.target.value = ""; }} />
            <button type="button" disabled={busy} onClick={() => poster.current?.click()} className="mt-2 w-full rounded border border-edge px-2 py-1 text-xs hover:border-accent">Upload poster</button>
          </div>
          <div className="min-w-0 flex-1">
            <div className="aspect-video overflow-hidden rounded-lg bg-gradient-to-br from-accent/40 to-panel">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              {m.has_backdrop && <img src={adminImg("backdrops", m.id, m.thumb_v, bust)} alt="" className="h-full w-full object-cover" />}
            </div>
            <input ref={backdrop} type="file" accept="image/jpeg,image/png,image/webp" hidden onChange={(e) => { upload("backdrop", e.target.files?.[0]); e.target.value = ""; }} />
            <button type="button" disabled={busy} onClick={() => backdrop.current?.click()} className="mt-2 rounded border border-edge px-3 py-1 text-xs hover:border-accent">Upload banner</button>
            <p className="mt-2 text-xs text-dim">The banner is the wide picture behind the movie. To use a frame from the film instead, use Thumbnails in the Library tab.</p>
          </div>
        </div>

        <Field label="Title"><input className={fieldClass} value={title} maxLength={200} onChange={(e) => setTitle(e.target.value)} required /></Field>
        <div className="grid grid-cols-2 gap-4">
          <Field label="Year (0 if unknown)"><input className={fieldClass} inputMode="numeric" value={year} onChange={(e) => setYear(e.target.value.replace(/\D/g, "").slice(0, 4))} /></Field>
          <Field label="Rating (0 to 10)"><input className={fieldClass} inputMode="decimal" value={rating} onChange={(e) => setRating(e.target.value.replace(/[^0-9.]/g, "").slice(0, 4))} /></Field>
        </div>
        <Field label="Genres (separated by commas)"><input className={fieldClass} value={genres} placeholder="Action, Drama" onChange={(e) => setGenres(e.target.value)} /></Field>
        <Field label="Synopsis"><textarea className={`${fieldClass} min-h-28`} value={synopsis} maxLength={5000} onChange={(e) => setSynopsis(e.target.value)} /></Field>
        <Toggle checked={!hidden} onChange={(v) => setHidden(!v)} label={hidden ? "Hidden from guests" : "Visible to guests"} />

        {msg && <Notice kind={msg.kind}>{msg.text}</Notice>}
        <div className="flex flex-wrap items-center gap-3 pt-1">
          <button disabled={busy} className="rounded-md bg-accent px-5 py-2 font-medium disabled:opacity-60">Save</button>
          <button type="button" onClick={onClose} className="rounded-md border border-edge px-5 py-2">Close</button>
          <button type="button" disabled={busy} onClick={lookup} className="ml-auto text-sm text-dim underline hover:text-ink">Look up on TMDB</button>
        </div>
        {m.edited && <p className="text-xs text-dim">Your edits are protected: rescans will not overwrite them.</p>}
      </form>
      <SubtitleManager movieId={m.id} />
    </Modal>
  );
}
