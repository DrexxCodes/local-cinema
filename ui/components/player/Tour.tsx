"use client";
import { RefObject, useEffect, useMemo, useState } from "react";
import { IOS_FULLSCREEN_NOTE, TOUR_STEPS } from "@/constants/tour";
import { isIOS } from "@/lib/device";
import { ChevronLeftIcon, ChevronRightIcon } from "@/components/ui/Icons";

type Props = { root: RefObject<HTMLElement | null>; onDone: () => void };

const find = (root: HTMLElement | null, key: string) => root?.querySelector<HTMLElement>(`[data-tour="${key}"]`) ?? null;
const visible = (el: HTMLElement | null) => !!el && el.getClientRects().length > 0;

/** Step by step tour of the player controls. The current control is lit up, the rest dimmed. */
export default function Tour({ root, onDone }: Props) {
  const steps = useMemo(() => {
    const touch = typeof matchMedia !== "undefined" && matchMedia("(pointer: coarse)").matches;
    return TOUR_STEPS.filter((s) => (s.touchOnly ? touch : true) && (!s.target || visible(find(root.current, s.target))));
  }, [root]);
  const [i, setI] = useState(0);
  const step = steps[i];
  const last = i === steps.length - 1;

  // Light up the control being explained.
  useEffect(() => {
    const el = step?.target ? find(root.current, step.target) : null;
    el?.classList.add("tour-spot");
    return () => el?.classList.remove("tour-spot");
  }, [step, root]);

  useEffect(() => {
    const key = (e: KeyboardEvent) => {
      if (e.key === "ArrowRight") setI((n) => Math.min(n + 1, steps.length - 1));
      else if (e.key === "ArrowLeft") setI((n) => Math.max(n - 1, 0));
      else if (e.key === "Escape") onDone();
      else return;
      e.preventDefault(); e.stopPropagation();
    };
    window.addEventListener("keydown", key, true);
    return () => window.removeEventListener("keydown", key, true);
  }, [steps.length, onDone]);

  if (!step) return null;
  const text = step.id === "fullscreen" && isIOS() ? step.text + IOS_FULLSCREEN_NOTE : step.text;

  return (
    <div className={`absolute inset-0 z-[90] ${step.target ? "" : "bg-black/70"}`} role="dialog" aria-label="Player guide" aria-live="polite">
      <div className="absolute left-1/2 top-[14%] w-[min(92%,24rem)] -translate-x-1/2 rounded-2xl border border-white/15 bg-panel p-5 shadow-2xl">
        <p className="text-xs text-dim">{i + 1} of {steps.length}</p>
        <h2 className="title-face mt-1 text-xl">{step.title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-ink/85">{text}</p>

        <div className="mt-4 flex gap-1.5" aria-hidden>
          {steps.map((s, n) => <span key={s.id} className={`h-1.5 flex-1 rounded-full ${n <= i ? "bg-accent" : "bg-white/20"}`} />)}
        </div>

        <div className="mt-4 flex items-center gap-2">
          <button onClick={onDone} className="mr-auto text-sm text-dim underline hover:text-ink">Skip tour</button>
          <button onClick={() => setI((n) => n - 1)} disabled={i === 0} aria-label="Previous"
            className="grid h-10 w-10 place-items-center rounded-full border border-white/20 disabled:opacity-30"><ChevronLeftIcon /></button>
          {last ? (
            <button onClick={onDone} className="rounded-full bg-accent px-5 py-2.5 text-sm font-bold">Done</button>
          ) : (
            <button onClick={() => setI((n) => n + 1)} aria-label="Next" className="grid h-10 w-10 place-items-center rounded-full bg-accent"><ChevronRightIcon /></button>
          )}
        </div>
      </div>
    </div>
  );
}
