import type { CSSProperties } from "react";

type Props = { value: number; max: number; buffered?: number; onChange: (v: number) => void; label: string; step?: number; className?: string; valueText?: string; tour?: string };

/** Range slider painted in the accent colour, with the buffered part in a lighter tone. */
export default function SeekBar({ value, max, buffered, onChange, label, step = 0.1, className = "", valueText, tour }: Props) {
  const top = max > 0 ? max : 1;
  const p = (Math.min(value, top) / top) * 100;
  const b = buffered === undefined ? p : Math.max(p, (Math.min(buffered, top) / top) * 100);
  return (
    <input type="range" className={`seek ${className}`} min={0} max={top} step={step} value={Math.min(value, top)}
      onChange={(e) => onChange(+e.target.value)} aria-label={label} aria-valuetext={valueText} data-tour={tour}
      style={{ "--p": `${p}%`, "--b": `${b}%` } as CSSProperties} />
  );
}
