"use client";
import Link from "next/link";
import { artUrl } from "@/lib/movies";
import type { Movie } from "@/types";
import { EyeIcon, HeartIcon, StarIcon } from "@/components/ui/Icons";

type Props = {
  m: Movie;
  /** 0 to 1 share watched, shows a bar under the poster. */
  progress?: number;
  watched?: boolean;
  inList?: boolean;
  onToggleList?: (id: string) => void;
  /** Fixed width for carousels; grids let the card fill its cell. */
  fixed?: boolean;
};

export default function MovieCard({ m, progress, watched, inList, onToggleList, fixed }: Props) {
  return (
    <div className={`group ${fixed ? "w-36 shrink-0 sm:w-44" : "w-full"}`}>
      <Link href={`/movie/?id=${m.id}`} className="block">
        <div className="relative aspect-[2/3] overflow-hidden rounded-xl bg-panel ring-1 ring-white/10 shadow-lg shadow-black/40 transition duration-300 group-hover:-translate-y-1 group-hover:ring-accent/60">
          {m.has_poster ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={artUrl(m, "poster")} alt={m.title} loading="lazy" className="h-full w-full object-cover" />
          ) : (
            <div className="flex h-full flex-col justify-end bg-gradient-to-br from-accent/40 to-panel p-3">
              <span className="title-face text-lg leading-tight">{m.title}</span>
              {m.year > 0 && <span className="text-xs text-dim">{m.year}</span>}
            </div>
          )}
          {progress !== undefined && progress > 0 && (
            <div className="absolute inset-x-0 bottom-0 h-1 bg-black/60">
              <div className="h-full bg-accent" style={{ width: `${Math.min(100, progress * 100)}%` }} />
            </div>
          )}
        </div>
        <p className="mt-2 truncate text-sm font-medium">{m.title}</p>
      </Link>
      <div className="mt-0.5 flex items-center gap-2 text-xs text-dim">
        <span>{m.year > 0 ? m.year : ""}</span>
        <span className="ml-auto flex items-center gap-2">
          {onToggleList && (
            <button onClick={() => onToggleList(m.id)} aria-pressed={inList} aria-label={inList ? "Remove from my list" : "Add to my list"}
              className={inList ? "text-red-500" : "hover:text-ink"}><HeartIcon filled={inList} className="h-3.5 w-3.5" /></button>
          )}
          <span className={watched ? "text-accent" : ""} title={watched ? "Watched" : undefined}><EyeIcon className="h-3.5 w-3.5" /></span>
          {m.rating > 0 && <span className="flex items-center gap-0.5 text-amber-400"><StarIcon className="h-3 w-3" />{m.rating.toFixed(1)}</span>}
        </span>
      </div>
    </div>
  );
}
