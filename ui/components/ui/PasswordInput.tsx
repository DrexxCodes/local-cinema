"use client";
import { InputHTMLAttributes, useState } from "react";
import { EyeIcon, EyeOffIcon } from "./Icons";

type Props = Omit<InputHTMLAttributes<HTMLInputElement>, "type">;

/** Password field with a button that shows or hides what was typed. */
export default function PasswordInput({ className = "", ...rest }: Props) {
  const [shown, setShown] = useState(false);
  // Top margins belong to the wrapper so the button stays centred on the field.
  const classes = className.split(/\s+/).filter(Boolean);
  const margin = classes.filter((c) => /^mt-/.test(c)).join(" ");
  const field = classes.filter((c) => !/^mt-/.test(c)).join(" ");
  return (
    <div className={`relative ${margin}`}>
      <input {...rest} type={shown ? "text" : "password"} className={`${field} pr-11`} />
      <button type="button" onClick={() => setShown((s) => !s)} aria-label={shown ? "Hide password" : "Show password"} aria-pressed={shown}
        className="absolute right-1 top-1/2 grid h-9 w-9 -translate-y-1/2 place-items-center rounded-md text-dim hover:text-ink">
        {shown ? <EyeOffIcon /> : <EyeIcon />}
      </button>
    </div>
  );
}
