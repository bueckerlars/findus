<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string;
    padding?: "none" | "sm" | "md" | "lg";
  }>(),
  { title: undefined, padding: "md" },
);
</script>

<template>
  <section class="fx-section">
    <header v-if="$slots.header || title || $slots.actions" class="fx-section__header">
      <slot name="header">
        <h2 v-if="title" class="fx-section__title">{{ title }}</h2>
      </slot>
      <div v-if="$slots.actions" class="flex flex-wrap items-center gap-2">
        <slot name="actions" />
      </div>
    </header>
    <div
      :class="
        padding === 'none'
          ? ''
          : padding === 'sm'
          ? 'p-3 sm:p-4'
          : padding === 'lg'
          ? 'p-5 sm:p-6'
          : 'fx-section__body'
      "
    >
      <slot />
    </div>
  </section>
</template>
