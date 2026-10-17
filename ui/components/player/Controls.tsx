"use client";
import { SKIP_SECONDS } from "@/constants/player";
import { formatClock } from "@/lib/format";
import type { usePlayer } from "@/hooks/usePlayer";
import type { SubTrack } from "@/types";
import {
  Back10Icon, CaptionsIcon, ExitFullscreenIcon, Forward10Icon, FullscreenIcon, MoreIcon, MuteIcon,
  PauseIcon, PlayIcon, RotateIcon, SpeedIcon, VolumeIcon,
} from "@/components/ui/Icons";
import PlayerMenu, { MenuKind } from "./PlayerMenu";
import SeekBar from "./SeekBar";

type Props = {
  pl: ReturnType<typeof usePlayer>;
  menu: MenuKind; onMenu: (m: MenuKind) => void;
  tracks: SubTrack[]; selected: string; onSelect: (key: string) => void;
  movieId: string;
  canFullscreen: boolean; fullscreen: boolean; onFullscreen: () => void;
  canRotate: boolean; onRotate: () => void;
  onPip?: () => void; onAirplay?: () => void;
};

const btn = "grid h-10 w-10 place-items-center rounded-full text-ink/90 transition-colors hover:bg-white/15 hover:text-accent";

export default function Controls(p: Props) {
  const { pl } = p;
  const pick = (m: MenuKind) => p.onMenu(p.menu === m ? null : m);
  return (
    <div className="pointer-events-auto bg-gradient-to-t from-black/85 via-black/40 to-transparent px-3 pb-3 pt-10 sm:px-6 sm:pb-5">
      <div className="flex items-center gap-3">
        <span className="w-12 text-right text-xs tabular-nums text-ink/90">{formatClock(pl.current)}</span>
        <SeekBar value={pl.current} max={pl.duration} buffered={pl.buffered} onChange={pl.seekTo} label="Seek"
          valueText={`${formatClock(pl.current)} of ${formatClock(pl.duration)}`} className="flex-1" tour="seek" />
        <span className="w-14 text-xs tabular-nums text-ink/70">-{formatClock(Math.max(0, pl.duration - pl.current))}</span>
      </div>

      <div className="relative mt-1 flex items-center gap-1">
        <button data-tour="play" onClick={pl.toggle} aria-label={pl.paused ? "Play" : "Pause"} className={`${btn} bg-accent hover:bg-accent hover:text-on-accent`}>
          {pl.paused ? <PlayIcon className="h-5 w-5" /> : <PauseIcon className="h-5 w-5" />}
        </button>
        <button data-tour="back" onClick={() => pl.seekBy(-SKIP_SECONDS)} aria-label={`Back ${SKIP_SECONDS} seconds`} className={btn}><Back10Icon /></button>
        <button data-tour="forward" onClick={() => pl.seekBy(SKIP_SECONDS)} aria-label={`Forward ${SKIP_SECONDS} seconds`} className={btn}><Forward10Icon /></button>
        <div data-tour="volume" className="group hidden items-center sm:flex">
          <button onClick={pl.toggleMute} aria-label={pl.muted ? "Unmute" : "Mute"} className={btn}>{pl.muted || pl.volume === 0 ? <MuteIcon /> : <VolumeIcon />}</button>
          <SeekBar value={pl.muted ? 0 : pl.volume} max={1} step={0.05} onChange={pl.setVolume} label="Volume" className="w-0 opacity-0 transition-all group-hover:w-20 group-hover:opacity-100 group-focus-within:w-20 group-focus-within:opacity-100" />
        </div>

        <span className="ml-auto" />
        <button data-tour="speed" onClick={() => pick("speed")} aria-label="Playback speed" aria-expanded={p.menu === "speed"} className={`${btn} ${pl.rate !== 1 ? "text-accent" : ""}`}><SpeedIcon /></button>
        <button data-tour="captions" onClick={() => pick("captions")} aria-label="Subtitles" aria-expanded={p.menu === "captions"} className={`${btn} ${p.selected ? "text-accent" : ""}`}><CaptionsIcon /></button>
        {p.canRotate && <button data-tour="rotate" onClick={p.onRotate} aria-label="Rotate screen" className={btn}><RotateIcon /></button>}
        {p.canFullscreen && <button data-tour="fullscreen" onClick={p.onFullscreen} aria-label={p.fullscreen ? "Exit full screen" : "Full screen"} className={btn}>{p.fullscreen ? <ExitFullscreenIcon /> : <FullscreenIcon />}</button>}
        <button data-tour="more" onClick={() => pick("more")} aria-label="More" aria-expanded={p.menu === "more"} className={btn}><MoreIcon /></button>

        <PlayerMenu menu={p.menu} rate={pl.rate} onRate={(r) => { pl.setRate(r); p.onMenu(null); }}
          tracks={p.tracks} selected={p.selected} onSelect={(k) => { p.onSelect(k); p.onMenu(null); }}
          movieId={p.movieId} onAirplay={p.onAirplay} onPip={p.onPip} />
      </div>
    </div>
  );
}
