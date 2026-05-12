<script setup lang="ts">
import { onMounted, ref } from "vue";
import { api, postJson } from "../api";

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
  await postJson("/api/admin/settings/registration", { mode: registrationMode.value });
  await load();
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
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
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
  </div>
</template>
