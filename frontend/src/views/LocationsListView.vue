<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { api } from "../api";
import { useCreateModals } from "../composables/useCreateModals";
import { useSession } from "../session";
import { PERM_LOCATIONS_WRITE } from "../permissions";
import LocationTreeRow, { type LocationTreeNode } from "../components/LocationTreeRow.vue";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";
import FxSkeletonList from "../components/primitives/FxSkeletonList.vue";

const tree = ref<LocationTreeNode[]>([]);
const loading = ref(true);
const expandedIds = ref(new Set<string>());
const { isAdmin, can } = useSession();
const canManageLocations = computed(() => isAdmin.value || can(PERM_LOCATIONS_WRITE));
const { openCreateLocation } = useCreateModals();

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
  try {
    const r = await api<{ tree: LocationTreeNode[] }>("/api/locations");
    tree.value = r.tree;
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <FxPageHeader :title="$t('locations.title')" :subtitle="$t('locations.subtitle')">
      <template #actions>
        <FxButton v-if="canManageLocations" variant="primary" size="sm" icon-left="plus" @click="openCreateLocation()">{{ $t("locations.newLocation") }}</FxButton>
      </template>
    </FxPageHeader>

    <div v-if="loading" class="fx-card overflow-hidden p-0">
      <FxSkeletonList :rows="6" :aria-label="$t('common.loadingAria')" />
    </div>
    <ul v-else-if="tree.length" class="fx-card divide-y divide-zinc-100 overflow-visible p-0" role="list">
      <LocationTreeRow
        v-for="node in tree"
        :key="node.ID"
        :node="node"
        :is-expanded="isExpanded"
        :toggle="toggle"
        :is-admin="canManageLocations"
      />
    </ul>
    <FxEmptyState v-else icon="mapPin" :title="$t('locations.noLocations')">
      <FxButton v-if="canManageLocations" variant="primary" size="sm" icon-left="plus" @click="openCreateLocation()">{{ $t("locations.createFirst") }}</FxButton>
    </FxEmptyState>
  </div>
</template>
