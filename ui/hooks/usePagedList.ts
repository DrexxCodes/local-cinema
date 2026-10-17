"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import type { Page } from "@/types";

/**
 * Offset paged list for infinite scroll. Everything reloads from the first page
 * when `key` changes (for example a new search text). Call `reload` to refresh.
 */
export function usePagedList<T extends { id: string }>(
  fetchPage: (offset: number) => Promise<Page<T>>,
  key: string,
) {
  const [items, setItems] = useState<T[]>([]);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const gen = useRef(0);
  const busy = useRef(false);
  const offset = useRef(0);
  const fetcher = useRef(fetchPage);
  useEffect(() => { fetcher.current = fetchPage; });

  const loadMore = useCallback(async () => {
    if (busy.current) return;
    busy.current = true;
    setLoading(true);
    const g = gen.current;
    try {
      const page = await fetcher.current(offset.current);
      if (g !== gen.current) return;
      setItems((prev) => {
        const seen = new Set(prev.map((p) => p.id));
        return [...prev, ...page.items.filter((n) => !seen.has(n.id))];
      });
      offset.current += page.items.length;
      setTotal(page.total);
      setHasMore(page.has_more && page.items.length > 0);
      setError("");
    } catch (e) {
      if (g !== gen.current) return;
      setError((e as Error).message);
      setHasMore(false);
    } finally {
      if (g === gen.current) { busy.current = false; setLoading(false); }
    }
  }, []);

  const reload = useCallback(() => {
    gen.current++;
    busy.current = false;
    offset.current = 0;
    setItems([]); setTotal(0); setHasMore(true); setError("");
    loadMore();
  }, [loadMore]);

  useEffect(() => { reload(); }, [key, reload]);

  return { items, total, hasMore, loading, error, loadMore, reload };
}
