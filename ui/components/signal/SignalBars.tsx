import type { SignalLevel } from "@/types";

const COLORS: Record<number, string> = { 4: "bg-emerald-400", 3: "bg-emerald-400", 2: "bg-amber-400", 1: "bg-red-400", 0: "bg-red-400" };

/** Four ascending bars like a phone signal indicator. */
export default function SignalBars({ level, ready = true, className = "" }: { level: SignalLevel; ready?: boolean; className?: string }) {
  const lit = ready ? level : 0;
  return (
    <span className={`inline-flex h-4 items-end gap-[3px] ${className}`} aria-hidden>
      {[1, 2, 3, 4].map((n) => (
        <span key={n} style={{ height: `${n * 25}%` }}
          className={`w-[3px] rounded-sm ${n <= lit ? COLORS[lit] : "bg-white/25"}`} />
      ))}
    </span>
  );
}
