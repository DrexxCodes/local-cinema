"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { HERO_INTERVAL_MS } from "@/constants/home";
import { formatDuration } from "@/lib/format";
import { artUrl } from "@/lib/movies";
import type { Movie } from "@/types";
import { CheckIcon, PlayIcon, PlusIcon, StarIcon } from "@/components/ui/Icons";

type Props = { movies: Movie[]; inList: (id: string) => boolean; onToggleList: (id: string) => void };

export default function Hero({ movies, inList, onToggleList }: Props) {
  const [i, setI] = useState(0);
  const count = movies.length;

  useEffect(() => {
    if (count < 2) return;
    const t = setTimeout(() => setI((n) => (n + 1) % count), HERO_INTERVAL_MS);
    return () => clearTimeout(t); // restarts whenever the slide changes, also after a manual pick
  }, [i, count]);

  const m = movies[Math.min(i, count - 1)];
  if (!m) return null;
  const saved = inList(m.id);
  const duration = formatDuration(m.duration_seconds);

  return (
    <section className="relative h-[68vh] min-h-[26rem] w-full overflow-hidden bg-panel sm:h-[78vh]" aria-roledescription="carousel">
      {movies.map((s, idx) => (
        <div key={s.id} className={`absolute inset-0 transition-opacity duration-700 ${idx === i ? "opacity-100" : "opacity-0"}`} aria-hidden={idx !== i}>
          {s.has_backdrop ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={artUrl(s, "backdrop")} alt="" className="h-full w-full object-cover" loading={idx === 0 ? "eager" : "lazy"} />
          ) : (
            <div className="h-full w-full bg-gradient-to-br from-accent/40 via-panel to-stage" />
          )}
        </div>
      ))}
      <div className="absolute inset-0 bg-gradient-to-r from-stage/90 via-stage/30 to-transparent" />
      <div className="absolute inset-0 bg-gradient-to-t from-stage via-transparent to-black/40" />

      <div className="absolute inset-x-0 bottom-0 px-4 pb-16 sm:px-8 sm:pb-24 lg:px-16">
        <div className="max-w-xl">
          <p className="text-xs text-ink/80 sm:text-sm">{duration ? `Duration: ${duration}` : ""}</p>
          <p className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-ink/80 sm:text-sm">
            {m.rating > 0 && <span className="flex items-center gap-1 font-semibold"><StarIcon className="h-4 w-4 text-amber-400" />{m.rating.toFixed(1)}</span>}
            {m.year > 0 && <span>{m.year}</span>}
            {m.genres.length > 0 && <span className="text-dim">{m.genres.slice(0, 3).join(" | ")}</span>}
          </p>
          <h1 className="title-face mt-2 text-4xl sm:text-6xl">{m.title}</h1>
          {m.synopsis && <p className="mt-3 line-clamp-3 max-w-md text-sm leading-relaxed text-ink/85">{m.synopsis}</p>}
          <div className="mt-6 flex flex-wrap items-center gap-3">
            <Link href={`/watch/?id=${m.id}`} className="inline-flex items-center gap-2 rounded-full bg-accent px-6 py-3 text-sm font-bold uppercase tracking-wide shadow-lg shadow-accent/30">
              <PlayIcon /> Watch
            </Link>
            <button onClick={() => onToggleList(m.id)} aria-pressed={saved}
              className="inline-flex items-center gap-2 rounded-full border border-white/20 bg-black/50 px-6 py-3 text-sm font-bold uppercase tracking-wide backdrop-blur hover:bg-white/10">
              {saved ? <CheckIcon /> : <PlusIcon />} {saved ? "In list" : "Add list"}
            </button>
          </div>
        </div>
      </div>

      {count > 1 && (
        <div className="absolute inset-x-0 bottom-6 flex justify-center gap-2 sm:bottom-10" role="tablist" aria-label="Featured movies">
          {movies.map((s, idx) => (
            <button key={s.id} role="tab" aria-selected={idx === i} aria-label={`Show ${s.title}`} onClick={() => setI(idx)}
              className="group py-2">
              <span className={`block h-[3px] w-10 rounded-full transition-colors sm:w-14 ${idx === i ? "bg-ink" : "bg-white/30 group-hover:bg-white/60"}`} />
            </button>
          ))}
        </div>
      )}
    </section>
  );
}
