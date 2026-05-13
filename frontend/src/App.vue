<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from "vue";
import { RouterView } from "vue-router";
import { Toaster } from "vue-sonner";
import "vue-sonner/style.css";
import FxAlertDialogHost from "./components/FxAlertDialogHost.vue";
import { useSession } from "./session";
import { isFxDarkThemeId } from "./themes";

type SonnerPosition = "top-center" | "bottom-right";

const sonnerTheme = ref<"light" | "dark">("light");
const sonnerPosition = ref<SonnerPosition>("bottom-right");
const { user } = useSession();

let removeToastMqListener: (() => void) | undefined;

function updateToastPosition() {
  sonnerPosition.value = window.matchMedia("(max-width: 639px)").matches ? "top-center" : "bottom-right";
}

watch(
  () => user.value?.theme,
  (t) => {
    sonnerTheme.value = t && isFxDarkThemeId(t) ? "dark" : "light";
  },
  { immediate: true },
);

onMounted(() => {
  const mq = window.matchMedia("(max-width: 639px)");
  updateToastPosition();
  mq.addEventListener("change", updateToastPosition);
  removeToastMqListener = () => mq.removeEventListener("change", updateToastPosition);
});

onUnmounted(() => {
  removeToastMqListener?.();
});
</script>

<template>
  <RouterView />
  <FxAlertDialogHost />
  <Toaster :position="sonnerPosition" :duration="4000" :rich-colors="true" :theme="sonnerTheme" />
</template>
