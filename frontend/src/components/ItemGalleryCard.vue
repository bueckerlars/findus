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
    photoPath?: string | null;
    badgeLabel?: string;
    badgeTone?: Tone;
  }>(),
  {
    locationName: undefined,
    photoPath: undefined,
    badgeLabel: undefined,
    badgeTone: "neutral",
  },
);
</script>

<template>
  <RouterLink
    :to="'/items/' + id"
    class="group fx-item-gallery flex flex-col overflow-hidden rounded-xl border border-zinc-200/80 bg-white shadow-sm ring-1 ring-zinc-950/[0.03]"
  >
    <div class="fx-item-gallery-media aspect-square bg-gradient-to-br from-zinc-50 to-zinc-100 ring-1 ring-zinc-100/80">
      <img
        v-if="photoPath"
        :src="'/items/' + id + '/photo'"
        alt=""
        class="fx-item-gallery-photo"
        @error="($event.target as HTMLImageElement).style.display = 'none'"
      />
      <div v-else class="fx-item-gallery-placeholder">
        <ItemPhotoPlaceholder :item-id="id" />
      </div>
      <div class="fx-item-gallery-shade" aria-hidden="true"></div>
      <span class="fx-item-gallery-fab" aria-hidden="true"><FxSvg name="chevronRight" class="fx-icon h-4 w-4" /></span>
    </div>
    <div class="relative z-[1] flex min-h-0 flex-1 flex-col gap-1 border-t border-zinc-100/90 p-2.5">
      <span class="line-clamp-2 text-sm font-medium leading-snug text-zinc-900 transition-colors duration-200 group-hover:text-sky-950">{{ name }}</span>
      <p
        v-if="locationName"
        class="flex min-w-0 items-start gap-1.5 text-xs font-medium leading-snug text-zinc-600 transition-colors group-hover:text-zinc-800"
      >
        <FxSvg name="mapPin" class="mt-0.5 h-3 w-3 shrink-0 text-sky-600" aria-hidden="true" />
        <span class="line-clamp-2 min-w-0">{{ locationName }}</span>
      </p>
      <FxBadge v-if="badgeLabel" :tone="badgeTone">{{ badgeLabel }}</FxBadge>
    </div>
  </RouterLink>
</template>
