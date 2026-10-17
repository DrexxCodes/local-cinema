type Tab<K extends string> = { key: K; label: string };

type Props<K extends string> = {
  title: string;
  icon: React.ReactNode;
  tabs?: Tab<K>[];
  active?: K;
  onTab?: (k: K) => void;
  aside?: React.ReactNode;
};

/** Section title with an accent dot under it and optional tabs on the right. */
export default function SectionHeader<K extends string>({ title, icon, tabs, active, onTab, aside }: Props<K>) {
  return (
    <div className="flex flex-wrap items-end justify-between gap-x-8 gap-y-3 border-b border-white/10 pb-3">
      <h2 className="relative flex items-center gap-3 text-xl font-semibold sm:text-2xl">
        {icon}{title}
        <span className="absolute -bottom-[14px] left-1 h-1.5 w-1.5 rounded-full bg-accent" />
      </h2>
      {tabs && (
        <div className="flex gap-5 text-sm sm:gap-10" role="tablist">
          {tabs.map((t) => (
            <button key={t.key} role="tab" aria-selected={active === t.key} onClick={() => onTab?.(t.key)}
              className={`transition-colors ${active === t.key ? "font-semibold text-ink" : "text-dim hover:text-ink"}`}>
              {t.label}
            </button>
          ))}
        </div>
      )}
      {aside}
    </div>
  );
}
