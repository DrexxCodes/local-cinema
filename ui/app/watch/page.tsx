"use client";
import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { SAVE_EVERY_MS } from "@/constants/player";
import { api, withAuth } from "@/lib/api";
import { artUrl } from "@/lib/movies";
import { getSession } from "@/lib/session";
import { useGuest } from "@/hooks/useGuest";
import type { Movie, SubTrack } from "@/types";
import Player from "@/components/player/Player";

function Watch() {
  const session = useGuest();
  const params = useSearchParams();
  const id = params.get("id") ?? "";
  const resume = params.get("resume") === "1";
  const video = useRef<HTMLVideoElement>(null);
  const [ready, setReady] = useState(false);
  const [startAt, setStartAt] = useState(0);
  const [movie, setMovie] = useState<Movie | null>(null);
  const [tracks, setTracks] = useState<SubTrack[]>([]);

  const save = useCallback((completed = false) => {
    const v = video.current, s = getSession();
    if (!v || !s || !id || !v.currentTime) return;
    fetch(`/api/playback/${id}`, {
      method: "PUT", keepalive: true,
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${s.token}`, "X-Session-Id": s.session_id },
      body: JSON.stringify({ position_seconds: v.currentTime, completed }),
    }).catch(() => {});
  }, [id]);

  useEffect(() => {
    if (!session || !id) return;
    (async () => {
      if (resume) {
        const p = await api<{ position_seconds: number }>(`/api/playback/${id}`).catch(() => null);
        setStartAt(p?.position_seconds ?? 0);
      }
      api<Movie>(`/api/movies/${id}`).then(setMovie).catch(() => {});
      api<SubTrack[]>(`/api/movies/${id}/subtitles`).then(setTracks).catch(() => {});
      setReady(true);
    })();
  }, [session, id, resume]);

  // Save every 10 seconds and when leaving the page.
  useEffect(() => {
    const t = setInterval(() => { if (video.current && !video.current.paused) save(); }, SAVE_EVERY_MS);
    const onHide = () => save();
    window.addEventListener("pagehide", onHide);
    return () => { clearInterval(t); window.removeEventListener("pagehide", onHide); save(); };
  }, [save]);

  if (!ready) return <main className="grid min-h-screen place-items-center bg-black text-dim">Loading</main>;
  return (
    <main className="bg-black">
      <Player movieId={id} title={movie?.title ?? ""} src={withAuth(`/stream/${id}`)} poster={movie?.has_backdrop ? artUrl(movie, "backdrop") : undefined}
        videoRef={video} tracks={tracks} startAt={startAt} onPause={() => save()} onEnded={() => save(true)} />
    </main>
  );
}

export default function WatchPage() { return <Suspense><Watch /></Suspense>; }
