"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { HEARTBEAT_MS } from "@/constants/home";
import { api } from "@/lib/api";
import { getSession } from "@/lib/session";
import { getSignal, subscribeSignal } from "@/lib/signal/monitor";
import type { Session } from "@/types";

// Gate for logged-in pages: redirects to /login, keeps the session alive and
// reports the measured link quality with every heartbeat.
export function useGuest() {
  const router = useRouter();
  const [session, setSession] = useState<Session | null>(null);

  useEffect(() => {
    const s = getSession();
    if (!s) { router.replace("/login/"); return; }
    setSession(s);
    const stopSignal = subscribeSignal(() => {}); // keeps the probe running for the heartbeat
    const beat = () => {
      const r = getSignal();
      api("/api/auth/heartbeat", {
        method: "POST",
        body: JSON.stringify({ rtt_ms: r.rtt ?? 0, signal: r.ready ? r.level : 0 }),
      }).catch(() => {});
    };
    beat();
    const t = setInterval(beat, HEARTBEAT_MS);
    return () => { clearInterval(t); stopSignal(); };
  }, [router]);

  return session;
}
