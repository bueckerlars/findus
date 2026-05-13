<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from "vue";
import { csrfToken } from "../api";
import { useI18n } from "vue-i18n";
import { toast } from "../composables/useToast";

const { t } = useI18n();

const from = ref(1);
const to = ref(24);
const loading = ref(false);
const err = ref("");
const pdfUrl = ref("");
const filename = ref("labels.pdf");

const maxBatch = 400;
const canGenerate = computed(() => Number.isInteger(from.value) && Number.isInteger(to.value) && from.value > 0 && to.value > 0 && from.value <= to.value);

onBeforeUnmount(() => {
  if (pdfUrl.value) URL.revokeObjectURL(pdfUrl.value);
});

function validateRange(): string {
  if (!Number.isInteger(from.value) || !Number.isInteger(to.value)) return t("adminLabelGenerator.rangeInteger");
  if (from.value < 1 || to.value < 1) return t("adminLabelGenerator.rangePositive");
  if (from.value > to.value) return t("adminLabelGenerator.rangeOrder");
  if (to.value - from.value + 1 > maxBatch) return t("adminLabelGenerator.rangeTooLarge", { n: maxBatch });
  return "";
}

async function generate() {
  err.value = "";
  const msg = validateRange();
  if (msg) {
    err.value = msg;
    toast.error(msg);
    return;
  }
  loading.value = true;
  try {
    const headers = new Headers({ "Content-Type": "application/json" });
    const csrf = csrfToken();
    if (csrf) headers.set("X-CSRF-Token", csrf);
    const res = await fetch("/api/admin/labels/generate", {
      method: "POST",
      credentials: "same-origin",
      headers,
      body: JSON.stringify({ from: from.value, to: to.value }),
    });
    if (!res.ok) {
      let text = res.statusText;
      try {
        const j = (await res.json()) as { error?: string };
        if (j.error) text = j.error;
      } catch {
        /* ignore */
      }
      throw new Error(text);
    }
    const blob = await res.blob();
    const nextName = res.headers.get("X-Filename") || `labels-${from.value}-${to.value}.pdf`;
    if (pdfUrl.value) URL.revokeObjectURL(pdfUrl.value);
    pdfUrl.value = URL.createObjectURL(blob);
    filename.value = nextName;
    toast.success(t("adminLabelGenerator.generated"));
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("adminLabelGenerator.generateFailed");
    toast.error(err.value);
  } finally {
    loading.value = false;
  }
}

function download() {
  if (!pdfUrl.value) return;
  const a = document.createElement("a");
  a.href = pdfUrl.value;
  a.download = filename.value;
  a.rel = "noopener";
  a.click();
}
</script>

<template>
  <div class="max-w-5xl space-y-6">
    <h1 class="text-2xl font-semibold text-zinc-900">{{ $t("adminLabelGenerator.title") }}</h1>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <p class="text-sm text-zinc-600">{{ $t("adminLabelGenerator.help") }}</p>
      <p class="mt-1 text-xs text-zinc-500">{{ $t("adminLabelGenerator.constraints", { n: maxBatch }) }}</p>
      <p v-if="err" class="mt-2 text-sm text-red-700">{{ err }}</p>
      <div class="mt-4 flex flex-wrap items-end gap-3">
        <label class="text-sm text-zinc-700">
          {{ $t("adminLabelGenerator.from") }}
          <input v-model.number="from" type="number" min="1" class="fx-input mt-1 w-32" />
        </label>
        <label class="text-sm text-zinc-700">
          {{ $t("adminLabelGenerator.to") }}
          <input v-model.number="to" type="number" min="1" class="fx-input mt-1 w-32" />
        </label>
        <button type="button" class="fx-btn-primary text-sm" :disabled="loading || !canGenerate" @click="generate">
          {{ loading ? $t("adminLabelGenerator.generating") : $t("adminLabelGenerator.generate") }}
        </button>
        <button type="button" class="fx-btn-secondary text-sm" :disabled="!pdfUrl" @click="download">
          {{ $t("adminLabelGenerator.download") }}
        </button>
      </div>
    </section>
    <section v-if="pdfUrl" class="rounded-2xl border border-zinc-200/80 bg-white p-4 shadow-sm">
      <p class="mb-2 text-sm text-zinc-600">{{ filename }}</p>
      <iframe :src="pdfUrl" class="h-[72vh] w-full rounded-lg border border-zinc-200" />
    </section>
  </div>
</template>
