"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import PasswordInput from "@/components/ui/PasswordInput";

type Folder = { path: string; valid?: boolean; count?: number; bytes?: number; warnings?: string[] };

const input = "mt-1 w-full rounded-md border border-edge bg-panel px-3 py-2.5";
const label = "mt-4 block text-sm text-dim";

export default function Setup() {
  const router = useRouter();
  const [step, setStep] = useState(1);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState("");
  const [name, setName] = useState("My Cinema");
  const [tagline, setTagline] = useState("");
  const [accent, setAccent] = useState("#7C3AED");
  const [folders, setFolders] = useState<Folder[]>([]);
  const [draft, setDraft] = useState("");
  const [email, setEmail] = useState("");
  const [pw, setPw] = useState("");
  const [pw2, setPw2] = useState("");
  const [maxStreams, setMaxStreams] = useState(10);
  const [mdns, setMdns] = useState("cinema");
  const [tmdb, setTmdb] = useState("");
  const [ips, setIps] = useState<{ name: string; ip: string }[]>([]);
  const [scan, setScan] = useState<{ running: boolean; total: number; done: number } | null>(null);

  useEffect(() => {
    api("/api/setup/status").then((s: any) => {
      if (s.setup_complete) { router.replace("/login/"); return; }
      setStep(Math.min(Math.max(s.setup_step || 1, 1), 4));
      setName(s.cinema_name || "My Cinema"); setTagline(s.tagline || "");
      setAccent(s.accent_color || "#7C3AED");
      setFolders((s.movie_folders || []).map((p: string) => ({ path: p })));
      setMaxStreams(s.global_max_streams || 10); setMdns(s.mdns_name || "cinema");
      setReady(true);
    }).catch((e) => setError(e.message));
    api("/api/setup/interfaces").then(setIps).catch(() => {});
  }, [router]);

  useEffect(() => { document.documentElement.style.setProperty("--accent", accent); }, [accent]);

  useEffect(() => {
    if (step !== 5) return;
    const t = setInterval(() => api("/api/setup/scan-status").then(setScan).catch(() => {}), 1000);
    return () => clearInterval(t);
  }, [step]);

  async function save(next: number, data: Record<string, unknown>) {
    setError("");
    try { await api("/api/setup/step", { method: "POST", body: JSON.stringify({ step: next, data }) }); setStep(next); }
    catch (e) { setError((e as ApiError).message); }
  }

  async function addFolder() {
    const path = draft.trim();
    if (!path) return;
    const v = await api<any>("/api/setup/validate-folder", { method: "POST", body: JSON.stringify({ path }) });
    setFolders((f) => [...f.filter((x) => x.path !== path), { path, valid: v.valid, count: v.video_count, bytes: v.total_bytes, warnings: v.warnings }]);
    setDraft("");
  }

  async function finish() {
    if (pw.length < 8) return setError("Use a password with at least 8 characters.");
    if (pw !== pw2) return setError("The passwords do not match.");
    setError("");
    try {
      await api("/api/setup/step", { method: "POST", body: JSON.stringify({ step: 4, data: { global_max_streams: maxStreams, mdns_name: mdns, tmdb_api_key: tmdb } }) });
      await api("/api/setup/complete", { method: "POST", body: JSON.stringify({ admin_email: email, admin_password: pw }) });
      setStep(5);
    } catch (e) { setError((e as ApiError).message); }
  }

  if (!ready) return <main className="grid page-fill place-items-center text-dim">{error || "Loading"}</main>;

  const validFolders = folders.filter((f) => f.valid !== false);
  const lan = ips[0]?.ip ?? "your-ip";
  const port = typeof window !== "undefined" ? window.location.port || "80" : "8765";

  return (
    <main className="mx-auto grid page-fill max-w-5xl gap-10 p-6 md:grid-cols-2 md:items-center">
      <section>
        <p className="text-sm text-dim">Step {Math.min(step, 4)} of 4</p>
        <div className="mt-2 flex gap-1" aria-hidden>{[1, 2, 3, 4].map((n) => <span key={n} className={`h-1 flex-1 rounded ${n <= step ? "bg-accent" : "bg-edge"}`} />)}</div>

        {step === 1 && (<>
          <h1 className="title-face mt-6 text-3xl">Name your cinema</h1>
          <label className={label} htmlFor="n">Cinema name</label>
          <input id="n" className={input} value={name} maxLength={50} onChange={(e) => setName(e.target.value)} />
          <label className={label} htmlFor="t">Tagline (optional)</label>
          <input id="t" className={input} value={tagline} maxLength={120} onChange={(e) => setTagline(e.target.value)} />
          <label className={label} htmlFor="a">Accent color</label>
          <input id="a" type="color" value={accent} onChange={(e) => setAccent(e.target.value)} className="mt-1 h-10 w-20 rounded border border-edge bg-panel" />
          <Nav onNext={() => name.trim() ? save(2, { cinema_name: name.trim(), tagline, accent_color: accent }) : setError("Give your cinema a name.")} />
        </>)}

        {step === 2 && (<>
          <h1 className="title-face mt-6 text-3xl">Where are your movies?</h1>
          <label className={label} htmlFor="f">Folder path on this computer</label>
          <div className="mt-1 flex gap-2">
            <input id="f" className={input.replace("mt-1 ", "")} value={draft} placeholder="D:\Movies or /mnt/movies" onChange={(e) => setDraft(e.target.value)} onKeyDown={(e) => e.key === "Enter" && addFolder()} />
            <button onClick={addFolder} className="rounded-md border border-edge px-4">Add</button>
          </div>
          <ul className="mt-4 space-y-2">
            {folders.map((f) => (
              <li key={f.path} className="rounded-md border border-edge bg-panel p-3 text-sm">
                <div className="flex justify-between gap-2"><span className="break-all">{f.path}</span>
                  <button className="text-dim" onClick={() => setFolders(folders.filter((x) => x.path !== f.path))}>Remove</button></div>
                {f.valid !== undefined && <p className={f.valid && f.count ? "text-dim" : "text-amber-400"}>
                  {f.valid ? `${f.count} videos, ${((f.bytes ?? 0) / 1e9).toFixed(1)} GB` : ""} {f.warnings?.join(" ")}</p>}
              </li>
            ))}
          </ul>
          <Nav onBack={() => setStep(1)} onNext={() => validFolders.length ? save(3, { movie_folders: validFolders.map((f) => f.path) }) : setError("Add at least one valid folder.")} />
        </>)}

        {step === 3 && (<>
          <h1 className="title-face mt-6 text-3xl">Create the admin account</h1>
          <label className={label} htmlFor="e">Email</label>
          <input id="e" type="email" className={input} value={email} onChange={(e) => setEmail(e.target.value)} />
          <label className={label} htmlFor="p">Password (8+ characters)</label>
          <PasswordInput id="p" className={input} value={pw} onChange={(e) => setPw(e.target.value)} />
          <label className={label} htmlFor="p2">Confirm password</label>
          <PasswordInput id="p2" className={input} value={pw2} onChange={(e) => setPw2(e.target.value)} />
          <Nav onBack={() => setStep(2)} onNext={() => !email.includes("@") ? setError("Enter a valid email.") : pw.length < 8 ? setError("Use a password with at least 8 characters.") : pw !== pw2 ? setError("The passwords do not match.") : (setError(""), save(4, {}))} />
        </>)}

        {step === 4 && (<>
          <h1 className="title-face mt-6 text-3xl">Limits and discovery</h1>
          <label className={label} htmlFor="g">Maximum devices watching at once</label>
          <input id="g" type="number" min={1} className={input} value={maxStreams} onChange={(e) => setMaxStreams(+e.target.value)} />
          {maxStreams > 20 && <p className="mt-1 text-sm text-amber-400">Above 20 streams, your WiFi is usually the bottleneck.</p>}
          <label className={label} htmlFor="m">Network name (guests open http://{mdns}.local)</label>
          <input id="m" className={input} value={mdns} onChange={(e) => setMdns(e.target.value.replace(/[^a-z0-9-]/gi, "").toLowerCase())} />
          <label className={label} htmlFor="k">TMDB API key (optional, adds posters and synopses)</label>
          <input id="k" className={input} value={tmdb} onChange={(e) => setTmdb(e.target.value)} />
          <Nav onBack={() => setStep(3)} nextLabel="Finish setup" onNext={finish} />
        </>)}

        {step === 5 && (<>
          <h1 className="title-face mt-6 text-3xl">{name} is ready</h1>
          <dl className="mt-4 space-y-2 text-sm">
            <div><dt className="text-dim">Guest address</dt><dd>http://{mdns}.local:{port} or http://{lan}:{port}</dd></div>
            <div><dt className="text-dim">Admin panel</dt><dd>http://{lan}:{port}/admin/</dd></div>
          </dl>
          {scan && <p className="mt-4 text-dim">{scan.running ? `Scanning: ${scan.done} of ${scan.total}` : `Scan finished, ${scan.total} movies found.`}</p>}
          <div className="mt-6 flex gap-3">
            <a href="/admin/" className="rounded-md bg-accent px-5 py-2.5 font-medium">Create profiles</a>
            <a href="/login/" className="rounded-md border border-edge px-5 py-2.5">Open the cinema</a>
          </div>
        </>)}

        {error && <p role="alert" className="mt-4 text-sm text-red-400">{error}</p>}
      </section>

      <aside className="hidden rounded-lg border border-edge bg-panel p-8 md:block" aria-label="Preview">
        <p className="text-xs text-dim">Preview</p>
        <h2 className="title-face mt-6 text-4xl">{name || "My Cinema"}</h2>
        <p className="mt-1 text-dim">{tagline}</p>
        <div className="mt-8 grid grid-cols-4 gap-2">
          {Array.from({ length: 8 }, (_, i) => <div key={i} className="aspect-[2/3] rounded bg-gradient-to-br from-accent/40 to-stage" />)}
        </div>
        <button className="mt-8 rounded-md bg-accent px-5 py-2 text-sm font-medium" tabIndex={-1}>Play</button>
      </aside>
    </main>
  );
}

function Nav({ onBack, onNext, nextLabel = "Next" }: { onBack?: () => void; onNext: () => void; nextLabel?: string }) {
  return (
    <div className="mt-8 flex gap-3">
      {onBack && <button onClick={onBack} className="rounded-md border border-edge px-5 py-2.5">Back</button>}
      <button onClick={onNext} className="rounded-md bg-accent px-5 py-2.5 font-medium">{nextLabel}</button>
    </div>
  );
}
