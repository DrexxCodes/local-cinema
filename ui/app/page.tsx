"use client";
import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { clearSession } from "@/lib/session";
import { parseDate } from "@/lib/format";
import { pickFeatured, search } from "@/lib/movies";
import { useCatalog } from "@/hooks/useCatalog";
import { useGuest } from "@/hooks/useGuest";
import { useMyList } from "@/hooks/useMyList";
import type { HomeView } from "@/types";
import Header from "@/components/layout/Header";
import ContinueBanner from "@/components/home/ContinueBanner";
import ContinueRow from "@/components/home/ContinueRow";
import Hero from "@/components/home/Hero";
import MovieGrid from "@/components/home/MovieGrid";
import MoviesSection from "@/components/home/MoviesSection";
import TrendsSection from "@/components/home/TrendsSection";

export default function Home() {
  const session = useGuest();
  const router = useRouter();
  const { movies, progressById } = useCatalog(session);
  const list = useMyList(session?.user.id);
  const [view, setView] = useState<HomeView>("home");
  const [q, setQ] = useState("");

  const featured = useMemo(() => (movies ? pickFeatured(movies, progressById) : []), [movies, progressById]);

  // Unfinished movies, most recently watched first. The first one gets the wide banner.
  const continuing = useMemo(() => {
    if (!movies) return [];
    return movies
      .filter((m) => { const p = progressById.get(m.id); return p && !p.completed && p.position_seconds > 30; })
      .sort((a, b) => parseDate(progressById.get(b.id)?.updated_at ?? "") - parseDate(progressById.get(a.id)?.updated_at ?? ""));
  }, [movies, progressById]);

  if (!session || movies === null) return <main className="page-fill grid place-items-center text-dim">Loading</main>;

  async function logout() {
    await api("/api/auth/logout", { method: "POST" }).catch(() => {});
    clearSession();
    router.replace("/login/");
  }

  const results = q.trim() ? search(movies, q) : null;
  const saved = movies.filter((m) => list.has(m.id));
  const shared = { progress: progressById, inList: list.has, onToggleList: list.toggle };
  const showHero = view === "home" && !results && featured.length > 0;
  const top = continuing[0];

  return (
    <div className="pb-12">
      <Header session={session} view={view} onView={setView} query={q} onQuery={setQ} onLogout={logout} overlay={showHero} />

      {movies.length === 0 && (
        <p className="mx-auto mt-40 max-w-md px-6 text-center text-dim">
          No movies yet. The host can add folders in the admin panel, then run a rescan.
        </p>
      )}

      {results ? (
        <section className="px-4 pt-32 sm:px-8 sm:pt-28 lg:px-16">
          <h2 className="mb-5 text-xl font-semibold">{results.length} result{results.length === 1 ? "" : "s"}</h2>
          <MovieGrid movies={results} {...shared} />
        </section>
      ) : view === "list" ? (
        <section className="px-4 pt-32 sm:px-8 sm:pt-28 lg:px-16">
          <h2 className="mb-5 text-xl font-semibold">My list</h2>
          {saved.length ? <MovieGrid movies={saved} {...shared} /> : (
            <p className="py-16 text-center text-dim">Tap the heart or Add list on a movie to keep it here.</p>
          )}
        </section>
      ) : view === "movies" ? (
        <div className="pt-28 sm:pt-20"><MoviesSection movies={movies} {...shared} /></div>
      ) : (
        <>
          {showHero && <Hero movies={featured} inList={list.has} onToggleList={list.toggle} />}
          {movies.length > 0 && (
            <div className={showHero ? "" : "pt-32 sm:pt-24"}>
              <TrendsSection movies={movies} {...shared} />
              {top && <ContinueBanner movie={top} progress={progressById.get(top.id)!} inList={list.has(top.id)} onToggleList={list.toggle} />}
              <ContinueRow movies={continuing.slice(1)} progress={progressById} inList={list.has} onToggleList={list.toggle} />
              <MoviesSection movies={movies} {...shared} />
            </div>
          )}
        </>
      )}
    </div>
  );
}
