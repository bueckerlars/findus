<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from "vue";
import { RouterView, useRouter, RouterLink } from "vue-router";
import AppSidebar from "../components/AppSidebar.vue";
import FxSvg from "../components/FxSvg.vue";
import CommandPalette from "../components/CommandPalette.vue";
import CreateModalsHost from "../components/CreateModalsHost.vue";

const router = useRouter();
const navOpen = ref(false);
const isNarrow = ref(false);
const sidebarRef = ref<InstanceType<typeof AppSidebar> | null>(null);

let removeMqListener: (() => void) | undefined;
let removeEscListener: (() => void) | undefined;
let stopAfterEach: (() => void) | undefined;

function closeNav() {
  navOpen.value = false;
}

function toggleNav() {
  navOpen.value = !navOpen.value;
}

function onEscape(e: KeyboardEvent) {
  if (e.key !== "Escape" || !navOpen.value) return;
  closeNav();
}

function updateDrawerScrollLock() {
  if (navOpen.value && isNarrow.value) {
    document.documentElement.classList.add("fx-shell-drawer-open");
  } else {
    document.documentElement.classList.remove("fx-shell-drawer-open");
  }
}

watch([navOpen, isNarrow], () => {
  updateDrawerScrollLock();
  if (navOpen.value && isNarrow.value) {
    void nextTick(() => {
      sidebarRef.value?.focusFirst?.();
    });
  }
});

onMounted(() => {
  const mq = window.matchMedia("(max-width: 639px)");
  isNarrow.value = mq.matches;
  const onMq = () => {
    isNarrow.value = mq.matches;
    if (!mq.matches) closeNav();
  };
  mq.addEventListener("change", onMq);
  removeMqListener = () => mq.removeEventListener("change", onMq);

  document.addEventListener("keydown", onEscape);
  removeEscListener = () => document.removeEventListener("keydown", onEscape);

  stopAfterEach = router.afterEach(() => {
    closeNav();
  });
});

onUnmounted(() => {
  document.documentElement.classList.remove("fx-shell-drawer-open");
  removeMqListener?.();
  removeEscListener?.();
  stopAfterEach?.();
});
</script>

<template>
  <div class="fx-page-shell flex min-h-screen flex-col font-sans antialiased sm:flex-row sm:items-stretch">
    <div
      v-if="navOpen"
      class="fixed inset-0 z-[30] bg-zinc-950/40 backdrop-blur-[2px] sm:hidden"
      aria-hidden="true"
      @pointerdown="closeNav"
    />
    <header
      class="sticky top-0 z-[35] flex shrink-0 items-center gap-3 border-b border-zinc-200/80 bg-white/95 px-2 py-1.5 backdrop-blur-md sm:hidden"
    >
      <button
        type="button"
        class="flex h-11 min-w-[2.75rem] items-center justify-center rounded-lg text-zinc-700 outline-offset-2 transition hover:bg-zinc-100/90 focus-visible:outline focus-visible:ring-2 focus-visible:ring-sky-400/35"
        :aria-expanded="navOpen"
        aria-controls="fx-app-sidebar"
        :aria-label="$t('shell.openNavigationMenu')"
        @click="toggleNav"
      >
        <FxSvg name="bars3" class="h-6 w-6 text-current" />
      </button>
      <RouterLink
        to="/"
        class="flex min-w-0 items-center gap-2 rounded-lg py-1.5 pr-2 text-sm font-semibold tracking-tight text-zinc-900 outline-offset-2 transition hover:text-sky-800"
        @click="closeNav"
      >
        <span
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-sky-500 to-sky-700 text-sm font-bold text-white shadow-md shadow-sky-900/20 ring-1 ring-white/20"
          >F</span
        >
        <span class="truncate">{{ $t("common.findus") }}</span>
      </RouterLink>
    </header>
    <AppSidebar
      ref="sidebarRef"
      :mobile-drawer-open="navOpen"
      :mobile-drawer-as-dialog="navOpen && isNarrow"
      :mobile-drawer-aria-hidden="isNarrow && !navOpen"
    />
    <main
      class="min-w-0 flex-1 overflow-y-auto px-4 py-6 max-sm:pb-[calc(5rem+env(safe-area-inset-bottom,0px))] sm:px-6 sm:py-8 lg:px-10"
    >
      <RouterView />
    </main>
    <CommandPalette />
    <CreateModalsHost />
  </div>
</template>
