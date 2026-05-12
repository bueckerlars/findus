<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { api, postJson } from "../api";
import { confirmAlert } from "../composables/useAlertDialog";
import { useCreateModals } from "../composables/useCreateModals";
import { setLocationDetailCommandHandlers } from "../composables/useLocationDetailCommandBridge";
import { useSession } from "../session";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import FxSvg from "../components/FxSvg.vue";
import ItemPhotoPlaceholder from "../components/ItemPhotoPlaceholder.vue";
import FxQrMenuButton from "../components/FxQrMenuButton.vue";

type Location = { ID: string; Name: string; Description: string; ParentID?: string | null };
type Item = { ID: string; Name: string; Description: string; PhotoPath?: string | null };
type Crumb = { ID: string; Name: string };

const route = useRoute();
const router = useRouter();
const { isAdmin } = useSession();
const { openCreateItem, openCreateLocation } = useCreateModals();

const loc = ref<Location | null>(null);
const children = ref<Location[]>([]);
const items = ref<Item[]>([]);
const breadcrumb = ref<Crumb[]>([]);
const backHref = ref("/locations");
const backLabel = ref("All locations");

const id = computed(() => route.params.id as string);
const itemsViewKey = computed(() => "location_items_" + id.value);

onMounted(load);
watch(id, () => {
  void load();
});
async function load() {
  const r = await api<{
    location: Location;
    children: Location[];
    items: Item[];
    breadcrumb: Crumb[];
    back_href: string;
    back_label: string;
  }>("/api/locations/" + id.value);
  loc.value = r.location;
  children.value = r.children;
  items.value = r.items;
  breadcrumb.value = r.breadcrumb;
  backHref.value = r.back_href;
  backLabel.value = r.back_label;
}

function qrDownloadFilename(): string {
  const raw = (loc.value?.Name || "location").replace(/[^\w\-._\s]+/g, "").trim().replace(/\s+/g, "-");
  const base = raw.length ? raw.slice(0, 80) : "location";
  return `findus-${base}-qr.png`;
}

function downloadLocationQrPng() {
  if (!loc.value) return;
  const a = document.createElement("a");
  a.href = "/locations/" + loc.value.ID + "/qr.png";
  a.download = qrDownloadFilename();
  a.rel = "noopener";
  document.body.appendChild(a);
  a.click();
  a.remove();
}

async function copyLocationPageLink() {
  try {
    await navigator.clipboard.writeText(window.location.href);
  } catch {
    /* ignore */
  }
}

watch(
  () => ({ l: loc.value, ad: isAdmin.value }),
  () => {
    if (!loc.value) {
      setLocationDetailCommandHandlers(null);
      return;
    }
    const shared = {
      downloadQrPng: downloadLocationQrPng,
      copyPageLink: copyLocationPageLink,
    };
    if (isAdmin.value) {
      setLocationDetailCommandHandlers({
        ...shared,
        deleteLocation: () => del(),
      });
    } else {
      setLocationDetailCommandHandlers(shared);
    }
  },
  { flush: "post" },
);

onUnmounted(() => {
  setLocationDetailCommandHandlers(null);
});

async function del() {
  const ok = await confirmAlert({
    title: "Delete this location?",
    message: "Only empty locations can be deleted. This cannot be undone.",
    confirmLabel: "Delete",
    variant: "danger",
  });
  if (!ok) return;
  await postJson("/api/locations/" + id.value + "/delete", {});
  await router.push("/locations");
}
</script>

<template>
  <div v-if="!loc" class="text-zinc-500">Loading…</div>
  <div v-else class="mx-auto max-w-3xl">
    <RouterLink :to="backHref" class="inline-flex items-center gap-1 text-sm font-medium text-sky-600 hover:text-sky-700">{{ backLabel }}</RouterLink>

    <nav class="mt-4 flex flex-wrap items-center gap-x-1 gap-y-1 text-sm text-zinc-500" aria-label="Breadcrumb">
      <template v-for="(c, i) in breadcrumb" :key="c.ID">
        <span v-if="i > 0" class="text-zinc-300">/</span>
        <RouterLink class="hover:text-sky-600" :to="'/locations/' + c.ID">{{ c.Name }}</RouterLink>
      </template>
    </nav>

    <div class="mt-4 fx-card p-6 sm:p-8">
      <div class="flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
        <div class="min-w-0 flex-1">
          <h1 class="text-2xl font-semibold tracking-tight text-zinc-900">{{ loc.Name }}</h1>
          <p v-if="loc.Description" class="mt-3 whitespace-pre-wrap text-zinc-600">{{ loc.Description }}</p>
          <p v-else class="mt-3 text-sm italic text-zinc-400">No description</p>
        </div>
        <div class="flex shrink-0 flex-row flex-wrap items-center gap-2 sm:justify-end">
          <FxQrMenuButton
            :png-url="'/locations/' + loc.ID + '/qr.png'"
            :download-name="loc.Name"
            hint="Scan to open this location on your phone (same account)."
          />
          <template v-if="isAdmin">
            <RouterLink :to="'/locations/' + loc.ID + '/edit'" class="fx-icon-btn" aria-label="Edit location" title="Edit">
              <FxSvg name="pencilSquare" />
            </RouterLink>
            <button type="button" class="fx-icon-btn-danger" aria-label="Delete location" title="Delete" @click="del">
              <FxSvg name="trash" />
            </button>
          </template>
        </div>
      </div>
    </div>

    <section class="mt-6 fx-card p-6">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-base font-semibold text-zinc-900">Inside this location</h2>
        <button
          v-if="isAdmin"
          type="button"
          class="text-sm font-medium text-sky-600 hover:text-sky-700"
          @click="openCreateLocation({ parentId: loc.ID })"
        >
          + Add sub-location
        </button>
      </div>
      <ul class="mt-4 space-y-2">
        <li v-for="ch in children" :key="ch.ID">
          <RouterLink :to="'/locations/' + ch.ID" class="group fx-list-row">
            <span class="font-medium">{{ ch.Name }}</span>
            <span class="text-zinc-400" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
          </RouterLink>
        </li>
        <li
          v-if="!children.length"
          class="rounded-xl border border-dashed border-zinc-200 bg-zinc-50/80 px-4 py-6 text-center text-sm text-zinc-500"
        >
          No sub-locations
        </li>
      </ul>
    </section>

    <ItemsViewToggle :storage-key="itemsViewKey" class="mt-4 fx-card p-6">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-zinc-900">Items here</h2>
          <div class="flex flex-wrap items-center gap-2">
            <ItemsViewModeToolbar />
            <button
              v-if="isAdmin"
              type="button"
              class="text-sm font-medium text-sky-600 hover:text-sky-700"
              @click="openCreateItem({ locationId: loc.ID })"
            >
              + Add item
            </button>
          </div>
        </div>
      </template>

      <div v-if="items.length" class="mt-4">
        <ul class="items-view-list-only space-y-2">
          <li v-for="it in items" :key="it.ID">
            <RouterLink :to="'/items/' + it.ID" class="group fx-item-row relative fx-list-row">
              <span class="fx-item-row-accent" aria-hidden="true"></span>
              <div class="relative z-[1] min-w-0 flex-1">
                <span class="font-medium text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{ it.Name }}</span>
              </div>
              <span class="fx-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon" /></span>
            </RouterLink>
          </li>
        </ul>
        <div class="items-view-gallery-only grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4">
          <RouterLink
            v-for="it in items"
            :key="it.ID + '-g'"
            :to="'/items/' + it.ID"
            class="group fx-item-gallery flex flex-col overflow-hidden rounded-xl border border-zinc-200/80 bg-white shadow-sm ring-1 ring-zinc-950/[0.03]"
          >
            <div class="fx-item-gallery-media aspect-square bg-gradient-to-br from-zinc-50 to-zinc-100 ring-1 ring-zinc-100/80">
              <img
                v-if="it.PhotoPath"
                :src="'/items/' + it.ID + '/photo'"
                alt=""
                class="fx-item-gallery-photo"
                @error="($event.target as HTMLImageElement).style.display = 'none'"
              />
              <div v-else class="fx-item-gallery-placeholder">
                <ItemPhotoPlaceholder :item-id="it.ID" />
              </div>
              <div class="fx-item-gallery-shade" aria-hidden="true"></div>
              <span class="fx-item-gallery-fab" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
            </div>
            <div class="relative z-[1] flex min-h-0 flex-1 flex-col gap-1.5 border-t border-zinc-100/90 p-3">
              <span class="line-clamp-2 font-medium leading-snug text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{ it.Name }}</span>
            </div>
          </RouterLink>
        </div>
      </div>
      <p v-else class="mt-4 rounded-xl border border-dashed border-zinc-200 bg-zinc-50/80 px-4 py-6 text-center text-sm text-zinc-500">
        No items in this place yet
      </p>
    </ItemsViewToggle>
  </div>
</template>
