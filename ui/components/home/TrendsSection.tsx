"use client";
import { useMemo, useState } from "react";
import { TREND_TABS } from "@/constants/home";
import { genreList, trending } from "@/lib/movies";
import type { Movie, Progress, TrendTab } from "@/types";
import MovieCard from "@/components/MovieCard";
import { TrendIcon } from "@/components/ui/Icons";
import Carousel from "./Carousel";
import GenrePills from "./GenrePills";
import SectionHeader from "./SectionHeader";

type Props = {
  movies: Movie[];
  progress: Map<string, Progress>;
  inList: (id: string) => boolean;
  onToggleList: (id: string) => void;
};

export default function TrendsSection({ movies, progress, inList, onToggleList }: Props) {
  const [tab, setTab] = useState<TrendTab>("popular");
  const [chosen, setChosen] = useState<string[]>([]);
  const genres = useMemo(() => genreList(movies), [movies]);
  const list = useMemo(() => trending(movies, tab, chosen), [movies, tab, chosen]);
  const toggle = (g: string) => setChosen((c) => (c.includes(g) ? c.filter((x) => x !== g) : [...c, g]));

  return (
    <section className="mt-10 sm:mt-14" aria-label="Trends now">
      <div className="px-4 sm:px-8 lg:px-16">
        <SectionHeader title="Trends Now" icon={<TrendIcon />} tabs={TREND_TABS} active={tab} onTab={setTab} />
        <div className="mt-5"><GenrePills genres={genres} chosen={chosen} onToggle={toggle} /></div>
      </div>
      <div className="mt-5">
        {list.length ? (
          <Carousel label="Trending movies">
            {list.map((m) => (
              <div key={m.id} role="listitem">
                <MovieCard fixed m={m} watched={!!progress.get(m.id)?.completed} inList={inList(m.id)} onToggleList={onToggleList} />
              </div>
            ))}
          </Carousel>
        ) : <p className="px-4 text-sm text-dim sm:px-8 lg:px-16">Nothing matches these genres.</p>}
      </div>
    </section>
  );
}
