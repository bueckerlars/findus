<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { api, csrfToken } from "../api";
import { useI18n } from "vue-i18n";
import { toast } from "../composables/useToast";
import FxAlert from "../components/primitives/FxAlert.vue";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";
import FxToggle from "../components/primitives/FxToggle.vue";
import LocationSelectorRow from "../components/LocationSelectorRow.vue";

const { t } = useI18n();

// --- Shared layout config ---
const cols = ref(2);
const rows = ref(0); // 0 = auto
const fullPage = ref(false);

// --- Item label generator ---
const from = ref(1);
const to = ref(24);
const itemLoading = ref(false);
const itemErr = ref("");
const itemPdfUrl = ref("");
const itemFilename = ref("labels.pdf");

const maxBatch = 400;
const canGenerateItems = computed(
  () =>
    Number.isInteger(from.value) &&
    Number.isInteger(to.value) &&
    from.value > 0 &&
    to.value > 0 &&
    from.value <= to.value,
);

onBeforeUnmount(() => {
  if (itemPdfUrl.value) URL.revokeObjectURL(itemPdfUrl.value);
  if (locPdfUrl.value) URL.revokeObjectURL(locPdfUrl.value);
});

function validateRange(): string {
  if (!Number.isInteger(from.value) || !Number.isInteger(to.value)) return t("adminLabelGenerator.rangeInteger");
  if (from.value < 1 || to.value < 1) return t("adminLabelGenerator.rangePositive");
  if (from.value > to.value) return t("adminLabelGenerator.rangeOrder");
  if (to.value - from.value + 1 > maxBatch) return t("adminLabelGenerator.rangeTooLarge", { n: maxBatch });
  return "";
}

async function generateItems() {
  itemErr.value = "";
  const msg = validateRange();
  if (msg) {
    itemErr.value = msg;
    toast.error(msg);
    return;
  }
  itemLoading.value = true;
  try {
    const headers = new Headers({ "Content-Type": "application/json" });
    const csrf = csrfToken();
    if (csrf) headers.set("X-CSRF-Token", csrf);
    const res = await fetch("/api/admin/labels/generate", {
      method: "POST",
      credentials: "same-origin",
      headers,
      body: JSON.stringify({
        from: from.value,
        to: to.value,
        cols: cols.value,
        rows: rows.value,
        full_page: fullPage.value,
      }),
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
    if (itemPdfUrl.value) URL.revokeObjectURL(itemPdfUrl.value);
    itemPdfUrl.value = URL.createObjectURL(blob);
    itemFilename.value = nextName;
    toast.success(t("adminLabelGenerator.generated"));
  } catch (e) {
    itemErr.value = e instanceof Error ? e.message : t("adminLabelGenerator.generateFailed");
    toast.error(itemErr.value);
  } finally {
    itemLoading.value = false;
  }
}

function downloadItems() {
  if (!itemPdfUrl.value) return;
  triggerDownload(itemPdfUrl.value, itemFilename.value);
}

// --- Location label generator ---
type LocationTreeNode = { ID: string; Name: string; children: LocationTreeNode[] };

const locTree = ref<LocationTreeNode[]>([]);
const locLoading = ref(false);
const locExpanded = ref(new Set<string>());
const locSelected = ref(new Set<string>());
const locGenLoading = ref(false);
const locErr = ref("");
const locPdfUrl = ref("");
const locFilename = ref("location-labels.pdf");

const canGenerateLocs = computed(() => locSelected.value.size > 0);
const allLocIDs = computed(() => collectIDs(locTree.value));

onMounted(async () => {
  locLoading.value = true;
  try {
    const data = await api<{ tree: LocationTreeNode[] }>("/api/locations");
    locTree.value = data.tree;
  } catch {
    /* ignore – tree stays empty */
  } finally {
    locLoading.value = false;
  }
});

function toggleLocExpanded(id: string) {
  const next = new Set(locExpanded.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  locExpanded.value = next;
}

function toggleLocSelected(id: string) {
  const next = new Set(locSelected.value);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  locSelected.value = next;
}

function collectIDs(nodes: LocationTreeNode[]): string[] {
  const ids: string[] = [];
  for (const n of nodes) {
    ids.push(n.ID);
    ids.push(...collectIDs(n.children));
  }
  return ids;
}

function toggleSelectAll() {
  const all = collectIDs(locTree.value);
  if (locSelected.value.size === all.length) {
    locSelected.value = new Set();
  } else {
    locSelected.value = new Set(all);
  }
}

async function generateLocations() {
  locErr.value = "";
  if (!canGenerateLocs.value) return;
  locGenLoading.value = true;
  try {
    const headers = new Headers({ "Content-Type": "application/json" });
    const csrf = csrfToken();
    if (csrf) headers.set("X-CSRF-Token", csrf);
    const res = await fetch("/api/admin/labels/generate-locations", {
      method: "POST",
      credentials: "same-origin",
      headers,
      body: JSON.stringify({
        location_ids: [...locSelected.value],
        cols: cols.value,
        rows: rows.value,
        full_page: fullPage.value,
      }),
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
    const nextName = res.headers.get("X-Filename") || "location-labels.pdf";
    if (locPdfUrl.value) URL.revokeObjectURL(locPdfUrl.value);
    locPdfUrl.value = URL.createObjectURL(blob);
    locFilename.value = nextName;
    toast.success(t("adminLocationLabelGenerator.generated"));
  } catch (e) {
    locErr.value = e instanceof Error ? e.message : t("adminLocationLabelGenerator.generateFailed");
    toast.error(locErr.value);
  } finally {
    locGenLoading.value = false;
  }
}

function downloadLocations() {
  if (!locPdfUrl.value) return;
  triggerDownload(locPdfUrl.value, locFilename.value);
}

function triggerDownload(url: string, name: string) {
  const a = document.createElement("a");
  a.href = url;
  a.download = name;
  a.rel = "noopener";
  a.click();
}
</script>

<template>
  <div class="w-full space-y-5">
    <FxPageHeader :title="$t('adminLabelGenerator.title')" />

    <!-- Shared layout config -->
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="mb-4 text-center text-sm font-semibold text-zinc-700">{{ $t("adminLabelGenerator.layoutTitle") }}</h2>
      <div class="mx-auto grid max-w-xl grid-cols-1 items-end justify-items-center gap-6 sm:grid-cols-3 sm:gap-x-10">
        <label class="flex w-full max-w-[7rem] flex-col items-center gap-1.5 text-sm text-zinc-700">
          <span class="font-medium">{{ $t("adminLabelGenerator.cols") }}</span>
          <input v-model.number="cols" type="number" min="1" max="8" class="fx-input w-full text-center" />
        </label>
        <label class="flex w-full max-w-[7rem] flex-col items-center gap-1.5 text-sm text-zinc-700">
          <span class="font-medium">{{ $t("adminLabelGenerator.rows") }}</span>
          <input v-model.number="rows" type="number" min="0" max="20" class="fx-input w-full text-center" />
        </label>
        <div class="flex justify-center pb-1">
          <FxToggle
            v-model="fullPage"
            size="md"
            :label="$t('adminLabelGenerator.fullPage')"
            :aria-label="$t('adminLabelGenerator.fullPage')"
          />
        </div>
      </div>
      <p class="mt-2 text-center text-xs text-zinc-400">{{ $t("adminLabelGenerator.rowsHint") }}</p>
      <p v-if="fullPage" class="mx-auto mt-1 max-w-md text-center text-xs text-zinc-500">
        {{ $t("adminLabelGenerator.fullPageHint") }}
      </p>
    </section>

    <!-- Item label generator -->
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="mb-1 text-sm font-semibold text-zinc-700">{{ $t("adminLabelGenerator.itemTitle") }}</h2>
      <p class="text-sm text-zinc-600">{{ $t("adminLabelGenerator.help") }}</p>
      <p class="mt-1 text-xs text-zinc-500">{{ $t("adminLabelGenerator.constraints", { n: maxBatch }) }}</p>
      <FxAlert v-if="itemErr" class="mt-3">{{ itemErr }}</FxAlert>
      <div class="mt-4 flex flex-wrap items-end gap-3">
        <label class="text-sm text-zinc-700">
          {{ $t("adminLabelGenerator.from") }}
          <input v-model.number="from" type="number" min="1" class="fx-input mt-1 w-32" />
        </label>
        <label class="text-sm text-zinc-700">
          {{ $t("adminLabelGenerator.to") }}
          <input v-model.number="to" type="number" min="1" class="fx-input mt-1 w-32" />
        </label>
        <button type="button" class="fx-btn-primary text-sm" :disabled="itemLoading || !canGenerateItems" @click="generateItems">
          {{ itemLoading ? $t("adminLabelGenerator.generating") : $t("adminLabelGenerator.generate") }}
        </button>
        <button type="button" class="fx-btn-secondary text-sm" :disabled="!itemPdfUrl" @click="downloadItems">
          {{ $t("adminLabelGenerator.download") }}
        </button>
      </div>
    </section>
    <section v-if="itemPdfUrl" class="rounded-2xl border border-zinc-200/80 bg-white p-4 shadow-sm">
      <p class="mb-2 text-sm text-zinc-600">{{ itemFilename }}</p>
      <iframe :src="itemPdfUrl" class="h-[72vh] w-full rounded-lg border border-zinc-200" />
    </section>

    <!-- Location label generator -->
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="mb-1 text-sm font-semibold text-zinc-700">{{ $t("adminLocationLabelGenerator.title") }}</h2>
      <p class="text-sm text-zinc-600">{{ $t("adminLocationLabelGenerator.help") }}</p>
      <FxAlert v-if="locErr" class="mt-3">{{ locErr }}</FxAlert>

      <div class="mt-4">
        <div class="mb-2 flex items-center justify-between">
          <span class="text-xs text-zinc-500">
            {{ $t("adminLocationLabelGenerator.selected", { n: locSelected.size }) }}
          </span>
          <button type="button" class="text-xs text-sky-600 hover:underline" @click="toggleSelectAll">
            {{ allLocIDs.length > 0 && locSelected.size === allLocIDs.length ? $t("adminLocationLabelGenerator.deselectAll") : $t("adminLocationLabelGenerator.selectAll") }}
          </button>
        </div>

        <div v-if="locLoading" class="fx-card p-4 text-center text-sm text-zinc-400">{{ $t("adminLocationLabelGenerator.loading") }}</div>
        <div v-else-if="locTree.length === 0" class="fx-card p-4 text-center text-sm text-zinc-400">{{ $t("adminLocationLabelGenerator.empty") }}</div>
        <ul v-else class="fx-card max-h-96 divide-y divide-zinc-100 overflow-y-auto overflow-visible p-0" role="list">
          <LocationSelectorRow
            v-for="node in locTree"
            :key="node.ID"
            :node="node"
            :selected="locSelected"
            :expanded="locExpanded"
            @toggle-select="toggleLocSelected"
            @toggle-expand="toggleLocExpanded"
          />
        </ul>
      </div>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button type="button" class="fx-btn-primary text-sm" :disabled="locGenLoading || !canGenerateLocs" @click="generateLocations">
          {{ locGenLoading ? $t("adminLabelGenerator.generating") : $t("adminLocationLabelGenerator.generate") }}
        </button>
        <button type="button" class="fx-btn-secondary text-sm" :disabled="!locPdfUrl" @click="downloadLocations">
          {{ $t("adminLabelGenerator.download") }}
        </button>
      </div>
    </section>
    <section v-if="locPdfUrl" class="rounded-2xl border border-zinc-200/80 bg-white p-4 shadow-sm">
      <p class="mb-2 text-sm text-zinc-600">{{ locFilename }}</p>
      <iframe :src="locPdfUrl" class="h-[72vh] w-full rounded-lg border border-zinc-200" />
    </section>
  </div>
</template>
