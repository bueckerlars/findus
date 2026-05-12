<script setup lang="ts">
import { inject, type Ref } from "vue";
import { useI18n } from "vue-i18n";
import FxSvg from "./FxSvg.vue";

const { t } = useI18n();

const mode = inject<Ref<"list" | "gallery">>("itemsViewMode")!;
const setMode = inject<(m: "list" | "gallery") => void>("itemsViewSetMode")!;
</script>

<template>
  <div
    class="inline-flex rounded-xl border border-zinc-200/90 bg-zinc-50/80 p-1 shadow-inner"
    role="toolbar"
    :aria-label="t('itemsView.toolbarAria')"
  >
    <button
      type="button"
      class="items-view-mode-btn"
      data-items-view-mode="list"
      :title="t('itemsView.listView')"
      :aria-label="t('itemsView.listView')"
      :aria-pressed="mode === 'list' ? 'true' : 'false'"
      @click="setMode('list')"
    >
      <FxSvg name="viewList" class="fx-icon shrink-0" />
    </button>
    <button
      type="button"
      class="items-view-mode-btn"
      data-items-view-mode="gallery"
      :title="t('itemsView.galleryView')"
      :aria-label="t('itemsView.galleryView')"
      :aria-pressed="mode === 'gallery' ? 'true' : 'false'"
      @click="setMode('gallery')"
    >
      <FxSvg name="viewGrid" class="fx-icon shrink-0" />
    </button>
  </div>
</template>
