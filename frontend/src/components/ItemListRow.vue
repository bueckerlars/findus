<script setup lang="ts">
import { RouterLink } from "vue-router";
import FxSvg from "./FxSvg.vue";
import FxBadge from "./primitives/FxBadge.vue";

type Tone = "neutral" | "success" | "info" | "warning" | "danger";

withDefaults(
  defineProps<{
    id: string;
    name: string;
    locationName?: string;
    timestamp?: string;
    timestampIso?: string;
    badgeLabel?: string;
    badgeTone?: Tone;
  }>(),
  {
    locationName: undefined,
    timestamp: undefined,
    timestampIso: undefined,
    badgeLabel: undefined,
    badgeTone: "neutral",
  },
);
</script>

<template>
  <RouterLink
    :to="'/items/' + id"
    class="group fx-item-row relative fx-list-row rounded-none border-0 shadow-none hover:shadow-sm"
  >
    <span class="fx-item-row-accent" aria-hidden="true"></span>
    <div class="relative z-[1] min-w-0 flex-1">
      <div class="flex flex-wrap items-center gap-2">
        <span class="font-medium text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{ name }}</span>
        <FxBadge v-if="badgeLabel" :tone="badgeTone">{{ badgeLabel }}</FxBadge>
      </div>
      <p
        v-if="locationName"
        class="mt-0.5 flex min-w-0 items-center gap-1.5 text-xs font-medium text-zinc-600 transition-colors duration-200 group-hover:text-zinc-800"
      >
        <FxSvg name="mapPin" class="h-3 w-3 shrink-0 text-sky-600" aria-hidden="true" />
        <span class="truncate">{{ locationName }}</span>
      </p>
    </div>
    <div class="relative z-[1] flex shrink-0 items-center gap-1.5">
      <time
        v-if="timestamp"
        :datetime="timestampIso"
        class="hidden text-xs tabular-nums text-zinc-400 transition-colors group-hover:text-zinc-500 sm:inline"
      >{{ timestamp }}</time>
      <span class="fx-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon" /></span>
    </div>
  </RouterLink>
</template>
