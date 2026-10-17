"use client";
import { useCallback, useEffect, useState } from "react";
import { MY_LIST_PREFIX } from "@/constants/storage";

/** "My list" is kept on this device, one list per room. */
export function useMyList(userId: string | undefined) {
  const key = userId ? MY_LIST_PREFIX + userId : "";
  const [ids, setIds] = useState<string[]>([]);

  useEffect(() => {
    if (!key) return;
    try { setIds(JSON.parse(localStorage.getItem(key) ?? "[]")); } catch { setIds([]); }
  }, [key]);

  const toggle = useCallback((id: string) => {
    setIds((prev) => {
      const next = prev.includes(id) ? prev.filter((x) => x !== id) : [id, ...prev];
      if (key) localStorage.setItem(key, JSON.stringify(next));
      return next;
    });
  }, [key]);

  return { ids, has: (id: string) => ids.includes(id), toggle };
}
