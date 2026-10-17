"use client";
import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useCinema } from "@/hooks/useCinema";

export default function Footer() {
  const { cinema } = useCinema();
  const path = usePathname();
  const [year, setYear] = useState(new Date().getFullYear());
  useEffect(() => { setYear(new Date().getFullYear()); }, []);

  if (path?.startsWith("/watch")) return null; // the player owns the whole screen

  return (
    <footer className="border-t border-white/10 px-4 py-5 text-center text-xs text-dim print:hidden sm:px-8">
      &copy; {year} {cinema?.name ?? ""}{cinema?.name ? ". " : ""}All rights reserved.
    </footer>
  );
}
