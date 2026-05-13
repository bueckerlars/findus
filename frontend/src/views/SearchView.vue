<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { api } from "../api";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import FxSvg from "../components/FxSvg.vue";
import ItemListRow from "../components/ItemListRow.vue";
import ItemGalleryCard from "../components/ItemGalleryCard.vue";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxAlert from "../components/primitives/FxAlert.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";
import FxSkeletonList from "../components/primitives/FxSkeletonList.vue";
import FxSkeletonGallery from "../components/primitives/FxSkeletonGallery.vue";

const { t } = useI18n();

type Item = { ID: string; Name: string; Description: string; location_name: string; PhotoPath?: string | null };

const route = useRoute();
const searchInputRef = ref<HTMLInputElement | null>(null);
const q = ref(typeof route.query.q === "string" ? route.query.q : "");
const results = ref<Item[]>([]);
const searchLoading = ref(false);
const searchError = ref<string | null>(null);
let debounceTimer: ReturnType<typeof setTimeout>;
let fetchSeq = 0;

function norm(s: string) {
  return (s || "").toLowerCase().trim();
}

const hasQuery = () => norm(q.value).length > 0;

async function run() {
  const query = q.value;
  if (!norm(query)) {
    fetchSeq++;
    searchLoading.value = false;
    searchError.value = null;
    results.value = [];
    return;
  }
  const seq = ++fetchSeq;
  searchLoading.value = true;
  searchError.value = null;
  try {
    const r = await api<{ query: string; results: Item[] }>("/api/search?q=" + encodeURIComponent(query));
    if (seq !== fetchSeq) return;
    results.value = r.results ?? [];
  } catch (e) {
    if (seq !== fetchSeq) return;
    results.value = [];
    searchError.value = (e as Error).message || t("search.failed");
  } finally {
    if (seq === fetchSeq) searchLoading.value = false;
  }
}

watch(q, () => {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    void run();
  }, 300);
});
watch(
  () => route.query.q,
  (v) => {
    if (typeof v === "string" && v !== q.value) q.value = v;
  },
);
void run();

onMounted(() => {
  nextTick(() => {
    searchInputRef.value?.focus();
  });
});

onUnmounted(() => {
  clearTimeout(debounceTimer);
  fetchSeq++;
});
</script>

<template>
  <ItemsViewToggle storage-key="page_search" class="mx-auto max-w-2xl">
    <template #header>
      <FxPageHeader :title="$t('search.title')" :subtitle="$t('search.liveHint')">
        <template #actions><ItemsViewModeToolbar /></template>
      </FxPageHeader>
      <div>
        <label class="sr-only" for="q">{{ $t("search.queryLabel") }}</label>
        <div class="relative">
          <span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" aria-hidden="true">
            <FxSvg name="magnifyingGlass" class="h-4 w-4" />
          </span>
          <input
            id="q"
            ref="searchInputRef"
            v-model="q"
            type="search"
            :placeholder="$t('search.placeholderShort')"
            class="fx-input !mt-0 fx-input--lg border-zinc-200 pl-10 shadow-sm"
            autocomplete="off"
            :aria-busy="searchLoading && hasQuery()"
          />
        </div>
      </div>
    </template>

    <div class="mt-5">
      <FxAlert v-if="hasQuery() && searchError" tone="warning" class="mb-3">{{ searchError }}</FxAlert>

      <div
        class="items-view-list-only"
        role="region"
        :aria-label="$t('search.resultsAria')"
        :aria-busy="searchLoading && hasQuery()"
      >
        <FxSkeletonList v-if="searchLoading && hasQuery()" :rows="3" :aria-label="$t('common.loadingAria')" />
        <template v-else-if="results.length">
          <div class="overflow-hidden rounded-xl border border-zinc-200/80 bg-white shadow-sm ring-1 ring-zinc-950/[0.03]">
            <div class="divide-y divide-zinc-100">
              <ItemListRow
                v-for="it in results"
                :key="it.ID"
                :id="it.ID"
                :name="it.Name"
                :location-name="it.location_name"
                :photo-path="it.PhotoPath"
              />
            </div>
          </div>
        </template>
        <FxEmptyState v-else-if="hasQuery()" icon="magnifyingGlass" :title="$t('search.noMatchesTry')" />
        <FxEmptyState v-else icon="magnifyingGlass" :title="$t('search.typeToFind')" />
      </div>

      <div
        class="items-view-gallery-only"
        role="region"
        :aria-label="$t('search.resultsAria')"
        :aria-busy="searchLoading && hasQuery()"
      >
        <FxSkeletonGallery v-if="searchLoading && hasQuery()" :tiles="6" :aria-label="$t('common.loadingAria')" />
        <template v-else-if="results.length">
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4">
            <ItemGalleryCard
              v-for="it in results"
              :key="it.ID + '-g'"
              :id="it.ID"
              :name="it.Name"
              :location-name="it.location_name"
              :photo-path="it.PhotoPath"
            />
          </div>
        </template>
        <FxEmptyState v-else-if="hasQuery()" icon="magnifyingGlass" :title="$t('search.noMatchesTry')" />
        <FxEmptyState v-else icon="magnifyingGlass" :title="$t('search.typeToFind')" />
      </div>
    </div>
  </ItemsViewToggle>
</template>
