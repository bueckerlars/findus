<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, useId, watch } from "vue";
import FxSvg from "./FxSvg.vue";

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
const menuId = useId();

const hintText = computed(
  () =>
    props.hint ||
    "Scan to open on your phone (same account).",
);

const downloadFilename = computed(() => {
  const raw = (props.downloadName || "item").replace(/[^\w\-._\s]+/g, "").trim().replace(/\s+/g, "-");
  const base = raw.length ? raw.slice(0, 80) : "item";
  return `findus-${base}-qr.png`;
});

function onDocPointerDown(e: PointerEvent) {
  if (!open.value) return;
  const t = e.target as Node | null;
  const el = root.value;
  if (el && t && !el.contains(t)) open.value = false;
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
  <div ref="root" class="relative inline-flex">
    <button
      type="button"
      class="fx-icon-btn"
      :class="{ 'border-sky-300/80 bg-sky-50/60 text-sky-800 ring-1 ring-sky-200/60': open }"
      aria-label="QR code"
      title="QR code"
      aria-haspopup="true"
      :aria-expanded="open ? 'true' : 'false'"
      :aria-controls="menuId"
      @click="toggle"
    >
      <FxSvg name="qr" />
    </button>
    <div
      v-show="open"
      :id="menuId"
      class="absolute right-0 top-full z-50 mt-2 w-[min(20rem,calc(100vw-2rem))] rounded-xl border border-zinc-200/90 bg-white p-4 shadow-lg shadow-zinc-900/10 ring-1 ring-zinc-950/[0.04]"
      role="region"
      aria-label="QR code"
      @click.stop
    >
      <p class="mb-3 text-center text-xs leading-snug text-zinc-500">
        {{ hintText }}
      </p>
      <div class="flex justify-center rounded-lg border border-zinc-100 bg-zinc-50/90 p-2">
        <img :src="pngUrl" width="200" height="200" alt="" class="h-48 w-48 max-w-full object-contain" />
      </div>
      <button type="button" class="fx-btn-primary mt-4 w-full text-sm" @click="downloadPng">Download PNG</button>
    </div>
  </div>
</template>
