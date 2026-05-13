<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import { useCreateModals } from "../composables/useCreateModals";
import { useSession } from "../session";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import ItemListRow from "../components/ItemListRow.vue";
import ItemGalleryCard from "../components/ItemGalleryCard.vue";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";
import FxSkeletonList from "../components/primitives/FxSkeletonList.vue";
import FxSkeletonGallery from "../components/primitives/FxSkeletonGallery.vue";

type Item = {
  ID: string;
  Name: string;
  Description: string;
  LocationID: string;
  location_name: string;
  PhotoPath?: string | null;
};

const items = ref<Item[]>([]);
const loading = ref(true);
const { isAdmin } = useSession();
const { openCreateItem } = useCreateModals();

onMounted(async () => {
  try {
    const r = await api<{ items: Item[] }>("/api/items");
    items.value = r.items;
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <ItemsViewToggle storage-key="page_items" class="mx-auto max-w-3xl">
    <template #header>
      <FxPageHeader :title="$t('items.title')" :subtitle="$t('items.subtitle')">
        <template #actions>
          <ItemsViewModeToolbar />
          <FxButton v-if="isAdmin" variant="primary" size="sm" icon-left="plus" @click="openCreateItem()">{{ $t("items.newItem") }}</FxButton>
          <FxButton variant="secondary" size="sm" :to="'/search'" icon-left="magnifyingGlass">{{ $t("common.search") }}</FxButton>
        </template>
      </FxPageHeader>
    </template>

    <template v-if="loading">
      <div class="items-view-list-only fx-card overflow-hidden p-0">
        <FxSkeletonList :rows="6" :aria-label="$t('common.loadingAria')" />
      </div>
      <div class="items-view-gallery-only fx-card overflow-hidden p-0">
        <FxSkeletonGallery :tiles="8" :aria-label="$t('common.loadingAria')" />
      </div>
    </template>
    <div v-else-if="items.length" class="fx-card overflow-hidden p-0">
      <div class="items-view-list-only divide-y divide-zinc-100">
        <ItemListRow
          v-for="it in items"
          :key="it.ID"
          :id="it.ID"
          :name="it.Name"
          :location-name="it.location_name"
          :photo-path="it.PhotoPath"
        />
      </div>
      <div class="items-view-gallery-only grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 sm:gap-4">
        <ItemGalleryCard
          v-for="it in items"
          :key="it.ID + '-g'"
          :id="it.ID"
          :name="it.Name"
          :location-name="it.location_name"
          :photo-path="it.PhotoPath"
        />
      </div>
    </div>
    <FxEmptyState v-else icon="cube" :title="$t('items.noItems')">
      <FxButton v-if="isAdmin" variant="primary" size="sm" icon-left="plus" @click="openCreateItem()">{{ $t("items.addFirst") }}</FxButton>
    </FxEmptyState>
  </ItemsViewToggle>
</template>
