"use client";
import type { Movie, Progress } from "@/types";
import MovieCard from "@/components/MovieCard";
import Carousel from "./Carousel";

type Props = {
  movies: Movie[];
  progress: Map<string, Progress>;
  inList: (id: string) => boolean;
  onToggleList: (id: string) => void;
};

export default function ContinueRow({ movies, progress, inList, onToggleList }: Props) {
  if (!movies.length) return null;
  return (
    <section className="mt-10" aria-label="More to continue">
      <h2 className="mb-3 px-4 text-lg font-semibold sm:px-8 lg:px-16">Continue watching</h2>
      <Carousel label="Continue watching">
        {movies.map((m) => {
          const p = progress.get(m.id)!;
          return (
            <div key={m.id} role="listitem">
              <MovieCard fixed m={m} inList={inList(m.id)} onToggleList={onToggleList}
                progress={m.duration_seconds ? p.position_seconds / m.duration_seconds : 0.1} />
            </div>
          );
        })}
      </Carousel>
    </section>
  );
}
