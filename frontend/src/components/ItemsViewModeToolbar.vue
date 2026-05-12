<script setup lang="ts">
import { inject, type Ref } from "vue";
import FxSvg from "./FxSvg.vue";

const mode = inject<Ref<"list" | "gallery">>("itemsViewMode")!;
const setMode = inject<(m: "list" | "gallery") => void>("itemsViewSetMode")!;
</script>

<template>
  <div
    class="inline-flex rounded-xl border border-zinc-200/90 bg-zinc-50/80 p-1 shadow-inner"
    role="toolbar"
    aria-label="Item layout"
  >
    <button
      type="button"
      class="items-view-mode-btn"
      data-items-view-mode="list"
      title="List view"
      aria-label="List view"
      :aria-pressed="mode === 'list' ? 'true' : 'false'"
      @click="setMode('list')"
    >
      <FxSvg name="viewList" class="fx-icon shrink-0" />
    </button>
    <button
      type="button"
      class="items-view-mode-btn"
      data-items-view-mode="gallery"
      title="Gallery view"
      aria-label="Gallery view"
      :aria-pressed="mode === 'gallery' ? 'true' : 'false'"
      @click="setMode('gallery')"
    >
      <FxSvg name="viewGrid" class="fx-icon shrink-0" />
    </button>
  </div>
</template>
