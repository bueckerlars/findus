<script setup lang="ts">
import { ref, watch } from "vue";
import { RouterView } from "vue-router";
import { Toaster } from "vue-sonner";
import "vue-sonner/style.css";
import FxAlertDialogHost from "./components/FxAlertDialogHost.vue";
import { useSession } from "./session";
import { isFxDarkThemeId } from "./themes";

const sonnerTheme = ref<"light" | "dark">("light");
const { user } = useSession();

watch(
  () => user.value?.theme,
  (t) => {
    sonnerTheme.value = t && isFxDarkThemeId(t) ? "dark" : "light";
  },
  { immediate: true },
);
</script>

<template>
  <RouterView />
  <FxAlertDialogHost />
  <Toaster position="bottom-right" :duration="4000" :rich-colors="true" :theme="sonnerTheme" />
</template>
