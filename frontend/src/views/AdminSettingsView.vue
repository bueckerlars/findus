<script setup lang="ts">
import { onMounted, ref } from "vue";
import { api, csrfToken, postJson } from "../api";
import { toast } from "../composables/useToast";

type ItemIdPolicy = {
  kind: string;
  prefix?: string;
  width?: number;
  next_seq?: number;
};

const registrationMode = ref("admin_only");
const itemIdPolicy = ref<ItemIdPolicy>({ kind: "sequential", prefix: "item", width: 4 });
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
      kind: ids.policy.kind || "sequential",
      prefix: ids.policy.prefix ?? "item",
      width: ids.policy.width && ids.policy.width > 0 ? ids.policy.width : 4,
      next_seq: ids.policy.next_seq,
    };
    itemCount.value = ids.item_count;
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Load failed";
  }
}

async function saveRegistrationMode() {
  err.value = "";
  try {
    await postJson("/api/admin/settings/registration", { mode: registrationMode.value });
    await load();
    toast.success("Registration mode saved.");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
    toast.error(err.value);
  }
}

async function saveItemIdPolicy() {
  err.value = "";
  try {
    await postJson("/api/admin/settings/item-ids", {
      kind: itemIdPolicy.value.kind,
      prefix: itemIdPolicy.value.prefix ?? "",
      width: itemIdPolicy.value.width ?? 4,
    });
    await load();
    toast.success("Item ID policy saved.");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
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
    toast.success(format === "json" ? "JSON export downloaded." : "CSV bundle downloaded.");
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Export failed";
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
      `Import complete: ${res.items_created} items created, ${res.items_updated} updated; ${res.locations_created} locations created, ${res.locations_updated} updated.`,
    );
    await load();
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Import failed";
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
    const t = csrfToken();
    if (t) headers.set("X-CSRF-Token", t);
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
      `Import complete: ${resBody.items_created} items created, ${resBody.items_updated} updated; ${resBody.locations_created} locations created, ${resBody.locations_updated} updated.`,
    );
    await load();
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Import failed";
    toast.error(err.value);
  }
}
</script>

<template>
  <div class="max-w-4xl space-y-10">
    <h1 class="text-2xl font-semibold text-zinc-900">Application settings</h1>
    <p v-if="err" class="text-sm text-red-700">{{ err }}</p>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">Registration mode</h2>
      <p class="mt-2 text-sm text-zinc-600">Who may create new accounts.</p>
      <div class="mt-4 flex flex-wrap items-end gap-3">
        <select v-model="registrationMode" class="fx-input max-w-xs">
          <option value="admin_only">Admin only</option>
          <option value="invite">Invite</option>
          <option value="open">Open</option>
        </select>
        <button type="button" class="fx-btn-primary text-sm" @click="saveRegistrationMode">Save</button>
      </div>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">Item IDs</h2>
      <p class="mt-2 text-sm text-zinc-600">
        Changing the scheme rewrites every item’s primary key and updates label links and image files. Bookmarks to
        <span class="font-mono">/items/…</span> will break. QR codes use a separate token and keep working.
      </p>
      <p class="mt-1 text-xs text-zinc-500">{{ itemCount }} items in the database.</p>
      <div class="mt-4 flex flex-wrap items-end gap-3">
        <select v-model="itemIdPolicy.kind" class="fx-input max-w-xs">
          <option value="sequential">Sequential (default)</option>
          <option value="ulid">ULID</option>
          <option value="uuid">UUID v4</option>
        </select>
        <template v-if="itemIdPolicy.kind === 'sequential'">
          <input v-model="itemIdPolicy.prefix" class="fx-input w-40" placeholder="Prefix (default: item)" />
          <label class="flex items-center gap-2 text-sm text-zinc-600">
            Width
            <input v-model.number="itemIdPolicy.width" type="number" min="1" max="12" class="fx-input w-20" />
          </label>
        </template>
        <button type="button" class="fx-btn-primary text-sm" @click="saveItemIdPolicy">Save</button>
      </div>
    </section>
    <section class="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-sm">
      <h2 class="text-lg font-semibold text-zinc-900">Inventory export / import</h2>
      <p class="mt-2 text-sm text-zinc-600">
        Export or merge-import locations, labels, item templates, and items (including item–label links). User accounts
        and registration settings are not included. Only image paths are in the data; use
        <a class="text-sky-700 underline hover:text-sky-800" href="/admin/backup.zip">full database backup (ZIP)</a>
        for a complete SQLite snapshot and image files.
      </p>
      <p class="mt-2 text-xs text-zinc-500">
        Import merges by primary key (updates existing rows; does not delete missing rows). Label names must remain
        unique across the database. JSON is the canonical round-trip format; CSV is delivered as a ZIP of spreadsheets.
      </p>
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button type="button" class="fx-btn-primary text-sm" @click="downloadExport('json')">Download JSON</button>
        <button type="button" class="fx-btn-primary text-sm" @click="downloadExport('csv')">Download CSV (ZIP)</button>
      </div>
      <div class="mt-6 flex flex-wrap items-end gap-8">
        <div>
          <p class="text-sm font-medium text-zinc-800">Import JSON</p>
          <p class="mt-1 text-xs text-zinc-500">Use the file produced by “Download JSON”.</p>
          <input
            ref="importJsonRef"
            type="file"
            accept=".json,application/json"
            class="hidden"
            @change="onImportJson"
          />
          <button type="button" class="fx-btn-secondary mt-2 text-sm" @click="triggerJsonPick">Choose file…</button>
        </div>
        <div>
          <p class="text-sm font-medium text-zinc-800">Import CSV bundle</p>
          <p class="mt-1 text-xs text-zinc-500">Use the ZIP from “Download CSV (ZIP)”.</p>
          <input ref="importZipRef" type="file" accept=".zip,application/zip" class="hidden" @change="onImportZip" />
          <button type="button" class="fx-btn-secondary mt-2 text-sm" @click="triggerZipPick">Choose ZIP…</button>
        </div>
      </div>
    </section>
  </div>
</template>
