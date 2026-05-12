<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { api, postJson } from "../api";
import { confirmAlert } from "../composables/useAlertDialog";
import { toast } from "../composables/useToast";
import TemplateFieldsEditor from "../components/TemplateFieldsEditor.vue";
import { parseTemplateFieldsJson, validateTemplateFields } from "../types/templateFields";

const route = useRoute();
const router = useRouter();
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
    err.value = parsed.error;
    return;
  }
  const ve = validateTemplateFields(parsed.fields);
  if (ve) {
    err.value = ve;
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
    toast.success("Template saved.");
    await router.push("/admin/templates");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
  }
}

async function del() {
  const ok = await confirmAlert({
    title: "Delete this template?",
    message: "Items may need reassignment to another template.",
    confirmLabel: "Delete",
    variant: "danger",
  });
  if (!ok) return;
  try {
    await postJson("/api/admin/templates/" + id.value + "/delete", {});
    toast.success("Template deleted.");
    await router.push("/admin/templates");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Delete failed";
  }
}
</script>

<template>
  <div class="max-w-3xl space-y-6">
    <p>
      <RouterLink to="/admin/templates" class="text-sm font-medium text-sky-700 hover:text-sky-800">← Templates</RouterLink>
    </p>
    <h1 class="text-2xl font-semibold text-zinc-900">{{ isNew ? "New template" : "Edit template" }}</h1>
    <p v-if="err" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
    <form class="space-y-4" @submit.prevent="save">
      <div v-if="isNew">
        <label class="fx-label" for="tid">Template ID</label>
        <input id="tid" v-model="tplId" class="fx-input font-mono text-sm" required placeholder="e.g. laptop" />
      </div>
      <div>
        <label class="fx-label" for="tdn">Display name</label>
        <input id="tdn" v-model="displayName" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="tso">Sort order</label>
        <input id="tso" v-model.number="sortOrder" type="number" class="fx-input w-32" />
      </div>
      <div id="tfe-root">
        <label class="fx-label">Template fields</label>
        <TemplateFieldsEditor v-model="fieldsJson" />
      </div>
      <div class="flex gap-2">
        <button type="submit" class="fx-btn-primary">Save</button>
        <button v-if="!isNew" type="button" class="rounded-xl border border-red-200 px-4 py-2 text-sm text-red-700" @click="del">Delete</button>
      </div>
    </form>
  </div>
</template>
