<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, patchJson, postJson } from "../api";
import { useSession } from "../session";
import { PERM_ITEMS_WRITE } from "../permissions";
import FxSvg, { type FxIconName } from "../components/FxSvg.vue";
import FxQrMenuButton from "../components/FxQrMenuButton.vue";
import { confirmAlert } from "../composables/useAlertDialog";
import { toast } from "../composables/useToast";
import { setItemDetailCommandHandlers } from "../composables/useItemDetailCommandBridge";
import FxAlert from "../components/primitives/FxAlert.vue";
import FxSkeleton from "../components/primitives/FxSkeleton.vue";

type Item = {
  ID: string;
  Name: string;
  Description: string;
  LocationID: string;
  TemplateType: string;
};
type Label = { ID: string; Name: string; Color: string };
type PathEl = { ID: string; Name: string };
type AttrRow = { Key: string; Label: string; DisplayValue: string; RawValue: string };
type LocOpt = { ID: string; Label: string };
type KV = { K: string; V: string };
type TemplateField = {
  key: string;
  label: string;
  widget: string;
  required: boolean;
  placeholder?: string;
  options?: { value: string; label: string }[];
};

type ItemAttachment = {
  ID: string;
  ItemID: string;
  OriginalFilename: string;
  Title: string;
  MimeType: string;
  SizeBytes: number;
  CreatedAt: string;
  DownloadURL: string;
};

const route = useRoute();
const router = useRouter();
const { isAdmin, can } = useSession();
const canEditItems = computed(() => isAdmin.value || can(PERM_ITEMS_WRITE));
const { t } = useI18n();

const item = ref<Item | null>(null);
const labels = ref<Label[]>([]);
const photoUrl = ref("");
const locationPath = ref<PathEl[]>([]);
const attrRows = ref<AttrRow[]>([]);
const systemRows = ref<{ label: string; value: string }[]>([]);
const qrMenuBtn = ref<{ close: () => void } | null>(null);

const editMode = ref(false);
const editLoading = ref(false);
const saveBusy = ref(false);
const saveErr = ref("");
let editBaseline = "";

const draftName = ref("");
const draftDesc = ref("");
const draftLocationId = ref("");
const locationsEdit = ref<LocOpt[]>([]);
const allLabelsEdit = ref<Label[]>([]);
const selectedLabelsEdit = ref<Record<string, boolean>>({});
const fieldsEdit = ref<TemplateField[]>([]);
const fieldValsEdit = ref<Record<string, string>>({});
const addPairsEdit = ref<{ k: string; v: string }[]>([]);
const photoFile = ref<File | null>(null);
const photoPendingPreview = ref<string | null>(null);
const labelAddMenuOpen = ref(false);
const labelPickerRoot = ref<HTMLElement | null>(null);

const attachments = ref<ItemAttachment[]>([]);
const attachmentBusy = ref(false);
const attachmentDragOver = ref(false);
const attachmentInputRef = ref<HTMLInputElement | null>(null);

const id = computed(() => route.params.id as string);
const qrPngUrl = computed(() => "/items/" + id.value + "/qr.png");
const breadcrumbTitle = computed(() => (editMode.value ? draftName.value : item.value?.Name) || "");
const heroImageSrc = computed(() => photoPendingPreview.value || photoUrl.value);

watch(photoFile, (f) => {
  if (photoPendingPreview.value) {
    URL.revokeObjectURL(photoPendingPreview.value);
    photoPendingPreview.value = null;
  }
  if (f) {
    photoPendingPreview.value = URL.createObjectURL(f);
  }
});

async function load() {
  const r = await api<{
    item: Item;
    labels: Label[];
    photo_url: string;
    location_path: PathEl[];
    attr_rows: AttrRow[];
    system_rows: { label: string; value: string }[];
    attachments?: ItemAttachment[];
  }>("/api/items/" + id.value);
  item.value = r.item;
  labels.value = r.labels ?? [];
  photoUrl.value = r.photo_url;
  locationPath.value = r.location_path ?? [];
  attrRows.value = r.attr_rows ?? [];
  systemRows.value = r.system_rows ?? [];
  attachments.value = r.attachments ?? [];
}

function applyMergedRows(rows: KV[]) {
  const merged = Object.fromEntries(rows.map((r) => [r.K, r.V]));
  const tplKeys = new Set(fieldsEdit.value.map((f) => f.key));
  for (const f of fieldsEdit.value) {
    fieldValsEdit.value[f.key] = merged[f.key] != null ? String(merged[f.key]) : "";
  }
  const extras: { k: string; v: string }[] = [];
  for (const r of rows) {
    if (!tplKeys.has(r.K)) {
      extras.push({ k: r.K, v: r.V });
    }
  }
  addPairsEdit.value = extras.length ? extras : [];
}

function addPairsForSnapshot() {
  return addPairsEdit.value.filter((r) => r.k.trim() !== "" || r.v.trim() !== "");
}

function systemRowLabel(label: string): string {
  const map: Record<string, string> = {
    "Item ID": t("itemDetail.systemLabels.itemId"),
    Created: t("itemDetail.systemLabels.created"),
    "Last updated": t("itemDetail.systemLabels.lastUpdated"),
    Template: t("itemDetail.systemLabels.template"),
  };
  return map[label] ?? label;
}

async function loadEditForm() {
  if (!item.value) return;
  const r = await api<{
    item: Item;
    locations: LocOpt[];
    all_labels: Label[];
    selected_labels: Record<string, boolean>;
    additional_rows: KV[];
  }>("/api/items/" + id.value + "/edit");
  draftName.value = r.item.Name;
  draftDesc.value = r.item.Description;
  draftLocationId.value = r.item.LocationID;
  locationsEdit.value = r.locations ?? [];
  allLabelsEdit.value = r.all_labels ?? [];
  const sl: Record<string, boolean> = {};
  for (const lb of r.all_labels) {
    sl[lb.ID] = !!r.selected_labels?.[lb.ID];
  }
  selectedLabelsEdit.value = sl;
  const ft = await api<{ fields: TemplateField[] }>(
    "/api/items/new/fields?template_type=" + encodeURIComponent(r.item.TemplateType),
  );
  fieldsEdit.value = ft.fields ?? [];
  fieldValsEdit.value = {};
  for (const f of ft.fields) {
    fieldValsEdit.value[f.key] = "";
  }
  applyMergedRows(r.additional_rows || []);
  photoFile.value = null;
  labelAddMenuOpen.value = false;
}

function editSnapshot(): string {
  const slKeys = Object.keys(selectedLabelsEdit.value).sort();
  const sl: Record<string, boolean> = {};
  for (const k of slKeys) sl[k] = !!selectedLabelsEdit.value[k];
  return JSON.stringify({
    n: draftName.value,
    d: draftDesc.value,
    loc: draftLocationId.value,
    fv: fieldValsEdit.value,
    ap: addPairsForSnapshot(),
    sl,
  });
}

function isDirty() {
  return editMode.value && editSnapshot() !== editBaseline;
}

async function enterEditMode() {
  if (!canEditItems.value || !item.value) return;
  saveErr.value = "";
  editLoading.value = true;
  try {
    await loadEditForm();
    editMode.value = true;
    editBaseline = editSnapshot();
  } catch (e) {
    saveErr.value = e instanceof Error ? e.message : t("common.couldNotLoadEditor");
  } finally {
    editLoading.value = false;
  }
}

/** Deep-link / command palette: `/items/:id?edit=1` (same component instance when only query changes). */
async function applyRouteEditIntent() {
  if (route.query.edit !== "1" || !canEditItems.value || !item.value) return;
  await router.replace({ path: "/items/" + id.value, query: {} });
  if (editMode.value) return;
  await enterEditMode();
}

async function exitEditMode() {
  if (isDirty()) {
    const ok = await confirmAlert({
      title: t("itemDetail.discardTitle"),
      message: t("itemDetail.discardMsg"),
      confirmLabel: t("common.discard"),
      variant: "default",
    });
    if (!ok) return;
  }
  editMode.value = false;
  saveErr.value = "";
  await load();
}

async function saveItem() {
  if (!item.value || saveBusy.value) return;
  saveErr.value = "";
  saveBusy.value = true;
  const fd = new FormData();
  fd.append("name", draftName.value);
  fd.append("description", draftDesc.value);
  fd.append("location_id", draftLocationId.value);
  fd.append("template_type", item.value.TemplateType);
  for (const lid of Object.keys(selectedLabelsEdit.value).filter((k) => selectedLabelsEdit.value[k])) {
    fd.append("label_id", lid);
  }
  for (const row of addPairsEdit.value) {
    const k = row.k.trim();
    if (k !== "") {
      fd.append("add_k", k);
      fd.append("add_v", row.v);
    }
  }
  for (const f of fieldsEdit.value) {
    fd.append(f.key, fieldValsEdit.value[f.key] ?? "");
  }
  if (photoFile.value) {
    fd.append("photo", photoFile.value);
  }
  try {
    await api<{ next: string }>("/api/items/" + id.value, { method: "POST", body: fd });
    editMode.value = false;
    await load();
    toast.success(t("toast.itemSaved"));
  } catch (e) {
    saveErr.value = e instanceof Error ? e.message : t("common.saveFailed");
  } finally {
    saveBusy.value = false;
  }
}

onMounted(async () => {
  document.addEventListener("pointerdown", onDocPointerDown, true);
  document.addEventListener("keydown", onGlobalKeydown, true);
  await load();
  await applyRouteEditIntent();
});
onUnmounted(() => {
  document.removeEventListener("pointerdown", onDocPointerDown, true);
  document.removeEventListener("keydown", onGlobalKeydown, true);
  setItemDetailCommandHandlers(null);
  if (photoPendingPreview.value) {
    URL.revokeObjectURL(photoPendingPreview.value);
  }
});

watch(
  () => ({ it: item.value, ed: editMode.value, ad: canEditItems.value }),
  () => {
    if (!item.value) {
      setItemDetailCommandHandlers(null);
      return;
    }
    const shared = {
      downloadQrPng,
      copyPageLink: copyItemPageLink,
    };
    if (!canEditItems.value) {
      setItemDetailCommandHandlers(shared);
      return;
    }
    if (editMode.value) {
      setItemDetailCommandHandlers({
        ...shared,
        save: () => saveItem(),
        cancel: () => exitEditMode(),
      });
    } else {
      setItemDetailCommandHandlers({
        ...shared,
        deleteItem: () => del(),
      });
    }
  },
  { flush: "post" },
);

watch(id, async () => {
  editMode.value = false;
  photoFile.value = null;
  clearAttachmentPick();
  labelAddMenuOpen.value = false;
  await load();
  await applyRouteEditIntent();
});

watch(editMode, (ed) => {
  if (!ed) {
    clearAttachmentPick();
  }
});

watch(
  () => route.query.edit,
  async (edit) => {
    if (edit !== "1") return;
    await applyRouteEditIntent();
  },
);

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

function qrDownloadFilename(): string {
  const raw = (item.value?.Name || "item").replace(/[^\w\-._\s]+/g, "").trim().replace(/\s+/g, "-");
  const base = raw.length ? raw.slice(0, 80) : "item";
  return `findus-${base}-qr.png`;
}

function downloadQrPng() {
  const a = document.createElement("a");
  a.href = qrPngUrl.value;
  a.download = qrDownloadFilename();
  a.rel = "noopener";
  document.body.appendChild(a);
  a.click();
  a.remove();
}

async function copyItemPageLink() {
  try {
    await navigator.clipboard.writeText(window.location.href);
  } catch {
    /* ignore */
  }
}

async function del() {
  const ok = await confirmAlert({
    title: t("itemDetail.deleteTitle"),
    message: t("itemDetail.deleteMsg"),
    confirmLabel: t("common.delete"),
    variant: "danger",
  });
  if (!ok) return;
  try {
    await postJson("/api/items/" + id.value + "/delete", {});
    toast.success(t("toast.itemDeleted"));
    await router.push("/items");
  } catch (e) {
    toast.error(e instanceof Error ? e.message : t("common.deleteFailed"));
  }
}

function onPhotoEdit(e: Event) {
  const t = (e.target as HTMLInputElement).files?.[0];
  photoFile.value = t || null;
}

function clearPhotoEdit() {
  photoFile.value = null;
  const inputs = document.querySelectorAll<HTMLInputElement>(".fx-item-photo-file-input");
  inputs.forEach((el) => {
    el.value = "";
  });
}

const selectedLabelsOrdered = computed(() => allLabelsEdit.value.filter((lb) => selectedLabelsEdit.value[lb.ID]));

const labelsAvailableToAdd = computed(() => allLabelsEdit.value.filter((lb) => !selectedLabelsEdit.value[lb.ID]));

function toggleLabelAddMenu() {
  qrMenuBtn.value?.close();
  labelAddMenuOpen.value = !labelAddMenuOpen.value;
}

function pickLabel(id: string) {
  selectedLabelsEdit.value = { ...selectedLabelsEdit.value, [id]: true };
  labelAddMenuOpen.value = false;
}

function removeLabel(id: string) {
  const next = { ...selectedLabelsEdit.value, [id]: false };
  selectedLabelsEdit.value = next;
}

function addCustomAttributeRow() {
  addPairsEdit.value = [...addPairsEdit.value, { k: "", v: "" }];
}

function removeCustomAttributeRow(i: number) {
  addPairsEdit.value = addPairsEdit.value.filter((_, j) => j !== i);
}

function attachmentMimeIcon(mime: string): FxIconName {
  if (mime.startsWith("image/")) return "photo";
  if (mime === "application/pdf") return "documentText";
  return "cube";
}

function formatAttachmentBytes(n: number): string {
  if (!Number.isFinite(n) || n < 0) return t("common.emDash");
  if (n < 1024) return `${Math.round(n)} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(n < 10 * 1024 ? 1 : 0)} KB`;
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}

function attachmentDisplayName(a: ItemAttachment): string {
  const tit = a.Title?.trim();
  return tit || a.OriginalFilename || t("itemDetail.documentFallback");
}

/** Default display title from file name (basename, extension stripped for known types). */
function defaultTitleFromFilename(name: string): string {
  const base = name.trim().replace(/^.*[/\\]/, "");
  const noExt = base.replace(/\.(pdf|png|jpe?g|gif|webp)$/i, "").trim();
  return noExt || base || t("itemDetail.documentFallback");
}

function attachmentRowTitle(a: ItemAttachment): string {
  return a.Title?.trim() ? a.Title.trim() : defaultTitleFromFilename(a.OriginalFilename);
}

function onAttachmentFilePick(e: Event) {
  const input = e.target as HTMLInputElement;
  const f = input.files?.[0];
  input.value = "";
  if (f) void uploadAttachment(f);
}

function onAttachmentDrop(e: DragEvent) {
  attachmentDragOver.value = false;
  const f = e.dataTransfer?.files?.[0];
  if (!f) return;
  void uploadAttachment(f);
}

function onDropZoneDragLeave(e: DragEvent) {
  const rel = e.relatedTarget as Node | null;
  const box = e.currentTarget as HTMLElement;
  if (rel && box.contains(rel)) return;
  attachmentDragOver.value = false;
}

function openAttachmentPicker() {
  attachmentInputRef.value?.click();
}

function clearAttachmentPick() {
  attachmentDragOver.value = false;
  if (attachmentInputRef.value) attachmentInputRef.value.value = "";
}

async function uploadAttachment(file: File) {
  if (!item.value || attachmentBusy.value) return;
  attachmentBusy.value = true;
  try {
    const fd = new FormData();
    fd.append("file", file);
    const title = defaultTitleFromFilename(file.name).trim();
    fd.append("title", title);
    await api("/api/items/" + id.value + "/attachments", { method: "POST", body: fd });
    clearAttachmentPick();
    await load();
    toast.success(t("toast.documentUploaded"));
  } catch (e) {
    toast.error(e instanceof Error ? e.message : t("toast.documentUploadFailed"));
  } finally {
    attachmentBusy.value = false;
  }
}

async function onAttachmentTitleBlur(a: ItemAttachment, ev: Event) {
  if (!item.value || !canEditItems.value || !editMode.value || attachmentBusy.value) return;
  const el = ev.target as HTMLInputElement;
  const next = el.value.trim();
  const prev = attachmentRowTitle(a);
  if (next === prev) return;
  attachmentBusy.value = true;
  try {
    await patchJson("/api/items/" + id.value + "/attachments/" + a.ID, { title: next });
    await load();
    toast.success(t("toast.documentTitleSaved"));
  } catch (e) {
    el.value = prev;
    toast.error(e instanceof Error ? e.message : t("toast.documentSaveFailed"));
  } finally {
    attachmentBusy.value = false;
  }
}

async function deleteAttachment(a: ItemAttachment) {
  if (!item.value || attachmentBusy.value) return;
  const ok = await confirmAlert({
    title: t("itemDetail.documentDeleteTitle"),
    message: t("itemDetail.documentDeleteMsg", { name: attachmentDisplayName(a) }),
    confirmLabel: t("itemDetail.documentRemove"),
    cancelLabel: t("common.cancel"),
    variant: "danger",
  });
  if (!ok) return;
  attachmentBusy.value = true;
  try {
    await api("/api/items/" + id.value + "/attachments/" + a.ID, { method: "DELETE" });
    await load();
    toast.success(t("toast.documentRemoved"));
  } catch (e) {
    toast.error(e instanceof Error ? e.message : t("toast.documentRemoveFailed"));
  } finally {
    attachmentBusy.value = false;
  }
}
</script>

<template>
  <div v-if="!item" class="mx-auto max-w-4xl space-y-5" :aria-label="$t('common.loadingAria')" aria-busy="true">
    <FxSkeleton width="6rem" height="0.75rem" />
    <div class="fx-card p-5">
      <div class="grid gap-4 sm:grid-cols-[minmax(0,11.25rem)_1fr]">
        <FxSkeleton width="100%" height="11.25rem" />
        <div class="space-y-2">
          <FxSkeleton width="16rem" height="1.5rem" />
          <FxSkeleton width="100%" height="0.75rem" />
          <FxSkeleton width="80%" height="0.75rem" />
          <FxSkeleton width="60%" height="0.75rem" />
        </div>
      </div>
    </div>
  </div>
  <div v-else class="mx-auto max-w-4xl space-y-6">
    <RouterLink to="/items" class="inline-flex items-center gap-1 text-sm font-medium text-sky-600 hover:text-sky-700">{{ $t("itemDetail.backItems") }}</RouterLink>

    <nav class="flex flex-wrap items-center gap-x-1 gap-y-1 text-sm text-zinc-500" :aria-label="$t('common.breadcrumb')">
      <RouterLink to="/items" class="hover:text-sky-600">{{ $t("itemDetail.breadcrumbItems") }}</RouterLink>
      <span class="text-zinc-300">/</span>
      <span class="truncate font-medium text-zinc-800">{{ breadcrumbTitle }}</span>
    </nav>

    <FxAlert v-if="saveErr">{{ saveErr }}</FxAlert>

    <div class="fx-card overflow-visible">
      <div class="grid gap-6 p-5 sm:grid-cols-[minmax(0,11.25rem)_1fr] sm:items-start sm:gap-8 sm:p-7 lg:p-8">
        <div class="mx-auto w-full max-w-[11.25rem] shrink-0 justify-self-center sm:mx-0 sm:justify-self-start">
          <template v-if="heroImageSrc">
            <div
              class="overflow-hidden rounded-2xl border border-zinc-200/90 bg-zinc-50 shadow-[inset_0_1px_0_rgba(255,255,255,0.6)] ring-1 ring-zinc-950/[0.04]"
            >
              <img :src="heroImageSrc" alt="" class="aspect-square w-full object-cover" />
            </div>
            <div v-if="editMode" class="mt-3 space-y-2">
              <label
                class="flex w-full cursor-pointer items-center justify-center gap-2 rounded-xl border border-zinc-200/90 bg-white px-3 py-2.5 text-xs font-semibold text-zinc-800 shadow-sm transition hover:border-sky-200 hover:bg-sky-50/70 hover:text-sky-900"
              >
                <FxSvg name="photo" class="h-4 w-4 shrink-0 text-zinc-500" />
                {{ photoUrl ? $t("itemDetail.replacePhoto") : $t("itemDetail.setPhoto") }}
                <input type="file" accept="image/*" class="fx-item-photo-file-input sr-only" @change="onPhotoEdit" />
              </label>
              <p v-if="photoFile" class="truncate text-center text-[11px] leading-tight text-zinc-500" :title="photoFile.name">{{ photoFile.name }}</p>
              <button
                v-if="photoFile"
                type="button"
                class="w-full text-center text-[11px] font-medium text-zinc-500 underline-offset-2 hover:text-red-600 hover:underline"
                @click="clearPhotoEdit"
              >
                {{ $t("itemDetail.discardNewImage") }}
              </button>
            </div>
          </template>
          <template v-else>
            <div
              class="fx-item-detail-photo-slot flex aspect-square items-center justify-center rounded-2xl bg-gradient-to-br from-zinc-50 to-zinc-100 text-zinc-400 ring-1 ring-zinc-950/[0.05]"
            >
              <FxSvg name="cube" class="h-14 w-14 opacity-40 sm:h-16 sm:w-16" />
            </div>
            <div v-if="editMode" class="mt-3 space-y-2">
              <label
                class="flex w-full cursor-pointer items-center justify-center gap-2 rounded-xl border border-zinc-200/90 bg-white px-3 py-2.5 text-xs font-semibold text-zinc-800 shadow-sm transition hover:border-sky-200 hover:bg-sky-50/70 hover:text-sky-900"
              >
                <FxSvg name="photo" class="h-4 w-4 shrink-0 text-zinc-500" />
                {{ $t("itemDetail.addPhoto") }}
                <input type="file" accept="image/*" class="fx-item-photo-file-input sr-only" @change="onPhotoEdit" />
              </label>
              <p v-if="photoFile" class="truncate text-center text-[11px] leading-tight text-zinc-500" :title="photoFile.name">{{ photoFile.name }}</p>
              <button
                v-if="photoFile"
                type="button"
                class="w-full text-center text-[11px] font-medium text-zinc-500 underline-offset-2 hover:text-red-600 hover:underline"
                @click="clearPhotoEdit"
              >
                {{ $t("itemDetail.discardNewImage") }}
              </button>
            </div>
          </template>
        </div>

        <div class="min-w-0">
          <div class="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between lg:gap-8">
            <div class="min-w-0 flex-1 space-y-4">
              <div class="flex flex-wrap items-end gap-2.5 gap-y-2">
                <template v-if="!editMode">
                  <h1 class="min-w-0 max-w-full text-2xl font-semibold tracking-tight text-zinc-900 sm:text-3xl">{{ item.Name }}</h1>
                </template>
                <template v-else>
                  <label class="sr-only" for="item-inline-name">{{ $t("itemDetail.name") }}</label>
                  <input
                    id="item-inline-name"
                    v-model="draftName"
                    type="text"
                    required
                    :placeholder="$t('itemDetail.itemNamePh')"
                    class="min-w-0 flex-1 basis-[min(100%,16rem)] rounded-xl border border-zinc-200 bg-white px-3 py-2 text-2xl font-semibold tracking-tight text-zinc-900 shadow-inner shadow-zinc-950/5 outline-none transition placeholder:text-zinc-400 focus:border-sky-400 focus:ring-2 focus:ring-sky-100 sm:text-3xl"
                  />
                </template>
              </div>
              <template v-if="!editMode">
                <p v-if="item.Description" class="whitespace-pre-wrap text-[15px] leading-relaxed text-zinc-600">
                  {{ item.Description }}
                </p>
                <p v-else class="text-sm italic text-zinc-400">{{ $t("common.noDescription") }}</p>
                <div v-if="locationPath.length" class="flex flex-wrap items-center gap-x-1 gap-y-1 text-sm text-zinc-500">
                  <FxSvg name="mapPin" class="mr-0.5 h-4 w-4 shrink-0 text-zinc-400" />
                  <span class="font-medium text-zinc-400">{{ $t("itemDetail.storedIn") }}</span>
                  <template v-for="(p, i) in locationPath" :key="p.ID">
                    <span v-if="i > 0" class="text-zinc-300">/</span>
                    <RouterLink :to="'/locations/' + p.ID" class="font-medium text-sky-700 hover:text-sky-800 hover:underline">{{ p.Name }}</RouterLink>
                  </template>
                </div>
              </template>
              <template v-else>
                <div>
                  <label class="fx-label" for="item-inline-desc">{{ $t("itemDetail.description") }}</label>
                  <textarea
                    id="item-inline-desc"
                    v-model="draftDesc"
                    rows="4"
                    class="fx-input mt-1.5 min-h-[6.5rem] resize-y text-sm leading-relaxed text-zinc-800"
                    :placeholder="$t('itemDetail.descPlaceholder')"
                  />
                </div>
                <div>
                  <label class="fx-label" for="item-inline-loc">{{ $t("itemDetail.storedIn") }}</label>
                  <div class="mt-1.5 flex items-center gap-2">
                    <FxSvg name="mapPin" class="h-4 w-4 shrink-0 text-zinc-400" aria-hidden="true" />
                    <select id="item-inline-loc" v-model="draftLocationId" class="fx-input mt-0 min-h-[2.5rem] flex-1 py-2 text-sm" required>
                      <option v-for="o in locationsEdit" :key="o.ID" :value="o.ID">{{ o.Label }}</option>
                    </select>
                  </div>
                </div>
              </template>
            </div>

            <div class="flex shrink-0 flex-row items-center gap-2 self-start lg:pt-0.5">
              <FxQrMenuButton
                ref="qrMenuBtn"
                :png-url="qrPngUrl"
                :download-name="item?.Name"
                :hint="$t('itemDetail.qrHint')"
                @open="labelAddMenuOpen = false"
              />
              <template v-if="canEditItems">
                <template v-if="!editMode">
                  <button
                    type="button"
                    class="fx-icon-btn"
                    :aria-label="$t('itemDetail.editItem')"
                    :title="$t('common.edit')"
                    :disabled="editLoading"
                    @click="enterEditMode"
                  >
                    <FxSvg name="pencilSquare" />
                  </button>
                  <button
                    type="button"
                    class="fx-icon-btn-danger"
                    :aria-label="$t('itemDetail.deleteItem')"
                    :title="$t('common.delete')"
                    @click="del"
                  >
                    <FxSvg name="trash" />
                  </button>
                </template>
                <template v-else>
                  <button type="button" class="fx-icon-btn" :aria-label="$t('itemDetail.viewMode')" :title="$t('common.view')" @click="exitEditMode">
                    <FxSvg name="eye" />
                  </button>
                  <button
                    type="button"
                    class="fx-icon-btn border-emerald-200/90 bg-emerald-50/80 text-emerald-800 hover:border-emerald-300 hover:bg-emerald-100"
                    :aria-label="$t('itemDetail.saveChanges')"
                    :title="$t('itemDetail.save')"
                    :disabled="saveBusy"
                    @click="saveItem"
                  >
                    <FxSvg name="check" />
                  </button>
                </template>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="labels.length && !editMode" class="fx-card px-5 py-4 sm:px-6">
      <h2 class="text-xs font-semibold uppercase tracking-wide text-zinc-400">{{ $t("itemDetail.labels") }}</h2>
      <div class="mt-3 flex flex-wrap gap-2">
        <span
          v-for="lb in labels"
          :key="lb.ID"
          class="inline-flex items-center gap-1.5 rounded-full border border-zinc-200/90 bg-zinc-100/80 px-2.5 py-1 text-xs font-medium text-zinc-800"
        >
          <span class="h-2 w-2 shrink-0 rounded-full ring-1 ring-zinc-300/80" :style="{ backgroundColor: lb.Color }"></span>
          {{ lb.Name }}
        </span>
      </div>
    </div>

    <div v-if="editMode" class="fx-card px-5 py-4 sm:px-6">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <h2 class="text-xs font-semibold uppercase tracking-wide text-zinc-400">{{ $t("itemDetail.labels") }}</h2>
        <div ref="labelPickerRoot" class="relative shrink-0">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-200/90 bg-white px-2.5 py-1.5 text-xs font-semibold text-zinc-700 shadow-sm transition hover:border-sky-200 hover:bg-sky-50/70 hover:text-sky-900 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="!labelsAvailableToAdd.length"
            aria-haspopup="listbox"
            :aria-expanded="labelAddMenuOpen ? 'true' : 'false'"
            aria-controls="item-label-add-dropdown"
            @click="toggleLabelAddMenu"
          >
            <FxSvg name="plus" />
            {{ $t("itemDetail.addLabel") }}
          </button>
          <div
            v-show="labelAddMenuOpen"
            id="item-label-add-dropdown"
            class="absolute right-0 top-full z-50 mt-2 max-h-56 w-[min(16rem,calc(100vw-2rem))] overflow-y-auto rounded-xl border border-zinc-200/90 bg-white py-1 shadow-lg shadow-zinc-900/10 ring-1 ring-zinc-950/[0.04]"
            role="listbox"
            :aria-label="$t('itemDetail.labelsToAdd')"
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
      <p v-if="!allLabelsEdit.length" class="mt-2 text-sm text-zinc-500">{{ $t("itemDetail.createLabelsFirst") }}</p>
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
              :aria-label="$t('itemDetail.removeLabel', { name: lb.Name })"
              @click="removeLabel(lb.ID)"
            >
              <span class="text-base leading-none" aria-hidden="true">×</span>
            </button>
          </span>
        </template>
        <p v-else class="text-sm text-zinc-400">{{ $t("itemDetail.noLabels") }}</p>
      </div>
    </div>

    <section class="fx-card overflow-hidden p-0">
      <div class="border-b border-zinc-100 bg-zinc-50/80 px-5 py-3.5">
        <h2 class="text-sm font-semibold text-zinc-800">{{ $t("itemDetail.documentsTitle") }}</h2>
        <p class="mt-1 text-xs text-zinc-500">{{ $t("itemDetail.documentsHint") }}</p>
      </div>
      <div
        v-if="canEditItems && editMode"
        class="border-b border-zinc-100 px-5 py-4 sm:px-6"
        @dragover.prevent
        @drop.prevent="onAttachmentDrop"
      >
        <div
          class="rounded-xl border-2 border-dashed px-4 py-8 text-center transition-colors"
          :class="
            attachmentDragOver
              ? 'border-sky-400 bg-sky-50/90 ring-2 ring-sky-200/60'
              : 'border-zinc-200 bg-zinc-50/50 hover:border-zinc-300'
          "
          :aria-busy="attachmentBusy ? 'true' : 'false'"
          @dragenter.prevent="attachmentDragOver = true"
          @dragleave="onDropZoneDragLeave"
          @click.self="!attachmentBusy && openAttachmentPicker()"
        >
          <input
            ref="attachmentInputRef"
            type="file"
            class="sr-only"
            accept="application/pdf,image/*"
            :disabled="attachmentBusy"
            @change="onAttachmentFilePick"
          />
          <FxSvg name="plus" class="mx-auto h-8 w-8 text-zinc-400" :class="attachmentBusy && 'animate-pulse text-sky-500'" aria-hidden="true" />
          <p class="mt-2 text-sm text-zinc-600">{{ $t("itemDetail.documentDropHint") }}</p>
          <p v-if="attachmentBusy" class="mt-1 text-xs font-medium text-sky-700">{{ $t("common.loading") }}</p>
          <button
            type="button"
            class="mt-3 inline-flex items-center justify-center rounded-lg border border-zinc-200/90 bg-white px-3 py-2 text-xs font-semibold text-zinc-800 shadow-sm transition hover:border-sky-200 hover:bg-sky-50/70 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="attachmentBusy"
            @click.stop="openAttachmentPicker"
          >
            {{ $t("itemDetail.documentPickFile") }}
          </button>
        </div>
      </div>
      <div v-if="attachments.length" class="divide-y divide-zinc-100 border-t border-zinc-100">
        <div
          v-for="a in attachments"
          :key="a.ID"
          class="flex min-w-0 items-center gap-2 px-4 py-3 sm:gap-3 sm:px-5 sm:py-3"
        >
          <span class="shrink-0 text-zinc-500" :title="a.MimeType">
            <FxSvg :name="attachmentMimeIcon(a.MimeType)" class="h-5 w-5" aria-hidden="true" />
          </span>
          <div class="min-w-0 flex-1">
            <template v-if="canEditItems && editMode">
              <label class="sr-only" :for="'att-title-' + a.ID">{{ $t("itemDetail.documentTitlePh") }}</label>
              <input
                :id="'att-title-' + a.ID"
                :key="a.ID + '-' + (a.Title || '') + '-' + a.OriginalFilename"
                type="text"
                maxlength="120"
                class="fx-input mt-0 w-full min-w-0 text-sm font-medium leading-snug"
                :defaultValue="attachmentRowTitle(a)"
                :aria-label="$t('itemDetail.documentTitlePh')"
                @blur="onAttachmentTitleBlur(a, $event)"
              />
            </template>
            <template v-else>
              <div class="break-words text-sm font-medium leading-snug text-zinc-800">{{ attachmentDisplayName(a) }}</div>
            </template>
          </div>
          <span
            class="shrink-0 whitespace-nowrap tabular-nums text-xs text-zinc-500 sm:text-sm"
            :title="formatAttachmentBytes(a.SizeBytes)"
          >{{ formatAttachmentBytes(a.SizeBytes) }}</span>
          <div class="flex shrink-0 items-center gap-0.5 sm:gap-1">
            <a
              :href="a.DownloadURL"
              class="fx-icon-btn"
              download
              target="_blank"
              rel="noopener"
              :title="$t('itemDetail.documentDownload')"
              :aria-label="$t('itemDetail.documentDownload')"
            >
              <FxSvg name="arrowDownTray" />
            </a>
            <button
              v-if="canEditItems && editMode"
              type="button"
              class="fx-icon-btn-danger"
              :disabled="attachmentBusy"
              :title="$t('itemDetail.documentRemove')"
              :aria-label="$t('itemDetail.documentRemove')"
              @click="deleteAttachment(a)"
            >
              <FxSvg name="trash" />
            </button>
          </div>
        </div>
      </div>
      <p v-else class="px-5 py-6 text-sm text-zinc-500 sm:px-6">{{ $t("itemDetail.documentsEmpty") }}</p>
    </section>

    <section v-if="attrRows.length && !editMode" class="fx-card overflow-hidden p-0">
      <h2 class="border-b border-zinc-100 bg-zinc-50/80 px-5 py-3.5 text-sm font-semibold text-zinc-800">{{ $t("itemDetail.details") }}</h2>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[20rem] text-left text-sm">
          <thead>
            <tr class="border-b border-zinc-200 bg-zinc-50/90">
              <th scope="col" class="px-5 py-3 font-medium text-zinc-500">{{ $t("common.field") }}</th>
              <th scope="col" class="px-5 py-3 font-medium text-zinc-500">{{ $t("common.value") }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-100">
            <tr v-for="(row, i) in attrRows" :key="i" class="align-top">
              <td class="whitespace-nowrap px-5 py-3.5 font-medium text-zinc-800">{{ row.Label }}</td>
              <td class="px-5 py-3.5 break-words text-zinc-700">{{ row.DisplayValue }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section v-if="editMode" class="fx-card overflow-hidden p-0">
      <h2 class="border-b border-zinc-100 bg-zinc-50/80 px-5 py-3.5 text-sm font-semibold text-zinc-800">{{ $t("itemDetail.details") }}</h2>
      <div v-if="fieldsEdit.length" class="overflow-x-auto">
        <table class="w-full min-w-[20rem] text-left text-sm">
          <thead>
            <tr class="border-b border-zinc-200 bg-zinc-50/90">
              <th scope="col" class="px-5 py-3 font-medium text-zinc-500">{{ $t("common.field") }}</th>
              <th scope="col" class="px-5 py-3 font-medium text-zinc-500">{{ $t("common.value") }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-100">
            <tr v-for="f in fieldsEdit" :key="f.key" class="align-top">
              <td class="whitespace-nowrap px-5 py-3.5 font-medium text-zinc-800">
                {{ f.label }}<span v-if="f.required" class="text-red-500"> *</span>
              </td>
              <td class="px-5 py-3.5">
                <select
                  v-if="f.widget === 'select'"
                  v-model="fieldValsEdit[f.key]"
                  class="fx-input mt-0 py-2 text-sm"
                  :required="f.required"
                >
                  <option value="">{{ $t("common.optionEmpty") }}</option>
                  <option v-for="opt in f.options || []" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
                </select>
                <input
                  v-else
                  v-model="fieldValsEdit[f.key]"
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
      <div class="px-5 py-4 sm:px-6" :class="fieldsEdit.length ? 'border-t border-zinc-100' : ''">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <h2 class="text-xs font-semibold uppercase tracking-wide text-zinc-400">{{ $t("itemDetail.customAttrs") }}</h2>
          <div class="relative shrink-0">
            <button
              type="button"
              class="inline-flex items-center gap-1.5 rounded-lg border border-zinc-200/90 bg-white px-2.5 py-1.5 text-xs font-semibold text-zinc-700 shadow-sm transition hover:border-sky-200 hover:bg-sky-50/70 hover:text-sky-900 disabled:cursor-not-allowed disabled:opacity-40"
              @click="addCustomAttributeRow"
            >
              <FxSvg name="plus" />
              {{ $t("itemDetail.addField") }}
            </button>
          </div>
        </div>
        <p class="mt-2 text-xs text-zinc-500">{{ $t("itemDetail.customAttrsHelp") }}</p>
        <div v-if="addPairsEdit.length" class="mt-3 space-y-2">
          <div
            v-for="(row, i) in addPairsEdit"
            :key="i"
            class="flex flex-col gap-2 rounded-xl border border-zinc-100 bg-zinc-50/40 p-3 sm:flex-row sm:items-center sm:gap-3 sm:p-2 sm:pr-2"
          >
            <input v-model="row.k" class="fx-input mt-0 flex-1 text-sm" :placeholder="$t('itemDetail.keyPh')" />
            <input v-model="row.v" class="fx-input mt-0 flex-1 text-sm" :placeholder="$t('itemDetail.valuePh')" />
            <button
              type="button"
              class="shrink-0 self-end rounded-lg px-2 py-1.5 text-xs font-semibold text-zinc-500 hover:bg-red-50 hover:text-red-700 sm:self-center"
              @click="removeCustomAttributeRow(i)"
            >
              {{ $t("common.remove") }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <section v-if="systemRows.length" class="fx-card overflow-hidden p-0">
      <h2 class="border-b border-zinc-100 bg-zinc-50/80 px-5 py-3.5 text-sm font-semibold text-zinc-800">{{ $t("itemDetail.system") }}</h2>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[20rem] text-left text-sm">
          <thead>
            <tr class="border-b border-zinc-200 bg-zinc-50/90">
              <th scope="col" class="px-5 py-3 font-medium text-zinc-500">{{ $t("common.field") }}</th>
              <th scope="col" class="px-5 py-3 font-medium text-zinc-500">{{ $t("common.value") }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-100">
            <tr v-for="(row, i) in systemRows" :key="i" class="align-top bg-zinc-50/50">
              <td class="whitespace-nowrap px-5 py-3.5 font-medium text-zinc-600">{{ systemRowLabel(row.label) }}</td>
              <td
                class="px-5 py-3.5 break-words text-zinc-700"
                :class="row.label === 'Item ID' ? 'font-mono text-xs break-all' : 'text-sm'"
              >
                {{ row.value }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
