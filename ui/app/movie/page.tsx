"use client";
import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { api } from "@/lib/api";
import { formatClock, formatDuration, formatSize } from "@/lib/format";
import { artUrl } from "@/lib/movies";
import { useGuest } from "@/hooks/useGuest";
import type { Movie } from "@/types";
import StreamButton from "@/components/stream/StreamButton";

function Detail() {
  const session = useGuest();
  const id = useSearchParams().get("id") ?? "";
  const [m, setM] = useState<Movie | null>(null);
  const [pos, setPos] = useState(0);
  const [missing, setMissing] = useState(false);

  useEffect(() => {
    if (!session || !id) return;
    api<Movie>(`/api/movies/${id}`).then(setM).catch(() => setMissing(true));
    api<{ position_seconds: number; completed: boolean }>(`/api/playback/${id}`)
      .then((p) => setPos(!p.completed && p.position_seconds > 30 ? p.position_seconds : 0)).catch(() => {});
  }, [session, id]);

  if (missing) return <main className="grid page-fill place-items-center text-dim">This movie is no longer available.</main>;
  if (!m) return <main className="grid page-fill place-items-center text-dim">Loading</main>;

  return (
    <main className="relative page-fill">
      {m.has_backdrop && (
        // eslint-disable-next-line @next/next/no-img-element
        <img src={artUrl(m, "backdrop")} alt="" className="absolute inset-0 h-[60vh] w-full object-cover opacity-40" />
      )}
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-stage/80 to-stage" />
      <div className="relative px-4 pt-4 sm:px-10">
        <Link href="/" className="text-sm text-dim hover:text-ink">Back to the cinema</Link>
        <div className="mt-10 flex flex-col gap-8 sm:flex-row">
          <div className="aspect-[2/3] w-44 shrink-0 overflow-hidden rounded-md bg-panel ring-1 ring-edge sm:w-56">
            {m.has_poster && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={artUrl(m, "poster")} alt={m.title} className="h-full w-full object-cover" />
            )}
          </div>
          <div className="max-w-2xl">
            <h1 className="title-face text-4xl sm:text-5xl">{m.title}</h1>
            <p className="mt-2 text-dim">
              {[m.year || null, m.rating ? `${m.rating.toFixed(1)} rating` : null, formatDuration(m.duration_seconds) || null, m.genres.join(", ") || null].filter(Boolean).join(", ")}
            </p>
            {m.synopsis && <p className="mt-4 leading-relaxed text-ink/85">{m.synopsis}</p>}
            <div className="mt-6 flex flex-wrap gap-3">
              <Link href={`/watch/?id=${m.id}`} className="rounded-md bg-accent px-6 py-2.5 font-medium">{pos ? "Start over" : "Play"}</Link>
              {pos > 0 && <Link href={`/watch/?id=${m.id}&resume=1`} className="rounded-md border border-edge px-6 py-2.5">Resume from {formatClock(pos)}</Link>}
              <StreamButton movieId={m.id} />
            </div>
            <dl className="mt-10 grid max-w-md grid-cols-[8rem_1fr] gap-y-1 text-sm">
              <dt className="text-dim">Container</dt><dd>{m.container.toUpperCase()}</dd>
              <dt className="text-dim">Video</dt><dd>{[m.video_codec, m.resolution].filter(Boolean).join(" ") || "Unknown"}</dd>
              <dt className="text-dim">Audio</dt><dd>{m.audio_codec || "Unknown"}</dd>
              <dt className="text-dim">File size</dt><dd>{formatSize(m.file_size)}</dd>
            </dl>
          </div>
        </div>
      </div>
    </main>
  );
}

export default function MoviePage() { return <Suspense><Detail /></Suspense>; }
