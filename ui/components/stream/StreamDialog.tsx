"use client";
import { useContext, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { useExternalStream } from "@/hooks/useExternalStream";
import { copyText } from "@/lib/clipboard";
import { CloseIcon, SpinnerIcon } from "@/components/ui/Icons";
import { OverlayContext } from "@/components/player/OverlayContext";
import LimitPanel from "./LimitPanel";

function CopyRow({ label, value }: { label?: string; value: string }) {
  const [done, setDone] = useState(false);
  const input = useRef<HTMLInputElement>(null);
  async function copy() {
    input.current?.select();
    setDone(await copyText(value));
    setTimeout(() => setDone(false), 2000);
  }
  return (
    <div>
      {label && <p className="mb-1 text-xs text-dim">{label}</p>}
      <div className="flex gap-2">
        <input ref={input} readOnly value={value} onFocus={(e) => e.currentTarget.select()} aria-label={label ?? "Stream URL"}
          className="min-w-0 flex-1 rounded-md border border-edge bg-panel px-3 py-2 text-xs" />
        <button onClick={copy} className="shrink-0 rounded-md bg-accent px-4 py-2 text-sm font-medium">{done ? "Copied" : "Copy"}</button>
      </div>
    </div>
  );
}

/** Creates a stream URL for another app and shows it, or explains why it can't. */
export default function StreamDialog({ movieId, onClose }: { movieId: string; onClose: () => void }) {
  const { state, request, disconnect, sendUpgrade } = useExternalStream(movieId);
  const { root, onModal } = useContext(OverlayContext);
  useEffect(() => { request(); }, [request]);
  useEffect(() => { onModal?.(true); return () => onModal?.(false); }, [onModal]);
  useEffect(() => {
    const esc = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", esc);
    return () => window.removeEventListener("keydown", esc);
  }, [onClose]);

  // Rendered into the player (or the page) so no blur or transform on a parent can trap it.
  const dialog = (
    <div role="dialog" aria-modal aria-label="Stream on another app" onClick={onClose}
      className="fixed inset-0 z-[100] flex overflow-y-auto bg-black/75 p-4 backdrop-blur-sm">
      <div onClick={(e) => e.stopPropagation()} className="glass relative m-auto w-full max-w-md rounded-2xl bg-panel/95 p-6 text-ink shadow-2xl">
        <button onClick={onClose} aria-label="Close" className="absolute right-4 top-4 text-dim hover:text-ink"><CloseIcon /></button>
        <h2 className="title-face pr-8 text-xl">Stream on another app</h2>

        <div className="mt-4 space-y-4">
          {(state.kind === "loading" || state.kind === "idle") && <p className="flex items-center gap-3 text-dim"><SpinnerIcon className="h-5 w-5" /> Creating your link</p>}

          {state.kind === "ready" && (
            <>
              <CopyRow value={state.link.url} />
              <p className="text-sm text-dim">Paste it into VLC (Media, Open Network Stream), Infuse, nPlayer or any player on the same Wi-Fi. It uses one of this profile&apos;s connections until you stop watching.</p>
              {state.link.subtitles.length > 0 && (
                <details className="rounded-lg border border-edge p-3 text-sm">
                  <summary className="cursor-pointer">Subtitle links ({state.link.subtitles.length})</summary>
                  <div className="mt-3 space-y-3">{state.link.subtitles.map((s) => <CopyRow key={s.url} label={s.label} value={s.url} />)}</div>
                </details>
              )}
            </>
          )}

          {state.kind === "limit" && <LimitPanel limit={state.limit} onDisconnect={disconnect} onUpgrade={sendUpgrade} />}

          {state.kind === "denied" && <p className="text-dim">{state.message}</p>}

          {state.kind === "error" && (
            <div>
              <p role="alert" className="text-red-400">{state.message}</p>
              <button onClick={request} className="mt-3 rounded-md border border-edge px-4 py-2 text-sm">Try again</button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
  return createPortal(dialog, root ?? document.body);
}
