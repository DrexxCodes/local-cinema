"use client";
import { createContext, useCallback, useContext, useEffect, useState } from "react";
import type { CinemaInfo } from "@/types";

type Ctx = { cinema: CinemaInfo | null; refresh: () => void };
const CinemaContext = createContext<Ctx>({ cinema: null, refresh: () => {} });

/** Loads the cinema name, tagline and accent colour once and applies the accent. */
export function CinemaProvider({ children }: { children: React.ReactNode }) {
  const [cinema, setCinema] = useState<CinemaInfo | null>(null);

  const refresh = useCallback(() => {
    fetch("/api/cinema").then((r) => r.json()).then((c: CinemaInfo & { name: string }) => {
      if (!c) return;
      setCinema(c);
      if (c.accent_color) document.documentElement.style.setProperty("--accent", c.accent_color);
      if (c.name) document.title = c.name;
    }).catch(() => {});
  }, []);

  useEffect(() => { refresh(); }, [refresh]);
  return <CinemaContext.Provider value={{ cinema, refresh }}>{children}</CinemaContext.Provider>;
}

export const useCinema = () => useContext(CinemaContext);
