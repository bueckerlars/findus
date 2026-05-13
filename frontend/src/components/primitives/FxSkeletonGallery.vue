<script setup lang="ts">
import FxSkeleton from "./FxSkeleton.vue";

withDefaults(
  defineProps<{
    tiles?: number;
    ariaLabel?: string;
  }>(),
  { tiles: 6, ariaLabel: undefined },
);
</script>

<template>
  <div
    class="grid grid-cols-2 gap-3 p-4 sm:grid-cols-3 sm:gap-4 lg:grid-cols-4"
    :aria-busy="true"
    :aria-label="ariaLabel"
    role="list"
  >
    <div v-for="n in tiles" :key="n" class="fx-skeleton-tile" role="listitem">
      <FxSkeleton class="aspect-square !rounded-none" width="100%" height="100%" />
      <div class="flex flex-col gap-1.5 p-3">
        <FxSkeleton width="80%" height="0.75rem" />
        <FxSkeleton width="50%" height="0.625rem" />
      </div>
    </div>
  </div>
</template>
