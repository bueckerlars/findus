import { LOCALE_MESSAGES, SUPPORTED_LOCALES } from "./bundle";

export { SUPPORTED_LOCALES };
/** Widens with each `frontend/locale/*.json` file (Vite glob keys are not literal-typed). */
export type SupportedLocale = string;

export const LOCALE_STORAGE_KEY = "findus.locale";

export function isSupportedLocale(x: string | null | undefined): boolean {
  return x != null && Object.prototype.hasOwnProperty.call(LOCALE_MESSAGES, x);
}

function metaFor(code: string): { bcp47?: string; label?: string } | undefined {
  const bundle = LOCALE_MESSAGES[code];
  if (!bundle || typeof bundle._meta !== "object" || bundle._meta === null) return undefined;
  const m = bundle._meta as Record<string, unknown>;
  const bcp47 = typeof m.bcp47 === "string" ? m.bcp47 : undefined;
  const label = typeof m.label === "string" ? m.label : undefined;
  if (bcp47 === undefined && label === undefined) return undefined;
  return { bcp47, label };
}

/** BCP 47 tag for Intl / toLocaleString; falls back to the UI code. */
export function bcp47ForUiLocale(ui: SupportedLocale | string): string {
  const fromMeta = metaFor(String(ui))?.bcp47;
  if (fromMeta) return fromMeta;
  if (String(ui).includes("-")) return String(ui);
  return String(ui);
}

/** Native display name from the locale file (`_meta.label`). */
export function localeMenuLabel(code: string): string {
  return metaFor(code)?.label ?? code;
}

function normalizeBrowserTag(raw: string): SupportedLocale | null {
  const lower = raw.trim().toLowerCase();
  if (!lower) return null;
  if (isSupportedLocale(lower)) return lower;
  const primary = lower.split("-")[0] ?? "";
  if (isSupportedLocale(primary)) return primary;
  return null;
}

function browserPreferredLocale(): SupportedLocale {
  if (typeof navigator === "undefined") {
    return (SUPPORTED_LOCALES.includes("en") ? "en" : SUPPORTED_LOCALES[0]!) as SupportedLocale;
  }
  const list = navigator.languages?.length ? [...navigator.languages] : [navigator.language];
  for (const lang of list) {
    const hit = normalizeBrowserTag(String(lang));
    if (hit) return hit;
  }
  return (SUPPORTED_LOCALES.includes("en") ? "en" : SUPPORTED_LOCALES[0]!) as SupportedLocale;
}

export function resolveInitialLocale(): SupportedLocale {
  if (typeof localStorage === "undefined") return browserPreferredLocale();
  try {
    const stored = localStorage.getItem(LOCALE_STORAGE_KEY);
    if (stored != null && isSupportedLocale(stored)) return stored;
  } catch {
    /* ignore */
  }
  return browserPreferredLocale();
}
