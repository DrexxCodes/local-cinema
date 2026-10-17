"use client";
import { useState } from "react";
import { LinkIcon } from "@/components/ui/Icons";
import StreamDialog from "./StreamDialog";

type Props = { movieId: string; label?: string; className?: string };

export default function StreamButton({ movieId, label = "Copy stream URL", className = "" }: Props) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button onClick={() => setOpen(true)}
        className={className || "inline-flex items-center gap-2 rounded-md border border-edge px-5 py-2.5 hover:border-accent"}>
        <LinkIcon className="h-4 w-4" /> {label}
      </button>
      {open && <StreamDialog movieId={movieId} onClose={() => setOpen(false)} />}
    </>
  );
}
