<script setup lang="ts">
import { ref, onMounted, watch, nextTick, provide } from "vue";
import { readItemsViewMode, writeItemsViewMode } from "../composables/itemsViewMode";

const props = defineProps<{
  storageKey: string;
}>();

const mode = ref<"list" | "gallery">("list");
const root = ref<HTMLElement | null>(null);

function setMode(m: "list" | "gallery") {
  mode.value = m;
}

provide("itemsViewMode", mode);
provide("itemsViewSetMode", setMode);

function syncAria() {
  root.value?.querySelectorAll("[data-items-view-mode]").forEach((btn) => {
    const el = btn as HTMLElement;
    const on = el.getAttribute("data-items-view-mode") === mode.value;
    el.setAttribute("aria-pressed", on ? "true" : "false");
  });
}

onMounted(() => {
  mode.value = readItemsViewMode(props.storageKey);
  nextTick(syncAria);
});

watch(
  () => props.storageKey,
  (k) => {
    mode.value = readItemsViewMode(k);
    nextTick(syncAria);
  },
);

watch(mode, (m) => {
  writeItemsViewMode(props.storageKey, m);
  nextTick(syncAria);
});
</script>

<template>
  <div
    ref="root"
    class="items-view-scope"
    :data-items-view="mode"
    :data-items-view-key="storageKey"
  >
    <slot name="header" />
    <slot />
  </div>
</template>
