import { createI18n } from "vue-i18n";
import { LOCALE_STORAGE_KEY, resolveInitialLocale, type SupportedLocale } from "./locale/constants";
import { LOCALE_MESSAGES } from "./locale/bundle";

const missingWarn = import.meta.env.DEV;

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: resolveInitialLocale(),
  fallbackLocale: "en",
  missingWarn,
  fallbackWarn: missingWarn,
  messages: LOCALE_MESSAGES,
} as never);

export function persistLocale(locale: SupportedLocale): void {
  try {
    localStorage.setItem(LOCALE_STORAGE_KEY, locale);
  } catch {
    /* ignore */
  }
}

export function syncHtmlLang(locale: string): void {
  if (typeof document !== "undefined") {
    document.documentElement.lang = locale;
  }
}
