import { GRID_PAGE_SIZE, HERO_COUNT, TREND_COUNT } from "@/constants/home";
import type { Movie, Progress, SortKey, TrendTab } from "@/types";
import { withAuth } from "./api";
import { parseDate } from "./format";

export function artUrl(m: Movie, kind: "poster" | "backdrop"): string {
  return withAuth(`/api/movies/${m.id}/${kind}?v=${m.thumb_v ?? 0}`);
}

const byAdded = (a: Movie, b: Movie) => parseDate(b.added_at) - parseDate(a.added_at);

/** Wide art for the hero: unwatched first, then by rating, then newest. */
export function pickFeatured(movies: Movie[], progress: Map<string, Progress>): Movie[] {
  const withArt = movies.filter((m) => m.has_backdrop);
  const pool = withArt.length ? withArt : movies;
  return [...pool]
    .sort((a, b) => {
      const wa = progress.has(a.id) ? 1 : 0, wb = progress.has(b.id) ? 1 : 0;
      return wa - wb || b.rating - a.rating || byAdded(a, b);
    })
    .slice(0, HERO_COUNT);
}

export function genreList(movies: Movie[]): string[] {
  const counts = new Map<string, number>();
  movies.forEach((m) => m.genres.forEach((g) => counts.set(g, (counts.get(g) ?? 0) + 1)));
  return [...counts.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0])).map(([g]) => g);
}

/** A movie matches when it has at least one of the chosen genres (no genres chosen means all). */
export function matchGenres(m: Movie, chosen: string[]): boolean {
  return chosen.length === 0 || m.genres.some((g) => chosen.includes(g));
}

export function trending(movies: Movie[], tab: TrendTab, genres: string[]): Movie[] {
  const list = movies.filter((m) => matchGenres(m, genres));
  const sorted =
    tab === "popular" ? list.sort((a, b) => b.rating - a.rating || byAdded(a, b))
    : tab === "premieres" ? list.sort((a, b) => b.year - a.year || byAdded(a, b))
    : list.sort(byAdded);
  return sorted.slice(0, TREND_COUNT);
}

/** Default order: latest and year newest first, A-Z alphabetical. `reversed` flips it. */
export function browse(movies: Movie[], opts: { genres: string[]; minRating: number; sort: SortKey; reversed: boolean }): Movie[] {
  const { genres, minRating, sort, reversed } = opts;
  const list = movies.filter((m) => matchGenres(m, genres) && (minRating <= 0 || m.rating >= minRating));
  const base =
    sort === "az" ? (a: Movie, b: Movie) => a.title.localeCompare(b.title)
    : sort === "year" ? (a: Movie, b: Movie) => b.year - a.year || a.title.localeCompare(b.title)
    : byAdded;
  return list.sort((a, b) => (reversed ? -base(a, b) : base(a, b)));
}

export function search(movies: Movie[], q: string): Movie[] {
  const needle = q.trim().toLowerCase();
  if (!needle) return [];
  return movies.filter((m) => [m.title, String(m.year), ...m.genres].join(" ").toLowerCase().includes(needle));
}

export function pageOf<T>(list: T[], page: number): T[] {
  return list.slice(0, page * GRID_PAGE_SIZE);
}
