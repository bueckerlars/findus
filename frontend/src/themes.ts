/** Must match backend domain.ValidUITheme allowlist. */

export const FX_DEFAULT_THEME = "default";

export type FxThemeEntry = {
  id: string;
  label: string;
  description: string;
  scheme: "light" | "dark";
  /** Tailwind bg-* classes for preview stripes */
  swatches: [string, string, string];
};

export const FX_THEMES: readonly FxThemeEntry[] = [
  {
    id: "default",
    label: "Sky",
    description: "Cool sky accents on a light zinc base.",
    scheme: "light",
    swatches: ["bg-sky-500", "bg-indigo-400", "bg-zinc-200"],
  },
  {
    id: "ocean",
    label: "Ocean",
    description: "Teal and cyan highlights with a slate undertone.",
    scheme: "light",
    swatches: ["bg-cyan-500", "bg-teal-600", "bg-slate-300"],
  },
  {
    id: "forest",
    label: "Forest",
    description: "Emerald and green accents.",
    scheme: "light",
    swatches: ["bg-emerald-600", "bg-green-500", "bg-lime-200"],
  },
  {
    id: "amber",
    label: "Amber",
    description: "Warm amber and orange highlights.",
    scheme: "light",
    swatches: ["bg-amber-500", "bg-orange-500", "bg-yellow-200"],
  },
  {
    id: "rose",
    label: "Rose",
    description: "Rose and fuchsia accents.",
    scheme: "light",
    swatches: ["bg-rose-500", "bg-fuchsia-500", "bg-pink-200"],
  },
  {
    id: "night",
    label: "Night",
    description: "Dark zinc shell with sky accents.",
    scheme: "dark",
    swatches: ["bg-sky-400", "bg-zinc-700", "bg-zinc-950"],
  },
  {
    id: "ocean-night",
    label: "Ocean night",
    description: "Deep slate with cyan highlights.",
    scheme: "dark",
    swatches: ["bg-cyan-400", "bg-slate-700", "bg-slate-950"],
  },
  {
    id: "forest-night",
    label: "Forest night",
    description: "Near-black with emerald accents.",
    scheme: "dark",
    swatches: ["bg-emerald-400", "bg-emerald-900", "bg-zinc-950"],
  },
  {
    id: "amber-night",
    label: "Amber night",
    description: "Warm dark browns with amber highlights.",
    scheme: "dark",
    swatches: ["bg-amber-400", "bg-amber-900", "bg-stone-950"],
  },
  {
    id: "rose-night",
    label: "Rose night",
    description: "Dark plum with rose highlights.",
    scheme: "dark",
    swatches: ["bg-rose-400", "bg-rose-950", "bg-zinc-950"],
  },
] as const;

const allowed = new Set(FX_THEMES.map((t) => t.id));

const darkIds = new Set(FX_THEMES.filter((t) => t.scheme === "dark").map((t) => t.id));

export function isFxThemeId(id: string): boolean {
  return allowed.has(id);
}

export function isFxDarkThemeId(id: string): boolean {
  return darkIds.has(id);
}

export function normalizeFxThemeId(id: string | undefined | null): string {
  if (id && isFxThemeId(id)) return id;
  return FX_DEFAULT_THEME;
}
