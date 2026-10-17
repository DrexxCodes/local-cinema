"use client";
import { RefObject, useCallback, useEffect, useState } from "react";
import { SKIP_SECONDS } from "@/constants/player";

export type PlayerState = {
  paused: boolean; current: number; duration: number; buffered: number;
  volume: number; muted: boolean; rate: number; waiting: boolean; ended: boolean;
};

const initial: PlayerState = { paused: true, current: 0, duration: 0, buffered: 0, volume: 1, muted: false, rate: 1, waiting: true, ended: false };

/** Mirrors the video element's state and exposes the actions the controls need. */
export function usePlayer(video: RefObject<HTMLVideoElement | null>, src: string) {
  const [st, setSt] = useState<PlayerState>(initial);

  useEffect(() => {
    const v = video.current;
    if (!v) return;
    const read = () => {
      let buffered = 0;
      for (let i = 0; i < v.buffered.length; i++) {
        if (v.buffered.start(i) <= v.currentTime + 0.5 && v.buffered.end(i) >= v.currentTime) buffered = v.buffered.end(i);
      }
      setSt((p) => ({
        ...p, paused: v.paused, current: v.currentTime, duration: Number.isFinite(v.duration) ? v.duration : 0, buffered,
        volume: v.volume, muted: v.muted, rate: v.playbackRate, ended: v.ended,
      }));
    };
    const waiting = () => setSt((p) => ({ ...p, waiting: true }));
    const playing = () => setSt((p) => ({ ...p, waiting: false }));
    const events = ["timeupdate", "progress", "durationchange", "play", "pause", "volumechange", "ratechange", "ended", "seeked", "loadedmetadata"];
    events.forEach((n) => v.addEventListener(n, read));
    v.addEventListener("waiting", waiting);
    v.addEventListener("seeking", waiting);
    ["playing", "canplay"].forEach((n) => v.addEventListener(n, playing));
    read();
    return () => {
      events.forEach((n) => v.removeEventListener(n, read));
      v.removeEventListener("waiting", waiting);
      v.removeEventListener("seeking", waiting);
      ["playing", "canplay"].forEach((n) => v.removeEventListener(n, playing));
    };
  }, [video, src]);

  const toggle = useCallback(() => {
    const v = video.current;
    if (!v) return;
    if (v.paused || v.ended) v.play().catch(() => {}); else v.pause();
  }, [video]);

  const seekTo = useCallback((t: number) => {
    const v = video.current;
    if (v) v.currentTime = Math.max(0, Math.min(t, v.duration || t));
  }, [video]);

  const seekBy = useCallback((d: number = SKIP_SECONDS) => {
    const v = video.current;
    if (v) v.currentTime = Math.max(0, Math.min(v.currentTime + d, v.duration || v.currentTime + d));
  }, [video]);

  const setVolume = useCallback((x: number) => {
    const v = video.current;
    if (v) { v.volume = x; v.muted = x === 0; }
  }, [video]);

  const toggleMute = useCallback(() => { const v = video.current; if (v) v.muted = !v.muted; }, [video]);
  const setRate = useCallback((r: number) => { const v = video.current; if (v) v.playbackRate = r; }, [video]);

  return { ...st, toggle, seekTo, seekBy, setVolume, toggleMute, setRate };
}
