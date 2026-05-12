<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, useId, watch } from "vue";
import { useI18n } from "vue-i18n";
import FxSvg from "./FxSvg.vue";

const { t } = useI18n();

const props = withDefaults(
  defineProps<{
    pngUrl: string;
    /** Sanitized for `findus-{base}-qr.png` download filename. */
    downloadName?: string;
    hint?: string;
  }>(),
  {
    downloadName: "",
    hint: "",
  },
);

const emit = defineEmits<{
  /** Fired when the menu opens (e.g. close other popovers). */
  open: [];
}>();

const open = ref(false);
const root = ref<HTMLElement | null>(null);
const triggerRef = ref<HTMLButtonElement | null>(null);
const panelRef = ref<HTMLElement | null>(null);
const menuId = useId();

/** Fixed panel position (viewport); Teleport avoids parent overflow / stacking contexts. */
const menuStyle = ref({ top: "0px", right: "0px" });

const hintText = computed(() => props.hint || t("qr.hintGeneric"));

const downloadFilename = computed(() => {
  const raw = (props.downloadName || "item").replace(/[^\w\-._\s]+/g, "").trim().replace(/\s+/g, "-");
  const base = raw.length ? raw.slice(0, 80) : "item";
  return `findus-${base}-qr.png`;
});

function updateMenuPosition() {
  const btn = triggerRef.value;
  if (!btn || !open.value) return;
  const r = btn.getBoundingClientRect();
  const margin = 8;
  const vw = window.innerWidth;
  const vh = window.innerHeight;
  let topPx = r.bottom + margin;
  let rightPx = vw - r.right;
  menuStyle.value = { top: `${topPx}px`, right: `${rightPx}px` };
  requestAnimationFrame(() => {
    const panel = panelRef.value;
    if (!panel || !open.value) return;
    const pr = panel.getBoundingClientRect();
    if (pr.left < margin) rightPx = vw - margin - pr.width;
    if (pr.right > vw - margin) rightPx = margin;
    if (pr.bottom > vh - margin && r.top - margin - pr.height >= margin) {
      topPx = r.top - margin - pr.height;
    }
    menuStyle.value = { top: `${topPx}px`, right: `${rightPx}px` };
  });
}

function detachPositionListeners() {
  window.removeEventListener("scroll", updateMenuPosition, true);
  window.removeEventListener("resize", updateMenuPosition);
}

function onDocPointerDown(e: PointerEvent) {
  if (!open.value) return;
  const t = e.target as Node | null;
  const rootEl = root.value;
  const panelEl = panelRef.value;
  if (t && (rootEl?.contains(t) || panelEl?.contains(t))) return;
  open.value = false;
}

function onGlobalKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && open.value) open.value = false;
}

function toggle() {
  if (!open.value) emit("open");
  open.value = !open.value;
}

function close() {
  open.value = false;
}

function downloadPng() {
  const a = document.createElement("a");
  a.href = props.pngUrl;
  a.download = downloadFilename.value;
  a.rel = "noopener";
  document.body.appendChild(a);
  a.click();
  a.remove();
  close();
}

onMounted(() => {
  document.addEventListener("pointerdown", onDocPointerDown, true);
  document.addEventListener("keydown", onGlobalKeydown, true);
});

onUnmounted(() => {
  document.removeEventListener("pointerdown", onDocPointerDown, true);
  document.removeEventListener("keydown", onGlobalKeydown, true);
  detachPositionListeners();
});

watch(open, async (isOpen) => {
  if (isOpen) {
    await nextTick();
    updateMenuPosition();
    window.addEventListener("scroll", updateMenuPosition, true);
    window.addEventListener("resize", updateMenuPosition);
  } else {
    detachPositionListeners();
  }
});

watch(
  () => props.pngUrl,
  () => {
    open.value = false;
  },
);

defineExpose({ close });
</script>

<template>
  <div ref="root" class="inline-flex">
    <button
      ref="triggerRef"
      type="button"
      class="fx-icon-btn"
      :class="{ 'border-sky-300/80 bg-sky-50/60 text-sky-800 ring-1 ring-sky-200/60': open }"
      :aria-label="t('qr.code')"
      :title="t('qr.code')"
      aria-haspopup="true"
      :aria-expanded="open ? 'true' : 'false'"
      :aria-controls="menuId"
      @click="toggle"
    >
      <FxSvg name="qr" />
    </button>
  </div>
  <Teleport to="body">
    <div
      v-show="open"
      :id="menuId"
      ref="panelRef"
      class="fixed z-[90] w-[min(20rem,calc(100vw-2rem))] rounded-xl border border-zinc-200/90 bg-white p-4 shadow-lg shadow-zinc-900/10 ring-1 ring-zinc-950/[0.04]"
      role="region"
      :aria-label="t('qr.code')"
      :style="{ top: menuStyle.top, right: menuStyle.right }"
      @click.stop
    >
      <p class="mb-3 text-center text-xs leading-snug text-zinc-500">
        {{ hintText }}
      </p>
      <div class="flex justify-center rounded-lg border border-zinc-100 bg-zinc-50/90 p-2">
        <img :src="pngUrl" width="200" height="200" alt="" class="h-48 w-48 max-w-full object-contain" />
      </div>
      <button type="button" class="fx-btn-primary mt-4 w-full text-sm" @click="downloadPng">{{ t("qr.downloadPng") }}</button>
    </div>
  </Teleport>
</template>
