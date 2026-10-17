/** Round-trip history. Failed probes show as red ticks at the bottom. */
export default function Sparkline({ values, className = "h-10 w-full" }: { values: (number | null)[]; className?: string }) {
  const W = 120, H = 32;
  const nums = values.filter((v): v is number => v !== null);
  const max = Math.max(100, ...nums);
  const step = values.length > 1 ? W / (values.length - 1) : W;
  const y = (v: number) => H - 2 - (v / max) * (H - 6);
  let path = "";
  let pen = false;
  values.forEach((v, i) => {
    if (v === null) { pen = false; return; }
    path += `${pen ? "L" : "M"}${(i * step).toFixed(1)} ${y(v).toFixed(1)} `;
    pen = true;
  });
  return (
    <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" className={className} aria-hidden>
      {path && <path d={path} fill="none" stroke="currentColor" strokeWidth="1.5" vectorEffect="non-scaling-stroke" strokeLinejoin="round" />}
      {values.map((v, i) => v === null && <rect key={i} x={i * step - 1} y={H - 4} width="2" height="4" className="fill-red-400" />)}
    </svg>
  );
}
