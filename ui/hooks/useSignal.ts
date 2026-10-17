"use client";
import { useSyncExternalStore } from "react";
import { getServerSignal, getSignal, subscribeSignal } from "@/lib/signal/monitor";
import type { SignalReading } from "@/types";

/** Live link quality between this device and the cinema server. */
export function useSignal(): SignalReading {
  return useSyncExternalStore(subscribeSignal, getSignal, getServerSignal);
}
