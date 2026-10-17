"use client";
import { RefObject, useCallback, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { CONTROLS_HIDE_MS, DOUBLE_TAP_MS, SKIP_SECONDS } from "@/constants/player";
import { IOS_HINT_KEY } from "@/constants/tour";
import { useFirstTimeGuide } from "@/hooks/useFirstTimeGuide";
import { usePlayer } from "@/hooks/usePlayer";
import { withAuth } from "@/lib/api";
import { isIOS } from "@/lib/device";
import type { SubTrack } from "@/types";
import SignalBadge from "@/components/signal/SignalBadge";
import StreamButton from "@/components/stream/StreamButton";
import { PlayIcon, SpinnerIcon } from "@/components/ui/Icons";
import Controls from "./Controls";
import IosFullscreenHint from "./IosFullscreenHint";
import { OverlayContext } from "./OverlayContext";
import Tour from "./Tour";
import type { MenuKind } from "./PlayerMenu";

type Props = {
  movieId: string; title: string; src: string; poster?: string;
  videoRef: RefObject<HTMLVideoElement | null>;
  tracks: SubTrack[]; startAt: number;
  onPause: () => void; onEnded: () => void;
};

type Rot = 0 | 90 | 270;
const ROT_CLASS: Record<Rot, string> = { 0: "player-rot0", 90: "player-rot90", 270: "player-rot270" };

// Fraction (0 to 1) of the way along the screen's current left to right direction.
function along(e: React.PointerEvent, r: DOMRect, rot: Rot): number {
  if (rot === 90) return (e.clientY - r.top) / r.height;
  if (rot === 270) return (r.bottom - e.clientY) / r.height;
  return (e.clientX - r.left) / r.width;
}

export default function Player({ movieId, title, src, poster, videoRef, tracks, startAt, onPause, onEnded }: Props) {
  const box = useRef<HTMLDivElement>(null);
  const [root, setRoot] = useState<HTMLElement | null>(null);
  const setBox = useCallback((el: HTMLDivElement | null) => { box.current = el; setRoot(el); }, []);
  const pl = usePlayer(videoRef, src);
  const [visible, setVisible] = useState(true);
  const [menu, setMenu] = useState<MenuKind>(null);
  const [rot, setRot] = useState<Rot>(0);
  const [fullscreen, setFullscreen] = useState(false);
  const [selected, setSelected] = useState(() => tracks.find((t) => t.default)?.key ?? "");
  const [error, setError] = useState("");
  const [iosHint, setIosHint] = useState(false);
  const guide = useFirstTimeGuide(pl.duration > 0 && !error);
  const resumeAfterGuide = useRef(false);
  const resumeAfterModal = useRef(false);
  const [flash, setFlash] = useState<{ side: "left" | "right" | "center"; text: string } | null>(null);
  const hideTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const tap = useRef<{ t: number; timer: ReturnType<typeof setTimeout> | null }>({ t: 0, timer: null });

  const canFullscreen = typeof document !== "undefined" && (document.fullscreenEnabled || "webkitEnterFullscreen" in HTMLVideoElement.prototype);
  const canPip = typeof document !== "undefined" && !!document.pictureInPictureEnabled;
  const canAirplay = typeof HTMLVideoElement !== "undefined" && "webkitShowPlaybackTargetPicker" in HTMLVideoElement.prototype;

  // The default subtitle track arrives after the first render.
  useEffect(() => { const d = tracks.find((t) => t.default)?.key; if (d) setSelected((s) => s || d); }, [tracks]);

  const applyTrack = useCallback(() => {
    const v = videoRef.current;
    if (!v) return;
    Array.from(v.textTracks).forEach((t, i) => { t.mode = tracks[i]?.key === selected ? "showing" : "disabled"; });
  }, [videoRef, tracks, selected]);
  useEffect(() => { applyTrack(); }, [applyTrack]);

  // Show the controls, then hide them again after a while if the movie is playing.
  const wake = useCallback(() => {
    setVisible(true);
    if (hideTimer.current) clearTimeout(hideTimer.current);
    hideTimer.current = setTimeout(() => setVisible(false), CONTROLS_HIDE_MS);
  }, []);
  useEffect(() => {
    if (pl.paused || menu || guide.show) { setVisible(true); if (hideTimer.current) clearTimeout(hideTimer.current); }
    else wake();
    return () => { if (hideTimer.current) clearTimeout(hideTimer.current); };
  }, [pl.paused, menu, guide.show, wake]);

  // The movie waits while the guide is open and carries on afterwards.
  useEffect(() => {
    const v = videoRef.current;
    if (!v || !guide.show) return;
    resumeAfterGuide.current = !v.paused;
    v.pause();
    return () => { if (resumeAfterGuide.current) v.play().catch(() => {}); };
  }, [guide.show, videoRef]);

  // Dialogs opened from the controls (stream URL) pause the movie too.
  const onModal = useCallback((open: boolean) => {
    const v = videoRef.current;
    if (!v) return;
    if (open) { resumeAfterModal.current = !v.paused; v.pause(); }
    else if (resumeAfterModal.current) { resumeAfterModal.current = false; v.play().catch(() => {}); }
  }, [videoRef]);
  const overlay = useMemo(() => ({ root, onModal }), [root, onModal]);

  const say = (side: "left" | "right" | "center", text: string) => { setFlash({ side, text }); setTimeout(() => setFlash(null), 600); };

  function onSurface(e: React.PointerEvent) {
    setMenu(null);
    if (e.pointerType === "mouse") { pl.toggle(); wake(); return; }
    const now = Date.now();
    const frac = along(e, e.currentTarget.getBoundingClientRect(), rot);
    if (now - tap.current.t < DOUBLE_TAP_MS) {
      if (tap.current.timer) clearTimeout(tap.current.timer);
      tap.current.t = 0;
      if (frac < 0.35) { pl.seekBy(-SKIP_SECONDS); say("left", `-${SKIP_SECONDS}s`); }
      else if (frac > 0.65) { pl.seekBy(SKIP_SECONDS); say("right", `+${SKIP_SECONDS}s`); }
      else pl.toggle();
      return;
    }
    tap.current.t = now;
    tap.current.timer = setTimeout(() => { tap.current.t = 0; setVisible((v) => !v); }, DOUBLE_TAP_MS);
  }

  const toggleFullscreen = useCallback(() => {
    const v = videoRef.current as (HTMLVideoElement & { webkitEnterFullscreen?: () => void }) | null;
    if (document.fullscreenElement) { document.exitFullscreen().catch(() => {}); return; }
    if (document.fullscreenEnabled && box.current) {
      box.current.requestFullscreen().then(() => {
        const o = screen.orientation as ScreenOrientation & { lock?: (o: string) => Promise<void> };
        o?.lock?.("landscape").catch(() => {}); // only works on some Android browsers
      }).catch(() => {});
    } else if (isIOS() && localStorage.getItem(IOS_HINT_KEY) !== "1") setIosHint(true);
    else v?.webkitEnterFullscreen?.(); // iPhone: Apple's own player
  }, [videoRef]);

  const finishIosHint = (dontShowAgain: boolean) => {
    if (dontShowAgain) { try { localStorage.setItem(IOS_HINT_KEY, "1"); } catch { /* ignore */ } }
    setIosHint(false);
  };

  useEffect(() => {
    const on = () => { setFullscreen(!!document.fullscreenElement); if (!document.fullscreenElement) (screen.orientation as { unlock?: () => void })?.unlock?.(); };
    document.addEventListener("fullscreenchange", on);
    return () => document.removeEventListener("fullscreenchange", on);
  }, []);

  useEffect(() => {
    const key = (e: KeyboardEvent) => {
      const t = e.target as HTMLElement;
      if (guide.show || t.closest?.('[role="dialog"]')) return; // the guide and dialogs handle their own keys
      if (t.tagName === "INPUT" && (t as HTMLInputElement).type !== "range") return;
      const k = e.key.toLowerCase();
      if (k === " " || k === "k") { e.preventDefault(); pl.toggle(); }
      else if (k === "arrowleft" || k === "j") { pl.seekBy(-SKIP_SECONDS); say("left", `-${SKIP_SECONDS}s`); }
      else if (k === "arrowright" || k === "l") { pl.seekBy(SKIP_SECONDS); say("right", `+${SKIP_SECONDS}s`); }
      else if (k === "arrowup") { e.preventDefault(); pl.setVolume(Math.min(1, pl.volume + 0.1)); }
      else if (k === "arrowdown") { e.preventDefault(); pl.setVolume(Math.max(0, pl.volume - 0.1)); }
      else if (k === "m") pl.toggleMute();
      else if (k === "f") toggleFullscreen();
      else if (k === "c") setSelected((s) => (s ? "" : tracks[0]?.key ?? ""));
      else return;
      wake();
    };
    window.addEventListener("keydown", key);
    return () => window.removeEventListener("keydown", key);
  });

  function onVideoError() {
    const code = videoRef.current?.error?.code;
    if (code === 2) setError("Lost connection. Check your Wi-Fi, then reload to continue.");
    else if (code === 3 || code === 4) setError("This device can't play this movie's format.");
    else setError("This movie is no longer available.");
  }

  return (
    <OverlayContext.Provider value={overlay}>
    <div ref={setBox} className={`${ROT_CLASS[rot]} select-none overflow-hidden bg-black text-ink`} onPointerMove={(e) => { if (e.pointerType === "mouse") wake(); }}>
      {src && (
        <video ref={videoRef} src={src} poster={poster} autoPlay playsInline className="absolute inset-0 h-full w-full object-contain"
          onLoadedMetadata={() => { const v = videoRef.current; if (v && startAt) v.currentTime = startAt; applyTrack(); }}
          onPause={onPause} onEnded={onEnded} onError={onVideoError}>
          {tracks.map((t) => (
            <track key={t.key} kind="subtitles" srcLang={t.lang === "und" ? "en" : t.lang} label={t.label}
              src={withAuth(`/api/movies/${movieId}/subtitles/${encodeURIComponent(t.key)}`)} />
          ))}
        </video>
      )}
      <div className="absolute inset-0" onPointerUp={onSurface} />

      {pl.waiting && !error && <div className="pointer-events-none absolute inset-0 grid place-items-center text-accent"><SpinnerIcon /></div>}
      {flash && (
        <div className={`pointer-events-none absolute top-1/2 -translate-y-1/2 rounded-full bg-black/60 px-5 py-3 text-lg font-bold ${flash.side === "left" ? "left-[12%]" : flash.side === "right" ? "right-[12%]" : "left-1/2 -translate-x-1/2"}`}>{flash.text}</div>
      )}
      {pl.paused && !pl.waiting && !error && (
        <button onClick={() => { pl.toggle(); wake(); }} aria-label="Play"
          className="absolute left-1/2 top-1/2 grid h-20 w-20 -translate-x-1/2 -translate-y-1/2 place-items-center rounded-full bg-accent shadow-2xl shadow-accent/40"><PlayIcon className="h-9 w-9" /></button>
      )}

      <div className={`pointer-events-none absolute inset-0 flex flex-col justify-between transition-opacity duration-300 ${visible || error ? "opacity-100" : "opacity-0"}`}>
        <div className="pointer-events-auto flex items-center gap-3 bg-gradient-to-b from-black/80 to-transparent px-3 pb-8 pt-3 sm:px-6">
          <Link href={`/movie/?id=${movieId}`} className="rounded-full bg-black/50 px-4 py-1.5 text-sm backdrop-blur hover:bg-black/70">Back</Link>
          <p className="min-w-0 flex-1 truncate text-sm font-semibold">{title}</p>
          <SignalBadge className="rounded-full bg-black/50 backdrop-blur" />
        </div>
        {!error && (
          <Controls pl={pl} menu={menu} onMenu={setMenu} tracks={tracks} selected={selected} onSelect={setSelected} movieId={movieId}
            canFullscreen={canFullscreen} fullscreen={fullscreen} onFullscreen={toggleFullscreen}
            canRotate={!fullscreen} onRotate={() => setRot((r) => (r === 0 ? 90 : r === 90 ? 270 : 0))}
            onPip={canPip ? () => { videoRef.current?.requestPictureInPicture().catch(() => {}); setMenu(null); } : undefined}
            onAirplay={canAirplay ? () => { (videoRef.current as unknown as { webkitShowPlaybackTargetPicker: () => void }).webkitShowPlaybackTargetPicker(); setMenu(null); } : undefined} />
        )}
      </div>

      {guide.show && <Tour root={box} onDone={guide.finish} />}
      {iosHint && (
        <IosFullscreenHint onClose={() => setIosHint(false)}
          onContinue={(d) => { finishIosHint(d); (videoRef.current as (HTMLVideoElement & { webkitEnterFullscreen?: () => void }) | null)?.webkitEnterFullscreen?.(); }}
          onRotate={(d) => { finishIosHint(d); setRot(90); }} />
      )}

      {error && (
        <div role="alert" className="absolute inset-x-0 top-1/2 mx-auto flex max-w-sm -translate-y-1/2 flex-col items-center gap-4 px-6 text-center">
          <p className="text-dim">{error}</p>
          <p className="text-sm text-dim">Another app may be able to play it.</p>
          <StreamButton movieId={movieId} label="Stream on an external app" className="inline-flex items-center gap-2 rounded-md bg-accent px-5 py-2.5 font-medium" />
        </div>
      )}
    </div>
    </OverlayContext.Provider>
  );
}
