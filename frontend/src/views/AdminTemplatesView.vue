<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";

type ItemTemplate = { ID: string; DisplayName: string; SortOrder: number };
type Row = { template: ItemTemplate; count: number; has_alternate: boolean; fallback_display: string };

const rows = ref<Row[]>([]);

onMounted(async () => {
  const r = await api<{ rows: Row[] }>("/api/admin/templates");
  rows.value = r.rows;
});
</script>

<template>
    <div class="w-full space-y-6">
    <div class="flex items-center justify-between gap-4">
      <h1 class="text-2xl font-semibold text-zinc-900">{{ $t("adminTemplates.title") }}</h1>
      <RouterLink to="/admin/templates/new" class="fx-btn-primary text-sm">{{ $t("adminTemplates.new") }}</RouterLink>
    </div>
    <ul v-if="rows.length" class="divide-y divide-zinc-100 rounded-2xl border border-zinc-200/80 bg-white shadow-sm">
      <li v-for="row in rows" :key="row.template.ID" class="flex flex-wrap items-center justify-between gap-3 px-4 py-4">
        <div>
          <p class="font-medium text-zinc-900">{{ row.template.DisplayName || row.template.ID }}</p>
          <p class="text-xs text-zinc-500">{{ row.template.ID }} · {{ $t("adminTemplates.itemCount", { n: row.count }) }}</p>
        </div>
        <RouterLink :to="'/admin/templates/' + row.template.ID + '/edit'" class="text-sm font-medium text-sky-700">{{ $t("adminTemplates.edit") }}</RouterLink>
      </li>
    </ul>
    <div v-else class="fx-card px-5 py-14 text-center">
      <p class="text-zinc-500">{{ $t("adminTemplates.empty") }}</p>
      <RouterLink to="/admin/templates/new" class="mt-5 inline-flex fx-btn-primary text-sm">{{ $t("adminTemplates.new") }}</RouterLink>
    </div>
  </div>
</template>
