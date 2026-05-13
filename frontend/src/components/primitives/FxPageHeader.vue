<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string;
    subtitle?: string;
    titleAs?: "h1" | "h2";
  }>(),
  { subtitle: undefined, titleAs: "h1" },
);
</script>

<template>
  <header class="fx-page-header">
    <div class="min-w-0">
      <component :is="titleAs" class="fx-page-header__title">{{ title }}</component>
      <p v-if="subtitle" class="fx-page-header__subtitle">{{ subtitle }}</p>
    </div>
    <div v-if="$slots.actions" class="fx-page-header__actions">
      <slot name="actions" />
    </div>
  </header>
</template>
