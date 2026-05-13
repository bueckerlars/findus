<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import type { RouteLocationRaw } from "vue-router";
import FxSvg from "../FxSvg.vue";
import type { FxIconName } from "../FxSvg.vue";

type Variant = "primary" | "secondary" | "danger" | "ghost";
type Size = "sm" | "md" | "lg";

const props = withDefaults(
  defineProps<{
    variant?: Variant;
    size?: Size;
    loading?: boolean;
    iconLeft?: FxIconName;
    iconRight?: FxIconName;
    type?: "button" | "submit" | "reset";
    to?: RouteLocationRaw;
    href?: string;
    disabled?: boolean;
    fullWidth?: boolean;
  }>(),
  {
    variant: "primary",
    size: "md",
    loading: false,
    iconLeft: undefined,
    iconRight: undefined,
    type: "button",
    to: undefined,
    href: undefined,
    disabled: false,
    fullWidth: false,
  },
);

defineEmits<{ click: [MouseEvent] }>();

const classes = computed(() => [
  props.variant === "primary"
    ? "fx-btn-primary"
    : props.variant === "secondary"
    ? "fx-btn-secondary"
    : props.variant === "danger"
    ? "fx-btn-danger"
    : "fx-btn-ghost",
  props.size === "sm" ? "fx-btn--sm" : props.size === "lg" ? "fx-btn--lg" : null,
  props.fullWidth ? "w-full" : null,
]);

const isDisabled = computed(() => props.disabled || props.loading);

const iconClass = computed(() =>
  props.size === "sm" ? "h-3.5 w-3.5 shrink-0" : props.size === "lg" ? "h-5 w-5 shrink-0" : "h-4 w-4 shrink-0",
);
</script>

<template>
  <RouterLink v-if="to" :to="to" :class="[...classes, isDisabled ? 'pointer-events-none opacity-60' : null]">
    <span v-if="loading" :class="['shrink-0 animate-spin', iconClass]" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" class="h-full w-full">
        <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-dasharray="42 14" />
      </svg>
    </span>
    <FxSvg v-else-if="iconLeft" :name="iconLeft" :class="iconClass" />
    <slot />
    <FxSvg v-if="iconRight && !loading" :name="iconRight" :class="iconClass" />
  </RouterLink>
  <a v-else-if="href" :href="href" :class="[...classes, isDisabled ? 'pointer-events-none opacity-60' : null]" @click="$emit('click', $event)">
    <span v-if="loading" :class="['shrink-0 animate-spin', iconClass]" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" class="h-full w-full">
        <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-dasharray="42 14" />
      </svg>
    </span>
    <FxSvg v-else-if="iconLeft" :name="iconLeft" :class="iconClass" />
    <slot />
    <FxSvg v-if="iconRight && !loading" :name="iconRight" :class="iconClass" />
  </a>
  <button
    v-else
    :type="type"
    :class="classes"
    :disabled="isDisabled"
    :aria-busy="loading || undefined"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" :class="['shrink-0 animate-spin', iconClass]" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none" class="h-full w-full">
        <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-dasharray="42 14" />
      </svg>
    </span>
    <FxSvg v-else-if="iconLeft" :name="iconLeft" :class="iconClass" />
    <slot />
    <FxSvg v-if="iconRight && !loading" :name="iconRight" :class="iconClass" />
  </button>
</template>
