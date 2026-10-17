export type SubTrack = { key: string; lang: string; label: string; default: boolean; source: "file" | "upload" | "embedded"; editable?: boolean };

export type StreamLink = { url: string; title: string; subtitles: { label: string; url: string }[] };

export type DeviceInfo = { id: string; device_name: string; ip: string; kind: string; current: boolean; started_at: string; last_seen: string };

export type UpgradeStatus = { status: "pending" | "approved" | "declined"; slots: number; granted: number } | null;

export type LimitInfo = { message: string; max: number; devices: DeviceInfo[]; request: UpgradeStatus };

export type StreamState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "ready"; link: StreamLink }
  | { kind: "limit"; limit: LimitInfo }
  | { kind: "denied"; message: string }
  | { kind: "error"; message: string };
