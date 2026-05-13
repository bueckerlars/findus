<script setup lang="ts">
import { computed, useId } from "vue";

const props = defineProps<{ itemId: string }>();

/** Stable per mount; avoids duplicate fragment ids across the page (gallery grid). */
const mountKey = useId().replace(/[^a-zA-Z0-9_-]/g, "") || "ph";

const ids = computed(() => {
  const s = props.itemId.replace(/[^a-zA-Z0-9]/g, "_");
  const p = `fxiph-${mountKey}-${s}`;
  return {
    surf: `${p}-surf`,
    b1: `${p}-b1`,
    b2: `${p}-b2`,
    b3: `${p}-b3`,
    sheen: `${p}-sheen`,
  };
});

function urlRef(id: string) {
  return `url(#${id})`;
}
</script>

<template>
  <svg
    class="fx-item-photo-placeholder-svg block h-full w-full min-h-[8rem] text-zinc-400"
    viewBox="0 0 400 400"
    xmlns="http://www.w3.org/2000/svg"
    preserveAspectRatio="xMidYMid meet"
    role="presentation"
    aria-hidden="true"
  >
    <defs>
      <!-- objectBoundingBox: reliable when the tile is large (gallery) vs small (list thumb) -->
      <linearGradient :id="ids.surf" x1="0" y1="0" x2="1" y2="1" gradientUnits="objectBoundingBox">
        <stop offset="0%" stop-color="#fafafa" />
        <stop offset="42%" stop-color="#e7e5e4" />
        <stop offset="100%" stop-color="#cbd5e1" />
      </linearGradient>
      <radialGradient :id="ids.b1" cx="0.3" cy="0.24" r="0.65" gradientUnits="objectBoundingBox">
        <stop offset="0%" stop-color="#38bdf8" stop-opacity="0.5" />
        <stop offset="55%" stop-color="#38bdf8" stop-opacity="0.12" />
        <stop offset="100%" stop-color="#38bdf8" stop-opacity="0" />
      </radialGradient>
      <radialGradient :id="ids.b2" cx="0.78" cy="0.21" r="0.55" gradientUnits="objectBoundingBox">
        <stop offset="0%" stop-color="#818cf8" stop-opacity="0.42" />
        <stop offset="60%" stop-color="#818cf8" stop-opacity="0.1" />
        <stop offset="100%" stop-color="#818cf8" stop-opacity="0" />
      </radialGradient>
      <radialGradient :id="ids.b3" cx="0.52" cy="0.9" r="0.5" gradientUnits="objectBoundingBox">
        <stop offset="0%" stop-color="#fb7185" stop-opacity="0.28" />
        <stop offset="70%" stop-color="#fb7185" stop-opacity="0" />
        <stop offset="100%" stop-color="#fb7185" stop-opacity="0" />
      </radialGradient>
      <linearGradient :id="ids.sheen" x1="0" y1="1" x2="1" y2="0" gradientUnits="objectBoundingBox">
        <stop offset="0%" stop-color="#0ea5e9" stop-opacity="0" />
        <stop offset="48%" stop-color="#0ea5e9" stop-opacity="0.09" />
        <stop offset="100%" stop-color="#6366f1" stop-opacity="0" />
      </linearGradient>
    </defs>
    <rect width="400" height="400" :fill="urlRef(ids.surf)" />
    <rect width="400" height="400" :fill="urlRef(ids.b1)" />
    <rect width="400" height="400" :fill="urlRef(ids.b2)" />
    <rect width="400" height="400" :fill="urlRef(ids.b3)" />
    <rect width="400" height="400" :fill="urlRef(ids.sheen)" />
    <path d="M-40 340 C 80 260 200 300 440 220 L 440 440 L -40 440 Z" fill="#0ea5e9" fill-opacity="0.04" />
    <ellipse cx="200" cy="198" rx="132" ry="118" fill="none" stroke="#0ea5e9" stroke-opacity="0.09" stroke-width="1.25" />
    <ellipse cx="200" cy="198" rx="96" ry="86" fill="none" stroke="#6366f1" stroke-opacity="0.11" stroke-width="1" />
    <path d="M-24 168 L424 48" stroke="#64748b" stroke-opacity="0.07" stroke-width="1.25" stroke-linecap="round" />
    <path d="M32 432 L388 -32" stroke="#94a3b8" stroke-opacity="0.06" stroke-width="1" stroke-linecap="round" />
    <circle cx="200" cy="198" r="4" fill="#0ea5e9" fill-opacity="0.2" />
  </svg>
</template>
