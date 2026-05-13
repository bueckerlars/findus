<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";
import { useSession } from "../session";
import FxSvg from "./FxSvg.vue";
import { useCreateModals } from "../composables/useCreateModals";
import { useItemDetailCommandHandlers } from "../composables/useItemDetailCommandBridge";
import { useLocationDetailCommandHandlers } from "../composables/useLocationDetailCommandBridge";
import { useI18n } from "vue-i18n";
import { buildContextCommands } from "./commandPaletteContext";

const router = useRouter();
const route = useRoute();
const { t, locale } = useI18n();
const { isAdmin } = useSession();
const { openCreateItem, openCreateLocation, openCreateLabel } = useCreateModals();
const itemDetailHandlers = useItemDetailCommandHandlers();
const locationDetailHandlers = useLocationDetailCommandHandlers();

const dialog = ref<HTMLDialogElement | null>(null);
const inputEl = ref<HTMLInputElement | null>(null);
const triggerBtn = ref<HTMLButtonElement | null>(null);
const q = ref("");
const selectedIdx = ref(0);
const searchItems = ref<{ id: string; name: string; location_name: string }[]>([]);
const showResultsWrap = ref(false);
const commandSearchLoading = ref(false);
let debounceTimer: ReturnType<typeof setTimeout>;
let fetchAbort: AbortController | null = null;
let searchReq = 0;
let lastFocus: Element | null = null;

const COMMAND_CLOSE_FALLBACK_MS = 320;

const isMac = computed(() => {
  const nav = navigator as Navigator & { userAgentData?: { platform?: string } };
  const ua = nav.userAgentData;
  return /Mac|iPhone|iPad|iPod/i.test(navigator.platform || "") || ua?.platform === "macOS";
});

const modLabel = computed(() => (isMac.value ? "⌘" : "Ctrl+"));

function norm(s: string) {
  return (s || "").toLowerCase().trim();
}

function staticItems(): HTMLButtonElement[] {
  return Array.from(
    dialog.value?.querySelectorAll<HTMLButtonElement>("[data-cmd-static][data-keywords]") || [],
  );
}

function searchResultButtons(): HTMLButtonElement[] {
  return Array.from(dialog.value?.querySelectorAll<HTMLButtonElement>("[data-cmd-search-hit]") || []);
}

function openSearchHitEls(): HTMLButtonElement[] {
  const o = document.getElementById("fx-command-open-search");
  return o && !o.classList.contains("hidden") ? [o as HTMLButtonElement] : [];
}

function visibleItems(): HTMLElement[] {
  const a = staticItems().filter((el) => !el.classList.contains("hidden"));
  const b = searchResultButtons();
  const out: HTMLElement[] = [...a, ...b];
  out.push(...openSearchHitEls());
  return out;
}

function syncGroupVisibility() {
  dialog.value?.querySelectorAll("[data-fx-cmd-group]").forEach((g) => {
    if (g.id === "fx-command-search-wrap") return;
    const any = g.querySelector("[data-cmd-static][data-keywords]:not(.hidden), [data-cmd-search-hit]");
    g.classList.toggle("hidden", !any);
  });
}

function filterStatic(query: string) {
  const nq = norm(query);
  staticItems().forEach((el) => {
    const kw = (el.getAttribute("data-keywords") || "") + " " + (el.textContent || "");
    if (!nq) {
      el.classList.remove("hidden");
      return;
    }
    el.classList.toggle("hidden", norm(kw).indexOf(nq) === -1);
  });
  syncGroupVisibility();
}

function setActive(items: HTMLElement[]) {
  items.forEach((el, i) => {
    const on = i === selectedIdx.value;
    el.setAttribute("aria-selected", on ? "true" : "false");
    el.classList.toggle("fx-command-item-active", on);
    if (on) {
      try {
        el.scrollIntoView({ block: "nearest", behavior: "smooth" });
      } catch {
        el.scrollIntoView(false);
      }
    }
  });
}

function clampSelection() {
  const items = visibleItems();
  if (items.length === 0) {
    selectedIdx.value = 0;
    return;
  }
  if (selectedIdx.value >= items.length) selectedIdx.value = items.length - 1;
  if (selectedIdx.value < 0) selectedIdx.value = 0;
  setActive(items);
}

function renderSearchItems(items: { id: string; name: string; location_name: string }[], query: string) {
  searchItems.value = items;
  const hasQ = norm(query).length > 0;
  showResultsWrap.value = items.length > 0 || hasQ;
  selectedIdx.value = 0;
  nextTick(() => clampSelection());
}

async function runSearch(query: string) {
  if (fetchAbort) fetchAbort.abort();
  const nq = norm(query);
  if (!nq) {
    searchReq++;
    commandSearchLoading.value = false;
    searchItems.value = [];
    showResultsWrap.value = false;
    filterStatic("");
    clampSelection();
    return;
  }
  const myReq = ++searchReq;
  commandSearchLoading.value = true;
  fetchAbort = new AbortController();
  try {
    const data = await api<{ items: { id: string; name: string; location_name: string }[] }>(
      "/api/command-search?q=" + encodeURIComponent(query.trim()),
      { signal: fetchAbort.signal },
    );
    if (myReq !== searchReq) return;
    const items = data?.items || [];
    filterStatic(query);
    renderSearchItems(items, query);
    clampSelection();
  } catch (e) {
    if ((e as Error).name === "AbortError") return;
    if (myReq !== searchReq) return;
    filterStatic(query);
    renderSearchItems([], query);
    clampSelection();
  } finally {
    if (myReq === searchReq) commandSearchLoading.value = false;
  }
}

function scheduleSearch() {
  clearTimeout(debounceTimer);
  filterStatic(q.value);
  debounceTimer = setTimeout(() => {
    void runSearch(q.value);
  }, 220);
}

function openPalette() {
  if (dialog.value?.open) return;
  dialog.value?.classList.remove("fx-command-dialog--closing");
  lastFocus = document.activeElement;
  dialog.value?.showModal();
  triggerBtn.value?.setAttribute("aria-expanded", "true");
  q.value = "";
  searchReq++;
  commandSearchLoading.value = false;
  filterStatic("");
  searchItems.value = [];
  showResultsWrap.value = false;
  selectedIdx.value = 0;
  nextTick(() => {
    inputEl.value?.focus();
    inputEl.value?.select();
    clampSelection();
  });
}

function closePalette() {
  const d = dialog.value;
  if (!d?.open) return;
  if (d.classList.contains("fx-command-dialog--closing")) return;

  triggerBtn.value?.setAttribute("aria-expanded", "false");

  const panel = d.querySelector(".fx-command-panel");
  const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (!panel || reduce) {
    d.close();
    return;
  }

  d.classList.add("fx-command-dialog--closing");
  void d.offsetHeight;

  let finished = false;
  const cleanup = () => {
    if (finished) return;
    finished = true;
    panel.removeEventListener("transitionend", onTransitionEnd);
    d.classList.remove("fx-command-dialog--closing");
    if (d.open) d.close();
  };

  const onTransitionEnd = (e: Event) => {
    if (!(e instanceof TransitionEvent)) return;
    if (e.target !== panel) return;
    if (e.propertyName !== "opacity") return;
    cleanup();
  };

  panel.addEventListener("transitionend", onTransitionEnd);
  window.setTimeout(cleanup, COMMAND_CLOSE_FALLBACK_MS);
}

function goHref(href: string) {
  lastFocus = null;
  const d = dialog.value;
  const navigate = () => {
    void router.push(href);
  };
  if (d?.open) {
    d.addEventListener("close", () => nextTick(navigate), { once: true });
    closePalette();
  } else {
    closePalette();
    nextTick(navigate);
  }
}

function focusTargetAfterModal(sel: string) {
  const apply = () => {
    nextTick(() => {
      const el = document.querySelector(sel) as HTMLElement | null;
      el?.focus();
      if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement) {
        try {
          el.select();
        } catch {
          /* ignore */
        }
      }
    });
  };
  const d = dialog.value;
  if (d?.open) {
    d.addEventListener("close", apply, { once: true });
  } else {
    apply();
  }
}

function runCmdAction(action: string) {
  if (!action) return;
  if (action.startsWith("focus:")) {
    const sel = action.slice("focus:".length).trim();
    lastFocus = null;
    focusTargetAfterModal(sel);
    closePalette();
    return;
  }
  if (action.startsWith("tab:")) {
    const url = action.slice(4);
    lastFocus = null;
    const d = dialog.value;
    const open = () => window.open(url, "_blank", "noopener,noreferrer");
    if (d?.open) d.addEventListener("close", () => nextTick(open), { once: true });
    else nextTick(open);
    closePalette();
    return;
  }
  if (action.startsWith("item-detail:")) {
    const sub = action.slice("item-detail:".length);
    const h = itemDetailHandlers.value;
    if (!h) return;
    lastFocus = null;
    const d = dialog.value;
    const run = () => {
      if (sub === "save" && h.save) void h.save();
      else if (sub === "cancel" && h.cancel) void h.cancel();
      else if (sub === "delete" && h.deleteItem) void h.deleteItem();
      else if (sub === "download-qr" && h.downloadQrPng) void h.downloadQrPng();
      else if (sub === "copy-link" && h.copyPageLink) void h.copyPageLink();
    };
    if (d?.open) d.addEventListener("close", () => nextTick(run), { once: true });
    else nextTick(run);
    closePalette();
    return;
  }
  if (action.startsWith("loc-detail:")) {
    const sub = action.slice("loc-detail:".length);
    const h = locationDetailHandlers.value;
    if (!h) return;
    lastFocus = null;
    const d = dialog.value;
    const run = () => {
      if (sub === "delete" && h.deleteLocation) void h.deleteLocation();
      else if (sub === "download-qr" && h.downloadQrPng) void h.downloadQrPng();
      else if (sub === "copy-link" && h.copyPageLink) void h.copyPageLink();
    };
    if (d?.open) d.addEventListener("close", () => nextTick(run), { once: true });
    else nextTick(run);
    closePalette();
    return;
  }
  closePalette();
  if (action === "back") {
    nextTick(() => void router.back());
  }
}

function goCreate(
  kind: "item" | "location" | "label",
  opts?: { itemLocationId?: string; locationParentId?: string },
) {
  closePalette();
  if (kind === "item") openCreateItem(opts?.itemLocationId ? { locationId: opts.itemLocationId } : undefined);
  else if (kind === "location") openCreateLocation(opts?.locationParentId ? { parentId: opts.locationParentId } : undefined);
  else openCreateLabel();
}

function handleCommandButton(btn: HTMLElement) {
  const clipboard = btn.getAttribute("data-clipboard");
  if (clipboard) {
    lastFocus = null;
    const d = dialog.value;
    const run = () => void navigator.clipboard.writeText(clipboard).catch(() => {});
    if (d?.open) d.addEventListener("close", () => nextTick(run), { once: true });
    else nextTick(run);
    closePalette();
    return true;
  }
  const ext = btn.getAttribute("data-external-href");
  if (ext) {
    lastFocus = null;
    const d = dialog.value;
    const go = () => {
      window.location.href = ext;
    };
    if (d?.open) d.addEventListener("close", go, { once: true });
    else go();
    closePalette();
    return true;
  }
  const durl = btn.getAttribute("data-download-url");
  if (durl) {
    const fn = btn.getAttribute("data-download-filename") || "download.png";
    lastFocus = null;
    const d = dialog.value;
    const run = () => {
      const a = document.createElement("a");
      a.href = durl;
      a.download = fn;
      a.rel = "noopener";
      document.body.appendChild(a);
      a.click();
      a.remove();
    };
    if (d?.open) d.addEventListener("close", () => nextTick(run), { once: true });
    else nextTick(run);
    closePalette();
    return true;
  }

  const create = btn.getAttribute("data-create");
  if (create === "item" || create === "location" || create === "label") {
    const itemLoc = btn.getAttribute("data-create-location-id");
    const locParent = btn.getAttribute("data-create-parent-id");
    if (create === "item" && itemLoc) goCreate("item", { itemLocationId: itemLoc });
    else if (create === "location" && locParent) goCreate("location", { locationParentId: locParent });
    else goCreate(create);
    return true;
  }
  const cmdAction = btn.getAttribute("data-cmd-action");
  if (cmdAction) {
    runCmdAction(cmdAction);
    return true;
  }
  const href = btn.getAttribute("data-href");
  if (href) {
    goHref(href);
    return true;
  }
  return false;
}

function activateSelected() {
  const items = visibleItems();
  const el = items[selectedIdx.value];
  if (!el) return;
  void handleCommandButton(el);
}

function onGlobalKeydown(e: KeyboardEvent) {
  const cmdk = (e.metaKey || e.ctrlKey) && (e.key === "k" || e.key === "K");
  if (!cmdk) return;
  e.preventDefault();
  if (dialog.value?.open) closePalette();
  else openPalette();
}

function onDialogCancel(e: Event) {
  e.preventDefault();
  closePalette();
}

function onDialogClose() {
  clearTimeout(debounceTimer);
  if (fetchAbort) fetchAbort.abort();
  triggerBtn.value?.setAttribute("aria-expanded", "false");
  const lf = lastFocus as HTMLElement | null;
  if (lf && typeof lf.focus === "function") {
    try {
      lf.focus();
    } catch {
      /* ignore */
    }
  }
}

function onInputKeydown(e: KeyboardEvent) {
  if (e.key === "ArrowDown") {
    e.preventDefault();
    const n = visibleItems().length;
    if (n) selectedIdx.value = (selectedIdx.value + 1) % n;
    clampSelection();
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    const m = visibleItems().length;
    if (m) selectedIdx.value = (selectedIdx.value - 1 + m) % m;
    clampSelection();
  } else if (e.key === "Enter") {
    e.preventDefault();
    activateSelected();
  }
}

function onBodyClick(e: MouseEvent) {
  const btn = (e.target as HTMLElement)?.closest?.(
    "[data-href], [data-create], [data-cmd-action], [data-clipboard], [data-external-href], [data-download-url]",
  ) as HTMLElement | null;
  if (!btn || !dialog.value?.contains(btn)) return;
  if (!btn.hasAttribute("data-cmd-static") && !btn.hasAttribute("data-cmd-search-hit") && btn.id !== "fx-command-open-search")
    return;
  e.preventDefault();
  void handleCommandButton(btn);
}

function onBodyMousemove(e: MouseEvent) {
  const btn = (e.target as HTMLElement)?.closest?.(
    "[data-href], [data-create], [data-cmd-action], [data-clipboard], [data-external-href], [data-download-url]",
  ) as HTMLElement | null;
  if (!btn || !dialog.value?.contains(btn)) return;
  const items = visibleItems();
  const idx = items.indexOf(btn as HTMLElement);
  if (idx >= 0) {
    selectedIdx.value = idx;
    setActive(items);
  }
}

onMounted(() => {
  document.addEventListener("keydown", onGlobalKeydown, true);
});
onUnmounted(() => {
  document.removeEventListener("keydown", onGlobalKeydown, true);
});

watch(q, () => scheduleSearch());

const openSearchQText = computed(() =>
  norm(q.value).length > 0 ? t("cpUi.openSearchFor", { q: q.value.trim() }) : "",
);

const inputPlaceholder = computed(() =>
  route.path === "/search" ? t("cpUi.placeholderFilter") : t("cpUi.placeholderSearch"),
);

const contextCommands = computed(() => {
  void locale.value;
  return buildContextCommands(
    route,
    {
      isAdmin: isAdmin.value,
      itemDetail: itemDetailHandlers.value,
      locationDetail: locationDetailHandlers.value,
    },
    t,
  );
});

type CreateKind = "item" | "location" | "label";

const createKindMeta = computed((): Record<CreateKind, { label: string; keywords: string }> => ({
  item: { label: t("cpUi.newItem"), keywords: t("cpUi.newItemKw") },
  location: { label: t("cpUi.newLocation"), keywords: t("cpUi.newLocationKw") },
  label: { label: t("cpUi.newLabel"), keywords: t("cpUi.newLabelKw") },
}));

const DEFAULT_CREATE_ORDER: CreateKind[] = ["item", "location", "label"];

const createCommandOrder = computed((): CreateKind[] => {
  const path = route.path;
  if (path === "/locations" || path.startsWith("/locations/")) {
    return ["location", "item", "label"];
  }
  if (path === "/items" || path.startsWith("/items/")) {
    return ["item", "location", "label"];
  }
  if (path === "/labels" || path.startsWith("/labels/")) {
    return ["label", "item", "location"];
  }
  return DEFAULT_CREATE_ORDER;
});

const createCommands = computed(() =>
  createCommandOrder.value.map((kind) => ({
    kind,
    ...createKindMeta.value[kind],
  })),
);
</script>

<template>
  <dialog
    id="fx-command-dialog"
    ref="dialog"
    class="fx-command-dialog m-auto w-[calc(100vw-1.5rem)] max-w-lg border-0 bg-transparent p-0 sm:w-[calc(100vw-2rem)]"
    aria-labelledby="fx-command-title"
    @cancel="onDialogCancel"
    @close="onDialogClose"
  >
    <div class="fx-command-panel fx-command-glass">
      <h2 id="fx-command-title" class="sr-only">{{ $t("cpUi.title") }}</h2>
      <div class="flex items-center gap-2 border-b border-zinc-400/15 px-3 py-2 sm:px-3.5">
        <span class="fx-command-panel-search-icon" aria-hidden="true"><FxSvg name="magnifyingGlass" class="fx-icon shrink-0" /></span>
        <input
          id="fx-command-input"
          ref="inputEl"
          v-model="q"
          type="search"
          enterkeyhint="search"
          autocomplete="off"
          autocorrect="off"
          spellcheck="false"
          :placeholder="inputPlaceholder"
          class="min-w-0 flex-1 border-0 bg-transparent text-sm text-zinc-800 outline-none placeholder:text-zinc-500 focus:ring-0"
          @keydown="onInputKeydown"
        />
        <kbd class="fx-command-kbd pointer-events-none hidden shrink-0 select-none sm:inline-flex" aria-hidden="true"
          ><span>{{ modLabel }}</span
          >K</kbd
        >
      </div>
      <div
        id="fx-command-body"
        class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-2 pb-1.5 pt-0.5 sm:px-2.5"
        @click="onBodyClick"
        @mousemove="onBodyMousemove"
      >
        <div v-if="contextCommands.length" data-fx-cmd-group="context" class="mb-2">
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("cpUi.onPage") }}</p>
          <div class="space-y-px" role="listbox" :aria-label="$t('cpUi.onPageAria')">
            <button
              v-for="cmd in contextCommands"
              :key="cmd.id"
              type="button"
              data-cmd-static
              :data-keywords="cmd.keywords"
              :data-href="cmd.href"
              :data-cmd-action="cmd.action"
              :data-clipboard="cmd.clipboard"
              :data-external-href="cmd.externalHref"
              :data-download-url="cmd.downloadUrl"
              :data-download-filename="cmd.downloadFilename"
              :data-create="cmd.createPreset?.kind"
              :data-create-location-id="cmd.createPreset?.locationId"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg :name="cmd.icon" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ cmd.label }}</span>
            </button>
          </div>
        </div>
        <div v-if="isAdmin" data-fx-cmd-group="create" :class="['mb-2', contextCommands.length ? 'border-t border-zinc-400/15 pt-2' : '']">
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("cpUi.create") }}</p>
          <div class="space-y-px" role="listbox" :aria-label="$t('cpUi.createAria')">
            <button
              v-for="row in createCommands"
              :key="row.kind"
              type="button"
              data-cmd-static
              :data-create="row.kind"
              :data-keywords="row.keywords"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="plus" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ row.label }}</span>
            </button>
          </div>
        </div>
        <div data-fx-cmd-group="go" :class="['mb-2', isAdmin || contextCommands.length ? 'border-t border-zinc-400/15 pt-2' : '']">
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("cpUi.goTo") }}</p>
          <div class="space-y-px" role="listbox" :aria-label="$t('cpUi.pagesAria')">
            <button
              type="button"
              data-cmd-static
              data-href="/"
              :data-keywords="$t('cpUi.goHomeKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="home" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goHome") }}</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/locations"
              :data-keywords="$t('cpUi.goLocationsKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="mapPin" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goLocations") }}</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/items"
              :data-keywords="$t('cpUi.goItemsKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="cube" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goItems") }}</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/labels"
              :data-keywords="$t('cpUi.goLabelsKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="tag" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goLabels") }}</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/search"
              :data-keywords="$t('cpUi.goSearchKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="magnifyingGlass" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goSearch") }}</span>
            </button>
            <button
              v-if="isAdmin"
              type="button"
              data-cmd-static
              data-href="/admin/users"
              :data-keywords="$t('cpUi.goAdminUsersKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="users" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goAdminUsers") }}</span>
            </button>
            <button
              v-if="isAdmin"
              type="button"
              data-cmd-static
              data-href="/admin/settings"
              :data-keywords="$t('cpUi.goAdminSettingsKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="gear" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goAdminSettings") }}</span>
            </button>
            <button
              v-if="isAdmin"
              type="button"
              data-cmd-static
              data-href="/admin/templates"
              :data-keywords="$t('cpUi.goAdminTemplatesKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="gear" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goAdminTemplates") }}</span>
            </button>
            <button
              v-if="isAdmin"
              type="button"
              data-cmd-static
              data-href="/admin/label-generator"
              :data-keywords="$t('cpUi.goAdminLabelGeneratorKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="qr" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goAdminLabelGenerator") }}</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/profile"
              :data-keywords="$t('cpUi.goProfileKw')"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="users" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ $t("cpUi.goProfile") }}</span>
            </button>
          </div>
        </div>
        <div
          id="fx-command-search-wrap"
          data-fx-cmd-group="results"
          class="border-t border-zinc-400/15 pt-2"
          :class="{ hidden: !showResultsWrap }"
        >
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">{{ $t("cpUi.resultsItems") }}</p>
          <div
            id="fx-command-search-results"
            class="space-y-px"
            role="listbox"
            :aria-label="$t('cpUi.matchingItems')"
            :aria-busy="commandSearchLoading && norm(q).length > 0"
          >
            <template v-if="commandSearchLoading && norm(q).length > 0">
              <div
                v-for="n in 3"
                :key="'cmd-sk-' + n"
                class="fx-command-item flex w-full items-center gap-2 rounded-lg px-2 py-1.5"
                aria-hidden="true"
              >
                <span class="flex h-7 w-7 shrink-0 animate-pulse rounded-md bg-zinc-200/80"></span>
                <span class="min-w-0 flex-1 space-y-1.5">
                  <span class="block h-3.5 w-[min(12rem,55%)] animate-pulse rounded bg-zinc-200/90"></span>
                  <span class="block h-3 w-[min(9rem,45%)] animate-pulse rounded bg-zinc-200/70"></span>
                </span>
              </div>
            </template>
            <template v-else>
              <button
                v-for="it in searchItems"
                :key="it.id"
                type="button"
                data-cmd-search-hit
                :data-href="'/items/' + it.id"
                class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
              >
                <span
                  class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                  ><FxSvg name="cube" class="h-3.5 w-3.5 shrink-0"
                /></span>
                <span class="min-w-0 flex-1">
                  <span class="block truncate font-medium text-zinc-900">{{ it.name }}</span>
                  <span class="mt-0.5 flex min-w-0 items-center gap-1 text-[11px] font-medium leading-snug text-zinc-500">
                    <FxSvg name="mapPin" class="h-3 w-3 shrink-0 text-sky-600/90" aria-hidden="true" />
                    <span class="truncate">{{ it.location_name }}</span>
                  </span>
                </span>
              </button>
            </template>
          </div>
          <button
            id="fx-command-open-search"
            type="button"
            class="fx-command-item mt-0.5 flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            :class="{ hidden: !norm(q) }"
            :data-href="'/search?q=' + encodeURIComponent(q.trim())"
          >
            <span class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500"
              ><FxSvg name="magnifyingGlass" class="h-3.5 w-3.5 shrink-0"
            /></span>
            <span class="min-w-0 flex-1 font-medium text-zinc-900"
              >{{ $t("cpUi.openFullSearch") }}<span id="fx-command-open-search-q" class="block truncate text-[11px] font-normal text-zinc-500">{{ openSearchQText }}</span></span
            >
          </button>
        </div>
      </div>
      <div class="flex shrink-0 items-center justify-between gap-2 border-t border-zinc-400/15 px-3 py-1.5 text-[10px] text-zinc-600 sm:px-3.5">
        <span>{{ $t("cpUi.footerHints") }}</span>
        <span class="hidden text-zinc-500 sm:inline">{{ $t("cpUi.quickFind") }}</span>
      </div>
    </div>
  </dialog>
  <button
    id="fx-command-trigger"
    ref="triggerBtn"
    type="button"
    class="fx-command-trigger fx-command-glass sm:gap-2 sm:px-3 sm:py-1.5"
    :aria-label="$t('cpUi.openCommandPalette')"
    aria-haspopup="dialog"
    aria-controls="fx-command-dialog"
    aria-expanded="false"
    aria-keyshortcuts="Meta+K Control+K"
    @click="openPalette"
  >
    <span class="fx-command-trigger-icon" aria-hidden="true"><FxSvg name="magnifyingGlass" class="h-4 w-4 shrink-0" /></span>
    <kbd class="fx-command-kbd pointer-events-none hidden min-[380px]:inline-flex" :title="$t('cpUi.keyboardShortcut')">
      <span>{{ modLabel }}</span
      ><span>K</span>
    </kbd>
  </button>
</template>
