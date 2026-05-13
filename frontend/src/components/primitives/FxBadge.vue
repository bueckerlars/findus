<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    tone?: "neutral" | "success" | "info" | "warning" | "danger";
  }>(),
  { tone: "neutral" },
);

const classes = computed(() => [
  "fx-badge",
  props.tone === "success"
    ? "fx-badge--success"
    : props.tone === "info"
    ? "fx-badge--info"
    : props.tone === "warning"
    ? "fx-badge--warning"
    : props.tone === "danger"
    ? "fx-badge--danger"
    : "fx-badge--neutral",
]);
</script>

<template>
  <span :class="classes"><slot /></span>
</template>
