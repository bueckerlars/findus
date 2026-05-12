<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api, postJson } from "../api";
import { confirmAlert } from "../composables/useAlertDialog";
import { toast } from "../composables/useToast";

type ItemTemplate = { ID: string; DisplayName: string };
type Label = { ID: string; Name: string; Color: string; DefaultTemplateType?: string | null };

const route = useRoute();
const router = useRouter();
const id = computed(() => route.params.id as string);

const name = ref("");
const color = ref("#6366f1");
const defaultTemplateType = ref("");
const templates = ref<ItemTemplate[]>([]);
const err = ref("");

onMounted(async () => {
  const r = await api<{ label: Label; templates: ItemTemplate[]; selected_template: string }>("/api/labels/" + id.value + "/edit");
  name.value = r.label.Name;
  color.value = r.label.Color;
  templates.value = r.templates;
  defaultTemplateType.value = r.selected_template || "";
});

async function save() {
  err.value = "";
  try {
    const r = await postJson<{ next: string }>("/api/labels/" + id.value, {
      name: name.value,
      color: color.value,
      default_template_type: defaultTemplateType.value,
    });
    toast.success("Label saved.");
    await router.push(r.next);
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
  }
}

async function del() {
  const ok = await confirmAlert({
    title: "Delete this label?",
    message: "This cannot be undone.",
    confirmLabel: "Delete",
    variant: "danger",
  });
  if (!ok) return;
  try {
    await postJson("/api/labels/" + id.value + "/delete", {});
    toast.success("Label deleted.");
    await router.push("/labels");
  } catch (e) {
    toast.error(e instanceof Error ? e.message : "Delete failed");
  }
}
</script>

<template>
  <div class="max-w-lg space-y-6">
    <h1 class="text-2xl font-semibold text-zinc-900">Edit label</h1>
    <p v-if="err" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
    <form class="space-y-4" @submit.prevent="save">
      <div>
        <label class="fx-label" for="ln">Name</label>
        <input id="ln" v-model="name" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="lc">Color</label>
        <input id="lc" v-model="color" type="color" class="h-10 w-20 cursor-pointer rounded border border-zinc-200" />
      </div>
      <div>
        <label class="fx-label" for="lt">Default template</label>
        <select id="lt" v-model="defaultTemplateType" class="fx-input">
          <option value="">— None —</option>
          <option v-for="t in templates" :key="t.ID" :value="t.ID">{{ t.DisplayName || t.ID }}</option>
        </select>
      </div>
      <div class="flex gap-2">
        <button type="submit" class="fx-btn-primary">Save</button>
        <button type="button" class="rounded-xl border border-red-200 px-4 py-2 text-sm text-red-700" @click="del">Delete</button>
      </div>
    </form>
  </div>
</template>
