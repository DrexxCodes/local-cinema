import type { TrendTab } from "@/types";

export const HERO_COUNT = 5;
export const HERO_INTERVAL_MS = 8000;
export const TREND_COUNT = 20;
export const GRID_PAGE_SIZE = 18;
export const HEARTBEAT_MS = 15000;
export const ROW_SCROLL_RATIO = 0.85;

export const TREND_TABS: { key: TrendTab; label: string }[] = [
  { key: "popular", label: "Popular" },
  { key: "premieres", label: "Premieres" },
  { key: "recent", label: "Recently Added" },
];

export const NAV_ITEMS = [
  { key: "home", label: "Home" },
  { key: "movies", label: "Movies" },
  { key: "list", label: "My list" },
] as const;
