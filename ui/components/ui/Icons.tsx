type P = { className?: string; filled?: boolean };

const base = (filled?: boolean) => ({
  viewBox: "0 0 24 24", fill: filled ? "currentColor" : "none", stroke: "currentColor",
  strokeWidth: 2, strokeLinecap: "round" as const, strokeLinejoin: "round" as const, "aria-hidden": true,
});

export const PlayIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base(true)} stroke="none" className={className}><path d="M8 5.5v13a1 1 0 0 0 1.5.86l10.5-6.5a1 1 0 0 0 0-1.72L9.5 4.64A1 1 0 0 0 8 5.5Z" /></svg>
);
export const PauseIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base(true)} stroke="none" className={className}><rect x="6" y="5" width="4" height="14" rx="1" /><rect x="14" y="5" width="4" height="14" rx="1" /></svg>
);
export const PlusIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base()} className={className}><path d="M12 5v14M5 12h14" /></svg>
);
export const CheckIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base()} className={className}><path d="m5 12.5 4.5 4.5L19 7.5" /></svg>
);
export const HeartIcon = ({ className = "h-4 w-4", filled }: P) => (
  <svg {...base(filled)} className={className}><path d="M12 20.5s-7.5-4.6-9.2-9.4C1.7 7.7 3.7 4.5 7 4.5c2 0 3.4 1.1 5 3 1.6-1.9 3-3 5-3 3.3 0 5.3 3.2 4.2 6.6-1.7 4.8-9.2 9.4-9.2 9.4Z" /></svg>
);
export const EyeIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base()} className={className}><path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12Z" /><circle cx="12" cy="12" r="3" /></svg>
);
export const StarIcon = ({ className = "h-3.5 w-3.5" }: P) => (
  <svg {...base(true)} stroke="none" className={className}><path d="m12 2.5 2.9 6 6.6.9-4.8 4.6 1.2 6.5L12 17.3 6.1 20.5l1.2-6.5L2.5 9.4l6.6-.9L12 2.5Z" /></svg>
);
export const SearchIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" /></svg>
);
export const ChevronLeftIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="m15 5-7 7 7 7" /></svg>
);
export const ChevronRightIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="m9 5 7 7-7 7" /></svg>
);
export const ChevronDownIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base()} className={className}><path d="m5 9 7 7 7-7" /></svg>
);
export const TrendIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="m3 17 6-6 4 4 8-8" /><path d="M15 7h6v6" /></svg>
);
export const ClapperIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base(true)} stroke="none" className={className}><path d="M3 9h18v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V9Zm.4-1.8 1.3-3.5 4.1 1.5L7.5 8.7 3.4 7.2Zm5.8 1.5L10.5 5.2l4.1 1.5-1.3 3.5-4.1-1.5Zm5.8 1.5 1.3-3.5 4.1 1.5-1.3 3.5-4.1-1.5Z" /></svg>
);
export const LogoutIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base()} className={className}><path d="M15 4h3a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-3M10 17l-5-5 5-5M5 12h11" /></svg>
);
export const CloseIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base()} className={className}><path d="M6 6l12 12M18 6 6 18" /></svg>
);

export const Back10Icon = ({ className = "h-6 w-6" }: P) => (
  <svg {...base()} className={className}><path d="M3 12a9 9 0 1 0 3-6.7" /><path d="M3 4v5h5" /><text x="8.2" y="15.5" fontSize="7" fill="currentColor" stroke="none" fontWeight="700">10</text></svg>
);
export const Forward10Icon = ({ className = "h-6 w-6" }: P) => (
  <svg {...base()} className={className}><path d="M21 12a9 9 0 1 1-3-6.7" /><path d="M21 4v5h-5" /><text x="8.2" y="15.5" fontSize="7" fill="currentColor" stroke="none" fontWeight="700">10</text></svg>
);
export const VolumeIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="M11 5 6 9H3v6h3l5 4V5Z" /><path d="M15.5 8.5a5 5 0 0 1 0 7M18.5 5.5a9 9 0 0 1 0 13" /></svg>
);
export const MuteIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="M11 5 6 9H3v6h3l5 4V5Z" /><path d="m16 9 5 6M21 9l-5 6" /></svg>
);
export const CaptionsIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><rect x="3" y="5" width="18" height="14" rx="2" /><path d="M10 10.5a2 2 0 1 0 0 3M17 10.5a2 2 0 1 0 0 3" /></svg>
);
export const SpeedIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="M12 14l4-4" /><path d="M3.5 17a9 9 0 1 1 17 0" /></svg>
);
export const RotateIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><rect x="6" y="3" width="12" height="18" rx="2" transform="rotate(0)" /><path d="M21 8.5a4 4 0 0 0-3-3M3 15.5a4 4 0 0 0 3 3" /></svg>
);
export const FullscreenIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5" /></svg>
);
export const ExitFullscreenIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="M9 4v5H4M15 4v5h5M9 20v-5H4M15 20v-5h5" /></svg>
);
export const PipIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><rect x="3" y="5" width="18" height="14" rx="2" /><rect x="12" y="12" width="7" height="5" rx="1" fill="currentColor" /></svg>
);
export const LinkIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base()} className={className}><path d="M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1" /><path d="M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1" /></svg>
);
export const MoreIcon = ({ className = "h-5 w-5" }: P) => (
  <svg {...base(true)} stroke="none" className={className}><circle cx="5" cy="12" r="1.8" /><circle cx="12" cy="12" r="1.8" /><circle cx="19" cy="12" r="1.8" /></svg>
);
export const SpinnerIcon = ({ className = "h-10 w-10" }: P) => (
  <svg {...base()} className={`${className} animate-spin`}><path d="M12 3a9 9 0 1 0 9 9" opacity=".9" /></svg>
);
export const EyeOffIcon = ({ className = "h-4 w-4" }: P) => (
  <svg {...base()} className={className}><path d="M3 3l18 18" /><path d="M10.6 5.1A9.8 9.8 0 0 1 12 5c6.4 0 10 7 10 7a17 17 0 0 1-3.2 4M6.5 6.600A17 17 0 0 0 2 12s3.6 7 10 7a9.9 9.9 0 0 0 4-.9" /><path d="M9.9 9.900a3 3 0 0 0 4.2 4.2" /></svg>
);
