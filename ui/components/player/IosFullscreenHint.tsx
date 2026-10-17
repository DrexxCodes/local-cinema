"use client";
import { useState } from "react";

type Props = { onContinue: (dontShowAgain: boolean) => void; onRotate: (dontShowAgain: boolean) => void; onClose: () => void };

/** Explains iPhone full screen: Apple's own controls take over, and landscape needs orientation lock off. */
export default function IosFullscreenHint({ onContinue, onRotate, onClose }: Props) {
  const [again, setAgain] = useState(false);
  return (
    <div role="dialog" aria-modal aria-label="Full screen on iPhone" onClick={onClose}
      className="absolute inset-0 z-[100] flex overflow-y-auto bg-black/75 p-4 backdrop-blur-sm">
      <div onClick={(e) => e.stopPropagation()} className="m-auto w-full max-w-sm rounded-2xl border border-white/15 bg-panel p-5 text-ink shadow-2xl">
        <h2 className="title-face text-xl">Full screen on iPhone</h2>
        <p className="mt-3 text-sm leading-relaxed text-ink/85">
          iPhone full screen uses Apple&apos;s own player, so these controls are replaced by the phone&apos;s.
        </p>
        <p className="mt-3 text-sm leading-relaxed text-ink/85">
          To watch with your phone sideways, turn off <strong>Portrait Orientation Lock</strong>: open Control Center
          (swipe down from the top right corner) and tap the lock icon, then turn your phone.
        </p>
        <p className="mt-3 text-sm leading-relaxed text-dim">Prefer to keep these controls? Use Rotate instead, it turns the picture without leaving the page.</p>
        <label className="mt-4 flex items-center gap-2 text-sm text-dim">
          <input type="checkbox" checked={again} onChange={(e) => setAgain(e.target.checked)} className="accent-[var(--accent)]" /> Don&apos;t show this again
        </label>
        <div className="mt-4 flex flex-wrap gap-2">
          <button onClick={() => onContinue(again)} className="rounded-md bg-accent px-4 py-2 text-sm font-medium">Continue to full screen</button>
          <button onClick={() => onRotate(again)} className="rounded-md border border-white/20 px-4 py-2 text-sm">Use Rotate</button>
        </div>
      </div>
    </div>
  );
}
