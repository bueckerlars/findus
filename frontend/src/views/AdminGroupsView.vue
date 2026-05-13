<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import { useI18n } from "vue-i18n";
import { api } from "../api";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxAlert from "../components/primitives/FxAlert.vue";

const { t } = useI18n();

type GroupRow = {
  id: string;
  name: string;
  permissions: string[];
  member_count: number;
  created_at: string;
  updated_at: string;
};

const groups = ref<GroupRow[]>([]);
const err = ref("");

onMounted(load);

async function load() {
  err.value = "";
  try {
    const j = await api<{ groups: GroupRow[] }>("/api/admin/groups");
    groups.value = j.groups ?? [];
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.loadFailed");
  }
}
</script>

<template>
  <div class="w-full space-y-6">
    <FxPageHeader :title="$t('adminGroups.pageTitle')">
      <template #actions>
        <FxButton variant="primary" size="sm" icon-left="plus" to="/admin/groups/new">{{ $t("adminGroups.newGroup") }}</FxButton>
      </template>
    </FxPageHeader>
    <FxAlert v-if="err">{{ err }}</FxAlert>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <p v-if="!groups.length" class="text-sm text-zinc-500">{{ $t("adminGroups.empty") }}</p>
      <table v-else class="w-full text-left text-sm">
        <thead>
          <tr class="border-b border-zinc-200 text-zinc-500">
            <th class="py-2 pr-2">{{ $t("adminGroups.colName") }}</th>
            <th class="py-2 pr-2">{{ $t("adminGroups.colMembers") }}</th>
            <th class="py-2 pr-2">{{ $t("adminGroups.colPermissions") }}</th>
            <th class="py-2 pr-2 w-24"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="g in groups" :key="g.id" class="border-b border-zinc-100">
            <td class="py-2 pr-2 font-medium text-zinc-900">{{ g.name }}</td>
            <td class="py-2 pr-2 text-zinc-600">{{ g.member_count }}</td>
            <td class="py-2 pr-2 text-xs text-zinc-500">{{ (g.permissions || []).join(", ") }}</td>
            <td class="py-2 pr-2">
              <RouterLink class="text-sky-700 hover:text-sky-800" :to="'/admin/groups/' + g.id + '/edit'">{{
                $t("common.edit")
              }}</RouterLink>
            </td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>
