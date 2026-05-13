<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, postJson } from "../api";
import { PERM_ITEMS_WRITE, PERM_LABELS_WRITE, PERM_LOCATIONS_WRITE } from "../permissions";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxAlert from "../components/primitives/FxAlert.vue";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const isNew = computed(() => route.path.endsWith("/groups/new"));
const groupId = computed(() => (typeof route.params.id === "string" ? route.params.id : ""));

const name = ref("");
const selected = reactive<Record<string, boolean>>({});
const err = ref("");
const busy = ref(false);

const permDefs = computed(() => [
  { key: PERM_ITEMS_WRITE, label: t("permissions.itemsWrite") },
  { key: PERM_LABELS_WRITE, label: t("permissions.labelsWrite") },
  { key: PERM_LOCATIONS_WRITE, label: t("permissions.locationsWrite") },
]);

onMounted(async () => {
  if (isNew.value) {
    for (const p of permDefs.value) selected[p.key] = false;
    return;
  }
  err.value = "";
  try {
    const j = await api<{ group: { id: string; name: string; permissions: string[] } }>("/api/admin/groups/" + groupId.value);
    name.value = j.group.name;
    const sel: Record<string, boolean> = {};
    for (const p of permDefs.value) sel[p.key] = j.group.permissions?.includes(p.key) ?? false;
    Object.assign(selected, sel);
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.loadFailed");
  }
});

function selectedList(): string[] {
  return permDefs.value.filter((p) => selected[p.key]).map((p) => p.key);
}

async function save() {
  busy.value = true;
  err.value = "";
  try {
    const body = { name: name.value.trim(), permissions: selectedList() };
    if (isNew.value) {
      await postJson<{ id: string; next: string }>("/api/admin/groups", body);
      await router.push("/admin/groups");
    } else {
      await postJson("/api/admin/groups/" + groupId.value, body);
      await router.push("/admin/groups");
    }
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.saveFailed");
  } finally {
    busy.value = false;
  }
}

async function remove() {
  if (!confirm(t("adminGroupForm.confirmDelete"))) return;
  busy.value = true;
  err.value = "";
  try {
    await postJson("/api/admin/groups/" + groupId.value + "/delete", {});
    await router.push("/admin/groups");
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.deleteFailed");
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <div class="w-full max-w-xl space-y-6">
    <FxPageHeader :title="isNew ? $t('adminGroupForm.newTitle') : $t('adminGroupForm.editTitle')" />
    <FxAlert v-if="err">{{ err }}</FxAlert>
    <form class="space-y-4 rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm" @submit.prevent="save">
      <div>
        <label class="block text-sm font-medium text-zinc-700">{{ $t("adminGroupForm.nameLabel") }}</label>
        <input v-model="name" class="fx-input mt-1 w-full" required maxlength="128" />
      </div>
      <fieldset>
        <legend class="text-sm font-medium text-zinc-700">{{ $t("adminGroupForm.permissionsLegend") }}</legend>
        <ul class="mt-2 space-y-2">
          <li v-for="p in permDefs" :key="p.key" class="flex items-center gap-2">
            <input :id="'perm-' + p.key" v-model="selected[p.key]" type="checkbox" class="size-4 rounded border-zinc-300" />
            <label :for="'perm-' + p.key" class="text-sm text-zinc-800">{{ p.label }}</label>
          </li>
        </ul>
      </fieldset>
      <div class="flex flex-wrap gap-2">
        <FxButton type="submit" variant="primary" :disabled="busy">{{ $t("common.save") }}</FxButton>
        <FxButton type="button" variant="secondary" to="/admin/groups">{{ $t("common.cancel") }}</FxButton>
        <FxButton v-if="!isNew" type="button" variant="danger" :disabled="busy" @click="remove()">{{
          $t("common.delete")
        }}</FxButton>
      </div>
    </form>
  </div>
</template>
