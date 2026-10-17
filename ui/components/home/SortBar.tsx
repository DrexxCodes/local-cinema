"use client";
import type { SortKey } from "@/types";
import { ChevronDownIcon, StarIcon } from "@/components/ui/Icons";

const SORTS: { key: SortKey; label: string }[] = [
  { key: "latest", label: "Latest" },
  { key: "year", label: "Year" },
  { key: "az", label: "A-Z" },
];

type Props = {
  sort: SortKey; reversed: boolean; onSort: (k: SortKey) => void;
  minRating: number; onRating: (n: number) => void; showRating: boolean;
};

/** Sort buttons (tap the active one to flip the order) and a minimum rating slider. */
export default function SortBar({ sort, reversed, onSort, minRating, onRating, showRating }: Props) {
  return (
    <div className="flex flex-wrap items-center gap-x-6 gap-y-3 text-sm">
      <div className="flex items-center gap-2">
        <span className="text-dim">Sort by:</span>
        {SORTS.map((s) => {
          const on = sort === s.key;
          return (
            <button key={s.key} onClick={() => onSort(s.key)} aria-pressed={on}
              className={`flex items-center gap-1 rounded-full px-4 py-1.5 text-xs font-semibold ${on ? "bg-accent" : "bg-white/10 hover:bg-white/20"}`}>
              {s.label}
              {on && <ChevronDownIcon className={`h-3 w-3 transition-transform ${reversed ? "rotate-180" : ""}`} />}
            </button>
          );
        })}
      </div>
      {showRating && (
        <label className="ml-auto flex items-center gap-3 text-dim">
          <StarIcon className="h-4 w-4 text-amber-400" />
          <input type="range" min={0} max={10} step={0.5} value={minRating} onChange={(e) => onRating(+e.target.value)}
            aria-label="Minimum rating" className="w-32 accent-[var(--accent)] sm:w-48" />
          <span className="w-8 text-right tabular-nums text-ink">{minRating.toFixed(1)}</span>
        </label>
      )}
    </div>
  );
}
