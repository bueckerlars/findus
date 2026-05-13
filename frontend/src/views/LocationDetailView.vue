<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, postJson } from "../api";
import { confirmAlert } from "../composables/useAlertDialog";
import { toast } from "../composables/useToast";
import { useCreateModals } from "../composables/useCreateModals";
import { setLocationDetailCommandHandlers } from "../composables/useLocationDetailCommandBridge";
import { useSession } from "../session";
import { PERM_ITEMS_WRITE, PERM_LOCATIONS_WRITE } from "../permissions";
import ItemsViewToggle from "../components/ItemsViewToggle.vue";
import ItemsViewModeToolbar from "../components/ItemsViewModeToolbar.vue";
import FxSvg from "../components/FxSvg.vue";
import ItemPhotoPlaceholder from "../components/ItemPhotoPlaceholder.vue";
import FxQrMenuButton from "../components/FxQrMenuButton.vue";
import FxSkeleton from "../components/primitives/FxSkeleton.vue";
import FxSkeletonList from "../components/primitives/FxSkeletonList.vue";
import FxEmptyState from "../components/primitives/FxEmptyState.vue";
import FxButton from "../components/primitives/FxButton.vue";
import ItemListRow from "../components/ItemListRow.vue";
import ItemGalleryCard from "../components/ItemGalleryCard.vue";

type Location = { ID: string; Name: string; Description: string; ParentID?: string | null };
type Item = { ID: string; Name: string; Description: string; PhotoPath?: string | null };
type Crumb = { ID: string; Name: string };

const route = useRoute();
const router = useRouter();
const { isAdmin, can } = useSession();
const canManageLocations = computed(() => isAdmin.value || can(PERM_LOCATIONS_WRITE));
const canCreateItemsHere = computed(() => isAdmin.value || can(PERM_ITEMS_WRITE));
const { openCreateItem, openCreateLocation } = useCreateModals();
const { t, locale } = useI18n();

const loc = ref<Location | null>(null);
const children = ref<Location[]>([]);
const items = ref<Item[]>([]);
const breadcrumb = ref<Crumb[]>([]);
const backHref = ref("/locations");
const backLabel = ref("All locations");

const displayBackLabel = computed(() => {
  void locale.value;
  return backLabel.value === "All locations" ? t("home.allLocations") : backLabel.value;
});

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
  // Backend JSON encodes empty Go slices as null; normalize for template .length / v-for.
  children.value = r.children ?? [];
  items.value = r.items ?? [];
  breadcrumb.value = r.breadcrumb ?? [];
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
  () => ({ l: loc.value, ad: canManageLocations.value }),
  () => {
    if (!loc.value) {
      setLocationDetailCommandHandlers(null);
      return;
    }
    const shared = {
      downloadQrPng: downloadLocationQrPng,
      copyPageLink: copyLocationPageLink,
    };
    if (canManageLocations.value) {
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
    title: t("locationDetail.deleteTitle"),
    message: t("locationDetail.deleteMsg"),
    confirmLabel: t("common.delete"),
    variant: "danger",
  });
  if (!ok) return;
  try {
    await postJson("/api/locations/" + id.value + "/delete", {});
    toast.success(t("toast.locationDeleted"));
    await router.push("/locations");
  } catch (e) {
    toast.error(e instanceof Error ? e.message : t("common.deleteFailed"));
  }
}
</script>

<template>
  <div v-if="!loc" class="mx-auto max-w-3xl space-y-5" :aria-label="$t('common.loadingAria')" aria-busy="true">
    <FxSkeleton width="6rem" height="0.75rem" />
    <div class="fx-card p-5 space-y-3">
      <FxSkeleton width="14rem" height="1.5rem" />
      <FxSkeleton width="80%" height="0.75rem" />
      <FxSkeleton width="60%" height="0.75rem" />
    </div>
    <div class="fx-card overflow-hidden p-0">
      <FxSkeletonList :rows="4" />
    </div>
  </div>
  <div v-else class="mx-auto max-w-3xl">
    <RouterLink :to="backHref" class="inline-flex items-center gap-1 text-sm font-medium text-sky-600 hover:text-sky-700">{{ displayBackLabel }}</RouterLink>

    <nav class="mt-4 flex flex-wrap items-center gap-x-1 gap-y-1 text-sm text-zinc-500" :aria-label="$t('common.breadcrumb')">
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
          <p v-else class="mt-3 text-sm italic text-zinc-400">{{ $t("common.noDescription") }}</p>
        </div>
        <div class="flex shrink-0 flex-row flex-wrap items-center gap-2 sm:justify-end">
          <FxQrMenuButton
            :png-url="'/locations/' + loc.ID + '/qr.png'"
            :download-name="loc.Name"
            :hint="$t('locationDetail.qrHint')"
          />
          <template v-if="canManageLocations">
            <RouterLink :to="'/locations/' + loc.ID + '/edit'" class="fx-icon-btn" :aria-label="$t('locationDetail.editLocation')" :title="$t('locationDetail.edit')">
              <FxSvg name="pencilSquare" />
            </RouterLink>
            <button type="button" class="fx-icon-btn-danger" :aria-label="$t('locationDetail.deleteLocation')" :title="$t('locationDetail.delete')" @click="del">
              <FxSvg name="trash" />
            </button>
          </template>
        </div>
      </div>
    </div>

    <section class="mt-6 fx-card p-6">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-base font-semibold text-zinc-900">{{ $t("locationDetail.inside") }}</h2>
        <button
          v-if="canManageLocations"
        >
          {{ $t("locationDetail.addSub") }}
        </button>
      </div>
      <ul class="mt-4 space-y-2">
        <li v-for="ch in children" :key="ch.ID">
          <RouterLink :to="'/locations/' + ch.ID" class="group fx-list-row">
            <span class="font-medium">{{ ch.Name }}</span>
            <span class="text-zinc-400" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
          </RouterLink>
        </li>
        <li v-if="!children.length">
          <FxEmptyState icon="mapPin" :title="$t('locationDetail.noSub')" />
        </li>
      </ul>
    </section>

    <ItemsViewToggle :storage-key="itemsViewKey" class="mt-4 fx-card p-6">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-base font-semibold text-zinc-900">{{ $t("locationDetail.itemsHere") }}</h2>
          <div class="flex flex-wrap items-center gap-2">
            <ItemsViewModeToolbar />
            <button
              v-if="canCreateItemsHere"
            >
              {{ $t("locationDetail.addItem") }}
            </button>
          </div>
        </div>
      </template>

      <div v-if="items.length" class="mt-4">
        <ul class="items-view-list-only divide-y divide-zinc-100 overflow-hidden rounded-xl border border-zinc-200/80 bg-white shadow-sm ring-1 ring-zinc-950/[0.03]">
          <li v-for="it in items" :key="it.ID">
            <ItemListRow :id="it.ID" :name="it.Name" :photo-path="it.PhotoPath" />
          </li>
        </ul>
        <div class="items-view-gallery-only grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4">
          <ItemGalleryCard
            v-for="it in items"
            :key="it.ID + '-g'"
            :id="it.ID"
            :name="it.Name"
            :photo-path="it.PhotoPath"
          />
        </div>
      </div>
      <FxEmptyState v-else class="mt-4" icon="cube" :title="$t('locationDetail.emptyItemsHere')" />
    </ItemsViewToggle>
  </div>
</template>
