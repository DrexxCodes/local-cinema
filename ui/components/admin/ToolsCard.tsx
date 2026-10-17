"use client";
import { useCallback, useEffect, useState } from "react";
import { adminApi } from "@/lib/adminApi";
import type { ToolsInfo } from "@/types";
import { Notice } from "./ui";

/** FFmpeg status. It powers thumbnails, Make compatible and subtitle extraction. */
export default function ToolsCard() {
  const [t, setT] = useState<ToolsInfo | null>(null);
  const [err, setErr] = useState("");

  const load = useCallback(() => adminApi<ToolsInfo>("/api/admin/tools").then(setT).catch((e) => setErr(e.message)), []);
  useEffect(() => { load(); }, [load]);

  // While the download runs, refresh often.
  const running = !!t?.install.running;
  useEffect(() => {
    if (!running) return;
    const id = setInterval(load, 1000);
    return () => clearInterval(id);
  }, [running, load]);

  async function install() {
    setErr("");
    try { await adminApi("/api/admin/tools/install", { method: "POST" }); await load(); }
    catch (e) { setErr((e as Error).message); }
  }

  return (
    <section className="rounded-2xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur">
      <h2 className="title-face text-lg">FFmpeg</h2>
      <p className="mt-1 text-sm text-dim">Needed for thumbnails, Make compatible and subtitle extraction.</p>
      {t && (
        <div className="mt-3 text-sm">
          {t.ffmpeg && t.ffprobe ? (
            <>
              <p className="text-emerald-400">Installed</p>
              <p className="mt-1 break-all text-xs text-dim">{t.ffmpeg_version}</p>
              <p className="break-all text-xs text-dim">{t.ffmpeg_path}</p>
            </>
          ) : <p className="text-amber-400">{t.ffmpeg ? "ffprobe is missing" : "Not installed"}</p>}

          {running && (
            <div className="mt-3">
              <p>{t.install.current || "Working"} {t.install.percent ? `${t.install.percent}%` : ""}</p>
              <div className="mt-1.5 h-1.5 w-full max-w-sm rounded-full bg-white/15"><div className="h-full rounded-full bg-accent transition-all" style={{ width: `${t.install.percent ?? 0}%` }} /></div>
            </div>
          )}
          {!running && t.message && <p className="mt-2 text-dim">{t.message}</p>}

          {!(t.ffmpeg && t.ffprobe) && !running && (
            t.can_install
              ? <button onClick={install} className="mt-4 rounded-md bg-accent px-4 py-2 text-sm font-medium">Install FFmpeg</button>
              : <p className="mt-3 text-xs text-dim">Install FFmpeg with your package manager (for example <code>brew install ffmpeg</code>) and restart the app.</p>
          )}
        </div>
      )}
      {err && <Notice kind="error">{err}</Notice>}
    </section>
  );
}
