<script setup lang="ts">
import { RouterView } from "vue-router";
import AppSidebar from "./components/AppSidebar.vue";
import CommandPalette from "./components/CommandPalette.vue";
import FxAlertDialogHost from "./components/FxAlertDialogHost.vue";
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const guest = computed(() => !!route.meta.guestOnly);
</script>

<template>
  <div v-if="guest" class="min-h-full bg-zinc-100 font-sans text-zinc-900 antialiased">
    <RouterView />
  </div>
  <div v-else class="flex min-h-screen items-stretch fx-page-shell bg-zinc-50 font-sans text-zinc-900 antialiased">
    <AppSidebar />
    <main class="min-w-0 flex-1 overflow-y-auto px-5 py-10 sm:px-8 sm:py-12 lg:px-12">
      <RouterView />
    </main>
    <CommandPalette />
  </div>
  <FxAlertDialogHost />
</template>
