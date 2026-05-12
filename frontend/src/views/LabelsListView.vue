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
      <div>
        <h1 class="text-2xl font-semibold text-zinc-900">{{ $t("labels.title") }}</h1>
        <p class="mt-1 max-w-xl text-sm text-zinc-500">{{ $t("labels.subtitle") }}</p>
      </div>
      <button v-if="isAdmin" type="button" class="fx-btn-primary text-sm" @click="openCreateLabel()">{{ $t("labels.newLabel") }}</button>
    </div>
    <ul v-if="rows.length" class="divide-y divide-zinc-100 rounded-2xl border border-zinc-200/80 bg-white shadow-sm">
      <li v-for="row in rows" :key="row.label.ID" class="flex flex-wrap items-center justify-between gap-3 px-4 py-4">
        <RouterLink :to="row.chip_href" class="flex items-center gap-3">
          <span class="h-3 w-3 shrink-0 rounded-full ring-1 ring-zinc-200/80" :style="{ backgroundColor: row.label.Color }" />
          <span class="font-medium text-zinc-900">{{ row.label.Name }}</span>
        </RouterLink>
        <span v-if="row.default_template_title" class="text-xs text-zinc-500">{{ $t("labels.defaultLine", { name: row.default_template_title }) }}</span>
        <RouterLink
          v-if="isAdmin"
          :to="'/labels/' + row.label.ID + '/edit'"
          class="text-sm font-medium text-sky-700 hover:text-sky-800"
          >{{ $t("labels.edit") }}</RouterLink
        >
      </li>
    </ul>
    <div v-else class="fx-card px-5 py-14 text-center">
      <p class="text-zinc-500">{{ $t("labels.noLabels") }}</p>
      <button v-if="isAdmin" type="button" class="mt-5 inline-flex fx-btn-primary" @click="openCreateLabel()">
        {{ $t("labels.createFirst") }}
      </button>
    </div>
  </div>
</template>
