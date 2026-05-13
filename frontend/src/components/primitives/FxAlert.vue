<script setup lang="ts">
import { computed } from "vue";
import FxSvg from "../FxSvg.vue";
import type { FxIconName } from "../FxSvg.vue";

type Tone = "info" | "success" | "warning" | "danger";

const props = withDefaults(
  defineProps<{
    tone?: Tone;
    title?: string;
    size?: "md" | "lg";
    icon?: FxIconName | "none";
    role?: "alert" | "status";
  }>(),
  { tone: "danger", title: undefined, size: "md", icon: undefined, role: undefined },
);

const classes = computed(() => [
  "fx-alert",
  props.tone === "info"
    ? "fx-alert--info"
    : props.tone === "success"
    ? "fx-alert--success"
    : props.tone === "warning"
    ? "fx-alert--warning"
    : "fx-alert--danger",
  props.size === "lg" ? "fx-alert--lg" : null,
]);

const resolvedIcon = computed<FxIconName | null>(() => {
  if (props.icon === "none") return null;
  if (props.icon) return props.icon;
  switch (props.tone) {
    case "success":
      return "checkCircle";
    case "warning":
      return "exclamationTriangle";
    case "info":
      return "informationCircle";
    case "danger":
    default:
      return "exclamationCircle";
  }
});

const computedRole = computed(() => props.role ?? (props.tone === "danger" || props.tone === "warning" ? "alert" : "status"));
</script>

<template>
  <div :class="classes" :role="computedRole">
    <FxSvg v-if="resolvedIcon" :name="resolvedIcon" class="fx-alert__icon" />
    <div class="fx-alert__body">
      <p v-if="title" class="fx-alert__title">{{ title }}</p>
      <div v-if="$slots.default" :class="title ? 'mt-0.5 leading-snug' : 'leading-snug'">
        <slot />
      </div>
    </div>
  </div>
</template>
