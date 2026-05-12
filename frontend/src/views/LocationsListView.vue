<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import { useSession } from "../session";
import LocationTreeRow, { type LocationTreeNode } from "../components/LocationTreeRow.vue";

const tree = ref<LocationTreeNode[]>([]);
const expandedIds = ref(new Set<string>());
const { isAdmin } = useSession();

function isExpanded(id: string) {
  return expandedIds.value.has(id);
}

function toggle(id: string) {
  const next = new Set(expandedIds.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  expandedIds.value = next;
}

onMounted(async () => {
  const r = await api<{ tree: LocationTreeNode[] }>("/api/locations");
  tree.value = r.tree;
});
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <div class="mb-8 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight text-zinc-900 sm:text-3xl">Locations</h1>
        <p class="mt-2 max-w-xl text-sm leading-relaxed text-zinc-500">
          Top-level places first. Use the chevron to show or hide sub-locations. Tap the name to open a location.
        </p>
      </div>
      <RouterLink
        v-if="isAdmin"
        to="/locations/new"
        class="fx-btn-primary shrink-0 self-start text-sm shadow-md sm:self-center"
      >
        New location
      </RouterLink>
    </div>

    <ul v-if="tree.length" class="fx-card divide-y divide-zinc-100 overflow-visible p-0" role="list">
      <LocationTreeRow
        v-for="node in tree"
        :key="node.ID"
        :node="node"
        :is-expanded="isExpanded"
        :toggle="toggle"
        :is-admin="isAdmin"
      />
    </ul>
    <div v-else class="fx-card px-5 py-14 text-center">
      <p class="text-zinc-500">No locations yet.</p>
      <RouterLink v-if="isAdmin" to="/locations/new" class="mt-5 inline-flex fx-btn-primary">Create your first location</RouterLink>
    </div>
  </div>
</template>
