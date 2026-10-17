import { describe, expect, it } from "vitest";
import { GRID_PAGE_SIZE, HERO_COUNT, TREND_COUNT } from "@/constants/home";
import { SESSION_KEY } from "@/constants/storage";
import type { Movie, Progress } from "@/types";
import { artUrl, browse, genreList, matchGenres, pageOf, pickFeatured, search, trending } from "./movies";

let n = 0;
function movie(over: Partial<Movie> = {}): Movie {
  n++;
  return {
    id: `m${n}`, title: `Movie ${n}`, year: 2000, duration_seconds: 5400, container: "mp4", video_codec: "h264",
    audio_codec: "aac", resolution: "1080p", synopsis: "", genres: [], rating: 0, file_size: 1e9,
    has_poster: true, has_backdrop: true, added_at: "2026-01-01 00:00:00.000Z", thumb_v: 0, ...over,
  };
}
const ids = (xs: Movie[]) => xs.map((m) => m.id);

describe("artUrl", () => {
  it("adds the version so a new thumbnail is fetched", () => {
    expect(artUrl(movie({ id: "a", thumb_v: 7 }), "poster")).toBe("/api/movies/a/poster?v=7");
  });
  it("adds the sign in details when signed in", () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify({ token: "t k", session_id: "s", user: {} }));
    expect(artUrl(movie({ id: "a" }), "backdrop")).toBe("/api/movies/a/backdrop?v=0&token=t%20k&sid=s");
  });
});

describe("pickFeatured", () => {
  it("prefers movies with wide art and puts unwatched first", () => {
    const noArt = movie({ id: "noart", has_backdrop: false, rating: 10 });
    const watched = movie({ id: "watched", rating: 9 });
    const fresh = movie({ id: "fresh", rating: 5 });
    const progress = new Map<string, Progress>([["watched", { movie: "watched", position_seconds: 10, completed: false }]]);
    expect(ids(pickFeatured([noArt, watched, fresh], progress))).toEqual(["fresh", "watched"]);
  });
  it("uses every movie when none has wide art", () => {
    const a = movie({ id: "a", has_backdrop: false, rating: 1 });
    const b = movie({ id: "b", has_backdrop: false, rating: 2 });
    expect(ids(pickFeatured([a, b], new Map()))).toEqual(["b", "a"]);
  });
  it("breaks rating ties with the newest movie and caps the count", () => {
    const old = movie({ id: "old", added_at: "2020-01-01 00:00:00.000Z" });
    const recent = movie({ id: "recent", added_at: "2026-06-01 00:00:00.000Z" });
    expect(ids(pickFeatured([old, recent], new Map()))).toEqual(["recent", "old"]);
    const many = Array.from({ length: HERO_COUNT + 4 }, () => movie());
    expect(pickFeatured(many, new Map())).toHaveLength(HERO_COUNT);
  });
  it("does not change the list it was given", () => {
    const list = [movie({ rating: 1 }), movie({ rating: 9 })];
    const copy = [...list];
    pickFeatured(list, new Map());
    expect(list).toEqual(copy);
  });
});

describe("genreList", () => {
  it("orders by how many movies have the genre, then by name", () => {
    const list = [
      movie({ genres: ["Drama", "Crime"] }), movie({ genres: ["Drama"] }),
      movie({ genres: ["Action"] }), movie({ genres: ["Drama", "Action"] }),
    ];
    expect(genreList(list)).toEqual(["Drama", "Action", "Crime"]);
  });
  it("is empty without movies", () => {
    expect(genreList([])).toEqual([]);
  });
});

describe("matchGenres", () => {
  const m = movie({ genres: ["Drama", "Crime"] });
  it("matches everything when no genre is chosen", () => expect(matchGenres(m, [])).toBe(true));
  it("matches when any chosen genre is present", () => expect(matchGenres(m, ["Action", "Crime"])).toBe(true));
  it("does not match otherwise", () => expect(matchGenres(m, ["Action"])).toBe(false));
});

describe("trending", () => {
  const a = movie({ id: "a", rating: 9, year: 1999, added_at: "2026-01-01 00:00:00.000Z", genres: ["Drama"] });
  const b = movie({ id: "b", rating: 7, year: 2024, added_at: "2026-03-01 00:00:00.000Z", genres: ["Action"] });
  const c = movie({ id: "c", rating: 8, year: 2010, added_at: "2026-02-01 00:00:00.000Z", genres: ["Drama"] });
  it("popular is by rating", () => expect(ids(trending([a, b, c], "popular", []))).toEqual(["a", "c", "b"]));
  it("premieres is by year, newest first", () => expect(ids(trending([a, b, c], "premieres", []))).toEqual(["b", "c", "a"]));
  it("recent is by date added", () => expect(ids(trending([a, b, c], "recent", []))).toEqual(["b", "c", "a"]));
  it("filters by genre", () => expect(ids(trending([a, b, c], "popular", ["Drama"]))).toEqual(["a", "c"]));
  it("is capped", () => {
    const many = Array.from({ length: TREND_COUNT + 5 }, () => movie());
    expect(trending(many, "recent", [])).toHaveLength(TREND_COUNT);
  });
});

describe("browse", () => {
  const a = movie({ id: "a", title: "Alpha", year: 2001, rating: 5, added_at: "2026-01-01 00:00:00.000Z", genres: ["Drama"] });
  const b = movie({ id: "b", title: "Bravo", year: 2020, rating: 8, added_at: "2026-03-01 00:00:00.000Z", genres: ["Action"] });
  const c = movie({ id: "c", title: "Charlie", year: 2020, rating: 2, added_at: "2026-02-01 00:00:00.000Z", genres: ["Drama"] });
  const base = { genres: [], minRating: 0, sort: "latest" as const, reversed: false };

  it("latest puts the newest first", () => expect(ids(browse([a, b, c], base))).toEqual(["b", "c", "a"]));
  it("year puts the newest year first, ties by title", () => expect(ids(browse([a, c, b], { ...base, sort: "year" }))).toEqual(["b", "c", "a"]));
  it("a-z is alphabetical", () => expect(ids(browse([c, a, b], { ...base, sort: "az" }))).toEqual(["a", "b", "c"]));
  it("reversed flips the order", () => expect(ids(browse([a, b, c], { ...base, sort: "az", reversed: true }))).toEqual(["c", "b", "a"]));
  it("filters by genre and minimum rating together", () => {
    expect(ids(browse([a, b, c], { ...base, genres: ["Drama"], minRating: 3 }))).toEqual(["a"]);
  });
  it("a minimum rating of 0 keeps unrated movies", () => {
    expect(ids(browse([movie({ id: "u", rating: 0 })], base))).toEqual(["u"]);
  });
});

describe("search", () => {
  const list = [movie({ id: "a", title: "The Matrix", year: 1999, genres: ["Sci-Fi"] }), movie({ id: "b", title: "Heat", year: 1995, genres: ["Crime"] })];
  it("finds by title, ignoring case", () => expect(ids(search(list, "matrix"))).toEqual(["a"]));
  it("finds by year and by genre", () => {
    expect(ids(search(list, "1995"))).toEqual(["b"]);
    expect(ids(search(list, "sci"))).toEqual(["a"]);
  });
  it("returns nothing for empty or blank text", () => {
    expect(search(list, "")).toEqual([]);
    expect(search(list, "   ")).toEqual([]);
  });
});

describe("pageOf", () => {
  const list = Array.from({ length: GRID_PAGE_SIZE * 3 }, (_, i) => i);
  it("grows by one page at a time", () => {
    expect(pageOf(list, 1)).toHaveLength(GRID_PAGE_SIZE);
    expect(pageOf(list, 2)).toHaveLength(GRID_PAGE_SIZE * 2);
    expect(pageOf(list, 9)).toHaveLength(list.length);
  });
});
