<script setup lang="ts">
import { onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { useCreateModals } from "../composables/useCreateModals";
import CreateItemModal from "./CreateItemModal.vue";
import CreateLocationModal from "./CreateLocationModal.vue";
import CreateLabelModal from "./CreateLabelModal.vue";

const route = useRoute();
const { createModalState, consumePending, openCreateItem, openCreateLocation, openCreateLabel } = useCreateModals();

function applyPending() {
  const p = consumePending();
  if (!p) return;
  if (p.kind === "item") {
    openCreateItem({ locationId: p.locationId });
  } else if (p.kind === "location") {
    openCreateLocation({ parentId: p.parentId });
  } else {
    openCreateLabel();
  }
}

onMounted(applyPending);
watch(() => route.fullPath, applyPending);
</script>

<template>
  <CreateItemModal
    :model-value="createModalState.itemOpen"
    :initial-location-id="createModalState.itemInitialLocationId"
    @update:model-value="createModalState.itemOpen = $event"
  />
  <CreateLocationModal
    :model-value="createModalState.locationOpen"
    :initial-parent-id="createModalState.locationInitialParentId"
    @update:model-value="createModalState.locationOpen = $event"
  />
  <CreateLabelModal :model-value="createModalState.labelOpen" @update:model-value="createModalState.labelOpen = $event" />
</template>
