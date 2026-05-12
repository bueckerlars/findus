<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";

const route = useRoute();
const templatesSection = computed(() => route.path.startsWith("/admin/templates"));
</script>

<template>
  <div class="max-w-4xl space-y-8">
    <div class="border-b border-zinc-200/90">
      <nav class="-mb-px flex flex-wrap gap-1" aria-label="Admin sections">
        <RouterLink
          to="/admin/users"
          class="inline-flex items-center rounded-t-lg border border-transparent px-4 py-2.5 text-sm font-medium text-zinc-600 transition hover:text-zinc-900"
          :class="
            !templatesSection
              ? 'border-zinc-200/90 border-b-white bg-white text-zinc-900 shadow-sm ring-1 ring-zinc-950/[0.04]'
              : ''
          "
        >
          User management
        </RouterLink>
        <RouterLink
          to="/admin/templates"
          class="inline-flex items-center rounded-t-lg border border-transparent px-4 py-2.5 text-sm font-medium text-zinc-600 transition hover:text-zinc-900"
          :class="
            templatesSection
              ? 'border-zinc-200/90 border-b-white bg-white text-zinc-900 shadow-sm ring-1 ring-zinc-950/[0.04]'
              : ''
          "
        >
          Templates
        </RouterLink>
      </nav>
    </div>
    <RouterView />
  </div>
</template>
