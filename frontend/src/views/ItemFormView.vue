<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";

type TemplateField = {
  key: string;
  label: string;
  widget: string;
  required: boolean;
  placeholder?: string;
  options?: { value: string; label: string }[];
};

type LocOpt = { ID: string; Label: string };
type Label = { ID: string; Name: string; Color: string };
type ItemTemplate = { ID: string; DisplayName: string };
type Item = {
  ID: string;
  Name: string;
  Description: string;
  LocationID: string;
  TemplateType: string;
};
type KV = { K: string; V: string };

const route = useRoute();
const router = useRouter();
const isNew = computed(() => route.path.endsWith("/new"));
const id = computed(() => (isNew.value ? "" : (route.params.id as string)));

const name = ref("");
const description = ref("");
const locationId = ref("");
const templateType = ref("");
const locations = ref<LocOpt[]>([]);
const templates = ref<ItemTemplate[]>([]);
const allLabels = ref<Label[]>([]);
const selectedLabels = ref<Record<string, boolean>>({});
const fields = ref<TemplateField[]>([]);
const fieldVals = ref<Record<string, string>>({});
const addPairs = ref<{ k: string; v: string }[]>([{ k: "", v: "" }]);
const photo = ref<File | null>(null);
const err = ref("");
const noLocations = ref(false);

async function onTplChange() {
  if (!isNew.value) {
    return;
  }
  fieldVals.value = {};
  await loadFields();
  addPairs.value = [{ k: "", v: "" }];
}

async function loadFields() {
  if (!templateType.value) {
    fields.value = [];
    fieldVals.value = {};
    return;
  }
  const r = await api<{ fields: TemplateField[] }>(
    "/api/items/new/fields?template_type=" + encodeURIComponent(templateType.value),
  );
  fields.value = r.fields;
  const next: Record<string, string> = {};
  for (const f of r.fields) {
    next[f.key] = fieldVals.value[f.key] ?? "";
  }
  fieldVals.value = next;
}

function applyMergedRows(rows: KV[]) {
  const merged = Object.fromEntries(rows.map((r) => [r.K, r.V]));
  const tplKeys = new Set(fields.value.map((f) => f.key));
  for (const f of fields.value) {
    fieldVals.value[f.key] = merged[f.key] != null ? String(merged[f.key]) : "";
  }
  const extras: { k: string; v: string }[] = [];
  for (const r of rows) {
    if (!tplKeys.has(r.K)) {
      extras.push({ k: r.K, v: r.V });
    }
  }
  addPairs.value = extras.length ? extras : [{ k: "", v: "" }];
}

async function load() {
  err.value = "";
  if (isNew.value) {
    const locQ = typeof route.query.location_id === "string" ? route.query.location_id : "";
    const r = await api<{
      no_locations?: boolean;
      locations?: LocOpt[];
      selected_location?: string;
      all_labels?: Label[];
      templates?: ItemTemplate[];
      default_template_id?: string;
    }>("/api/items/new" + (locQ ? "?location_id=" + encodeURIComponent(locQ) : ""));
    if (r.no_locations) {
      noLocations.value = true;
      return;
    }
    noLocations.value = false;
    locations.value = r.locations || [];
    locationId.value = r.selected_location || locQ || "";
    allLabels.value = r.all_labels || [];
    templates.value = r.templates || [];
    templateType.value = r.default_template_id || (r.templates && r.templates[0]?.ID) || "";
    const sl: Record<string, boolean> = {};
    for (const lb of allLabels.value) {
      sl[lb.ID] = false;
    }
    selectedLabels.value = sl;
    await loadFields();
    addPairs.value = [{ k: "", v: "" }];
  } else {
    noLocations.value = false;
    const r = await api<{
      item: Item;
      locations: LocOpt[];
      all_labels: Label[];
      selected_labels: Record<string, boolean>;
      additional_rows: KV[];
    }>("/api/items/" + id.value + "/edit");
    name.value = r.item.Name;
    description.value = r.item.Description;
    locationId.value = r.item.LocationID;
    locations.value = r.locations;
    allLabels.value = r.all_labels;
    const sl: Record<string, boolean> = {};
    for (const lb of allLabels.value) {
      sl[lb.ID] = !!r.selected_labels[lb.ID];
    }
    selectedLabels.value = sl;
    templateType.value = String(r.item.TemplateType);
    await loadFields();
    applyMergedRows(r.additional_rows || []);
  }
}

onMounted(load);
watch(
  () => route.fullPath,
  () => {
    void load();
  },
);

function onPhoto(e: Event) {
  const t = (e.target as HTMLInputElement).files?.[0];
  photo.value = t || null;
}

async function submit() {
  err.value = "";
  const fd = new FormData();
  fd.append("name", name.value);
  fd.append("description", description.value);
  fd.append("location_id", locationId.value);
  fd.append("template_type", templateType.value);
  for (const lid of Object.keys(selectedLabels.value).filter((k) => selectedLabels.value[k])) {
    fd.append("label_id", lid);
  }
  for (const row of addPairs.value) {
    const k = row.k.trim();
    if (k !== "") {
      fd.append("add_k", k);
      fd.append("add_v", row.v);
    }
  }
  for (const f of fields.value) {
    fd.append(f.key, fieldVals.value[f.key] ?? "");
  }
  if (photo.value) {
    fd.append("photo", photo.value);
  }
  try {
    const r = await api<{ next: string }>(isNew.value ? "/api/items" : "/api/items/" + id.value, {
      method: "POST",
      body: fd,
    });
    await router.push(r.next);
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
  }
}
</script>

<template>
  <div class="max-w-2xl space-y-6">
    <h1 class="text-2xl font-semibold text-zinc-900">{{ isNew ? "New item" : "Edit item" }}</h1>
    <p v-if="noLocations" class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
      Create a location before adding items.
    </p>
    <template v-else>
      <p v-if="err" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
      <form class="space-y-4" @submit.prevent="submit">
        <div>
          <label class="fx-label" for="iname">Name</label>
          <input id="iname" v-model="name" class="fx-input" required />
        </div>
        <div>
          <label class="fx-label" for="idesc">Description</label>
          <textarea id="idesc" v-model="description" class="fx-input min-h-[80px]" />
        </div>
        <div>
          <label class="fx-label" for="iloc">Location</label>
          <select id="iloc" v-model="locationId" class="fx-input" required>
            <option v-for="o in locations" :key="o.ID" :value="o.ID">{{ o.Label }}</option>
          </select>
        </div>
        <div>
          <label class="fx-label" for="itpl">Template</label>
          <select id="itpl" v-model="templateType" class="fx-input" :disabled="!isNew" @change="onTplChange">
            <option v-for="t in templates" :key="t.ID" :value="t.ID">{{ t.DisplayName || t.ID }}</option>
          </select>
          <p v-if="!isNew" class="mt-1 text-xs text-zinc-500">Template type cannot be changed after creation.</p>
        </div>
        <div v-for="f in fields" :key="f.key" class="space-y-1">
          <label class="fx-label" :for="'f-' + f.key">{{ f.label }}{{ f.required ? " *" : "" }}</label>
          <select v-if="f.widget === 'select'" :id="'f-' + f.key" v-model="fieldVals[f.key]" class="fx-input" :required="f.required">
            <option value="">—</option>
            <option v-for="opt in f.options || []" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
          <input
            v-else
            :id="'f-' + f.key"
            v-model="fieldVals[f.key]"
            class="fx-input"
            :required="f.required"
            :placeholder="f.placeholder"
          />
        </div>
        <fieldset class="space-y-2 rounded-xl border border-zinc-200 p-4">
          <legend class="text-sm font-medium text-zinc-700">Additional fields</legend>
          <div v-for="(row, i) in addPairs" :key="i" class="flex gap-2">
            <input v-model="row.k" class="fx-input flex-1" placeholder="Key" />
            <input v-model="row.v" class="fx-input flex-1" placeholder="Value" />
          </div>
          <button type="button" class="text-sm text-sky-700 hover:underline" @click="addPairs.push({ k: '', v: '' })">Add pair</button>
        </fieldset>
        <fieldset class="space-y-2">
          <legend class="fx-label">Labels</legend>
          <label v-for="lb in allLabels" :key="lb.ID" class="flex items-center gap-2 text-sm">
            <input v-model="selectedLabels[lb.ID]" type="checkbox" class="rounded border-zinc-300" />
            <span>{{ lb.Name }}</span>
          </label>
        </fieldset>
        <div>
          <label class="fx-label" for="iph">Photo</label>
          <input id="iph" type="file" accept="image/*" class="block text-sm" @change="onPhoto" />
        </div>
        <button type="submit" class="fx-btn-primary">Save</button>
      </form>
    </template>
  </div>
</template>
