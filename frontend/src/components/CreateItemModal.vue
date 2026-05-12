<script setup lang="ts">
import { ref, watch, computed, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { api } from "../api";
import { toast } from "../composables/useToast";
import FxModal from "./FxModal.vue";
import FxSvg from "./FxSvg.vue";

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

const props = defineProps<{
  modelValue: boolean;
  initialLocationId?: string;
}>();

const emit = defineEmits<{ "update:modelValue": [boolean] }>();

const router = useRouter();

const name = ref("");
const description = ref("");
const locationId = ref("");
const templateType = ref("");
const locations = ref<LocOpt[]>([]);
const templates = ref<ItemTemplate[]>([]);
const allLabels = ref<Label[]>([]);
const selectedLabels = ref<Record<string, boolean>>({});
const labelAddMenuOpen = ref(false);
const labelPickerRoot = ref<HTMLElement | null>(null);
const fields = ref<TemplateField[]>([]);
const fieldVals = ref<Record<string, string>>({});
const addPairs = ref<{ k: string; v: string }[]>([]);
const photo = ref<File | null>(null);
const photoPreviewUrl = ref<string | null>(null);
const photoInputRef = ref<HTMLInputElement | null>(null);
const photoDragging = ref(false);
const err = ref("");
const noLocations = ref(false);
const busy = ref(false);

async function onTplChange() {
  fieldVals.value = {};
  await loadFields();
  addPairs.value = [];
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

async function load() {
  err.value = "";
  if (!props.modelValue) return;
  const locQ = props.initialLocationId?.trim() || "";
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
  labelAddMenuOpen.value = false;
  await loadFields();
  addPairs.value = [];
  name.value = "";
  description.value = "";
  revokePhotoPreview();
  photo.value = null;
  photoDragging.value = false;
}

const labelsAvailableToAdd = computed(() => allLabels.value.filter((lb) => !selectedLabels.value[lb.ID]));

const selectedLabelsOrdered = computed(() => allLabels.value.filter((lb) => selectedLabels.value[lb.ID]));

function revokePhotoPreview() {
  if (photoPreviewUrl.value) {
    URL.revokeObjectURL(photoPreviewUrl.value);
    photoPreviewUrl.value = null;
  }
}

function onDocPointerDown(e: PointerEvent) {
  const t = e.target as Node | null;
  if (labelAddMenuOpen.value) {
    const lp = labelPickerRoot.value;
    if (lp && t && !lp.contains(t)) labelAddMenuOpen.value = false;
  }
}

function onGlobalKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    labelAddMenuOpen.value = false;
  }
}

onMounted(() => {
  document.addEventListener("pointerdown", onDocPointerDown, true);
  document.addEventListener("keydown", onGlobalKeydown, true);
});

onUnmounted(() => {
  document.removeEventListener("pointerdown", onDocPointerDown, true);
  document.removeEventListener("keydown", onGlobalKeydown, true);
  revokePhotoPreview();
});

watch(
  () => props.modelValue,
  (open) => {
    if (!open) labelAddMenuOpen.value = false;
    if (open) void load();
  },
);

watch(
  () => props.initialLocationId,
  () => {
    if (props.modelValue) void load();
  },
);

function setPhotoFromFile(file: File | undefined | null) {
  if (!file || !file.type.startsWith("image/")) {
    return;
  }
  revokePhotoPreview();
  photo.value = file;
  photoPreviewUrl.value = URL.createObjectURL(file);
}

function openPhotoPicker() {
  photoInputRef.value?.click();
}

function onPhotoDragEnter(e: DragEvent) {
  e.preventDefault();
  if (e.dataTransfer?.types && Array.from(e.dataTransfer.types).includes("Files")) {
    photoDragging.value = true;
  }
}

function onPhotoDragLeave(e: DragEvent) {
  e.preventDefault();
  const el = e.currentTarget as HTMLElement;
  const rel = e.relatedTarget as Node | null;
  if (rel && el.contains(rel)) return;
  photoDragging.value = false;
}

function onPhotoDragOver(e: DragEvent) {
  e.preventDefault();
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = "copy";
  }
}

function onPhotoDrop(e: DragEvent) {
  e.preventDefault();
  photoDragging.value = false;
  const f = e.dataTransfer?.files?.[0];
  setPhotoFromFile(f);
}

function onPhoto(e: Event) {
  const input = e.target as HTMLInputElement;
  const t = input.files?.[0];
  setPhotoFromFile(t);
  input.value = "";
}

function clearPhoto() {
  revokePhotoPreview();
  photo.value = null;
  photoDragging.value = false;
  if (photoInputRef.value) photoInputRef.value.value = "";
}

function toggleLabelAddMenu() {
  labelAddMenuOpen.value = !labelAddMenuOpen.value;
}

function pickLabel(id: string) {
  selectedLabels.value = { ...selectedLabels.value, [id]: true };
  labelAddMenuOpen.value = false;
}

function removeLabel(id: string) {
  selectedLabels.value = { ...selectedLabels.value, [id]: false };
}

function addCustomAttributeRow() {
  addPairs.value = [...addPairs.value, { k: "", v: "" }];
}

function removeCustomAttributeRow(i: number) {
  addPairs.value = addPairs.value.filter((_, j) => j !== i);
}

function close() {
  emit("update:modelValue", false);
}

async function submit() {
  err.value = "";
  busy.value = true;
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
    const r = await api<{ next: string }>("/api/items", {
      method: "POST",
      body: fd,
    });
    toast.success("Item created.");
    close();
    await router.push(r.next);
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <FxModal
    :model-value="modelValue"
    title="New item"
    max-width-class="max-w-2xl"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <p v-if="noLocations" class="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
      Create a location before adding items.
    </p>
    <template v-else>
      <p v-if="err" class="mb-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
      <form id="fx-create-item-form" class="space-y-4" @submit.prevent="submit">
        <div>
          <label class="fx-label" for="fx-create-item-name">Name</label>
          <input id="fx-create-item-name" v-model="name" class="fx-input" required />
        </div>
        <div>
          <label class="fx-label" for="fx-create-item-desc">Description</label>
          <textarea id="fx-create-item-desc" v-model="description" class="fx-input min-h-[80px]" />
        </div>
        <div>
          <label class="fx-label" for="fx-create-item-loc">Location</label>
          <select id="fx-create-item-loc" v-model="locationId" class="fx-input" required>
            <option v-for="o in locations" :key="o.ID" :value="o.ID">{{ o.Label }}</option>
          </select>
        </div>
        <div>
          <label class="fx-label" for="fx-create-item-tpl">Template</label>
          <select id="fx-create-item-tpl" v-model="templateType" class="fx-input" @change="onTplChange">
            <option v-for="t in templates" :key="t.ID" :value="t.ID">{{ t.DisplayName || t.ID }}</option>
          </select>
        </div>

        <section class="fx-card overflow-hidden p-0">
          <h2 class="border-b border-zinc-100 bg-zinc-50/80 px-5 py-3.5 text-sm font-semibold text-zinc-800">Details</h2>
          <div v-if="fields.length" class="overflow-x-auto">
            <table class="w-full min-w-[20rem] text-left text-sm">
              <thead>
                <tr class="border-b border-zinc-200 bg-zinc-50/90">
                  <th scope="col" class="px-5 py-3 font-medium text-zinc-500">Field</th>
                  <th scope="col" class="px-5 py-3 font-medium text-zinc-500">Value</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-zinc-100">
                <tr v-for="f in fields" :key="f.key" class="align-top">
                  <td class="whitespace-nowrap px-5 py-3.5 font-medium text-zinc-800">
                    {{ f.label }}<span v-if="f.required" class="text-red-500"> *</span>
                  </td>
                  <td class="px-5 py-3.5">
                    <select
                      v-if="f.widget === 'select'"
                      v-model="fieldVals[f.key]"
                      class="fx-input mt-0 py-2 text-sm"
                      :required="f.required"
                    >
                      <option value="">—</option>
                      <option v-for="opt in f.options || []" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                    </select>
                    <input
                      v-else
                      v-model="fieldVals[f.key]"
                      type="text"
                      class="fx-input mt-0 py-2 text-sm"
                      :required="f.required"
                      :placeholder="f.placeholder"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="px-5 py-4 sm:px-6" :class="fields.length ? 'border-t border-zinc-100' : ''">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <h2 class="text-xs font-semibold uppercase tracking-wide text-zinc-400">Custom attributes</h2>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-200/90 bg-white px-2.5 py-1.5 text-xs font-semibold text-zinc-700 shadow-sm transition hover:border-sky-200 hover:bg-sky-50/70 hover:text-sky-900"
                @click="addCustomAttributeRow"
              >
                <FxSvg name="plus" />
                Add Field
              </button>
            </div>
            <p class="mt-2 text-xs text-zinc-500">Extra keys not covered by the preset fields above.</p>
            <div v-if="addPairs.length" class="mt-3 space-y-2">
              <div
                v-for="(row, i) in addPairs"
                :key="i"
                class="flex flex-col gap-2 rounded-xl border border-zinc-100 bg-zinc-50/40 p-3 sm:flex-row sm:items-center sm:gap-3 sm:p-2 sm:pr-2"
              >
                <input v-model="row.k" class="fx-input mt-0 flex-1 text-sm" placeholder="Key" />
                <input v-model="row.v" class="fx-input mt-0 flex-1 text-sm" placeholder="Value" />
                <button
                  type="button"
                  class="shrink-0 self-end rounded-lg px-2 py-1.5 text-xs font-semibold text-zinc-500 hover:bg-red-50 hover:text-red-700 sm:self-center"
                  @click="removeCustomAttributeRow(i)"
                >
                  Remove
                </button>
              </div>
            </div>
          </div>
        </section>
        <div class="fx-card px-5 py-4 sm:px-6">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <h2 class="text-xs font-semibold uppercase tracking-wide text-zinc-400">Labels</h2>
            <div ref="labelPickerRoot" class="relative shrink-0">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-200/90 bg-white px-2.5 py-1.5 text-xs font-semibold text-zinc-700 shadow-sm transition hover:border-sky-200 hover:bg-sky-50/70 hover:text-sky-900 disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="!labelsAvailableToAdd.length"
                aria-haspopup="listbox"
                :aria-expanded="labelAddMenuOpen ? 'true' : 'false'"
                aria-controls="fx-create-item-label-add-dropdown"
                @click="toggleLabelAddMenu"
              >
                <FxSvg name="plus" />
                Add label
              </button>
              <div
                v-show="labelAddMenuOpen"
                id="fx-create-item-label-add-dropdown"
                class="absolute right-0 top-full z-[110] mt-2 max-h-56 w-[min(16rem,calc(100vw-2rem))] overflow-y-auto rounded-xl border border-zinc-200/90 bg-white py-1 shadow-lg shadow-zinc-900/10 ring-1 ring-zinc-950/[0.04]"
                role="listbox"
                aria-label="Labels to add"
                @click.stop
              >
                <button
                  v-for="lb in labelsAvailableToAdd"
                  :key="lb.ID"
                  type="button"
                  role="option"
                  class="flex w-full items-center gap-2.5 px-3 py-2 text-left text-sm text-zinc-800 hover:bg-zinc-50"
                  @click="pickLabel(lb.ID)"
                >
                  <span class="h-2.5 w-2.5 shrink-0 rounded-full ring-1 ring-zinc-200/80" :style="{ backgroundColor: lb.Color }" />
                  {{ lb.Name }}
                </button>
              </div>
            </div>
          </div>
          <p v-if="!allLabels.length" class="mt-2 text-sm text-zinc-500">Create labels first under Labels in the sidebar.</p>
          <div v-else class="mt-3 flex min-h-[2.25rem] flex-wrap items-center gap-2">
            <template v-if="selectedLabelsOrdered.length">
              <span
                v-for="lb in selectedLabelsOrdered"
                :key="lb.ID"
                class="inline-flex max-w-full items-center gap-1 rounded-full border border-zinc-200/90 bg-zinc-100/80 py-1 pl-2.5 pr-0.5 text-xs font-medium text-zinc-800"
              >
                <span class="h-2 w-2 shrink-0 rounded-full ring-1 ring-zinc-300/80" :style="{ backgroundColor: lb.Color }" />
                <span class="min-w-0 truncate">{{ lb.Name }}</span>
                <button
                  type="button"
                  class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-zinc-500 transition hover:bg-zinc-200/90 hover:text-zinc-800"
                  :aria-label="'Remove label ' + lb.Name"
                  @click="removeLabel(lb.ID)"
                >
                  <span class="text-base leading-none" aria-hidden="true">×</span>
                </button>
              </span>
            </template>
            <p v-else class="text-sm text-zinc-400">No labels</p>
          </div>
        </div>

        <div class="fx-card px-5 py-4 sm:px-6">
          <h2 class="text-xs font-semibold uppercase tracking-wide text-zinc-400">Photo</h2>
          <p class="mt-1 text-xs leading-relaxed text-zinc-500">
            Optional. Drag an image into the area below or click to browse. Common formats such as PNG, JPG, or WebP are
            supported.
          </p>
          <div
            class="group relative mt-4 flex min-h-[10.5rem] cursor-pointer flex-col overflow-hidden rounded-2xl transition-all duration-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sky-500 sm:min-h-[12rem]"
            :class="
              photoDragging
                ? 'border-2 border-sky-400 bg-sky-50/80 ring-2 ring-sky-300/45'
                : photoPreviewUrl
                  ? 'border-2 border-solid border-zinc-200/90 bg-zinc-50/50 shadow-[inset_0_1px_0_rgba(255,255,255,0.5)]'
                  : 'border-2 border-dashed border-zinc-200/90 bg-gradient-to-b from-zinc-50/90 to-zinc-100/35 hover:border-zinc-300 hover:bg-zinc-50/90'
            "
            role="button"
            tabindex="0"
            :aria-label="photo ? 'Replace image — drop a file or click to browse' : 'Add image — drop a file or click to browse'"
            @click="openPhotoPicker"
            @keydown.enter.prevent="openPhotoPicker"
            @keydown.space.prevent="openPhotoPicker"
            @dragenter.prevent="onPhotoDragEnter"
            @dragleave.prevent="onPhotoDragLeave"
            @dragover.prevent="onPhotoDragOver"
            @drop.prevent="onPhotoDrop"
          >
            <input
              id="fx-create-item-photo-input"
              ref="photoInputRef"
              type="file"
              accept="image/*"
              class="sr-only"
              tabindex="-1"
              @change="onPhoto"
            />
            <template v-if="photoPreviewUrl">
              <div class="relative flex min-h-[10rem] w-full items-center justify-center bg-zinc-100/40">
                <img
                  :src="photoPreviewUrl"
                  alt=""
                  class="max-h-[min(16rem,42vh)] w-full object-contain"
                />
                <div
                  class="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-zinc-950/60 via-zinc-950/25 to-transparent px-3 pb-2.5 pt-10 text-center"
                >
                  <p class="text-[11px] font-medium tracking-wide text-white/95 drop-shadow-sm">Drop or click to replace</p>
                </div>
              </div>
            </template>
            <div v-else class="flex flex-1 flex-col items-center justify-center gap-2 px-5 py-9 text-center">
              <span
                class="flex h-12 w-12 items-center justify-center rounded-2xl bg-white text-sky-600 shadow-sm ring-1 ring-zinc-200/80 transition group-hover:ring-sky-200/70"
                aria-hidden="true"
              >
                <FxSvg name="photo" class="h-6 w-6" />
              </span>
              <p class="text-sm font-semibold text-zinc-800">Drop image here</p>
              <p class="max-w-sm text-xs leading-relaxed text-zinc-500">Or click anywhere in this area to choose a file</p>
            </div>
          </div>
          <div v-if="photo" class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-zinc-100/90 pt-3">
            <span class="min-w-0 flex-1 truncate text-xs font-medium text-zinc-600" :title="photo.name">{{ photo.name }}</span>
            <button
              type="button"
              class="shrink-0 rounded-lg px-2.5 py-1.5 text-xs font-semibold text-zinc-500 transition hover:bg-red-50 hover:text-red-700"
              @click.stop="clearPhoto"
            >
              Remove
            </button>
          </div>
        </div>
      </form>
    </template>

    <template #footer>
      <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <button type="button" class="fx-btn-secondary w-full sm:w-auto" :disabled="busy" @click="close">Cancel</button>
        <button
          v-if="!noLocations"
          type="submit"
          form="fx-create-item-form"
          class="fx-btn-primary w-full sm:w-auto"
          :disabled="busy"
        >
          Save
        </button>
      </div>
    </template>
  </FxModal>
</template>
