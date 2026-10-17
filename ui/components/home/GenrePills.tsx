"use client";
import { useRef } from "react";
import { ChevronRightIcon } from "@/components/ui/Icons";

type Props = { genres: string[]; chosen: string[]; onToggle: (g: string) => void };

/** Multi select genre chips. Chosen chips fill with the accent colour. */
export default function GenrePills({ genres, chosen, onToggle }: Props) {
  const row = useRef<HTMLDivElement>(null);
  if (!genres.length) return null;
  return (
    <div className="flex items-center gap-2">
      <div ref={row} className="row-scroll flex gap-2 overflow-x-auto py-1" role="group" aria-label="Genres">
        {genres.map((g) => {
          const on = chosen.includes(g);
          return (
            <button key={g} onClick={() => onToggle(g)} aria-pressed={on}
              className={`shrink-0 rounded-full px-5 py-1.5 text-xs font-semibold transition-colors ${on ? "bg-accent" : "bg-white/10 text-ink/90 hover:bg-white/20"}`}>
              {g}
            </button>
          );
        })}
      </div>
      <button onClick={() => row.current?.scrollBy({ left: 240, behavior: "smooth" })} aria-label="More genres"
        className="hidden shrink-0 text-dim hover:text-ink sm:block"><ChevronRightIcon className="h-4 w-4" /></button>
    </div>
  );
}
