<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import { useCreateModals } from "../composables/useCreateModals";
import { useSession } from "../session";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";
import FxSkeletonList from "../components/primitives/FxSkeletonList.vue";

type Label = { ID: string; Name: string; Color: string };
type Row = { label: Label; chip_href: string; default_template_title?: string };

const rows = ref<Row[]>([]);
const loading = ref(true);
const { isAdmin } = useSession();
const { openCreateLabel } = useCreateModals();

onMounted(async () => {
  try {
    const r = await api<{ label_rows: Row[] }>("/api/labels");
    rows.value = r.label_rows;
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="max-w-3xl space-y-5">
    <FxPageHeader :title="$t('labels.title')" :subtitle="$t('labels.subtitle')">
      <template #actions>
        <FxButton v-if="isAdmin" variant="primary" size="sm" icon-left="plus" @click="openCreateLabel()">{{ $t("labels.newLabel") }}</FxButton>
      </template>
    </FxPageHeader>

    <div v-if="loading" class="fx-card overflow-hidden p-0">
      <FxSkeletonList :rows="6" :aria-label="$t('common.loadingAria')" />
    </div>
    <ul v-else-if="rows.length" class="divide-y divide-zinc-100 fx-card overflow-hidden p-0">
      <li v-for="row in rows" :key="row.label.ID" class="flex flex-wrap items-center justify-between gap-3 px-4 py-2.5 sm:px-5">
        <RouterLink :to="row.chip_href" class="flex min-w-0 items-center gap-2.5">
          <span class="h-3 w-3 shrink-0 rounded-full ring-1 ring-zinc-200/80" :style="{ backgroundColor: row.label.Color }" />
          <span class="truncate text-sm font-medium text-zinc-900">{{ row.label.Name }}</span>
        </RouterLink>
        <div class="flex shrink-0 items-center gap-3">
          <span v-if="row.default_template_title" class="text-xs text-zinc-500">{{ $t("labels.defaultLine", { name: row.default_template_title }) }}</span>
          <RouterLink
            v-if="isAdmin"
            :to="'/labels/' + row.label.ID + '/edit'"
            class="text-xs font-semibold text-sky-700 hover:text-sky-800"
            >{{ $t("labels.edit") }}</RouterLink
          >
        </div>
      </li>
    </ul>
    <FxEmptyState v-else icon="tag" :title="$t('labels.noLabels')">
      <FxButton v-if="isAdmin" variant="primary" size="sm" icon-left="plus" @click="openCreateLabel()">{{ $t("labels.createFirst") }}</FxButton>
    </FxEmptyState>
  </div>
</template>
