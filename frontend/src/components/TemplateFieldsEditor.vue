<script setup lang="ts">
import { nextTick, ref, watch } from "vue";
import FxSvg from "./FxSvg.vue";
import type { TemplateField, TemplateWidget } from "../types/templateFields";
import {
  emptyTemplateField,
  parseTemplateFieldsJson,
  serializeTemplateFields,
} from "../types/templateFields";

const props = defineProps<{
  modelValue: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const rows = ref<TemplateField[]>([]);
const parseError = ref("");
const syncingFromParent = ref(false);

function loadFromProp(s: string) {
  const r = parseTemplateFieldsJson(s);
  if (!r.ok) {
    parseError.value = r.error;
    return;
  }
  parseError.value = "";
  syncingFromParent.value = true;
  rows.value = r.fields;
  nextTick(() => {
    syncingFromParent.value = false;
  });
}

watch(
  () => props.modelValue,
  (s) => {
    const current = serializeTemplateFields(rows.value);
    if (current === s) return;
    loadFromProp(s);
  },
  { immediate: true },
);

watch(
  rows,
  () => {
    if (syncingFromParent.value) return;
    parseError.value = "";
    emit("update:modelValue", serializeTemplateFields(rows.value));
  },
  { deep: true },
);

function addField() {
  rows.value = [...rows.value, emptyTemplateField()];
}

function removeField(index: number) {
  rows.value = rows.value.filter((_, i) => i !== index);
}

function moveField(index: number, delta: number) {
  const j = index + delta;
  if (j < 0 || j >= rows.value.length) return;
  const next = rows.value.slice();
  const t = next[index]!;
  next[index] = next[j]!;
  next[j] = t;
  rows.value = next;
}

function onWidgetChange(f: TemplateField, w: TemplateWidget) {
  f.widget = w;
  if (w === "text") {
    f.options = [];
  } else {
    f.pattern = "";
    f.min_int = undefined;
    f.max_int = undefined;
    f.max_len = undefined;
    if (!f.options?.length) f.options = [{ value: "", label: "" }];
  }
}

function addOption(f: TemplateField) {
  if (!f.options) f.options = [];
  f.options.push({ value: "", label: "" });
}

function removeOption(f: TemplateField, optIndex: number) {
  if (!f.options) return;
  f.options = f.options.filter((_, i) => i !== optIndex);
}

function minIntInput(f: TemplateField, raw: string) {
  const t = raw.trim();
  if (t === "") {
    f.min_int = undefined;
    return;
  }
  const n = Number(t);
  f.min_int = Number.isFinite(n) ? Math.trunc(n) : undefined;
}

function maxIntInput(f: TemplateField, raw: string) {
  const t = raw.trim();
  if (t === "") {
    f.max_int = undefined;
    return;
  }
  const n = Number(t);
  f.max_int = Number.isFinite(n) ? Math.trunc(n) : undefined;
}

function maxLenInput(f: TemplateField, raw: string) {
  const t = raw.trim();
  if (t === "") {
    f.max_len = undefined;
    return;
  }
  const n = Number(t);
  f.max_len = Number.isFinite(n) && n > 0 ? Math.trunc(n) : undefined;
}
</script>

<template>
  <div class="space-y-3">
    <p v-if="parseError" class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
      {{ parseError }}
    </p>
    <div class="space-y-3">
      <div v-if="!rows.length" class="rounded-xl border border-dashed border-zinc-200 bg-zinc-50/60 px-4 py-8 text-center text-sm text-zinc-500">
        No template fields yet. Add a field to collect structured details for items.
      </div>

      <div v-for="(f, i) in rows" :key="i" class="rounded-xl border border-zinc-200/90 bg-white p-4 shadow-sm">
        <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs font-semibold uppercase tracking-wide text-zinc-400">Field {{ i + 1 }}</span>
          <div class="flex flex-wrap items-center gap-1">
            <button
              type="button"
              class="rounded-lg border border-zinc-200 px-2 py-1 text-xs font-medium text-zinc-600 hover:bg-zinc-50 disabled:opacity-40"
              :disabled="i === 0"
              aria-label="Move field up"
              @click="moveField(i, -1)"
            >
              Up
            </button>
            <button
              type="button"
              class="rounded-lg border border-zinc-200 px-2 py-1 text-xs font-medium text-zinc-600 hover:bg-zinc-50 disabled:opacity-40"
              :disabled="i === rows.length - 1"
              aria-label="Move field down"
              @click="moveField(i, 1)"
            >
              Down
            </button>
            <button
              type="button"
              class="inline-flex items-center gap-1 rounded-lg border border-red-200 px-2 py-1 text-xs font-medium text-red-700 hover:bg-red-50"
              @click="removeField(i)"
            >
              <FxSvg name="trash" class="h-3.5 w-3.5" />
              Remove
            </button>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <div>
            <label class="fx-label" :for="`tfk-${i}`">Key</label>
            <input :id="`tfk-${i}`" v-model="f.key" class="fx-input font-mono text-sm" placeholder="e.g. serial" />
          </div>
          <div>
            <label class="fx-label" :for="`tfl-${i}`">Label</label>
            <input :id="`tfl-${i}`" v-model="f.label" class="fx-input text-sm" placeholder="Shown in forms" />
          </div>
          <div>
            <label class="fx-label" :for="`tfw-${i}`">Widget</label>
            <select
              :id="`tfw-${i}`"
              class="fx-input text-sm"
              :value="f.widget"
              @change="onWidgetChange(f, ($event.target as HTMLSelectElement).value as TemplateWidget)"
            >
              <option value="text">Text</option>
              <option value="select">Select</option>
            </select>
          </div>
          <div class="flex items-end pb-1">
            <label class="flex cursor-pointer items-center gap-2 text-sm text-zinc-700">
              <input v-model="f.required" type="checkbox" class="rounded border-zinc-300" />
              Required
            </label>
          </div>
        </div>

        <div class="mt-3">
          <label class="fx-label" :for="`tfp-${i}`">Placeholder</label>
          <input :id="`tfp-${i}`" v-model="f.placeholder" class="fx-input text-sm" placeholder="Optional" />
        </div>

        <template v-if="f.widget === 'text'">
          <div class="mt-3 grid gap-3 sm:grid-cols-2">
            <div class="sm:col-span-2">
              <label class="fx-label" :for="`tfpat-${i}`">Pattern (regex)</label>
              <input
                :id="`tfpat-${i}`"
                v-model="f.pattern"
                class="fx-input font-mono text-sm"
                placeholder="Optional validation"
                spellcheck="false"
              />
            </div>
            <div>
              <label class="fx-label" :for="`tfmin-${i}`">Min integer</label>
              <input
                :id="`tfmin-${i}`"
                class="fx-input text-sm"
                :value="f.min_int === undefined ? '' : String(f.min_int)"
                inputmode="numeric"
                placeholder="Optional"
                @input="minIntInput(f, ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div>
              <label class="fx-label" :for="`tfmax-${i}`">Max integer</label>
              <input
                :id="`tfmax-${i}`"
                class="fx-input text-sm"
                :value="f.max_int === undefined ? '' : String(f.max_int)"
                inputmode="numeric"
                placeholder="Optional"
                @input="maxIntInput(f, ($event.target as HTMLInputElement).value)"
              />
            </div>
            <div>
              <label class="fx-label" :for="`tfml-${i}`">Max length (characters)</label>
              <input
                :id="`tfml-${i}`"
                class="fx-input text-sm"
                :value="f.max_len === undefined ? '' : String(f.max_len)"
                inputmode="numeric"
                placeholder="Default 512 if empty"
                @input="maxLenInput(f, ($event.target as HTMLInputElement).value)"
              />
            </div>
          </div>
        </template>

        <div v-else class="mt-3 space-y-2">
          <div class="flex items-center justify-between gap-2">
            <span class="text-xs font-semibold uppercase tracking-wide text-zinc-400">Options</span>
            <button
              type="button"
              class="inline-flex items-center gap-1 rounded-lg border border-zinc-200 bg-white px-2 py-1 text-xs font-semibold text-zinc-700 hover:border-sky-200 hover:bg-sky-50/70"
              @click="addOption(f)"
            >
              <FxSvg name="plus" />
              Add option
            </button>
          </div>
          <div v-for="(opt, oi) in f.options || []" :key="oi" class="flex flex-col gap-2 rounded-lg border border-zinc-100 bg-zinc-50/50 p-3 sm:flex-row sm:items-center">
            <input v-model="opt.value" class="fx-input flex-1 font-mono text-sm" placeholder="Value" />
            <input v-model="opt.label" class="fx-input flex-1 text-sm" placeholder="Label" />
            <button
              type="button"
              class="shrink-0 self-end rounded-lg px-2 py-1 text-xs font-medium text-zinc-500 hover:bg-red-50 hover:text-red-700 sm:self-center"
              :disabled="(f.options?.length ?? 0) <= 1"
              @click="removeOption(f, oi)"
            >
              Remove
            </button>
          </div>
        </div>
      </div>

      <button
        type="button"
        class="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-dashed border-zinc-300 bg-zinc-50/50 px-4 py-3 text-sm font-medium text-zinc-700 hover:border-sky-300 hover:bg-sky-50/50 hover:text-sky-900"
        @click="addField"
      >
        <FxSvg name="plus" />
        Add field
      </button>
    </div>
  </div>
</template>

