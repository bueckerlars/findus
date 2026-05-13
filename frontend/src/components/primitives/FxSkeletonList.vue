<script setup lang="ts">
import FxSkeleton from "./FxSkeleton.vue";

withDefaults(
  defineProps<{
    rows?: number;
    ariaLabel?: string;
  }>(),
  { rows: 5, ariaLabel: undefined },
);
</script>

<template>
  <ul class="divide-y divide-zinc-100" role="list" :aria-busy="true" :aria-label="ariaLabel">
    <li v-for="n in rows" :key="n" class="flex items-center gap-3 px-4 py-3 sm:px-5">
      <FxSkeleton shape="circle" width="1.75rem" height="1.75rem" />
      <span class="flex min-w-0 flex-1 flex-col gap-1.5">
        <FxSkeleton width="55%" height="0.75rem" />
        <FxSkeleton width="35%" height="0.625rem" />
      </span>
      <FxSkeleton width="3rem" height="0.625rem" />
    </li>
  </ul>
</template>
