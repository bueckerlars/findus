<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import type { User } from "../api";
import { useSession } from "../session";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import FxSvg from "../components/FxSvg.vue";
import ItemPhotoPlaceholder from "../components/ItemPhotoPlaceholder.vue";
import { formatItemUpdatedAt } from "../utils/datetime";

type Item = {
  ID: string;
  Name: string;
  Description: string;
  LocationID: string;
  PhotoPath?: string | null;
  CreatedAt: string;
  UpdatedAt: string;
};

type Label = { ID: string; Name: string; Color: string };

type Location = {
  ID: string;
  Name: string;
  Description: string;
};

type RecentRow = { item: Item; location_name: string; recently_added: boolean };
type LocRow = { location: Location; sub_location_count: number; recently_added: boolean };

const data = ref<{
  user: User;
  item_count: number;
  location_count: number;
  label_count: number;
  recent_items: RecentRow[];
  home_locations: LocRow[];
  all_labels: Label[];
} | null>(null);

const { isAdmin } = useSession();

onMounted(async () => {
  data.value = await api("/api/home");
});
</script>

<template>
  <div v-if="!data" class="text-zinc-500">Loading…</div>
  <div v-else class="mx-auto max-w-5xl space-y-6">
    <div class="mb-8">
      <h1 class="text-2xl font-semibold tracking-tight text-zinc-900 sm:text-3xl">Hello, {{ data.user.username }}</h1>
      <p class="mt-2 max-w-xl text-zinc-500">Your Inventory at a glance</p>
    </div>

    <section class="fx-home-stats mb-8" aria-label="Inventory totals">
      <div class="fx-card fx-home-stat">
        <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="cube" class="fx-icon" /></span>
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">Items</p>
          <p class="mt-0.5 text-2xl font-semibold tabular-nums leading-none text-zinc-900">{{ data.item_count }}</p>
        </div>
      </div>
      <div class="fx-card fx-home-stat">
        <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="mapPin" class="fx-icon" /></span>
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">Locations</p>
          <p class="mt-0.5 text-2xl font-semibold tabular-nums leading-none text-zinc-900">{{ data.location_count }}</p>
        </div>
      </div>
      <div class="fx-card fx-home-stat">
        <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="tag" class="fx-icon" /></span>
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">Labels</p>
          <p class="mt-0.5 text-2xl font-semibold tabular-nums leading-none text-zinc-900">{{ data.label_count }}</p>
        </div>
      </div>
    </section>

    <ItemsViewToggle storage-key="home_recent_items" class="fx-card overflow-hidden p-0">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-zinc-100 px-5 py-4">
          <h2 id="home-recent-items" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">Recent items</h2>
          <div class="flex flex-wrap items-center gap-2">
            <ItemsViewModeToolbar />
            <RouterLink to="/items" class="text-sm font-medium text-sky-600 hover:text-sky-700">All items</RouterLink>
          </div>
        </div>
      </template>

      <template v-if="data.recent_items.length">
        <ul class="items-view-list-only divide-y divide-zinc-100">
          <li v-for="row in data.recent_items" :key="row.item.ID">
            <RouterLink
              :to="'/items/' + row.item.ID"
              class="group fx-home-item-row flex items-start gap-3 px-5 py-3.5"
            >
              <span class="fx-item-row-accent" aria-hidden="true"></span>
              <div class="relative z-[1] min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span class="font-medium text-zinc-900 transition-colors duration-200 group-hover:text-sky-900">{{ row.item.Name }}</span>
                  <span
                    v-if="row.recently_added"
                    class="rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 transition-transform duration-200 group-hover:scale-105"
                    >New</span
                  >
                  <span
                    v-else
                    class="rounded-md bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800 transition-transform duration-200 group-hover:scale-105"
                    >Updated</span
                  >
                </div>
                <p
                  class="mt-1.5 flex min-w-0 items-center gap-1.5 text-sm font-medium text-zinc-700 transition-colors group-hover:text-zinc-800"
                >
                  <FxSvg name="mapPin" class="h-3.5 w-3.5 shrink-0 text-sky-600" aria-hidden="true" />
                  <span class="truncate">{{ row.location_name }}</span>
                </p>
              </div>
              <div class="relative z-[1] flex shrink-0 items-start gap-1.5 pt-0.5">
                <time
                  class="text-right text-xs tabular-nums text-zinc-400 transition-colors group-hover:text-zinc-500"
                  :datetime="row.item.UpdatedAt"
                  >{{ formatItemUpdatedAt(row.item.UpdatedAt) }}</time
                >
                <span class="fx-home-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
              </div>
            </RouterLink>
          </li>
        </ul>
        <div class="items-view-gallery-only grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 sm:gap-4 lg:grid-cols-4">
          <RouterLink
            v-for="row in data.recent_items"
            :key="row.item.ID + '-g'"
            :to="'/items/' + row.item.ID"
            class="group fx-item-gallery flex flex-col overflow-hidden rounded-xl border border-zinc-200/80 bg-white shadow-sm ring-1 ring-zinc-950/[0.03]"
          >
            <div class="fx-item-gallery-media aspect-square bg-gradient-to-br from-zinc-50 to-zinc-100 ring-1 ring-zinc-100/80">
              <img
                v-if="row.item.PhotoPath"
                :src="'/items/' + row.item.ID + '/photo'"
                alt=""
                class="fx-item-gallery-photo"
                @error="($event.target as HTMLImageElement).style.display = 'none'"
              />
              <div v-else class="fx-item-gallery-placeholder">
                <ItemPhotoPlaceholder :item-id="row.item.ID" />
              </div>
              <div class="fx-item-gallery-shade" aria-hidden="true"></div>
              <span class="fx-item-gallery-fab" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
            </div>
            <div class="relative z-[1] flex min-h-0 flex-1 flex-col gap-1.5 border-t border-zinc-100/90 p-3">
              <span class="line-clamp-2 font-medium leading-snug text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{
                row.item.Name
              }}</span>
              <p class="flex min-w-0 items-start gap-1.5 text-sm font-medium leading-snug text-zinc-700 transition-colors group-hover:text-zinc-800">
                <FxSvg name="mapPin" class="mt-0.5 h-3.5 w-3.5 shrink-0 text-sky-600" aria-hidden="true" />
                <span class="line-clamp-2 min-w-0">{{ row.location_name }}</span>
              </p>
              <div class="flex flex-wrap items-center gap-1.5">
                <span
                  v-if="row.recently_added"
                  class="inline-flex rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 transition-transform duration-200 group-hover:scale-105"
                  >New</span
                >
                <span
                  v-else
                  class="inline-flex rounded-md bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800 transition-transform duration-200 group-hover:scale-105"
                  >Updated</span
                >
              </div>
            </div>
          </RouterLink>
        </div>
      </template>
      <div v-else class="px-5 py-10 text-center">
        <p class="text-sm text-zinc-500">No items yet.</p>
        <RouterLink v-if="isAdmin" to="/items/new" class="mt-3 inline-flex text-sm font-semibold text-sky-600 hover:text-sky-700"
          >Add an item</RouterLink
        >
      </div>
    </ItemsViewToggle>

    <section class="fx-card overflow-hidden p-0" aria-labelledby="home-locations">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-zinc-100 px-5 py-4">
        <h2 id="home-locations" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">Locations</h2>
        <RouterLink to="/locations" class="text-sm font-medium text-sky-600 hover:text-sky-700">All locations</RouterLink>
      </div>
      <ul v-if="data.home_locations.length" class="divide-y divide-zinc-100" role="list">
        <li v-for="row in data.home_locations" :key="row.location.ID">
          <RouterLink :to="'/locations/' + row.location.ID" class="group fx-home-loc-row flex items-center gap-3 px-5 py-3.5">
            <span class="fx-home-loc-icon" aria-hidden="true"><FxSvg name="mapPin" class="fx-icon" /></span>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-base font-semibold leading-snug text-zinc-900 transition-colors group-hover:text-sky-950">{{
                  row.location.Name
                }}</span>
                <span
                  v-if="row.recently_added"
                  class="rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 transition-transform duration-200 group-hover:scale-[1.02]"
                  >New</span
                >
                <span
                  v-else
                  class="rounded-md bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-800 transition-transform duration-200 group-hover:scale-[1.02]"
                  >Updated</span
                >
              </div>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <span
                v-if="row.sub_location_count > 0"
                class="fx-home-loc-count-badge"
                :aria-label="row.sub_location_count === 1 ? '1 sub-location' : row.sub_location_count + ' sub-locations'"
                >{{ row.sub_location_count }}</span
              >
              <span class="fx-home-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
            </div>
          </RouterLink>
        </li>
      </ul>
      <div v-else class="px-5 py-10 text-center">
        <p class="text-sm text-zinc-500">No locations yet.</p>
        <RouterLink v-if="isAdmin" to="/locations/new" class="mt-3 inline-flex text-sm font-semibold text-sky-600 hover:text-sky-700"
          >Create a location</RouterLink
        >
      </div>
    </section>

    <section class="fx-card overflow-hidden p-0" aria-labelledby="home-labels-heading">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-zinc-100 px-5 py-4">
        <h2 id="home-labels-heading" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">Labels</h2>
        <div class="flex flex-wrap items-center gap-2">
          <RouterLink v-if="isAdmin" to="/labels/new" class="text-sm font-medium text-sky-600 hover:text-sky-700">New label</RouterLink>
          <RouterLink to="/labels" class="text-sm font-medium text-sky-600 hover:text-sky-700">All labels</RouterLink>
        </div>
      </div>
      <div v-if="data.all_labels?.length" class="max-h-80 overflow-y-auto overscroll-y-contain px-5 py-4">
        <div class="flex flex-wrap gap-2.5">
          <RouterLink
            v-for="lb in data.all_labels"
            :key="lb.ID"
            :to="{ path: '/search', query: { q: lb.Name } }"
            class="group fx-home-label-chip max-w-[15rem]"
          >
            <span class="fx-home-label-chip-dot" :style="{ backgroundColor: lb.Color }" aria-hidden="true"></span>
            <span class="min-w-0 truncate">{{ lb.Name }}</span>
          </RouterLink>
        </div>
      </div>
      <div v-else class="px-5 py-10 text-center">
        <p class="text-sm text-zinc-500">No labels configured.</p>
        <RouterLink v-if="isAdmin" to="/labels/new" class="mt-3 inline-flex text-sm font-semibold text-sky-600 hover:text-sky-700"
          >Create a label</RouterLink
        >
      </div>
    </section>
  </div>
</template>
