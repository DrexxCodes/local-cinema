import type { CompatStatus } from "@/types";

const LOOK: Record<CompatStatus, { text: string; cls: string } | null> = {
  ok: null,
  unknown: null,
  convert: { text: "May not play", cls: "bg-amber-400 text-black" },
  converted: { text: "Converted", cls: "bg-emerald-400 text-black" },
};

/** Small label on a movie tile. Movies that play everywhere get no label. */
export default function CompatBadge({ status, hidden, missing }: { status: CompatStatus; hidden?: boolean; missing?: boolean }) {
  const look = missing ? { text: "File missing", cls: "bg-red-500 text-white" }
    : hidden ? { text: "Hidden", cls: "bg-white text-black" }
    : LOOK[status];
  if (!look) return null;
  return <span className={`absolute bottom-1.5 left-1.5 rounded px-1.5 py-0.5 text-[10px] font-bold ${look.cls}`}>{look.text}</span>;
}
