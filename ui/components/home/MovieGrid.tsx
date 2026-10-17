"use client";
import type { Movie, Progress } from "@/types";
import MovieCard from "@/components/MovieCard";

type Props = {
  movies: Movie[];
  progress: Map<string, Progress>;
  inList: (id: string) => boolean;
  onToggleList: (id: string) => void;
};

export const GRID_CLASS = "grid grid-cols-2 gap-x-4 gap-y-7 sm:grid-cols-3 md:grid-cols-4 xl:grid-cols-6";

export default function MovieGrid({ movies, progress, inList, onToggleList }: Props) {
  return (
    <div className={GRID_CLASS} role="list">
      {movies.map((m) => (
        <div key={m.id} role="listitem">
          <MovieCard m={m} watched={!!progress.get(m.id)?.completed} inList={inList(m.id)} onToggleList={onToggleList} />
        </div>
      ))}
    </div>
  );
}
