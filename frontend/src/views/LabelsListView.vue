<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { api } from "../api";
import { useCreateModals } from "../composables/useCreateModals";
import { useSession } from "../session";

type Label = { ID: string; Name: string; Color: string };
type Row = { label: Label; chip_href: string; default_template_title?: string };

const rows = ref<Row[]>([]);
const { isAdmin } = useSession();
const { openCreateLabel } = useCreateModals();

onMounted(async () => {
  const r = await api<{ label_rows: Row[] }>("/api/labels");
  rows.value = r.label_rows;
});
</script>

<template>
  <div class="max-w-3xl space-y-6">
    <div class="flex flex-wrap items-end justify-between gap-4">
      <h1 class="text-2xl font-semibold text-zinc-900">Labels</h1>
      <button v-if="isAdmin" type="button" class="fx-btn-primary text-sm" @click="openCreateLabel()">New label</button>
    </div>
    <ul class="divide-y divide-zinc-100 rounded-2xl border border-zinc-200/80 bg-white shadow-sm">
      <li v-for="row in rows" :key="row.label.ID" class="flex flex-wrap items-center justify-between gap-3 px-4 py-4">
        <RouterLink :to="row.chip_href" class="flex items-center gap-3">
          <span class="h-3 w-3 shrink-0 rounded-full ring-1 ring-zinc-200/80" :style="{ backgroundColor: row.label.Color }" />
          <span class="font-medium text-zinc-900">{{ row.label.Name }}</span>
        </RouterLink>
        <span v-if="row.default_template_title" class="text-xs text-zinc-500">Default: {{ row.default_template_title }}</span>
        <RouterLink
          v-if="isAdmin"
          :to="'/labels/' + row.label.ID + '/edit'"
          class="text-sm font-medium text-sky-700 hover:text-sky-800"
          >Edit</RouterLink
        >
      </li>
    </ul>
  </div>
</template>
