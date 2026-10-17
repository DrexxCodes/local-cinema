import type { AdminStats } from "@/types";

const Bar = ({ percent }: { percent: number }) => (
  <div className="mt-1.5 h-1.5 w-full max-w-md rounded-full bg-white/15">
    <div className="h-full rounded-full bg-accent transition-all" style={{ width: `${percent}%` }} />
  </div>
);

/** Live status of the scan, thumbnail and conversion jobs. Nothing shows while all are idle. */
export default function JobProgress({ stats, onCancelConvert }: { stats: AdminStats | null; onCancelConvert: () => void }) {
  if (!stats) return null;
  const { scan, thumbs, convert } = stats;
  return (
    <div className="mt-3 space-y-3 text-sm">
      {scan.running && <p>Scanning {scan.done} of {scan.total}: <span className="text-dim">{scan.current}</span></p>}
      {thumbs.running && <p>Thumbnails {thumbs.done} of {thumbs.total}: <span className="text-dim">{thumbs.current}</span></p>}
      {convert.running && (
        <div>
          <p className="flex flex-wrap items-center gap-3">
            <span>Converting {Math.min(convert.done + 1, convert.total)} of {convert.total}: <span className="text-dim">{convert.current}</span> {convert.percent ? `${convert.percent}%` : ""}</span>
            <button onClick={onCancelConvert} className="rounded border border-edge px-2.5 py-0.5 text-xs hover:border-red-400">Cancel</button>
          </p>
          <Bar percent={convert.percent ?? 0} />
        </div>
      )}
      {!scan.running && stats.scan_issues?.length > 0 && (
        <details className="rounded-lg border border-amber-400/40 p-3 text-amber-300">
          <summary className="cursor-pointer">The last scan skipped {stats.scan_issues.length} file{stats.scan_issues.length === 1 ? "" : "s"}</summary>
          <ul className="mt-2 space-y-1 break-all text-xs text-dim">{stats.scan_issues.map((i) => <li key={i}>{i}</li>)}</ul>
        </details>
      )}
      {!convert.running && stats.convert_error && <p role="alert" className="text-red-400">Conversion problem: {stats.convert_error}</p>}
    </div>
  );
}
