<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import { useSession } from "../session";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import FxSvg from "../components/FxSvg.vue";
import ItemPhotoPlaceholder from "../components/ItemPhotoPlaceholder.vue";

type Item = {
  ID: string;
  Name: string;
  Description: string;
  LocationID: string;
  location_name: string;
  PhotoPath?: string | null;
};

const items = ref<Item[]>([]);
const { isAdmin } = useSession();

onMounted(async () => {
  const r = await api<{ items: Item[] }>("/api/items");
  items.value = r.items;
});
</script>

<template>
  <ItemsViewToggle storage-key="page_items" class="mx-auto max-w-3xl">
    <template #header>
      <div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Items</h1>
          <p class="mt-1 text-sm text-zinc-500">Everything you track. Use search to find something quickly.</p>
        </div>
        <div class="flex shrink-0 flex-wrap items-center gap-2">
          <ItemsViewModeToolbar />
          <RouterLink to="/search" class="fx-btn-secondary text-sm">Search</RouterLink>
        </div>
      </div>
    </template>

    <div v-if="items.length" class="fx-card overflow-hidden p-0">
      <div class="items-view-list-only divide-y divide-zinc-100">
        <RouterLink
          v-for="it in items"
          :key="it.ID"
          :to="'/items/' + it.ID"
          class="group fx-item-row relative fx-list-row rounded-none border-0 shadow-none hover:shadow-sm"
        >
          <span class="fx-item-row-accent" aria-hidden="true"></span>
          <div class="relative z-[1] min-w-0 flex-1">
            <span class="font-medium text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{ it.Name }}</span>
            <p
              class="mt-1 flex min-w-0 items-center gap-1.5 text-sm font-medium text-zinc-700 transition-colors duration-200 group-hover:text-zinc-800"
            >
              <FxSvg name="mapPin" class="h-3.5 w-3.5 shrink-0 text-sky-600" aria-hidden="true" />
              <span class="truncate">{{ it.location_name }}</span>
            </p>
          </div>
          <span class="fx-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon" /></span>
        </RouterLink>
      </div>
      <div class="items-view-gallery-only grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 sm:gap-4">
        <RouterLink
          v-for="it in items"
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
      </div>
    </div>
    <div v-else class="fx-card px-5 py-12 text-center">
      <p class="text-zinc-500">No items yet.</p>
      <RouterLink v-if="isAdmin" to="/items/new" class="mt-4 inline-flex fx-btn-primary">Add your first item</RouterLink>
    </div>
  </ItemsViewToggle>
</template>
