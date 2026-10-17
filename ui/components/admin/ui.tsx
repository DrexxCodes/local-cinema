// Small shared pieces for the admin tabs.
export const fieldClass = "mt-1 w-full rounded-md border border-edge bg-panel px-3 py-2";

export function Field({ label, children, hint, className = "" }: { label: string; children: React.ReactNode; hint?: string; className?: string }) {
  return (
    <label className={`block text-sm text-dim ${className}`}>
      {label}{children}
      {hint && <span className="mt-1 block text-xs">{hint}</span>}
    </label>
  );
}

export function Toggle({ checked, onChange, label }: { checked: boolean; onChange: (v: boolean) => void; label: string }) {
  return (
    <button type="button" role="switch" aria-checked={checked} onClick={() => onChange(!checked)}
      className="flex items-center gap-3 text-sm">
      <span className={`relative h-6 w-11 rounded-full transition-colors ${checked ? "bg-accent" : "bg-edge"}`}>
        <span className={`absolute top-0.5 h-5 w-5 rounded-full bg-white transition-all ${checked ? "left-[22px]" : "left-0.5"}`} />
      </span>
      {label}
    </button>
  );
}

export function Notice({ kind, children }: { kind: "error" | "ok"; children: React.ReactNode }) {
  return <p role={kind === "error" ? "alert" : "status"} className={`mt-3 text-sm ${kind === "error" ? "text-red-400" : "text-emerald-400"}`}>{children}</p>;
}

export function Modal({ title, onClose, children, wide }: { title: string; onClose: () => void; children: React.ReactNode; wide?: boolean }) {
  return (
    <div role="dialog" aria-modal aria-label={title} className="fixed inset-0 z-50 grid place-items-center bg-black/70 p-4 backdrop-blur-sm" onClick={onClose}>
      <div className={`max-h-[90vh] w-full ${wide ? "max-w-2xl" : "max-w-md"} overflow-y-auto rounded-2xl border border-white/10 bg-panel p-6 shadow-2xl`} onClick={(e) => e.stopPropagation()}>
        <h2 className="title-face text-xl">{title}</h2>
        {children}
      </div>
    </div>
  );
}
