<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";
import { api } from "../api";
import type { User } from "../api";
import { useCreateModals } from "../composables/useCreateModals";
import { useSession } from "../session";
import { PERM_ITEMS_WRITE, PERM_LABELS_WRITE, PERM_LOCATIONS_WRITE } from "../permissions";
import { bcp47ForUiLocale, type SupportedLocale } from "../locale/constants";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import FxSvg from "../components/FxSvg.vue";
import ItemPhotoPlaceholder from "../components/ItemPhotoPlaceholder.vue";
import FxSkeleton from "../components/primitives/FxSkeleton.vue";
import FxSkeletonList from "../components/primitives/FxSkeletonList.vue";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";
import FxButton from "../components/primitives/FxButton.vue";
import ItemListRow from "../components/ItemListRow.vue";
import ItemGalleryCard from "../components/ItemGalleryCard.vue";
import LocationListRow from "../components/LocationListRow.vue";
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

const { isAdmin, can } = useSession();
const { openCreateItem, openCreateLocation, openCreateLabel } = useCreateModals();
const { locale } = useI18n();

const dateLocale = computed(() => bcp47ForUiLocale(locale.value as SupportedLocale));

function fmtItemTime(iso: string): string {
  return formatItemUpdatedAt(iso, dateLocale.value);
}

onMounted(async () => {
  data.value = await api("/api/home");
});
</script>

<template>
  <div v-if="!data" class="mx-auto max-w-5xl space-y-6" :aria-label="$t('common.loadingAria')" aria-busy="true">
    <div class="space-y-2">
      <FxSkeleton width="14rem" height="1.5rem" />
      <FxSkeleton width="20rem" height="0.75rem" />
    </div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-3 sm:gap-4">
      <div v-for="n in 3" :key="n" class="fx-card flex items-center gap-3 px-4 py-3">
        <FxSkeleton shape="circle" width="2.5rem" height="2.5rem" />
        <span class="flex flex-1 flex-col gap-1.5">
          <FxSkeleton width="4rem" height="0.5rem" />
          <FxSkeleton width="3rem" height="1.1rem" />
        </span>
      </div>
    </div>
    <div class="fx-card overflow-hidden p-0">
      <FxSkeletonList :rows="4" />
    </div>
  </div>
  <div v-else class="mx-auto max-w-5xl space-y-6">
    <FxPageHeader :title="$t('home.greeting', { name: data.user.username })" :subtitle="$t('home.tagline')" />

    <section class="fx-home-stats mb-6" :aria-label="$t('home.statsAria')">
      <div class="fx-card fx-home-stat">
        <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="cube" class="fx-icon" /></span>
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("home.items") }}</p>
          <p class="mt-0.5 text-2xl font-semibold tabular-nums leading-none text-zinc-900">{{ data.item_count }}</p>
        </div>
      </div>
      <div class="fx-card fx-home-stat">
        <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="mapPin" class="fx-icon" /></span>
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("home.locations") }}</p>
          <p class="mt-0.5 text-2xl font-semibold tabular-nums leading-none text-zinc-900">{{ data.location_count }}</p>
        </div>
      </div>
      <div class="fx-card fx-home-stat">
        <span class="fx-home-stat-icon" aria-hidden="true"><FxSvg name="tag" class="fx-icon" /></span>
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("home.labels") }}</p>
          <p class="mt-0.5 text-2xl font-semibold tabular-nums leading-none text-zinc-900">{{ data.label_count }}</p>
        </div>
      </div>
    </section>

    <ItemsViewToggle storage-key="home_recent_items" class="fx-card overflow-hidden p-0">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-zinc-100 px-5 py-4">
          <h2 id="home-recent-items" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">{{ $t("home.recentItems") }}</h2>
          <div class="flex flex-wrap items-center gap-2">
            <ItemsViewModeToolbar />
            <RouterLink to="/items" class="text-sm font-medium text-sky-600 hover:text-sky-700">{{ $t("home.allItems") }}</RouterLink>
          </div>
        </div>
      </template>

      <template v-if="data.recent_items.length">
        <ul class="items-view-list-only divide-y divide-zinc-100">
          <li v-for="row in data.recent_items" :key="row.item.ID">
            <ItemListRow
              :id="row.item.ID"
              :name="row.item.Name"
              :location-name="row.location_name"
              :photo-path="row.item.PhotoPath"
              :timestamp="fmtItemTime(row.item.UpdatedAt)"
              :timestamp-iso="row.item.UpdatedAt"
              :badge-label="row.recently_added ? $t('home.new') : $t('home.updated')"
              :badge-tone="row.recently_added ? 'success' : 'warning'"
            />
          </li>
        </ul>
        <div class="items-view-gallery-only grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 sm:gap-4 lg:grid-cols-4">
          <ItemGalleryCard
            v-for="row in data.recent_items"
            :key="row.item.ID + '-g'"
            :id="row.item.ID"
            :name="row.item.Name"
            :location-name="row.location_name"
            :photo-path="row.item.PhotoPath"
            :badge-label="row.recently_added ? $t('home.new') : $t('home.updated')"
            :badge-tone="row.recently_added ? 'success' : 'warning'"
          />
        </div>
      </template>
      <FxEmptyState v-else icon="cube" :title="$t('home.noItems')">
        <FxButton v-if="isAdmin || can(PERM_ITEMS_WRITE)" variant="primary" size="sm" icon-left="plus" @click="openCreateItem()">{{ $t("home.addItem") }}</FxButton>
      </FxEmptyState>
    </ItemsViewToggle>

    <section class="fx-card overflow-hidden p-0" aria-labelledby="home-locations">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-zinc-100 px-5 py-4">
        <h2 id="home-locations" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">{{ $t("home.locations") }}</h2>
        <RouterLink to="/locations" class="text-sm font-medium text-sky-600 hover:text-sky-700">{{ $t("home.allLocations") }}</RouterLink>
      </div>
      <ul v-if="data.home_locations.length" class="divide-y divide-zinc-100" role="list">
        <li v-for="row in data.home_locations" :key="row.location.ID">
          <LocationListRow
            :id="row.location.ID"
            :name="row.location.Name"
            :sub-count="row.sub_location_count"
            :sub-count-aria="row.sub_location_count === 1 ? $t('home.subLocationOne') : $t('home.subLocationMany', { n: row.sub_location_count })"
            :badge-label="row.recently_added ? $t('home.new') : $t('home.updated')"
            :badge-tone="row.recently_added ? 'success' : 'warning'"
          />
        </li>
      </ul>
      <FxEmptyState v-else icon="mapPin" :title="$t('home.noLocations')">
        <FxButton v-if="isAdmin || can(PERM_LOCATIONS_WRITE)" variant="primary" size="sm" icon-left="plus" @click="openCreateLocation()">{{ $t("home.createLocation") }}</FxButton>
      </FxEmptyState>
    </section>

    <section class="fx-card overflow-hidden p-0" aria-labelledby="home-labels-heading">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-zinc-100 px-5 py-4">
        <h2 id="home-labels-heading" class="text-sm font-semibold uppercase tracking-wide text-zinc-500">{{ $t("home.labels") }}</h2>
        <div class="flex flex-wrap items-center gap-2">
          <button v-if="isAdmin || can(PERM_LABELS_WRITE)" type="button" class="text-sm font-medium text-sky-600 hover:text-sky-700" @click="openCreateLabel()">
            {{ $t("home.newLabel") }}
          </button>
          <RouterLink to="/labels" class="text-sm font-medium text-sky-600 hover:text-sky-700">{{ $t("home.allLabels") }}</RouterLink>
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
        <p class="text-sm text-zinc-500">{{ $t("home.noLabels") }}</p>
        <button
          v-if="isAdmin || can(PERM_LABELS_WRITE)"
        >
          {{ $t("home.createLabel") }}
        </button>
      </div>
    </section>
  </div>
</template>
