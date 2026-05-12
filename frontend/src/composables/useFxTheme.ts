import { FX_DEFAULT_THEME, isFxDarkThemeId, normalizeFxThemeId } from "../themes";

/** Applies persisted UI theme to the document root (`data-fx-theme`). */
export function applyFxTheme(id: string | undefined | null): void {
  const t = normalizeFxThemeId(id);
  document.documentElement.dataset.fxTheme = t;
  if (isFxDarkThemeId(t)) {
    document.documentElement.setAttribute("data-fx-scheme", "dark");
  } else {
    document.documentElement.removeAttribute("data-fx-scheme");
  }
}

export function resetFxTheme(): void {
  document.documentElement.dataset.fxTheme = FX_DEFAULT_THEME;
  document.documentElement.removeAttribute("data-fx-scheme");
}
