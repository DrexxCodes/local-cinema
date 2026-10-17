"use client";
import Link from "next/link";
import { formatClock } from "@/lib/format";
import { artUrl } from "@/lib/movies";
import type { Movie, Progress } from "@/types";
import { CheckIcon, EyeIcon, HeartIcon, PlayIcon, PlusIcon } from "@/components/ui/Icons";

type Props = { movie: Movie; progress: Progress; inList: boolean; onToggleList: (id: string) => void };

/** Wide "now watching" card for the most recent unfinished movie. */
export default function ContinueBanner({ movie: m, progress: p, inList, onToggleList }: Props) {
  const share = m.duration_seconds ? Math.min(1, p.position_seconds / m.duration_seconds) : 0.1;
  const left = m.duration_seconds ? Math.max(0, m.duration_seconds - p.position_seconds) : 0;
  return (
    <section className="mt-12 px-4 sm:mt-16 sm:px-8 lg:px-16" aria-label="Continue watching">
      <div className="relative min-h-60 overflow-hidden rounded-2xl bg-panel shadow-2xl shadow-black/60 ring-1 ring-white/10 sm:aspect-[21/8]">
        {m.has_backdrop ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={artUrl(m, "backdrop")} alt="" className="absolute inset-0 h-full w-full object-cover" loading="lazy" />
        ) : <div className="absolute inset-0 bg-gradient-to-br from-accent/40 to-panel" />}
        <div className="absolute inset-0 bg-black/55" />
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-black/30" />

        <div className="absolute inset-x-0 top-0 flex items-start justify-between gap-4 p-5 sm:p-8">
          <div className="min-w-0">
            <h2 className="title-face truncate text-2xl sm:text-3xl">{m.title}{m.year > 0 ? ` (${m.year})` : ""}</h2>
            <p className="mt-1 flex items-center gap-2 text-sm text-dim"><EyeIcon /> Watching</p>
          </div>
          <div className="flex shrink-0 items-center gap-4 text-ink">
            <button onClick={() => onToggleList(m.id)} aria-pressed={inList} aria-label={inList ? "Remove from my list" : "Add to my list"}>
              <HeartIcon filled={inList} className={`h-6 w-6 ${inList ? "text-red-500" : ""}`} />
            </button>
            <button onClick={() => onToggleList(m.id)} aria-hidden tabIndex={-1} className="hidden sm:block">
              {inList ? <CheckIcon className="h-6 w-6" /> : <PlusIcon className="h-6 w-6" />}
            </button>
          </div>
        </div>

        <div className="absolute inset-x-0 bottom-0 flex items-center gap-3 p-4 sm:gap-5 sm:p-8">
          <Link href={`/watch/?id=${m.id}&resume=1`} aria-label={`Resume ${m.title}`}
            className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-accent shadow-lg shadow-accent/40"><PlayIcon className="h-5 w-5" /></Link>
          <span className="text-xs tabular-nums text-dim">{formatClock(p.position_seconds)}</span>
          <div className="relative h-1 flex-1 rounded-full bg-white/25">
            <div className="absolute inset-y-0 left-0 rounded-full bg-accent" style={{ width: `${share * 100}%` }} />
            <div className="absolute top-1/2 h-3.5 w-3.5 -translate-x-1/2 -translate-y-1/2 rounded-full bg-ink" style={{ left: `${share * 100}%` }} />
          </div>
          {left > 0 && <span className="text-xs tabular-nums text-dim">-{formatClock(left)}</span>}
        </div>
      </div>
    </section>
  );
}
