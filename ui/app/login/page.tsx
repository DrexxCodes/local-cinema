"use client";
import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { deviceId, deviceName } from "@/lib/device";
import { getSession, setSession } from "@/lib/session";
import type { Session } from "@/types";
import PasswordInput from "@/components/ui/PasswordInput";

function LoginForm() {
  const router = useRouter();
  const params = useSearchParams();
  const [cinema, setCinema] = useState<{ name: string; tagline: string } | null>(null);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [capModal, setCapModal] = useState<string | null>(null);
  const [full, setFull] = useState(false);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (getSession()) { router.replace("/"); return; }
    fetch("/api/cinema").then((r) => r.json()).then((c) => {
      if (!c.setup_complete) router.replace("/setup/");
      else setCinema({ name: c.name, tagline: c.tagline });
    }).catch(() => setError("Can't reach the cinema server. Check your WiFi."));
    if (params.get("ended")) setError("Your session ended. Please log in again.");
  }, [router, params]);

  async function submit(e?: React.FormEvent) {
    e?.preventDefault();
    setBusy(true); setError(""); setCapModal(null); setFull(false);
    try {
      const s = await api<Session>("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password, device_id: deviceId(), device_name: deviceName() }),
      });
      setSession(s);
      router.replace("/");
    } catch (err) {
      const a = err as ApiError;
      if (a.code === "max_connections_reached") setCapModal(a.message);
      else if (a.status === 503 && a.code === "server_at_capacity") setFull(true);
      else setError(a.message);
    } finally { setBusy(false); }
  }

  // Cinema full: retry every 30 seconds.
  useEffect(() => {
    if (!full) return;
    const t = setTimeout(() => submit(), 30000);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [full]);

  if (full) return (
    <main className="grid page-fill place-items-center p-6 text-center">
      <div>
        <h1 className="title-face text-3xl">The cinema is full</h1>
        <p className="mt-2 text-dim">Every seat is taken. We will try again in 30 seconds.</p>
        <button onClick={() => submit()} className="mt-6 rounded-md bg-accent px-5 py-2.5 font-medium">Try now</button>
      </div>
    </main>
  );

  return (
    <main className="grid page-fill place-items-center p-6">
      <form onSubmit={submit} className="w-full max-w-sm">
        <h1 className="title-face text-4xl">{cinema?.name ?? " "}</h1>
        {cinema?.tagline && <p className="mt-1 text-dim">{cinema.tagline}</p>}
        <label className="mt-8 block text-sm text-dim" htmlFor="room">Profile or email</label>
        <input id="room" value={email} onChange={(e) => setEmail(e.target.value)} autoCapitalize="none" autoComplete="username"
          className="mt-1 w-full rounded-md border border-edge bg-panel px-3 py-2.5" placeholder="profile204" />
        <label className="mt-4 block text-sm text-dim" htmlFor="pw">Password</label>
        <PasswordInput id="pw" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password"
          className="mt-1 w-full rounded-md border border-edge bg-panel px-3 py-2.5" />
        {error && <p role="alert" className="mt-3 text-sm text-red-400">{error}</p>}
        <button disabled={busy} className="mt-6 w-full rounded-md bg-accent py-2.5 font-medium disabled:opacity-60">
          {busy ? "Signing in" : "Sign in"}
        </button>
        <p className="mt-4 text-sm text-dim">Your profile number is your username. Ask the front desk if you are unsure.</p>
      </form>

      {capModal && (
        <div role="dialog" aria-modal className="fixed inset-0 grid place-items-center bg-black/70 p-6">
          <div className="max-w-sm rounded-lg border border-edge bg-panel p-6">
            <h2 className="title-face text-xl">Too many devices</h2>
            <p className="mt-2 text-dim">{capModal}</p>
            <div className="mt-6 flex gap-3">
              <button onClick={() => submit()} className="rounded-md bg-accent px-4 py-2 font-medium">Retry</button>
              <button onClick={() => setCapModal(null)} className="rounded-md border border-edge px-4 py-2">Cancel</button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

export default function LoginPage() {
  return <Suspense><LoginForm /></Suspense>;
}
