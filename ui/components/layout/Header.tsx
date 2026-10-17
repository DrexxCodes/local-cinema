"use client";
import { useEffect, useRef, useState } from "react";
import { NAV_ITEMS } from "@/constants/home";
import { useCinema } from "@/hooks/useCinema";
import type { HomeView, Session } from "@/types";
import { CloseIcon, LogoutIcon, SearchIcon } from "@/components/ui/Icons";
import SignalBadge from "@/components/signal/SignalBadge";

type Props = {
  session: Session;
  view: HomeView;
  onView: (v: HomeView) => void;
  query: string;
  onQuery: (q: string) => void;
  onLogout: () => void;
  /** Floats over the hero until the page scrolls. */
  overlay: boolean;
};

export default function Header({ session, view, onView, query, onQuery, onLogout, overlay }: Props) {
  const { cinema } = useCinema();
  const [scrolled, setScrolled] = useState(false);
  const [searching, setSearching] = useState(false);
  const [menu, setMenu] = useState(false);
  const input = useRef<HTMLInputElement>(null);
  const menuBox = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const on = () => setScrolled(window.scrollY > 24);
    on();
    window.addEventListener("scroll", on, { passive: true });
    return () => window.removeEventListener("scroll", on);
  }, []);

  useEffect(() => { if (searching) input.current?.focus(); }, [searching]);

  useEffect(() => {
    if (!menu) return;
    const close = (e: PointerEvent) => { if (!menuBox.current?.contains(e.target as Node)) setMenu(false); };
    document.addEventListener("pointerdown", close);
    return () => document.removeEventListener("pointerdown", close);
  }, [menu]);

  const solid = scrolled || !overlay;
  const showSearch = searching || query !== "";

  return (
    <header className={`fixed inset-x-0 top-0 z-40 transition-colors duration-300 ${solid ? "border-b border-white/10 bg-stage/70 backdrop-blur-xl" : "bg-gradient-to-b from-black/80 to-transparent"}`}>
      <div className="flex items-center gap-4 px-4 py-3 sm:gap-8 sm:px-8">
        <button onClick={() => { onQuery(""); setSearching(false); onView("home"); window.scrollTo({ top: 0, behavior: "smooth" }); }}
          className="truncate text-lg font-extrabold uppercase tracking-wide text-accent sm:text-2xl">
          {cinema?.name ?? ""}
        </button>

        <nav className="hidden items-center gap-6 text-sm sm:flex" aria-label="Main">
          {NAV_ITEMS.map((n) => (
            <button key={n.key} onClick={() => { onQuery(""); setSearching(false); onView(n.key); }}
              className={`relative py-1 transition-colors ${view === n.key && !query ? "text-ink" : "text-dim hover:text-ink"}`}>
              {n.label}
              {view === n.key && !query && <span className="absolute -bottom-1 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-accent" />}
            </button>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-1 sm:gap-3">
          {showSearch ? (
            <div className="flex items-center rounded-full border border-white/15 bg-black/40 pl-3 pr-1 backdrop-blur">
              <SearchIcon className="h-4 w-4 text-dim" />
              <input ref={input} value={query} onChange={(e) => onQuery(e.target.value)} placeholder="Title, genre or year" aria-label="Search"
                className="w-32 bg-transparent px-2 py-1.5 text-sm outline-none sm:w-56" />
              <button onClick={() => { onQuery(""); setSearching(false); }} aria-label="Close search" className="rounded-full p-1.5 text-dim hover:text-ink"><CloseIcon /></button>
            </div>
          ) : (
            <button onClick={() => setSearching(true)} aria-label="Search" className="rounded-full p-2 text-ink/90 hover:text-accent"><SearchIcon /></button>
          )}
          <SignalBadge />
          <div ref={menuBox} className="relative">
            <button onClick={() => setMenu((m) => !m)} aria-expanded={menu} aria-label="Account"
              className="grid h-9 w-9 place-items-center rounded-full bg-accent text-sm font-bold">
              {(session.user.room_number || session.user.display_name || "?").slice(0, 3)}
            </button>
            {menu && (
              <div className="absolute right-0 top-full mt-2 w-48 rounded-xl border border-white/10 bg-panel/95 p-2 shadow-xl backdrop-blur-xl">
                <p className="px-3 py-2 text-sm text-dim">{session.user.display_name}</p>
                <button onClick={onLogout} className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm hover:bg-white/10">
                  <LogoutIcon /> Log out
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <nav className="flex gap-1 overflow-x-auto px-4 pb-2 sm:hidden" aria-label="Main">
        {NAV_ITEMS.map((n) => (
          <button key={n.key} onClick={() => { onQuery(""); setSearching(false); onView(n.key); }}
            className={`rounded-full px-4 py-1 text-sm ${view === n.key && !query ? "bg-accent font-medium" : "bg-white/10 text-dim"}`}>
            {n.label}
          </button>
        ))}
      </nav>
    </header>
  );
}
