"use client";
import { useSignal } from "@/hooks/useSignal";
import SignalBars from "./SignalBars";
import Sparkline from "./Sparkline";

const Stat = ({ label, value }: { label: string; value: string }) => (
  <div><dt className="text-xs text-dim">{label}</dt><dd className="text-sm tabular-nums">{value}</dd></div>
);

/** Detailed link quality for this device. */
export default function SignalPanel({ title = "Your connection", className = "" }: { title?: string; className?: string }) {
  const r = useSignal();
  return (
    <section className={`rounded-xl border border-edge bg-panel/80 p-4 ${className}`} aria-label={title}>
      <div className="flex items-center gap-3">
        <SignalBars level={r.level} ready={r.ready} className="h-5" />
        <div>
          <p className="text-sm font-medium">{r.label}</p>
          <p className="text-xs text-dim">{title}</p>
        </div>
      </div>
      <div className="mt-3 text-accent"><Sparkline values={r.history} /></div>
      <dl className="mt-3 grid grid-cols-3 gap-3">
        <Stat label="Latency" value={r.rtt === null ? "n/a" : `${r.rtt} ms`} />
        <Stat label="Jitter" value={`${r.jitter} ms`} />
        <Stat label="Lost" value={`${Math.round(r.loss * 100)}%`} />
      </dl>
    </section>
  );
}
