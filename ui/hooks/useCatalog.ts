"use client";
import { useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";
import type { Movie, Progress, Session } from "@/types";

/** Loads the movie list and this room's watch progress. `movies` is null while loading. */
export function useCatalog(session: Session | null) {
  const [movies, setMovies] = useState<Movie[] | null>(null);
  const [progress, setProgress] = useState<Progress[]>([]);

  useEffect(() => {
    if (!session) return;
    api<Movie[]>("/api/movies").then(setMovies).catch(() => setMovies([]));
    api<Progress[]>("/api/playback").then(setProgress).catch(() => {});
  }, [session]);

  const progressById = useMemo(() => new Map(progress.map((p) => [p.movie, p])), [progress]);
  return { movies, progress, progressById };
}
