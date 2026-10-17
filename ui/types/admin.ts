import type { Movie } from "./movie";

export type ScanState = { running: boolean; total: number; done: number; current: string; percent?: number };

export type AdminStats = {
  active_sessions: number; movies: number; uptime_seconds: number;
  scan: ScanState; thumbs: ScanState; convert: ScanState; convert_error: string; scan_issues: string[]; subs: ScanState; subs_msg: string; pending_upgrades: number; update_available: string; ffmpeg_install: ScanState; ffmpeg: boolean;
};

export type AdminSession = {
  id: string; room_number: string; device_name: string; ip: string;
  started_at: string; last_seen: string; rtt_ms: number; signal: number; kind: string;
};

export type AdminUser = {
  id: string; room_number: string; username: string; email: string; display_name: string;
  max_connections: number; active: boolean; allow_external: boolean; notes: string; online: number; created: string;
};

export type Page<T> = { items: T[]; total: number; has_more: boolean };

export type Room = { room: string; email: string; password: string };
export type RoomsResult = { created: Room[]; skipped: string[] };

export type AdminSettings = {
  cinema_name: string; tagline: string; accent_color: string;
  movie_folders: string[]; watch_folders: boolean; global_max_streams: number;
  mdns_name: string; has_tmdb_key: boolean; admin_email: string; version: string;
  lan_ips: { name: string; ip: string }[];
};

export type FolderCheck = { valid: boolean; video_count: number; total_bytes: number; warnings: string[] };

/** ok: plays everywhere. convert: some devices may not play it. converted: a universal MP4 copy is in use. */
export type CompatStatus = "ok" | "convert" | "converted" | "unknown";
export type AdminMovie = Movie & {
  compat: CompatStatus;
  /** Removed from the stream by the admin. */
  hidden: boolean;
  /** The admin typed details or art, so rescans leave them alone. */
  edited: boolean;
  /** False when the file is no longer on disk. */
  available: boolean;
  tmdb_id: number;
  file_name: string;
};
export type MovieFilter = "" | "hidden" | "missing";

export type UpgradeRequest = {
  id: string; room_number: string; display_name: string; max_connections: number; online: number;
  slots: number; granted: number; status: "pending" | "approved" | "declined"; created: string;
};

export type UpdateInfo = {
  enabled: boolean; current: string; latest: string; available: boolean; url: string; notes: string;
  published_at: string; asset_name: string; asset_url: string; checked_at: string; error: string;
};

export type ToolsInfo = {
  ffmpeg: boolean; ffprobe: boolean; ffmpeg_version: string; ffmpeg_path: string;
  can_install: boolean; install: ScanState; message: string; platform: string;
};
