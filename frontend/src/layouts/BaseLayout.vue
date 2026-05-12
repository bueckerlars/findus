<script setup lang="ts">
import { RouterView } from "vue-router";
import AppSidebar from "../components/AppSidebar.vue";
import CommandPalette from "../components/CommandPalette.vue";
import CreateModalsHost from "../components/CreateModalsHost.vue";
</script>

<template>
  <div class="fx-page-shell flex min-h-screen items-stretch font-sans antialiased">
    <AppSidebar />
    <main class="min-w-0 flex-1 overflow-y-auto px-5 py-10 sm:px-8 sm:py-12 lg:px-12">
      <RouterView />
    </main>
    <CommandPalette />
    <CreateModalsHost />
  </div>
</template>
