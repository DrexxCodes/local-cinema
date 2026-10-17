"use client";
import { useEffect, useRef, useState } from "react";
import { useSignal } from "@/hooks/useSignal";
import SignalBars from "./SignalBars";
import SignalPanel from "./SignalPanel";

/** Signal bars with the latency beside them. Tap for the detailed panel. */
export default function SignalBadge({ className = "", align = "right" }: { className?: string; align?: "left" | "right" }) {
  const r = useSignal();
  const [open, setOpen] = useState(false);
  const box = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const close = (e: PointerEvent) => { if (!box.current?.contains(e.target as Node)) setOpen(false); };
    document.addEventListener("pointerdown", close);
    return () => document.removeEventListener("pointerdown", close);
  }, [open]);

  return (
    <div ref={box} className={`relative ${className}`}>
      <button onClick={() => setOpen((o) => !o)} aria-expanded={open} aria-label={`Signal strength: ${r.label}`}
        title={r.rtt === null ? r.label : `${r.label}, ${r.rtt} ms`}
        className="flex items-center gap-2 rounded-full px-2 py-1.5 text-xs text-dim hover:text-ink">
        <SignalBars level={r.level} ready={r.ready} />
        <span className="hidden tabular-nums md:inline">{r.rtt === null ? (r.ready ? "offline" : "...") : `${r.rtt} ms`}</span>
      </button>
      {open && (
        <div className={`absolute top-full z-50 mt-2 w-64 ${align === "right" ? "right-0" : "left-0"}`}>
          <SignalPanel />
        </div>
      )}
    </div>
  );
}
