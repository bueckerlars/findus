/**
 * Eager-loads every `*.json` in `frontend/locale/`.
 * Add a language: add `fr.json` (etc.) with a root `_meta` object, e.g.
 * `"_meta": { "bcp47": "fr-FR", "label": "Français" }`.
 */
const rawModules = import.meta.glob<{ default: Record<string, unknown> }>("../../locale/*.json", {
  eager: true,
});

function localeCodeFromGlobPath(filePath: string): string {
  const base = filePath.split("/").pop() ?? "";
  return base.replace(/\.json$/i, "");
}

function moduleMessages(mod: unknown): Record<string, unknown> {
  if (mod && typeof mod === "object" && "default" in mod) {
    const d = (mod as { default: unknown }).default;
    if (d && typeof d === "object" && !Array.isArray(d)) return d as Record<string, unknown>;
  }
  if (mod && typeof mod === "object" && !Array.isArray(mod)) return mod as Record<string, unknown>;
  return {};
}

const built: Record<string, Record<string, unknown>> = {};
for (const [path, mod] of Object.entries(rawModules)) {
  const code = localeCodeFromGlobPath(path);
  if (!code) continue;
  built[code] = moduleMessages(mod);
}

function sortLocaleCodes(codes: string[]): string[] {
  return [...codes].sort((a, b) => {
    if (a === "en") return b === "en" ? 0 : -1;
    if (b === "en") return 1;
    return a.localeCompare(b);
  });
}

/** UI locale codes from filenames (e.g. en, de, fr). */
export const SUPPORTED_LOCALES = Object.freeze(sortLocaleCodes(Object.keys(built))) as readonly string[];

export const LOCALE_MESSAGES: Readonly<Record<string, Record<string, unknown>>> = Object.freeze(built);
