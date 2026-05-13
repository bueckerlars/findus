import { isFxDarkThemeId, isFxThemeId, normalizeFxThemeId } from "../themes";

/** localStorage key for the last signed-in user's UI theme (guest auth shell / first paint). */
export const LAST_USER_FX_THEME_STORAGE_KEY = "findus.lastUserUiTheme";

/** Read a previously stored theme id; invalid values are removed from storage. */
export function readPersistedLastUserFxThemeId(): string | null {
  if (typeof localStorage === "undefined") return null;
  try {
    const raw = localStorage.getItem(LAST_USER_FX_THEME_STORAGE_KEY);
    if (!raw) return null;
    if (isFxThemeId(raw)) return raw;
    localStorage.removeItem(LAST_USER_FX_THEME_STORAGE_KEY);
  } catch {
    /* ignore (private mode, quota, etc.) */
  }
  return null;
}

/** Persist the last signed-in user's theme for guest pages and first load. */
export function persistLastUserFxTheme(id: string | undefined | null): void {
  if (typeof localStorage === "undefined") return;
  try {
    localStorage.setItem(LAST_USER_FX_THEME_STORAGE_KEY, normalizeFxThemeId(id));
  } catch {
    /* ignore */
  }
}

/** Applies UI theme to the document root (`data-fx-theme` / `data-fx-scheme`). */
export function applyFxTheme(id: string | undefined | null): void {
  const t = normalizeFxThemeId(id);
  document.documentElement.dataset.fxTheme = t;
  if (isFxDarkThemeId(t)) {
    document.documentElement.setAttribute("data-fx-scheme", "dark");
  } else {
    document.documentElement.removeAttribute("data-fx-scheme");
  }
}

/**
 * When no session: apply last user's stored theme, or the app default (via {@link normalizeFxThemeId}).
 * Used after logout and for guest routes so login/register match prior colors.
 */
export function resetFxTheme(): void {
  applyFxTheme(readPersistedLastUserFxThemeId());
}

/** Sync document theme from localStorage before the first router/session tick (reduces flash). */
export function hydrateFxThemeFromStorage(): void {
  resetFxTheme();
}
