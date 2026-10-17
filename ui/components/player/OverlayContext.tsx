"use client";
import { createContext } from "react";

/**
 * Dialogs opened from inside the player render into the player's own element, so they
 * stay visible in full screen and rotate with the picture. (A dialog inside the control
 * bar would be trapped by its blur effect, which breaks position: fixed.)
 */
export type OverlayCtx = {
  root: HTMLElement | null;
  /** Called with true when a dialog opens and false when it closes. The player pauses meanwhile. */
  onModal?: (open: boolean) => void;
};

export const OverlayContext = createContext<OverlayCtx>({ root: null });
