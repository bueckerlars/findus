<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, postJson } from "../api";
import { confirmAlert } from "../composables/useAlertDialog";
import { toast } from "../composables/useToast";
import TemplateFieldsEditor from "../components/TemplateFieldsEditor.vue";
import { parseTemplateFieldsJson, validateTemplateFields } from "../types/templateFields";
import { formatParseFieldsIssue, formatValidateFieldsIssue } from "../utils/fieldValidationMessages";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const isNew = computed(() => route.path.endsWith("/new"));
const id = computed(() => (isNew.value ? "" : (route.params.id as string)));

const tplId = ref("");
const displayName = ref("");
const sortOrder = ref(10);
const fieldsJson = ref("[]");
const err = ref("");

onMounted(async () => {
  if (isNew.value) {
    const r = await api<{ sort_order: number; fields_json: string }>("/api/admin/templates/new");
    sortOrder.value = r.sort_order;
    fieldsJson.value = r.fields_json;
  } else {
    const r = await api<{ template: { ID: string; DisplayName: string; SortOrder: number }; fields_json: string }>(
      "/api/admin/templates/" + id.value,
    );
    tplId.value = r.template.ID;
    displayName.value = r.template.DisplayName;
    sortOrder.value = r.template.SortOrder;
    fieldsJson.value = r.fields_json;
  }
});

async function save() {
  err.value = "";
  const parsed = parseTemplateFieldsJson(fieldsJson.value);
  if (!parsed.ok) {
    err.value = formatParseFieldsIssue(parsed.issue, t);
    return;
  }
  const ve = validateTemplateFields(parsed.fields);
  if (ve) {
    err.value = formatValidateFieldsIssue(ve, t);
    return;
  }
  try {
    if (isNew.value) {
      await postJson("/api/admin/templates", {
        id: tplId.value.trim(),
        display_name: displayName.value,
        fields_json: fieldsJson.value,
        sort_order: sortOrder.value,
      });
    } else {
      await postJson("/api/admin/templates/" + id.value, {
        id: id.value,
        display_name: displayName.value,
        fields_json: fieldsJson.value,
        sort_order: sortOrder.value,
      });
    }
    toast.success(t("toast.templateSaved"));
    await router.push("/admin/templates");
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.saveFailed");
  }
}

async function del() {
  const ok = await confirmAlert({
    title: t("adminTpl.deleteTitle"),
    message: t("adminTpl.deleteMsg"),
    confirmLabel: t("common.delete"),
    variant: "danger",
  });
  if (!ok) return;
  try {
    await postJson("/api/admin/templates/" + id.value + "/delete", {});
    toast.success(t("toast.templateDeleted"));
    await router.push("/admin/templates");
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.deleteFailed");
  }
}
</script>

<template>
  <div class="max-w-3xl space-y-6">
    <p>
      <RouterLink to="/admin/templates" class="text-sm font-medium text-sky-700 hover:text-sky-800">{{ $t("adminTpl.back") }}</RouterLink>
    </p>
    <h1 class="text-2xl font-semibold text-zinc-900">{{ isNew ? $t("adminTpl.newTitle") : $t("adminTpl.editTitle") }}</h1>
    <p v-if="err" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
    <form class="space-y-4" @submit.prevent="save">
      <div v-if="isNew">
        <label class="fx-label" for="tid">{{ $t("adminTpl.templateId") }}</label>
        <input id="tid" v-model="tplId" class="fx-input font-mono text-sm" required :placeholder="$t('adminTpl.idPlaceholder')" />
      </div>
      <div>
        <label class="fx-label" for="tdn">{{ $t("adminTpl.displayName") }}</label>
        <input id="tdn" v-model="displayName" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="tso">{{ $t("adminTpl.sortOrder") }}</label>
        <input id="tso" v-model.number="sortOrder" type="number" class="fx-input w-32" />
      </div>
      <div id="tfe-root">
        <label class="fx-label">{{ $t("adminTpl.fieldsEditorLabel") }}</label>
        <TemplateFieldsEditor v-model="fieldsJson" />
      </div>
      <div class="flex gap-2">
        <button type="submit" class="fx-btn-primary">{{ $t("adminTpl.save") }}</button>
        <button v-if="!isNew" type="button" class="rounded-xl border border-red-200 px-4 py-2 text-sm text-red-700" @click="del">{{ $t("adminTpl.delete") }}</button>
      </div>
    </form>
  </div>
</template>
