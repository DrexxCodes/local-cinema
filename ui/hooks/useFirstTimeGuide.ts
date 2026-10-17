"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { FIRST_TIME_KEY } from "@/constants/tour";

/**
 * Decides whether to show the first-time player guide. When a movie starts, the
 * browser value `isFirstTime` is created as "true" if it does not exist yet. The
 * guide shows while it is "true" and `finish` sets it to "false".
 */
export function useFirstTimeGuide(movieStarted: boolean) {
  const [show, setShow] = useState(false);
  const checked = useRef(false);

  useEffect(() => {
    if (!movieStarted || checked.current) return;
    checked.current = true;
    try {
      let v = localStorage.getItem(FIRST_TIME_KEY);
      if (v === null) { localStorage.setItem(FIRST_TIME_KEY, "true"); v = "true"; }
      setShow(v === "true");
    } catch { /* storage blocked: no guide */ }
  }, [movieStarted]);

  const finish = useCallback(() => {
    try { localStorage.setItem(FIRST_TIME_KEY, "false"); } catch { /* ignore */ }
    setShow(false);
  }, []);

  return { show, finish };
}
