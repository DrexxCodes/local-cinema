"use client";
import { useEffect, useRef, useState } from "react";

/**
 * Calls onVisible when the returned ref element scrolls near the viewport.
 * Pass `watch` (for example the loaded item count) so the check runs again after
 * new content loads and the element is still on screen.
 */
export function useSentinel(onVisible: () => void, enabled: boolean, watch: unknown, rootMargin = "400px") {
  const [el, setEl] = useState<Element | null>(null);
  const cb = useRef(onVisible);
  useEffect(() => { cb.current = onVisible; });

  useEffect(() => {
    if (!enabled || !el) return;
    const io = new IntersectionObserver((entries) => {
      if (entries.some((e) => e.isIntersecting)) cb.current();
    }, { rootMargin });
    io.observe(el);
    return () => io.disconnect();
  }, [el, enabled, watch, rootMargin]);

  return setEl;
}
