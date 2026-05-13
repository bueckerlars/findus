<script setup lang="ts">
import FxSvg from "../FxSvg.vue";
import type { FxIconName } from "../FxSvg.vue";

withDefaults(
  defineProps<{
    icon?: FxIconName;
    title: string;
    body?: string;
  }>(),
  { icon: "inbox", body: undefined },
);
</script>

<template>
  <div class="fx-empty-state">
    <span v-if="icon" class="fx-empty-state__icon" aria-hidden="true"><FxSvg :name="icon" /></span>
    <p class="fx-empty-state__title">{{ title }}</p>
    <p v-if="body" class="fx-empty-state__body">{{ body }}</p>
    <div v-if="$slots.default" class="fx-empty-state__actions">
      <slot />
    </div>
  </div>
</template>
