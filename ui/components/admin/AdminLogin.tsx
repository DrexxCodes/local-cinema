"use client";
import { useState } from "react";
import { setAdminToken } from "@/lib/adminApi";
import PasswordInput from "./../ui/PasswordInput";

export default function AdminLogin({ onDone }: { onDone: () => void }) {
  const [email, setEmail] = useState("");
  const [pw, setPw] = useState("");
  const [err, setErr] = useState("");

  async function go(e: React.FormEvent) {
    e.preventDefault(); setErr("");
    const res = await fetch("/api/collections/_superusers/auth-with-password", {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ identity: email, password: pw }),
    }).catch(() => null);
    const body = await res?.json().catch(() => null);
    if (!res?.ok || !body?.token) return setErr("Wrong email or password.");
    setAdminToken(body.token);
    onDone();
  }

  return (
    <main className="grid page-fill place-items-center p-6">
      <form onSubmit={go} className="glass w-full max-w-sm rounded-2xl p-6">
        <h1 className="title-face text-3xl">Admin sign in</h1>
        <input className="mt-6 w-full rounded-md border border-edge bg-panel px-3 py-2.5" placeholder="Admin email" value={email} onChange={(e) => setEmail(e.target.value)} aria-label="Admin email" autoComplete="username" />
        <PasswordInput className="mt-3 w-full rounded-md border border-edge bg-panel px-3 py-2.5" placeholder="Password" value={pw} onChange={(e) => setPw(e.target.value)} aria-label="Password" autoComplete="current-password" />
        {err && <p role="alert" className="mt-3 text-sm text-red-400">{err}</p>}
        <button className="mt-6 w-full rounded-md bg-accent py-2.5 font-medium">Sign in</button>
      </form>
    </main>
  );
}
