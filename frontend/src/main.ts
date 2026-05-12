import type { WritableComputedRef } from "vue";
import { createApp, watch } from "vue";
import App from "./App.vue";
import { i18n, persistLocale, syncHtmlLang } from "./i18n";
import type { SupportedLocale } from "./locale/constants";
import { router } from "./router";
import "./style.css";
import "./fx-dark-overrides.css";

const app = createApp(App);
app.use(i18n);
app.use(router);
watch(
  () => (i18n.global.locale as WritableComputedRef<string>).value,
  (loc) => {
    const s = String(loc);
    syncHtmlLang(s);
    persistLocale(s as SupportedLocale);
  },
  { immediate: true },
);
app.mount("#app");
