"use client";
import { useEffect, useState } from "react";
import QRCode from "qrcode";
import { adminApi } from "@/lib/adminApi";
import type { RoomsResult } from "@/types";
import { Field, fieldClass, Notice, Toggle } from "./ui";

const plural = (n: number) => `${n} profile${n === 1 ? "" : "s"}`;

export default function RoomsTab() {
  const [prefix, setPrefix] = useState("profile");
  const [from, setFrom] = useState(101);
  const [to, setTo] = useState(110);
  const [max, setMax] = useState(2);
  const [password, setPassword] = useState("");
  const [external, setExternal] = useState(false);
  const [result, setResult] = useState<RoomsResult | null>(null);
  const [qr, setQr] = useState("");
  const [lan, setLan] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState("");
  const [making, setMaking] = useState(0);

  useEffect(() => {
    fetch("/api/setup/interfaces").then((r) => r.json()).then((i) => setLan(i?.[0]?.ip ?? "")).catch(() => {});
  }, []);

  const port = typeof window !== "undefined" ? window.location.port : "";
  const ipUrl = lan ? `http://${lan}${port ? ":" + port : ""}` : typeof window !== "undefined" ? window.location.origin : "";

  async function create(e: React.FormEvent) {
    e.preventDefault();
    if (busy) return;
    setErr(""); setDone(""); setResult(null);
    setMaking(Math.max(0, to - from + 1)); setBusy(true);
    try {
      const r = await adminApi<RoomsResult>("/api/admin/rooms/bulk", {
        method: "POST", body: JSON.stringify({ prefix, from, to, max_connections: max, password, allow_external: external }),
      });
      setResult(r);
      setQr(await QRCode.toDataURL(ipUrl, { margin: 1, width: 220 }));
      setDone(r.created.length > 0
        ? `${plural(r.created.length)} created.`
        : "Nothing was created: every profile in that range already exists.");
    } catch (e) { setErr((e as Error).message); } finally { setBusy(false); }
  }

  return (
    <section className="mt-6">
      <form onSubmit={create} className="grid max-w-2xl grid-cols-2 gap-4 print:hidden sm:grid-cols-3">
        <Field label="Prefix"><input className={fieldClass} value={prefix} onChange={(e) => setPrefix(e.target.value)} /></Field>
        <Field label="First profile number"><input type="number" className={fieldClass} value={from} onChange={(e) => setFrom(+e.target.value)} /></Field>
        <Field label="Last profile number"><input type="number" className={fieldClass} value={to} onChange={(e) => setTo(+e.target.value)} /></Field>
        <Field label="Devices per profile"><input type="number" min={1} className={fieldClass} value={max} onChange={(e) => setMax(+e.target.value)} /></Field>
        <Field label="Shared password (blank makes one per profile)" className="col-span-1 sm:col-span-2">
          <input className={fieldClass} value={password} onChange={(e) => setPassword(e.target.value)} />
        </Field>
        <div className="col-span-full">
          <Toggle checked={external} onChange={setExternal} label="Allow external streaming (stream URLs for VLC and other players)" />
          <p className="mt-1 text-xs text-dim">You can change this for any profile later in the Users tab.</p>
        </div>
        <button disabled={busy} aria-busy={busy} className="col-span-full flex w-fit items-center gap-2 rounded-md bg-accent px-5 py-2.5 font-medium disabled:opacity-60">
          {busy && <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white" aria-hidden />}
          {busy ? "Creating profiles..." : "Create profiles"}
        </button>
      </form>
      {busy && (
        <p role="status" aria-live="polite" className="mt-4 flex items-center gap-3 rounded-lg border border-edge bg-panel px-4 py-3 text-sm print:hidden">
          <span className="h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-accent/40 border-t-accent" aria-hidden />
          Creating {plural(making)}. Please keep this page open, large batches can take a minute.
        </p>
      )}
      {err && <p role="alert" className="mt-3 text-sm text-red-400">{err}</p>}
      {done && !busy && <div className="print:hidden"><Notice kind="ok">{done}</Notice></div>}

      {result && (
        <div className="mt-8">
          <div className="flex items-center gap-4 print:hidden">
            <p className="text-sm text-dim">{result.created.length} created{result.skipped.length ? `, ${result.skipped.length} skipped (already exist): ${result.skipped.join(", ")}` : ""}. Passwords are shown only now.</p>
            {result.created.length > 0 && <button onClick={() => window.print()} className="rounded-md border border-edge px-4 py-1.5 text-sm">Print profile cards</button>}
          </div>
          <div className="mt-4 grid gap-4 sm:grid-cols-2 print:grid-cols-2">
            {result.created.map((r) => (
              <article key={r.email} className="break-inside-avoid rounded-lg border border-edge bg-panel p-4 print:border-black print:bg-white print:text-black">
                <h2 className="title-face text-xl">Profile {r.room}</h2>
                <div className="mt-3 flex items-center gap-4">
                  {qr && /* eslint-disable-next-line @next/next/no-img-element */ <img src={qr} alt="QR code to open the cinema" className="h-28 w-28 rounded bg-white p-1" />}
                  <dl className="text-sm">
                    <dt className="text-dim print:text-black">Username</dt><dd>{r.email.split("@")[0]}</dd>
                    <dt className="mt-1 text-dim print:text-black">Password</dt><dd className="font-mono">{r.password}</dd>
                    <dt className="mt-1 text-dim print:text-black">Address</dt><dd className="break-all">{ipUrl}</dd>
                  </dl>
                </div>
              </article>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}
