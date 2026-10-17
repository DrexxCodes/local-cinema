"use client";
import { useRef } from "react";
import { ROW_SCROLL_RATIO } from "@/constants/home";
import { ChevronLeftIcon, ChevronRightIcon } from "@/components/ui/Icons";

/** Horizontal scroller with edge arrows on larger screens. */
export default function Carousel({ children, label }: { children: React.ReactNode; label: string }) {
  const el = useRef<HTMLDivElement>(null);
  const go = (dir: 1 | -1) => el.current?.scrollBy({ left: dir * el.current.clientWidth * ROW_SCROLL_RATIO, behavior: "smooth" });
  const arrow = "absolute top-1/3 z-10 hidden h-11 w-11 -translate-y-1/2 place-items-center rounded-full bg-black/70 text-ink opacity-0 backdrop-blur transition-opacity hover:bg-black group-hover/row:opacity-100 sm:grid";
  return (
    <div className="group/row relative">
      <button onClick={() => go(-1)} aria-label={`Scroll ${label} left`} className={`${arrow} left-2`}><ChevronLeftIcon /></button>
      <div ref={el} className="row-scroll flex gap-4 overflow-x-auto px-4 pb-3 pt-1 sm:px-8 lg:px-16" role="list" aria-label={label}>{children}</div>
      <button onClick={() => go(1)} aria-label={`Scroll ${label} right`} className={`${arrow} right-2`}><ChevronRightIcon /></button>
    </div>
  );
}
