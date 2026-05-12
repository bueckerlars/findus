/** Must match backend domain.ValidUITheme allowlist. */

export const FX_DEFAULT_THEME = "default";

export type FxThemeEntry = {
  id: string;
  scheme: "light" | "dark";
  /** Tailwind bg-* classes for preview stripes */
  swatches: [string, string, string];
};

export const FX_THEMES: readonly FxThemeEntry[] = [
  {
    id: "default",
    scheme: "light",
    swatches: ["bg-sky-500", "bg-indigo-400", "bg-zinc-200"],
  },
  {
    id: "ocean",
    scheme: "light",
    swatches: ["bg-cyan-500", "bg-teal-600", "bg-slate-300"],
  },
  {
    id: "forest",
    scheme: "light",
    swatches: ["bg-emerald-600", "bg-green-500", "bg-lime-200"],
  },
  {
    id: "amber",
    scheme: "light",
    swatches: ["bg-amber-500", "bg-orange-500", "bg-yellow-200"],
  },
  {
    id: "rose",
    scheme: "light",
    swatches: ["bg-rose-500", "bg-fuchsia-500", "bg-pink-200"],
  },
  {
    id: "night",
    scheme: "dark",
    swatches: ["bg-sky-400", "bg-zinc-700", "bg-zinc-950"],
  },
  {
    id: "ocean-night",
    scheme: "dark",
    swatches: ["bg-cyan-400", "bg-slate-700", "bg-slate-950"],
  },
  {
    id: "forest-night",
    scheme: "dark",
    swatches: ["bg-emerald-400", "bg-emerald-900", "bg-zinc-950"],
  },
  {
    id: "amber-night",
    scheme: "dark",
    swatches: ["bg-amber-400", "bg-amber-900", "bg-stone-950"],
  },
  {
    id: "rose-night",
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
