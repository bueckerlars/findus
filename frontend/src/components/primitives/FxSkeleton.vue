<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    shape?: "rect" | "circle";
    width?: string;
    height?: string;
  }>(),
  { shape: "rect", width: "100%", height: "0.75rem" },
);

const classes = computed(() => [
  "fx-skeleton",
  props.shape === "circle" ? "fx-skeleton--circle" : null,
]);
</script>

<template>
  <span :class="classes" :style="{ width, height }" aria-hidden="true" />
</template>
