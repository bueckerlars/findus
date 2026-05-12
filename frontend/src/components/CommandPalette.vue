<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api } from "../api";
import { useSession } from "../session";
import FxSvg from "./FxSvg.vue";
import type { FxIconName } from "./FxSvg.vue";
import { useCreateModals } from "../composables/useCreateModals";
import { useItemEditCommandHandlers } from "../composables/useItemEditCommandBridge";

const router = useRouter();
const route = useRoute();
const { isAdmin } = useSession();
const { openCreateItem, openCreateLocation, openCreateLabel } = useCreateModals();
const itemEditHandlers = useItemEditCommandHandlers();

const dialog = ref<HTMLDialogElement | null>(null);
const inputEl = ref<HTMLInputElement | null>(null);
const triggerBtn = ref<HTMLButtonElement | null>(null);
const q = ref("");
const selectedIdx = ref(0);
const searchItems = ref<{ id: string; name: string; location_name: string }[]>([]);
const showResultsWrap = ref(false);
let debounceTimer: ReturnType<typeof setTimeout>;
let fetchAbort: AbortController | null = null;
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
    searchItems.value = [];
    showResultsWrap.value = false;
    filterStatic("");
    clampSelection();
    return;
  }
  fetchAbort = new AbortController();
  try {
    const data = await api<{ items: { id: string; name: string; location_name: string }[] }>(
      "/api/command-search?q=" + encodeURIComponent(query.trim()),
      { signal: fetchAbort.signal },
    );
    const items = data?.items || [];
    filterStatic(query);
    renderSearchItems(items, query);
    clampSelection();
  } catch (e) {
    if ((e as Error).name === "AbortError") return;
    filterStatic(query);
    renderSearchItems([], query);
    clampSelection();
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

  const onTransitionEnd = (e: TransitionEvent) => {
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
  if (action.startsWith("focus:")) {
    const sel = action.slice("focus:".length).trim();
    lastFocus = null;
    focusTargetAfterModal(sel);
    closePalette();
    return;
  }
  if (action === "item-edit:save" || action === "item-edit:cancel") {
    lastFocus = null;
    closePalette();
    const h = itemEditHandlers.value;
    if (h) {
      if (action === "item-edit:save") void h.save();
      else void h.cancel();
    }
    return;
  }
  closePalette();
  if (action === "back") {
    nextTick(() => void router.back());
  }
}

function goCreate(kind: "item" | "location" | "label") {
  closePalette();
  if (kind === "item") openCreateItem();
  else if (kind === "location") openCreateLocation();
  else openCreateLabel();
}

function activateSelected() {
  const items = visibleItems();
  const el = items[selectedIdx.value];
  if (!el) return;
  const create = el.getAttribute("data-create");
  if (create === "item" || create === "location" || create === "label") {
    goCreate(create);
    return;
  }
  const cmdAction = el.getAttribute("data-cmd-action");
  if (cmdAction) {
    runCmdAction(cmdAction);
    return;
  }
  const href = el.getAttribute("data-href");
  if (href) goHref(href);
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
  const btn = (e.target as HTMLElement)?.closest?.("[data-href], [data-create], [data-cmd-action]") as HTMLElement | null;
  if (!btn || !dialog.value?.contains(btn)) return;
  if (!btn.hasAttribute("data-cmd-static") && !btn.hasAttribute("data-cmd-search-hit") && btn.id !== "fx-command-open-search")
    return;
  e.preventDefault();
  const create = btn.getAttribute("data-create");
  if (create === "item" || create === "location" || create === "label") {
    goCreate(create);
    return;
  }
  const cmdAction = btn.getAttribute("data-cmd-action");
  if (cmdAction) {
    runCmdAction(cmdAction);
    return;
  }
  const href = btn.getAttribute("data-href");
  if (href) goHref(href);
}

function onBodyMousemove(e: MouseEvent) {
  const btn = (e.target as HTMLElement)?.closest?.("[data-href], [data-create], [data-cmd-action]") as HTMLElement | null;
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

const openSearchQText = computed(() => (norm(q.value).length > 0 ? ` for “${q.value.trim()}”` : ""));

const inputPlaceholder = computed(() =>
  route.path === "/search" ? "Filter or jump to a page…" : "Search items, jump to a page…",
);

type ContextCommand = {
  id: string;
  label: string;
  keywords: string;
  icon: FxIconName;
  href?: string;
  action?: string;
};

const contextCommands = computed((): ContextCommand[] => {
  const path = route.path;
  const admin = isAdmin.value;
  const out: ContextCommand[] = [];

  const itemsDetail = /^\/items\/([^/]+)$/.exec(path);
  if (itemsDetail?.[1] && itemsDetail[1] !== "new") {
    const id = itemsDetail[1];
    if (admin && itemEditHandlers.value) {
      out.push(
        {
          id: "ctx-item-save",
          label: "Save item",
          keywords: "save commit apply submit store persist",
          icon: "check",
          action: "item-edit:save",
        },
        {
          id: "ctx-item-cancel-edit",
          label: "Cancel editing",
          keywords: "cancel discard exit close abort revert view discard changes",
          icon: "eye",
          action: "item-edit:cancel",
        },
      );
    } else if (admin) {
      out.push({
        id: "ctx-item-edit",
        label: "Edit item",
        keywords: "edit modify save form",
        icon: "pencilSquare",
        href: `/items/${id}?edit=1`,
      });
    }
    out.push(
      { id: "ctx-item-all", label: "All items", keywords: "items list inventory back", icon: "cube", href: "/items" },
      {
        id: "ctx-item-search",
        label: "Search",
        keywords: "search find lookup filter",
        icon: "magnifyingGlass",
        href: "/search",
      },
    );
    return out;
  }

  const locDetail = /^\/locations\/([^/]+)$/.exec(path);
  if (locDetail?.[1] && locDetail[1] !== "new") {
    const id = locDetail[1];
    if (admin) {
      out.push(
        {
          id: "ctx-loc-edit",
          label: "Edit location",
          keywords: "edit modify rename",
          icon: "pencilSquare",
          href: `/locations/${id}/edit`,
        },
        {
          id: "ctx-loc-subloc",
          label: "Add sub-location",
          keywords: "new child sub place room",
          icon: "plus",
          href: `/locations/new?parent_id=${encodeURIComponent(id)}`,
        },
      );
    }
    out.push({
      id: "ctx-loc-all",
      label: "All locations",
      keywords: "locations places tree map back",
      icon: "mapPin",
      href: "/locations",
    });
    return out;
  }

  const locEdit = /^\/locations\/([^/]+)\/edit$/.exec(path);
  if (locEdit?.[1]) {
    const id = locEdit[1];
    out.push({
      id: "ctx-loc-view",
      label: "View location",
      keywords: "back cancel detail read",
      icon: "eye",
      href: `/locations/${id}`,
    });
    return out;
  }

  if (path === "/locations/new") {
    const pid = route.query.parent_id;
    if (typeof pid === "string" && pid.trim()) {
      out.push({
        id: "ctx-loc-parent",
        label: "Open parent location",
        keywords: "parent back navigate",
        icon: "mapPin",
        href: `/locations/${pid.trim()}`,
      });
    }
    out.push({
      id: "ctx-loc-new-all",
      label: "All locations",
      keywords: "locations list",
      icon: "mapPin",
      href: "/locations",
    });
    return out;
  }

  if (path === "/items") {
    out.push({
      id: "ctx-items-search",
      label: "Open search",
      keywords: "search find lookup filter",
      icon: "magnifyingGlass",
      href: "/search",
    });
    return out;
  }

  if (path === "/") {
    out.push({
      id: "ctx-home-search",
      label: "Open search",
      keywords: "search find lookup filter",
      icon: "magnifyingGlass",
      href: "/search",
    });
    return out;
  }

  if (path === "/search") {
    out.push(
      {
        id: "ctx-search-focus",
        label: "Focus search field",
        keywords: "focus type query input cursor",
        icon: "magnifyingGlass",
        action: "focus:#q",
      },
      { id: "ctx-search-home", label: "Home", keywords: "dashboard start", icon: "home", href: "/" },
    );
    return out;
  }

  if (path === "/labels") {
    if (admin) {
      out.push({
        id: "ctx-labels-new",
        label: "New label",
        keywords: "add create tag",
        icon: "plus",
        href: "/labels/new",
      });
    }
    return out;
  }

  const labelEdit = /^\/labels\/([^/]+)\/edit$/.exec(path);
  if (labelEdit?.[1]) {
    out.push({
      id: "ctx-label-all",
      label: "All labels",
      keywords: "labels list tags back",
      icon: "tag",
      href: "/labels",
    });
    return out;
  }

  if (path === "/labels/new") {
    out.push({
      id: "ctx-label-new-all",
      label: "All labels",
      keywords: "labels list cancel back",
      icon: "tag",
      href: "/labels",
    });
    return out;
  }

  if (path === "/admin/users") {
    out.push(
      {
        id: "ctx-admin-tpl",
        label: "Templates",
        keywords: "item template fields editor",
        icon: "gear",
        href: "/admin/templates",
      },
      { id: "ctx-admin-home", label: "Home", keywords: "dashboard", icon: "home", href: "/" },
    );
    return out;
  }

  if (path === "/admin/templates") {
    out.push(
      {
        id: "ctx-tpl-users",
        label: "Users",
        keywords: "admin invites user management",
        icon: "users",
        href: "/admin/users",
      },
      {
        id: "ctx-tpl-new",
        label: "New template",
        keywords: "add create",
        icon: "plus",
        href: "/admin/templates/new",
      },
    );
    return out;
  }

  if (path === "/admin/templates/new") {
    out.push({
      id: "ctx-tpl-new-list",
      label: "All templates",
      keywords: "templates list back cancel",
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  const tplEdit = /^\/admin\/templates\/([^/]+)\/edit$/.exec(path);
  if (tplEdit?.[1]) {
    out.push({
      id: "ctx-tpl-edit-list",
      label: "All templates",
      keywords: "templates list back cancel",
      icon: "gear",
      href: "/admin/templates",
    });
    return out;
  }

  if (path === "/items/new") {
    out.push({
      id: "ctx-itemform-items",
      label: "All items",
      keywords: "items list cancel back",
      icon: "cube",
      href: "/items",
    });
    return out;
  }

  if (path === "/profile") {
    out.push({
      id: "ctx-profile-home",
      label: "Home",
      keywords: "dashboard",
      icon: "home",
      href: "/",
    });
    return out;
  }

  return out;
});

type CreateKind = "item" | "location" | "label";

const createKindMeta: Record<CreateKind, { label: string; keywords: string }> = {
  item: { label: "New item", keywords: "new item add create" },
  location: { label: "New location", keywords: "new location place room shelf add create" },
  label: { label: "New label", keywords: "new label tag add create" },
};

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
    ...createKindMeta[kind],
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
      <h2 id="fx-command-title" class="sr-only">Command palette</h2>
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
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">On this page</p>
          <div class="space-y-px" role="listbox" aria-label="On this page">
            <button
              v-for="cmd in contextCommands"
              :key="cmd.id"
              type="button"
              data-cmd-static
              :data-keywords="cmd.keywords"
              :data-href="cmd.href"
              :data-cmd-action="cmd.action"
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
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">Create</p>
          <div class="space-y-px" role="listbox" aria-label="Create">
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
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">Go to</p>
          <div class="space-y-px" role="listbox" aria-label="Pages">
            <button
              type="button"
              data-cmd-static
              data-href="/"
              data-keywords="home start dashboard"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="home" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Home</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/locations"
              data-keywords="locations places rooms shelves map tree"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="mapPin" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Locations</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/items"
              data-keywords="items inventory things stuff"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="cube" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Items</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/labels"
              data-keywords="labels tags categories"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="tag" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Labels</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/search"
              data-keywords="search find lookup filter"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="magnifyingGlass" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Search</span>
            </button>
            <button
              v-if="isAdmin"
              type="button"
              data-cmd-static
              data-href="/admin/users"
              data-keywords="admin users user management invites registration backup settings"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="gear" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Admin · Users</span>
            </button>
            <button
              v-if="isAdmin"
              type="button"
              data-cmd-static
              data-href="/admin/templates"
              data-keywords="admin templates item template fields editor"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="gear" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Admin · Templates</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/profile"
              data-keywords="profile account user password avatar email"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="users" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Profile</span>
            </button>
          </div>
        </div>
        <div
          id="fx-command-search-wrap"
          data-fx-cmd-group="results"
          class="border-t border-zinc-400/15 pt-2"
          :class="{ hidden: !showResultsWrap }"
        >
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">Items</p>
          <div id="fx-command-search-results" class="space-y-px" role="listbox" aria-label="Matching items">
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
              >Open full search<span id="fx-command-open-search-q" class="block truncate text-[11px] font-normal text-zinc-500">{{ openSearchQText }}</span></span
            >
          </button>
        </div>
      </div>
      <div class="flex shrink-0 items-center justify-between gap-2 border-t border-zinc-400/15 px-3 py-1.5 text-[10px] text-zinc-600 sm:px-3.5">
        <span
          ><span class="font-medium text-zinc-700">↑↓</span> select · <span class="font-medium text-zinc-700">↵</span> open ·
          <span class="font-medium text-zinc-700">esc</span> close</span
        >
        <span class="hidden text-zinc-500 sm:inline">Quick find</span>
      </div>
    </div>
  </dialog>
  <button
    id="fx-command-trigger"
    ref="triggerBtn"
    type="button"
    class="fx-command-trigger fx-command-glass sm:gap-2 sm:px-3 sm:py-1.5"
    aria-label="Open command palette"
    aria-haspopup="dialog"
    aria-controls="fx-command-dialog"
    aria-expanded="false"
    aria-keyshortcuts="Meta+K Control+K"
    @click="openPalette"
  >
    <span class="fx-command-trigger-icon" aria-hidden="true"><FxSvg name="magnifyingGlass" class="h-4 w-4 shrink-0" /></span>
    <kbd class="fx-command-kbd pointer-events-none hidden min-[380px]:inline-flex" title="Keyboard shortcut">
      <span>{{ modLabel }}</span
      ><span>K</span>
    </kbd>
  </button>
</template>
