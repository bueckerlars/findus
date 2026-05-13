<script setup lang="ts">
import { RouterLink } from "vue-router";
import FxSvg from "./FxSvg.vue";
import FxBadge from "./primitives/FxBadge.vue";
import ItemPhotoPlaceholder from "./ItemPhotoPlaceholder.vue";

type Tone = "neutral" | "success" | "info" | "warning" | "danger";

withDefaults(
  defineProps<{
    id: string;
    name: string;
    locationName?: string;
    timestamp?: string;
    timestampIso?: string;
    photoPath?: string | null;
    badgeLabel?: string;
    badgeTone?: Tone;
  }>(),
  {
    locationName: undefined,
    timestamp: undefined,
    timestampIso: undefined,
    photoPath: undefined,
    badgeLabel: undefined,
    badgeTone: "neutral",
  },
);
</script>

<template>
  <RouterLink
    :to="'/items/' + id"
    class="group fx-item-row relative fx-list-row !items-stretch gap-3 rounded-none border-0 py-0 pl-0 pr-3 shadow-none hover:shadow-sm"
  >
    <span class="fx-item-row-accent" aria-hidden="true"></span>
    <div
      class="relative z-[1] flex shrink-0 items-center self-stretch py-2 pl-2.5 pr-1 sm:pl-3"
      aria-hidden="true"
    >
      <div
        class="fx-item-row-thumb-media relative h-10 w-10 shrink-0 overflow-hidden rounded-lg bg-gradient-to-br from-zinc-50 to-zinc-100 shadow-sm ring-1 ring-zinc-200/80 sm:h-11 sm:w-11"
      >
        <img
          v-if="photoPath"
          :src="'/items/' + id + '/photo'"
          alt=""
          class="h-full w-full object-cover"
          @error="($event.target as HTMLImageElement).style.display = 'none'"
        />
        <div v-else class="fx-item-row-thumb-placeholder flex h-full w-full">
          <ItemPhotoPlaceholder :item-id="id" />
        </div>
      </div>
    </div>
    <div class="relative z-[1] flex min-h-[2.75rem] min-w-0 flex-1 flex-col justify-center py-2">
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
    <div class="relative z-[1] flex shrink-0 items-center gap-1.5 self-stretch py-2">
      <time
        v-if="timestamp"
        :datetime="timestampIso"
        class="hidden text-xs tabular-nums text-zinc-400 transition-colors group-hover:text-zinc-500 sm:inline"
      >{{ timestamp }}</time>
      <span class="fx-item-row-chevron" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon" /></span>
    </div>
  </RouterLink>
</template>
