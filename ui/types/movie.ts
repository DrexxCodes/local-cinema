export type Movie = {
  id: string; title: string; year: number; duration_seconds: number;
  container: string; video_codec: string; audio_codec: string; resolution: string;
  synopsis: string; genres: string[]; rating: number; file_size: number;
  has_poster: boolean; has_backdrop: boolean; added_at: string;
  /** Changes whenever the thumbnail is regenerated; used to bust the image cache. */
  thumb_v: number;
};

export type Progress = {
  movie: string; position_seconds: number; completed: boolean; updated_at?: string;
};

export type HomeView = "home" | "movies" | "list";
export type TrendTab = "popular" | "premieres" | "recent";
export type SortKey = "latest" | "year" | "az";
