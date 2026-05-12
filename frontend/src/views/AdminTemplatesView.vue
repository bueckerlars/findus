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
  <div class="max-w-4xl space-y-6">
    <div class="flex items-center justify-between gap-4">
      <h1 class="text-2xl font-semibold text-zinc-900">Item templates</h1>
      <RouterLink to="/admin/templates/new" class="fx-btn-primary text-sm">New template</RouterLink>
    </div>
    <ul class="divide-y divide-zinc-100 rounded-2xl border border-zinc-200/80 bg-white shadow-sm">
      <li v-for="row in rows" :key="row.template.ID" class="flex flex-wrap items-center justify-between gap-3 px-4 py-4">
        <div>
          <p class="font-medium text-zinc-900">{{ row.template.DisplayName || row.template.ID }}</p>
          <p class="text-xs text-zinc-500">{{ row.template.ID }} · {{ row.count }} items</p>
        </div>
        <RouterLink :to="'/admin/templates/' + row.template.ID + '/edit'" class="text-sm font-medium text-sky-700">Edit</RouterLink>
      </li>
    </ul>
  </div>
</template>
