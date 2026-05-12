<script setup lang="ts">
import { ref, watch } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { api } from "../api";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import FxSvg from "../components/FxSvg.vue";
import ItemPhotoPlaceholder from "../components/ItemPhotoPlaceholder.vue";

type Item = { ID: string; Name: string; Description: string; location_name: string; PhotoPath?: string | null };

const route = useRoute();
const q = ref(typeof route.query.q === "string" ? route.query.q : "");
const results = ref<Item[]>([]);
let t: ReturnType<typeof setTimeout>;

async function run() {
  const r = await api<{ query: string; results: Item[] }>("/api/search?q=" + encodeURIComponent(q.value));
  results.value = r.results ?? [];
}

watch(q, () => {
  clearTimeout(t);
  t = setTimeout(run, 300);
});
watch(
  () => route.query.q,
  (v) => {
    if (typeof v === "string" && v !== q.value) q.value = v;
  },
);
void run();
</script>

<template>
  <ItemsViewToggle storage-key="page_search" class="mx-auto max-w-2xl">
    <template #header>
      <div class="flex flex-wrap items-center justify-between gap-3">
        <h1 class="text-2xl font-semibold tracking-tight">Search</h1>
        <ItemsViewModeToolbar />
      </div>
      <p class="mt-1 text-sm text-zinc-500">Results update as you type.</p>
      <div class="mt-6">
        <label class="sr-only" for="q">Search query</label>
        <div class="relative">
          <span class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-400" aria-hidden="true"
            ><FxSvg name="magnifyingGlass" class="fx-icon h-5 w-5"
          /></span>
          <input
            id="q"
            v-model="q"
            type="search"
            placeholder="Search items…"
            class="fx-input !mt-0 border-zinc-200 pl-11 text-base shadow-sm"
            autocomplete="off"
          />
        </div>
      </div>
    </template>

    <div class="mt-6">
      <div class="items-view-list-only space-y-2">
        <template v-if="results.length">
          <RouterLink
            v-for="it in results"
            :key="it.ID"
            :to="'/items/' + it.ID"
            class="group fx-item-row relative fx-list-row"
          >
            <span class="fx-item-row-accent" aria-hidden="true"></span>
            <div class="relative z-[1] min-w-0 flex-1">
              <div class="font-medium text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{ it.Name }}</div>
              <p
                class="mt-1 flex min-w-0 items-center gap-1.5 text-sm font-medium text-zinc-700 transition-colors duration-200 group-hover:text-zinc-800"
              >
                <FxSvg name="mapPin" class="h-3.5 w-3.5 shrink-0 text-sky-600" aria-hidden="true" />
                <span class="truncate">{{ it.location_name }}</span>
              </p>
            </div>
            <span class="fx-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon" /></span>
          </RouterLink>
        </template>
        <div v-else class="rounded-xl border border-dashed border-zinc-200 bg-zinc-50 px-4 py-8 text-center text-sm text-zinc-500">
          No matches — try another word.
        </div>
      </div>
      <div class="items-view-gallery-only grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4">
        <template v-if="results.length">
          <RouterLink
            v-for="it in results"
            :key="it.ID + '-g'"
            :to="'/items/' + it.ID"
            class="group fx-item-gallery flex flex-col overflow-hidden rounded-xl border border-zinc-200/80 bg-white shadow-sm ring-1 ring-zinc-950/[0.03]"
          >
            <div class="fx-item-gallery-media aspect-square bg-gradient-to-br from-zinc-50 to-zinc-100 ring-1 ring-zinc-100/80">
              <img
                v-if="it.PhotoPath"
                :src="'/items/' + it.ID + '/photo'"
                alt=""
                class="fx-item-gallery-photo"
                @error="($event.target as HTMLImageElement).style.display = 'none'"
              />
              <div v-else class="fx-item-gallery-placeholder">
                <ItemPhotoPlaceholder :item-id="it.ID" />
              </div>
              <div class="fx-item-gallery-shade" aria-hidden="true"></div>
              <span class="fx-item-gallery-fab" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
            </div>
            <div class="relative z-[1] flex min-h-0 flex-1 flex-col gap-1.5 border-t border-zinc-100/90 p-3">
              <span class="line-clamp-2 font-medium leading-snug text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{ it.Name }}</span>
              <p class="flex min-w-0 items-start gap-1.5 text-sm font-medium leading-snug text-zinc-700 transition-colors group-hover:text-zinc-800">
                <FxSvg name="mapPin" class="mt-0.5 h-3.5 w-3.5 shrink-0 text-sky-600" aria-hidden="true" />
                <span class="line-clamp-2 min-w-0">{{ it.location_name }}</span>
              </p>
            </div>
          </RouterLink>
        </template>
        <div v-else class="col-span-full rounded-xl border border-dashed border-zinc-200 bg-zinc-50 px-4 py-8 text-center text-sm text-zinc-500">
          No matches — try another word.
        </div>
      </div>
    </div>
  </ItemsViewToggle>
</template>
