<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import type { RouteLocationRaw } from "vue-router";
import FxSvg from "../FxSvg.vue";
import type { FxIconName } from "../FxSvg.vue";

const props = withDefaults(
  defineProps<{
    icon: FxIconName;
    tone?: "neutral" | "danger";
    size?: "sm" | "md";
    type?: "button" | "submit" | "reset";
    to?: RouteLocationRaw;
    href?: string;
    disabled?: boolean;
    loading?: boolean;
    ariaLabel: string;
    title?: string;
  }>(),
  {
    tone: "neutral",
    size: "md",
    type: "button",
    to: undefined,
    href: undefined,
    disabled: false,
    loading: false,
    title: undefined,
  },
);

defineEmits<{ click: [MouseEvent] }>();

const classes = computed(() => [
  props.tone === "danger" ? "fx-icon-btn-danger" : "fx-icon-btn",
  props.size === "sm" ? "fx-icon-btn--sm" : null,
]);

const isDisabled = computed(() => props.disabled || props.loading);
</script>

<template>
  <RouterLink
    v-if="to"
    :to="to"
    :class="[...classes, isDisabled ? 'pointer-events-none opacity-60' : null]"
    :aria-label="ariaLabel"
    :title="title ?? ariaLabel"
  >
    <FxSvg :name="icon" />
  </RouterLink>
  <a
    v-else-if="href"
    :href="href"
    :class="[...classes, isDisabled ? 'pointer-events-none opacity-60' : null]"
    :aria-label="ariaLabel"
    :title="title ?? ariaLabel"
    @click="$emit('click', $event)"
  >
    <FxSvg :name="icon" />
  </a>
  <button
    v-else
    :type="type"
    :class="classes"
    :disabled="isDisabled"
    :aria-label="ariaLabel"
    :aria-busy="loading || undefined"
    :title="title ?? ariaLabel"
    @click="$emit('click', $event)"
  >
    <FxSvg :name="icon" />
  </button>
</template>
