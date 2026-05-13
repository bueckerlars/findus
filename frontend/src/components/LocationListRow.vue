<script setup lang="ts">
import { RouterLink } from "vue-router";
import FxSvg from "./FxSvg.vue";
import FxBadge from "./primitives/FxBadge.vue";

type Tone = "neutral" | "success" | "info" | "warning" | "danger";

withDefaults(
  defineProps<{
    id: string;
    name: string;
    subCount?: number;
    subCountAria?: string;
    badgeLabel?: string;
    badgeTone?: Tone;
  }>(),
  {
    subCount: 0,
    subCountAria: undefined,
    badgeLabel: undefined,
    badgeTone: "neutral",
  },
);
</script>

<template>
  <RouterLink
    :to="'/locations/' + id"
    class="group fx-home-loc-row flex items-center gap-3 px-4 py-3 sm:px-5"
  >
    <span class="fx-home-loc-icon" aria-hidden="true"><FxSvg name="mapPin" class="fx-icon" /></span>
    <div class="min-w-0 flex-1">
      <div class="flex flex-wrap items-center gap-2">
        <span class="text-sm font-semibold leading-snug text-zinc-900 transition-colors group-hover:text-sky-950">{{ name }}</span>
        <FxBadge v-if="badgeLabel" :tone="badgeTone">{{ badgeLabel }}</FxBadge>
      </div>
    </div>
    <div class="flex shrink-0 items-center gap-2">
      <span
        v-if="subCount > 0"
        class="fx-home-loc-count-badge fx-home-loc-count-badge--sm"
        :aria-label="subCountAria"
      >{{ subCount }}</span>
      <span class="fx-home-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
    </div>
  </RouterLink>
</template>
