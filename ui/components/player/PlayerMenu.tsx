"use client";
import { SPEEDS } from "@/constants/player";
import type { SubTrack } from "@/types";
import StreamButton from "@/components/stream/StreamButton";
import { CheckIcon } from "@/components/ui/Icons";

export type MenuKind = "speed" | "captions" | "more" | null;

type Props = {
  menu: MenuKind; rate: number; onRate: (r: number) => void;
  tracks: SubTrack[]; selected: string; onSelect: (key: string) => void;
  movieId: string; onAirplay?: () => void; onPip?: () => void;
};

const Item = ({ active, onClick, children }: { active?: boolean; onClick: () => void; children: React.ReactNode }) => (
  <button onClick={onClick} className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm hover:bg-white/10">
    <span className="w-4">{active && <CheckIcon className="h-4 w-4 text-accent" />}</span>{children}
  </button>
);

/** Popover above the control bar. */
export default function PlayerMenu({ menu, rate, onRate, tracks, selected, onSelect, movieId, onAirplay, onPip }: Props) {
  if (!menu) return null;
  return (
    <div className="pointer-events-auto absolute bottom-full right-0 mb-2 max-h-64 w-56 overflow-y-auto rounded-xl border border-white/10 bg-panel/95 p-1.5 shadow-2xl backdrop-blur-xl" role="menu">
      {menu === "speed" && SPEEDS.map((r) => <Item key={r} active={r === rate} onClick={() => onRate(r)}>{r === 1 ? "Normal" : `${r}x`}</Item>)}
      {menu === "captions" && (
        <>
          <Item active={selected === ""} onClick={() => onSelect("")}>Off</Item>
          {tracks.map((t) => <Item key={t.key} active={selected === t.key} onClick={() => onSelect(t.key)}>{t.label}</Item>)}
          {tracks.length === 0 && <p className="px-3 py-2 text-sm text-dim">No subtitles for this movie.</p>}
        </>
      )}
      {menu === "more" && (
        <div className="space-y-1 p-1">
          {onPip && <button onClick={onPip} className="w-full rounded-lg px-3 py-2 text-left text-sm hover:bg-white/10">Picture in picture</button>}
          {onAirplay && <button onClick={onAirplay} className="w-full rounded-lg px-3 py-2 text-left text-sm hover:bg-white/10">AirPlay</button>}
          <StreamButton movieId={movieId} label="Copy stream URL"
            className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm hover:bg-white/10" />
        </div>
      )}
    </div>
  );
}
