<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { api, csrfToken, postJson } from "../api";
import { toast } from "../composables/useToast";

const { t } = useI18n();

type ItemIdPolicy = {
  prefix?: string;
  width?: number;
  next_seq?: number;
};

const registrationMode = ref("admin_only");
const itemIdPolicy = ref<ItemIdPolicy>({ prefix: "item", width: 4 });
const itemCount = ref(0);
const err = ref("");
const importJsonRef = ref<HTMLInputElement | null>(null);
const importZipRef = ref<HTMLInputElement | null>(null);

onMounted(load);

async function load() {
  err.value = "";
  try {
    const [reg, ids] = await Promise.all([
      api<{ registration_mode: string }>("/api/admin/settings/registration"),
      api<{ policy: ItemIdPolicy; item_count: number }>("/api/admin/settings/item-ids"),
    ]);
    registrationMode.value = reg.registration_mode;
    itemIdPolicy.value = {
      prefix: ids.policy.prefix ?? "item",
      width: ids.policy.width && ids.policy.width > 0 ? ids.policy.width : 4,
      next_seq: ids.policy.next_seq,
    };
    itemCount.value = ids.item_count;
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.loadFailed");
  }
}

async function saveRegistrationMode() {
  err.value = "";
  try {
    await postJson("/api/admin/settings/registration", { mode: registrationMode.value });
    await load();
    toast.success(t("toast.registrationModeSaved"));
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.saveFailed");
    toast.error(err.value);
  }
}

async function saveItemIdPolicy() {
  err.value = "";
  try {
    await postJson("/api/admin/settings/item-ids", {
      prefix: itemIdPolicy.value.prefix ?? "",
      width: itemIdPolicy.value.width ?? 4,
    });
    await load();
    toast.success(t("toast.itemIdPolicySaved"));
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.saveFailed");
    toast.error(err.value);
  }
}

type ImportSummary = {
  templates_created: number;
  templates_updated: number;
  labels_created: number;
  labels_updated: number;
  locations_created: number;
  locations_updated: number;
  items_created: number;
  items_updated: number;
  item_labels_replaced: number;
};

function triggerJsonPick() {
  importJsonRef.value?.click();
}

function triggerZipPick() {
  importZipRef.value?.click();
}

async function downloadExport(format: "json" | "csv") {
  err.value = "";
  try {
    const res = await fetch(`/api/admin/inventory-export?format=${format}`, { credentials: "same-origin" });
    if (!res.ok) {
      let msg = res.statusText;
      const ct = res.headers.get("Content-Type") || "";
      if (ct.includes("application/json")) {
        try {
          const j = (await res.json()) as { error?: string };
          if (j.error) msg = j.error;
        } catch {
          /* ignore */
        }
      }
      throw new Error(msg);
    }
    const blob = await res.blob();
    const name = format === "json" ? "findus-inventory.json" : "findus-inventory.zip";
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = name;
    a.rel = "noopener";
    a.click();
    URL.revokeObjectURL(url);
    toast.success(format === "json" ? t("toast.jsonExportDownloaded") : t("toast.csvBundleDownloaded"));
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("toast.exportFailed");
    toast.error(err.value);
  }
}

async function onImportJson(ev: Event) {
  const input = ev.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file) return;
  err.value = "";
  try {
    const text = await file.text();
    const body = JSON.parse(text) as unknown;
    const res = await postJson<ImportSummary>("/api/admin/inventory-import", body);
    toast.success(
      t("toast.importSummary", {
        itemsCreated: res.items_created,
        itemsUpdated: res.items_updated,
        locsCreated: res.locations_created,
        locsUpdated: res.locations_updated,
      }),
    );
    await load();
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("toast.importFailed");
    toast.error(err.value);
  }
}

async function onImportZip(ev: Event) {
  const input = ev.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file) return;
  err.value = "";
  try {
    const fd = new FormData();
    fd.append("file", file);
    const headers = new Headers();
    const csrf = csrfToken();
    if (csrf) headers.set("X-CSRF-Token", csrf);
    const res = await fetch("/api/admin/inventory-import", {
      method: "POST",
      credentials: "same-origin",
      headers,
      body: fd,
    });
    if (!res.ok) {
      let msg = res.statusText;
      try {
        const j = (await res.json()) as { error?: string };
        if (j.error) msg = j.error;
      } catch {
        /* ignore */
      }
      throw new Error(msg);
    }
    const resBody = (await res.json()) as ImportSummary;
    toast.success(
      t("toast.importSummary", {
        itemsCreated: resBody.items_created,
        itemsUpdated: resBody.items_updated,
        locsCreated: resBody.locations_created,
        locsUpdated: resBody.locations_updated,
      }),
    );
    await load();
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("toast.importFailed");
    toast.error(err.value);
  }
}
</script>

<template>
  <div class="max-w-4xl space-y-10">
    <h1 class="text-2xl font-semibold text-zinc-900">{{ $t("adminSettings.title") }}</h1>
    <p v-if="err" class="text-sm text-red-700">{{ err }}</p>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">{{ $t("adminSettings.registrationModeTitle") }}</h2>
      <p class="mt-2 text-sm text-zinc-600">{{ $t("adminSettings.registrationHelp") }}</p>
      <div class="mt-4 flex flex-wrap items-end gap-3">
        <select v-model="registrationMode" class="fx-input max-w-xs">
          <option value="admin_only">{{ $t("adminSettings.regAdminOnly") }}</option>
          <option value="invite">{{ $t("adminSettings.regInvite") }}</option>
          <option value="open">{{ $t("adminSettings.regOpen") }}</option>
        </select>
        <button type="button" class="fx-btn-primary text-sm" @click="saveRegistrationMode">{{ $t("common.save") }}</button>
      </div>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">{{ $t("adminSettings.itemIdsTitle") }}</h2>
      <p class="mt-2 text-sm text-zinc-600">
        {{ $t("adminSettings.itemIdsWarning") }}
      </p>
      <p class="mt-1 text-xs text-zinc-500">{{ $t("adminSettings.itemsInDb", { n: itemCount }) }}</p>
      <div class="mt-4 flex flex-wrap items-end gap-3">
        <input v-model="itemIdPolicy.prefix" class="fx-input w-40" :placeholder="$t('adminSettings.prefixPlaceholder')" />
        <label class="flex items-center gap-2 text-sm text-zinc-600">
          {{ $t("common.width") }}
          <input v-model.number="itemIdPolicy.width" type="number" min="1" max="12" class="fx-input w-20" />
        </label>
        <button type="button" class="fx-btn-primary text-sm" @click="saveItemIdPolicy">{{ $t("common.save") }}</button>
      </div>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">{{ $t("adminSettings.importExportTitle") }}</h2>
      <p class="mt-2 text-sm text-zinc-600">
        {{ $t("adminSettings.importExportHelpBeforeLink") }}
        <a class="text-sky-700 underline hover:text-sky-800" href="/admin/backup.zip">{{ $t("adminSettings.backupZipLinkLabel") }}</a>
        {{ $t("adminSettings.importExportHelpAfterLink") }}
      </p>
      <p class="mt-2 text-xs text-zinc-500">
        {{ $t("adminSettings.importExportNote") }}
      </p>
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button type="button" class="fx-btn-primary text-sm" @click="downloadExport('json')">{{ $t("adminSettings.downloadJson") }}</button>
        <button type="button" class="fx-btn-primary text-sm" @click="downloadExport('csv')">{{ $t("adminSettings.downloadCsvZip") }}</button>
      </div>
      <div class="mt-6 flex flex-wrap items-end gap-8">
        <div>
          <p class="text-sm font-medium text-zinc-800">{{ $t("adminSettings.importJsonTitle") }}</p>
          <p class="mt-1 text-xs text-zinc-500">{{ $t("adminSettings.importJsonHint") }}</p>
          <input ref="importJsonRef" type="file" accept=".json,application/json" class="hidden" @change="onImportJson" />
          <button type="button" class="fx-btn-secondary mt-2 text-sm" @click="triggerJsonPick">{{ $t("common.chooseFile") }}</button>
        </div>
        <div>
          <p class="text-sm font-medium text-zinc-800">{{ $t("adminSettings.importCsvTitle") }}</p>
          <p class="mt-1 text-xs text-zinc-500">{{ $t("adminSettings.importCsvHint") }}</p>
          <input ref="importZipRef" type="file" accept=".zip,application/zip" class="hidden" @change="onImportZip" />
          <button type="button" class="fx-btn-secondary mt-2 text-sm" @click="triggerZipPick">{{ $t("common.chooseZip") }}</button>
        </div>
      </div>
    </section>
  </div>
</template>
