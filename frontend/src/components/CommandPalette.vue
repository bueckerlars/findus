<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRouter } from "vue-router";
import { api } from "../api";
import { useSession } from "../session";
import FxSvg from "./FxSvg.vue";

const router = useRouter();
const { isAdmin } = useSession();

const dialog = ref<HTMLDialogElement | null>(null);
const inputEl = ref<HTMLInputElement | null>(null);
const triggerBtn = ref<HTMLButtonElement | null>(null);
const q = ref("");
const selectedIdx = ref(0);
const searchItems = ref<{ id: string; name: string; type: string }[]>([]);
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

function renderSearchItems(items: { id: string; name: string; type: string }[], query: string) {
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
    const data = await api<{ items: { id: string; name: string; type: string }[] }>(
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
  closePalette();
  void router.push(href);
}

function activateSelected() {
  const items = visibleItems();
  const el = items[selectedIdx.value];
  if (!el) return;
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
  const btn = (e.target as HTMLElement)?.closest?.("[data-href]") as HTMLElement | null;
  if (!btn || !dialog.value?.contains(btn)) return;
  if (!btn.hasAttribute("data-cmd-static") && !btn.hasAttribute("data-cmd-search-hit") && btn.id !== "fx-command-open-search")
    return;
  e.preventDefault();
  const href = btn.getAttribute("data-href");
  if (href) goHref(href);
}

function onBodyMousemove(e: MouseEvent) {
  const btn = (e.target as HTMLElement)?.closest?.("[data-href]") as HTMLElement | null;
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
          placeholder="Search items, jump to a page…"
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
        <div v-if="isAdmin" data-fx-cmd-group="create" class="mb-2">
          <p class="px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-zinc-400">Create</p>
          <div class="space-y-px" role="listbox" aria-label="Create">
            <button
              type="button"
              data-cmd-static
              data-href="/items/new"
              data-keywords="new item add create"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="plus" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">New item</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/locations/new"
              data-keywords="new location place room shelf add create"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="plus" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">New location</span>
            </button>
            <button
              type="button"
              data-cmd-static
              data-href="/labels/new"
              data-keywords="new label tag add create"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="plus" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">New label</span>
            </button>
          </div>
        </div>
        <div data-fx-cmd-group="go" :class="['mb-2', isAdmin ? 'border-t border-zinc-400/15 pt-2' : '']">
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
              data-href="/admin"
              data-keywords="admin settings users invites backup templates"
              class="fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25"
            >
              <span
                class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70"
                ><FxSvg name="gear" class="h-3.5 w-3.5 shrink-0"
              /></span>
              <span class="min-w-0 flex-1 font-medium text-zinc-900">Admin</span>
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
              <span class="min-w-0 flex-1 font-medium text-zinc-900">{{ it.name }}</span>
              <span
                class="shrink-0 rounded px-1.5 py-0.5 text-[10px] font-medium capitalize leading-none text-zinc-600 ring-1 ring-zinc-200/60 bg-zinc-100/50"
                >{{ it.type }}</span
              >
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
