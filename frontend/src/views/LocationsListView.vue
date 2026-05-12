<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import { useSession } from "../session";
import FxSvg from "../components/FxSvg.vue";

type Location = { ID: string; Name: string; Description: string };

const roots = ref<Location[]>([]);
const { isAdmin } = useSession();

onMounted(async () => {
  const r = await api<{ roots: Location[] }>("/api/locations");
  roots.value = r.roots;
});
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Locations</h1>
        <p class="mt-1 text-sm text-zinc-500">Tap a place to open it. Use QR labels to jump here from your phone.</p>
      </div>
    </div>
    <div v-if="roots.length" class="fx-card divide-y divide-zinc-100 overflow-hidden p-0">
      <RouterLink
        v-for="loc in roots"
        :key="loc.ID"
        :to="'/locations/' + loc.ID"
        class="group fx-list-row rounded-none border-0 shadow-none hover:shadow-none"
      >
        <div class="min-w-0">
          <div class="font-medium text-zinc-900">{{ loc.Name }}</div>
          <div v-if="loc.Description" class="mt-0.5 line-clamp-2 text-sm text-zinc-500">{{ loc.Description }}</div>
        </div>
        <span class="shrink-0 text-zinc-400" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
      </RouterLink>
    </div>
    <div v-else class="fx-card px-5 py-12 text-center">
      <p class="text-zinc-500">No locations yet.</p>
      <RouterLink v-if="isAdmin" to="/locations/new" class="mt-4 inline-flex fx-btn-primary">Create your first location</RouterLink>
    </div>
  </div>
</template>
