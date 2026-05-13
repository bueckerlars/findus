<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";

type ItemTemplate = { ID: string; DisplayName: string; SortOrder: number };
type Row = { template: ItemTemplate; count: number; has_alternate: boolean; fallback_display: string };

const rows = ref<Row[]>([]);

onMounted(async () => {
  const r = await api<{ rows: Row[] }>("/api/admin/templates");
  rows.value = r.rows;
});
</script>

<template>
  <div class="w-full space-y-5">
    <FxPageHeader :title="$t('adminTemplates.title')">
      <template #actions>
        <FxButton variant="primary" size="sm" icon-left="plus" :to="'/admin/templates/new'">{{ $t("adminTemplates.new") }}</FxButton>
      </template>
    </FxPageHeader>
    <ul v-if="rows.length" class="divide-y divide-zinc-100 fx-card overflow-hidden p-0">
      <li v-for="row in rows" :key="row.template.ID" class="flex flex-wrap items-center justify-between gap-3 px-4 py-3 sm:px-5">
        <div class="min-w-0">
          <p class="truncate text-sm font-medium text-zinc-900">{{ row.template.DisplayName || row.template.ID }}</p>
          <p class="text-xs text-zinc-500">{{ row.template.ID }} · {{ $t("adminTemplates.itemCount", { n: row.count }) }}</p>
        </div>
        <RouterLink :to="'/admin/templates/' + row.template.ID + '/edit'" class="text-xs font-semibold text-sky-700 hover:text-sky-800">{{ $t("adminTemplates.edit") }}</RouterLink>
      </li>
    </ul>
    <FxEmptyState v-else icon="documentText" :title="$t('adminTemplates.empty')">
      <FxButton variant="primary" size="sm" icon-left="plus" :to="'/admin/templates/new'">{{ $t("adminTemplates.new") }}</FxButton>
    </FxEmptyState>
  </div>
</template>
