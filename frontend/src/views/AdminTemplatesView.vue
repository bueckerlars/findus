<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";
import { api } from "../api";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";

type ItemTemplate = { ID: string; DisplayName: string; SortOrder: number };
type Row = { template: ItemTemplate; count: number; has_alternate: boolean; fallback_display: string };

const { t } = useI18n();
const rows = ref<Row[]>([]);

/** Slug prefix only when it adds information beyond the title (same name with different spelling is omitted). */
function templateSlugForSubtitle(t: ItemTemplate): string {
  const id = (t.ID ?? "").trim();
  const dn = (t.DisplayName ?? "").trim();
  if (!id) return "";
  if (!dn || dn.toLocaleLowerCase("en-US") === id.toLocaleLowerCase("en-US")) return "";
  return id;
}

function itemCountForRow(row: Row): number {
  const n = row.count as unknown;
  if (typeof n === "number" && Number.isFinite(n)) return n;
  if (typeof n === "string" && n.trim() !== "") {
    const parsed = Number(n);
    if (Number.isFinite(parsed)) return parsed;
  }
  return 0;
}

function templateRowMetaLine(row: Row): string {
  const slug = templateSlugForSubtitle(row.template);
  const itemsPart = t("adminTemplates.itemCount", { n: itemCountForRow(row) });
  return slug ? `${slug} · ${itemsPart}` : itemsPart;
}

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
          <p class="text-xs text-zinc-500">{{ templateRowMetaLine(row) }}</p>
        </div>
        <RouterLink :to="'/admin/templates/' + row.template.ID + '/edit'" class="text-xs font-semibold text-sky-700 hover:text-sky-800">{{ $t("adminTemplates.edit") }}</RouterLink>
      </li>
    </ul>
    <FxEmptyState v-else icon="documentText" :title="$t('adminTemplates.empty')">
      <FxButton variant="primary" size="sm" icon-left="plus" :to="'/admin/templates/new'">{{ $t("adminTemplates.new") }}</FxButton>
    </FxEmptyState>
  </div>
</template>
