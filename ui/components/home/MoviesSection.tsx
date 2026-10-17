"use client";
import { useEffect, useMemo, useState } from "react";
import { useSentinel } from "@/hooks/useSentinel";
import { GRID_PAGE_SIZE } from "@/constants/home";
import { browse, genreList, pageOf } from "@/lib/movies";
import type { Movie, Progress, SortKey } from "@/types";
import { ClapperIcon, PlusIcon } from "@/components/ui/Icons";
import GenrePills from "./GenrePills";
import MovieGrid from "./MovieGrid";
import SectionHeader from "./SectionHeader";
import SortBar from "./SortBar";

type Props = {
  movies: Movie[];
  progress: Map<string, Progress>;
  inList: (id: string) => boolean;
  onToggleList: (id: string) => void;
};

export default function MoviesSection({ movies, progress, inList, onToggleList }: Props) {
  const [chosen, setChosen] = useState<string[]>([]);
  const [sort, setSort] = useState<SortKey>("latest");
  const [reversed, setReversed] = useState(false);
  const [minRating, setMinRating] = useState(0);
  const [pages, setPages] = useState(1);

  const genres = useMemo(() => genreList(movies), [movies]);
  const rated = useMemo(() => movies.some((m) => m.rating > 0), [movies]);
  const list = useMemo(() => browse(movies, { genres: chosen, minRating, sort, reversed }), [movies, chosen, minRating, sort, reversed]);
  const shown = pageOf(list, pages);
  const more = shown.length < list.length;

  useEffect(() => { setPages(1); }, [chosen, sort, reversed, minRating]);
  const sentinel = useSentinel(() => setPages((p) => p + 1), more, shown.length);

  const onSort = (k: SortKey) => {
    if (k === sort) setReversed((r) => !r);
    else { setSort(k); setReversed(false); }
  };

  return (
    <section className="mt-12 px-4 sm:mt-16 sm:px-8 lg:px-16" aria-label="Movies">
      <SectionHeader title="Movies" icon={<ClapperIcon />}
        aside={<span className="text-sm text-dim">{list.length} {list.length === 1 ? "movie" : "movies"}</span>} />
      <div className="mt-5 space-y-4">
        <GenrePills genres={genres} chosen={chosen} onToggle={(g) => setChosen((c) => (c.includes(g) ? c.filter((x) => x !== g) : [...c, g]))} />
        <SortBar sort={sort} reversed={reversed} onSort={onSort} minRating={minRating} onRating={setMinRating} showRating={rated} />
      </div>

      <div className="mt-6">
        {shown.length ? (
          <MovieGrid movies={shown} progress={progress} inList={inList} onToggleList={onToggleList} />
        ) : <p className="py-10 text-center text-dim">No movies match these filters.</p>}
      </div>

      {more && (
        <div ref={sentinel} className="mt-10 flex justify-center">
          <button onClick={() => setPages((p) => p + 1)} aria-label={`Show ${GRID_PAGE_SIZE} more movies`}
            className="grid h-12 w-12 place-items-center rounded-full border border-white/15 bg-white/5 backdrop-blur hover:bg-white/15"><PlusIcon /></button>
        </div>
      )}
    </section>
  );
}
