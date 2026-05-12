<script setup lang="ts">
import { RouterLink, useRoute } from "vue-router";

const route = useRoute();
</script>

<template>
  <header class="sticky top-0 z-50 border-b border-zinc-200/80 bg-white/90 shadow-sm backdrop-blur-md">
    <div class="mx-auto flex max-w-6xl items-center justify-between gap-3 px-4 py-3 sm:px-5">
      <RouterLink to="/" class="flex shrink-0 items-center gap-2 rounded-lg py-1 pr-2 text-zinc-900 transition hover:text-sky-700">
        <span
          class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-sky-500 to-sky-700 text-sm font-bold text-white shadow-sm"
          >F</span
        >
        <span class="font-semibold tracking-tight">{{ $t("common.findus") }}</span>
      </RouterLink>
      <nav class="flex shrink-0 items-center gap-2" :aria-label="$t('guestNav.account')">
        <RouterLink v-if="route.path !== '/login'" to="/login" class="fx-btn-secondary min-h-0 rounded-full px-4 py-2 text-sm"
          >{{ $t("guestNav.signIn") }}</RouterLink
        >
        <RouterLink v-if="route.path !== '/register'" to="/register" class="fx-btn-secondary min-h-0 rounded-full px-4 py-2 text-sm"
          >{{ $t("guestNav.register") }}</RouterLink
        >
      </nav>
    </div>
  </header>
</template>
