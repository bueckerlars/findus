<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick, provide } from "vue";
import { readItemsViewMode, writeItemsViewMode, FINDUS_ITEMS_VIEW_MODE_EVENT } from "../composables/itemsViewMode";

const props = defineProps<{
  storageKey: string;
}>();

const mode = ref<"list" | "gallery">("list");
const root = ref<HTMLElement | null>(null);

function setMode(m: "list" | "gallery") {
  mode.value = m;
}

function onExternalItemsViewMode(e: Event) {
  const ce = e as CustomEvent<{ storageKey: string; mode: "list" | "gallery" }>;
  const d = ce.detail;
  if (!d || d.storageKey !== props.storageKey) return;
  setMode(d.mode);
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
  window.addEventListener(FINDUS_ITEMS_VIEW_MODE_EVENT, onExternalItemsViewMode);
});

onUnmounted(() => {
  window.removeEventListener(FINDUS_ITEMS_VIEW_MODE_EVENT, onExternalItemsViewMode);
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
